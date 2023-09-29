package hevc

import (
	"bytes"
	"fmt"

	"github.com/jaypadia-frame/mp4ff/avc"
	"github.com/jaypadia-frame/mp4ff/bits"
)

// SPS - HEVC SPS parameters
// ISO/IEC 23008-2 Sec. 7.3.2.3
type PPS struct {
	PPSPicParameterSetID byte
	PPSSeqParameterSetID byte
	DependentSliceSegmentsEnabledFlag bool
	OutputFlagPresentFlag bool
	NumExtraSliceHeaderBits byte
	SignDataHidingEnabledFlag bool
	CabacInitPresentFlag bool
	NumRefIdxL0DefaultActiveMinus1 byte
	NumRefIdxL1DefaultActiveMinus1 byte
	InitQPMinus26 byte
	ConstrainedIntraPredFlag bool
	TransformSkipEnabledFlag bool
	CUQPDeltaEnabledFlag bool
	DiffCUQPDeltaDepth byte
	PPSCbQPOffset byte
	PPSCrQPOffset byte
	PPSSliceChromaQPOffsetsPresentFlag bool
	WeightedPredFlag bool
	WeightedBipredFlag bool
	78

