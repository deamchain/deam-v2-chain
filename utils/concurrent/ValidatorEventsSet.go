package concurrent

import (
	"sync"

	"github.com/deamchain/deam-v2-base/hash"
	"github.com/deamchain/deam-v2-base/inter/idx"
)

type ValidatorEventsSet struct {
	sync.RWMutex
	Val map[idx.ValidatorID]hash.Event
}

func WrapValidatorEventsSet(v map[idx.ValidatorID]hash.Event) *ValidatorEventsSet {
	return &ValidatorEventsSet{
		RWMutex: sync.RWMutex{},
		Val:     v,
	}
}
