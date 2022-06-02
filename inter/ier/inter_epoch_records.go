package ier

import (
	"github.com/deamchain/deam-v2-base/hash"
	"github.com/deamchain/deam-v2-base/inter/idx"

	"go-galaxy/inter/iblockproc"
)

type LlrFullEpochRecord struct {
	BlockState iblockproc.BlockState
	EpochState iblockproc.EpochState
}

type LlrIdxFullEpochRecord struct {
	LlrFullEpochRecord
	Idx idx.Epoch
}

func (er LlrFullEpochRecord) Hash() hash.Hash {
	return hash.Of(er.BlockState.Hash().Bytes(), er.EpochState.Hash().Bytes())
}
