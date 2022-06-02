package galaxy

import (
	"math/big"

	"github.com/deamchain/deam-v2-base/hash"
	"github.com/deamchain/deam-v2-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"

	"go-galaxy/inter"
	"go-galaxy/galaxy/genesis"
	"go-galaxy/galaxy/genesis/gpos"
)

type Genesis struct {
	Accounts    genesis.Accounts
	Storage     genesis.Storage
	Delegations genesis.Delegations
	Blocks      genesis.Blocks
	RawEvmItems genesis.RawEvmItems
	Validators  gpos.Validators

	FirstEpoch    idx.Epoch
	PrevEpochTime inter.Timestamp
	Time          inter.Timestamp
	ExtraData     []byte

	TotalSupply *big.Int

	DriverOwner common.Address

	Rules Rules

	Hash func() hash.Hash
}
