package gossip

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/deamchain/deam-v2-base/hash"
	"github.com/deamchain/deam-v2-base/inter/dag"
	"github.com/deamchain/deam-v2-base/inter/idx"
	"github.com/deamchain/deam-v2-base/lachesis"
	"github.com/deamchain/deam-v2-base/utils/workers"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/protocols/snap"
	"github.com/ethereum/go-ethereum/event"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/dnsdisc"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rpc"

	"go-galaxy/ethapi"
	"go-galaxy/eventcheck"
	"go-galaxy/eventcheck/basiccheck"
	"go-galaxy/eventcheck/epochcheck"
	"go-galaxy/eventcheck/gaspowercheck"
	"go-galaxy/eventcheck/heavycheck"
	"go-galaxy/eventcheck/parentscheck"
	"go-galaxy/evmcore"
	"go-galaxy/galaxy"
	"go-galaxy/gossip/blockproc"
	"go-galaxy/gossip/blockproc/drivermodule"
	"go-galaxy/gossip/blockproc/eventmodule"
	"go-galaxy/gossip/blockproc/evmmodule"
	"go-galaxy/gossip/blockproc/sealmodule"
	"go-galaxy/gossip/blockproc/verwatcher"
	"go-galaxy/gossip/emitter"
	"go-galaxy/gossip/filters"
	"go-galaxy/gossip/gasprice"
	"go-galaxy/gossip/proclogger"
	snapsync "go-galaxy/gossip/protocols/snap"
	"go-galaxy/inter"
	"go-galaxy/logger"
	"go-galaxy/utils/gsignercache"
	"go-galaxy/utils/wgmutex"
	"go-galaxy/valkeystore"
	"go-galaxy/vecmt"
)

type ServiceFeed struct {
	scope notify.SubscriptionScope

	newEpoch        notify.Feed
	newPack         notify.Feed
	newEmittedEvent notify.Feed
	newBlock        notify.Feed
	newLogs         notify.Feed
}

func (f *ServiceFeed) SubscribeNewEpoch(ch chan<- idx.Epoch) notify.Subscription {
	return f.scope.Track(f.newEpoch.Subscribe(ch))
}

func (f *ServiceFeed) SubscribeNewEmitted(ch chan<- *inter.EventPayload) notify.Subscription {
	return f.scope.Track(f.newEmittedEvent.Subscribe(ch))
}

func (f *ServiceFeed) SubscribeNewBlock(ch chan<- evmcore.ChainHeadNotify) notify.Subscription {
	return f.scope.Track(f.newBlock.Subscribe(ch))
}

func (f *ServiceFeed) SubscribeNewLogs(ch chan<- []*types.Log) notify.Subscription {
	return f.scope.Track(f.newLogs.Subscribe(ch))
}

type BlockProc struct {
	SealerModule        blockproc.SealerModule
	TxListenerModule    blockproc.TxListenerModule
	GenesisTxTransactor blockproc.TxTransactor
	PreTxTransactor     blockproc.TxTransactor
	PostTxTransactor    blockproc.TxTransactor
	EventsModule        blockproc.ConfirmedEventsModule
	EVMModule           blockproc.EVM
}

func DefaultBlockProc(g galaxy.Genesis) BlockProc {
	return BlockProc{
		SealerModule:        sealmodule.New(),
		TxListenerModule:    drivermodule.NewDriverTxListenerModule(),
		GenesisTxTransactor: drivermodule.NewDriverTxGenesisTransactor(g),
		PreTxTransactor:     drivermodule.NewDriverTxPreTransactor(),
		PostTxTransactor:    drivermodule.NewDriverTxTransactor(),
		EventsModule:        eventmodule.New(),
		EVMModule:           evmmodule.New(),
	}
}

// Service implements go-ethereum/node.Service interface.
type Service struct {
	config Config

	// server
	p2pServer *p2p.Server
	Name      string

	accountManager *accounts.Manager

	// application
	store               *Store
	engine              lachesis.Consensus
	dagIndexer          *vecmt.Index
	engineMu            *sync.RWMutex
	emitters            []*emitter.Emitter
	txpool              TxPool
	heavyCheckReader    HeavyCheckReader
	gasPowerCheckReader GasPowerCheckReader
	checkers            *eventcheck.Checkers
	uniqueEventIDs      uniqueID

	// version watcher
	verWatcher *verwatcher.VerWarcher

	blockProcWg        sync.WaitGroup
	blockProcTasks     *workers.Workers
	blockProcTasksDone chan struct{}
	blockProcModules   BlockProc

	blockBusyFlag uint32
	eventBusyFlag uint32

	feed     ServiceFeed
	eventMux *event.TypeMux

	gpo *gasprice.Oracle

	// application protocol
	handler *handler

	galaxyDialCandidates enode.Iterator
	snapDialCandidates   enode.Iterator

	EthAPI        *EthAPIBackend
	netRPCService *ethapi.PublicNetAPI

	procLogger *proclogger.Logger

	stopped bool

	tflusher PeriodicFlusher

	logger.Instance
}

func NewService(stack *node.Node, config Config, store *Store, blockProc BlockProc, engine lachesis.Consensus, dagIndexer *vecmt.Index, newTxPool func(evmcore.StateReader) TxPool) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	svc, err := newService(config, store, blockProc, engine, dagIndexer, newTxPool)
	if err != nil {
		return nil, err
	}

	svc.p2pServer = stack.Server()
	svc.accountManager = stack.AccountManager()
	svc.eventMux = stack.EventMux()
	// Create the net API service
	svc.netRPCService = ethapi.NewPublicNetAPI(svc.p2pServer, store.GetRules().NetworkID)

	return svc, nil
}

func newService(config Config, store *Store, blockProc BlockProc, engine lachesis.Consensus, dagIndexer *vecmt.Index, newTxPool func(evmcore.StateReader) TxPool) (*Service, error) {
	svc := &Service{
		config:             config,
		blockProcTasksDone: make(chan struct{}),
		Name:               fmt.Sprintf("Node-%d", rand.Int()),
		store:              store,
		engine:             engine,
		blockProcModules:   blockProc,
		dagIndexer:         dagIndexer,
		engineMu:           new(sync.RWMutex),
		uniqueEventIDs:     uniqueID{new(big.Int)},
		procLogger:         proclogger.NewLogger(),
		Instance:           logger.New("gossip-service"),
	}

	svc.blockProcTasks = workers.New(new(sync.WaitGroup), svc.blockProcTasksDone, 1)

	// load epoch DB
	svc.store.loadEpochStore(svc.store.GetEpoch())
	es := svc.store.getEpochStore(svc.store.GetEpoch())
	svc.dagIndexer.Reset(svc.store.GetValidators(), es.table.DagIndex, func(id hash.Event) dag.Event {
		return svc.store.GetEvent(id)
	})
	// load caches for mutable values to avoid race condition
	svc.store.GetBlockEpochState()
	svc.store.GetHighestLamport()
	svc.store.GetLastBVs()
	svc.store.GetLastEVs()
	svc.store.GetLlrState()

	// create GPO
	svc.gpo = gasprice.NewOracle(&GPOBackend{store}, svc.config.GPO)

	// create checkers
	net := store.GetRules()
	txSigner := gsignercache.Wrap(types.LatestSignerForChainID(net.EvmChainConfig().ChainID))
	svc.heavyCheckReader.Store = store
	svc.heavyCheckReader.Pubkeys.Store(readEpochPubKeys(svc.store, svc.store.GetEpoch()))                                          // read pub keys of current epoch from DB
	svc.gasPowerCheckReader.Ctx.Store(NewGasPowerContext(svc.store, svc.store.GetValidators(), svc.store.GetEpoch(), net.Economy)) // read gaspower check data from DB
	svc.checkers = makeCheckers(config.HeavyCheck, txSigner, &svc.heavyCheckReader, &svc.gasPowerCheckReader, svc.store)

	// create tx pool
	stateReader := svc.GetEvmStateReader()
	svc.txpool = newTxPool(stateReader)

	// init dialCandidates
	dnsclient := dnsdisc.NewClient(dnsdisc.Config{})
	var err error
	svc.galaxyDialCandidates, err = dnsclient.NewIterator(config.GalaxyDiscoveryURLs...)
	if err != nil {
		return nil, err
	}
	svc.snapDialCandidates, err = dnsclient.NewIterator(config.SnapDiscoveryURLs...)
	if err != nil {
		return nil, err
	}

	// create protocol manager
	svc.handler, err = newHandler(handlerConfig{
		config:   config,
		notifier: &svc.feed,
		txpool:   svc.txpool,
		engineMu: svc.engineMu,
		checkers: svc.checkers,
		s:        store,
		process: processCallback{
			Event: func(event *inter.EventPayload) error {
				done := svc.procLogger.EventConnectionStarted(event, false)
				defer done()
				return svc.processEvent(event)
			},
			SwitchEpochTo:    svc.SwitchEpochTo,
			PauseEvmSnapshot: svc.PauseEvmSnapshot,
			BVs:              svc.ProcessBlockVotes,
			BR:               svc.ProcessFullBlockRecord,
			EV:               svc.ProcessEpochVote,
			ER:               svc.ProcessFullEpochRecord,
		},
	})
	if err != nil {
		return nil, err
	}

	// create API backend
	svc.EthAPI = &EthAPIBackend{config.ExtRPCEnabled, svc, stateReader, txSigner, config.AllowUnprotectedTxs}

	svc.verWatcher = verwatcher.New(config.VersionWatcher, verwatcher.NewStore(store.table.NetworkVersion))
	svc.tflusher = svc.makePeriodicFlusher()

	return svc, nil
}

// makeCheckers builds event checkers
func makeCheckers(heavyCheckCfg heavycheck.Config, txSigner types.Signer, heavyCheckReader *HeavyCheckReader, gasPowerCheckReader *GasPowerCheckReader, store *Store) *eventcheck.Checkers {
	// create signatures checker
	heavyCheck := heavycheck.New(heavyCheckCfg, heavyCheckReader, txSigner)

	// create gaspower checker
	gaspowerCheck := gaspowercheck.New(gasPowerCheckReader)

	return &eventcheck.Checkers{
		Basiccheck:    basiccheck.New(),
		Epochcheck:    epochcheck.New(store),
		Parentscheck:  parentscheck.New(),
		Heavycheck:    heavyCheck,
		Gaspowercheck: gaspowerCheck,
	}
}

// makePeriodicFlusher makes PeriodicFlusher
func (s *Service) makePeriodicFlusher() PeriodicFlusher {
	// Normally the diffs are committed by message processing. Yet, since we have async EVM snapshots generation,
	// we need to periodically commit its data
	return PeriodicFlusher{
		period: 10 * time.Millisecond,
		callback: PeriodicFlusherCallaback{
			busy: func() bool {
				// try to lock engineMu/blockProcWg pair as rarely as possible to not hurt
				// events/blocks pipeline concurrency
				return atomic.LoadUint32(&s.eventBusyFlag) != 0 || atomic.LoadUint32(&s.blockBusyFlag) != 0
			},
			commitNeeded: func() bool {
				// use slightly higher size threshold to avoid locking the mutex/wg pair and hurting events/blocks concurrency
				// PeriodicFlusher should mostly commit only data generated by async EVM snapshots generation
				return s.store.isCommitNeeded(120, 100)
			},
			commit: func() {
				s.engineMu.Lock()
				defer s.engineMu.Unlock()
				// Note: blockProcWg.Wait() is already called by s.commit
				if s.store.isCommitNeeded(120, 100) {
					s.commit(false)
				}
			},
		},
		wg:   sync.WaitGroup{},
		quit: make(chan struct{}),
	}
}

func (s *Service) EmitterWorld(signer valkeystore.SignerI) emitter.World {
	return emitter.World{
		External: &emitterWorld{
			emitterWorldProc: emitterWorldProc{s},
			emitterWorldRead: emitterWorldRead{s.store},
			WgMutex:          wgmutex.New(s.engineMu, &s.blockProcWg),
		},
		TxPool:   s.txpool,
		Signer:   signer,
		TxSigner: s.EthAPI.signer,
	}
}

// RegisterEmitter must be called before service is started
func (s *Service) RegisterEmitter(em *emitter.Emitter) {
	s.emitters = append(s.emitters, em)
}

// MakeProtocols constructs the P2P protocol definitions for `galaxy`.
func MakeProtocols(svc *Service, backend *handler, disc enode.Iterator) []p2p.Protocol {
	protocols := make([]p2p.Protocol, len(ProtocolVersions))
	for i, version := range ProtocolVersions {
		version := version // Closure

		protocols[i] = p2p.Protocol{
			Name:    ProtocolName,
			Version: version,
			Length:  protocolLengths[version],
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				// wait until handler has started
				backend.started.Wait()
				peer := newPeer(version, p, rw, backend.config.Protocol.PeerCache)
				defer peer.Close()

				select {
				case <-backend.quitSync:
					return p2p.DiscQuitting
				default:
					backend.wg.Add(1)
					defer backend.wg.Done()
					err := backend.handle(peer)
					return err
				}
			},
			NodeInfo: func() interface{} {
				return backend.NodeInfo()
			},
			PeerInfo: func(id enode.ID) interface{} {
				if p := backend.peers.Peer(id.String()); p != nil {
					return p.Info()
				}
				return nil
			},
			Attributes:     []enr.Entry{currentENREntry(svc)},
			DialCandidates: disc,
		}
	}
	return protocols
}

// Protocols returns protocols the service can communicate on.
func (s *Service) Protocols() []p2p.Protocol {
	protos := append(
		MakeProtocols(s, s.handler, s.galaxyDialCandidates),
		snap.MakeProtocols((*snapHandler)(s.handler), s.snapDialCandidates)...)
	return protos
}

// APIs returns api methods the service wants to expose on rpc channels.
func (s *Service) APIs() []rpc.API {
	apis := ethapi.GetAPIs(s.EthAPI)

	apis = append(apis, []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.EthAPI, s.config.FilterAPI),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   snapsync.NewPublicDownloaderAPI(s.handler.snapLeecher, s.eventMux),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)

	return apis
}

// Start method invoked when the node is ready to start the service.
func (s *Service) Start() error {
	// start tflusher before starting snapshots generation
	s.tflusher.Start()
	// start snapshots generation
	if s.store.evm.IsEvmSnapshotPaused() && !s.config.AllowSnapsync {
		return errors.New("cannot halt snapsync and start fullsync")
	}
	root := s.store.GetBlockState().FinalizedStateRoot
	if !s.store.evm.HasStateDB(root) {
		if !s.config.AllowSnapsync {
			return errors.New("fullsync isn't possible because state root is missing")
		}
		root = hash.Zero
	}
	_ = s.store.GenerateSnapshotAt(common.Hash(root), true)

	// start blocks processor
	s.blockProcTasks.Start(1)

	// start p2p
	StartENRUpdater(s, s.p2pServer.LocalNode())
	s.handler.Start(s.p2pServer.MaxPeers)

	// start emitters
	for _, em := range s.emitters {
		em.Start()
	}

	s.verWatcher.Start()

	return nil
}

// WaitBlockEnd waits until parallel block processing is complete (if any)
func (s *Service) WaitBlockEnd() {
	s.blockProcWg.Wait()
}

// Stop method invoked when the node terminates the service.
func (s *Service) Stop() error {
	defer log.Info("Galaxy service stopped")
	s.verWatcher.Stop()
	for _, em := range s.emitters {
		em.Stop()
	}

	// Stop all the peer-related stuff first.
	s.galaxyDialCandidates.Close()
	s.snapDialCandidates.Close()

	s.handler.Stop()
	s.feed.scope.Close()
	s.eventMux.Stop()
	// it's safe to stop tflusher only before locking engineMu
	s.tflusher.Stop()

	// flush the state at exit, after all the routines stopped
	s.engineMu.Lock()
	defer s.engineMu.Unlock()
	s.stopped = true

	s.blockProcWg.Wait()
	close(s.blockProcTasksDone)
	s.store.evm.Flush(s.store.GetBlockState(), s.store.GetBlock)
	return s.store.Commit()
}

// AccountManager return node's account manager
func (s *Service) AccountManager() *accounts.Manager {
	return s.accountManager
}