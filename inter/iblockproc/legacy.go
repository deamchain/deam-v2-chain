package iblockproc

import (
	"github.com/deamchain/deam-v2-base/hash"
	"github.com/deamchain/deam-v2-base/inter/idx"
	"github.com/deamchain/deam-v2-base/inter/pos"

	"go-galaxy/inter"
	"go-galaxy/galaxy"
)

type ValidatorEpochStateV0 struct {
	GasRefund      uint64
	PrevEpochEvent hash.Event
}

type EpochStateV0 struct {
	Epoch          idx.Epoch
	EpochStart     inter.Timestamp
	PrevEpochStart inter.Timestamp

	EpochStateRoot hash.Hash

	Validators        *pos.Validators
	ValidatorStates   []ValidatorEpochStateV0
	ValidatorProfiles ValidatorProfiles

	Rules galaxy.Rules
}
