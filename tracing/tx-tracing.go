package tracing

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/opentracing/opentracing-go"
)

var (
	enabled   bool
	txSpans   = make(map[common.Hash]opentracing.Span)
	txSpansMu sync.RWMutex

	noopSpan = opentracing.NoopTracer{}.StartSpan("")
)

func SetEnabled(val bool) {
	enabled = val
}

func Enabled() bool {
	return enabled
}

func StartTx(tx common.Hash, galaxytion string) {
	if !enabled {
		return
	}

	txSpansMu.Lock()
	defer txSpansMu.Unlock()

	if _, ok := txSpans[tx]; ok {
		return
	}

	span := opentracing.StartSpan("lifecycle")
	span.SetTag("txhash", tx.String())
	span.SetTag("enter", galaxytion)
	txSpans[tx] = span
}

func FinishTx(tx common.Hash, galaxytion string) {
	if !enabled {
		return
	}

	txSpansMu.Lock()
	defer txSpansMu.Unlock()

	span, ok := txSpans[tx]
	if !ok {
		return
	}

	span.SetTag("exit", galaxytion)
	span.Finish()
	delete(txSpans, tx)
}

func CheckTx(tx common.Hash, galaxytion string) opentracing.Span {
	if !enabled {
		return noopSpan
	}

	txSpansMu.RLock()
	defer txSpansMu.RUnlock()

	span, ok := txSpans[tx]

	if !ok {
		return noopSpan
	}

	return opentracing.GlobalTracer().StartSpan(
		galaxytion,
		opentracing.ChildOf(span.Context()),
	)
}
