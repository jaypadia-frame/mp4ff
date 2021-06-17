package avc

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"github.com/Eyevinn/mp4ff/bits"
)

// Errors for parsing and handling AVC slices
var (
	ErrNoSliceHeader      = errors.New("No slice header")
	ErrInvalidSliceType   = errors.New("Invalid slice type")
	ErrTooFewBytesToParse = errors.New("Too few bytes to parse symbol")
	ErrParserNotImplemented = errors.New("Parser not implemented")
)

// SliceType - AVC slice type
type SliceType uint

func (s SliceType) String() string {
	switch s {
	case SLICE_I:
		return "I"
	case SLICE_P:
		return "P"
	case SLICE_B:
		return "B"
	case SLICE_SI:
		return "SI"
	case SLICE_SP:
		return "SP"
	default:
		return ""
	}
}

// AVC slice types
const (
	SLICE_P  = SliceType(0)
	SLICE_B  = SliceType(1)
	SLICE_I  = SliceType(2)
	SLICE_SP = SliceType(3)
	SLICE_SI = SliceType(4)
)

// SliceHeader - AVC Slice header
type SliceHeader struct {
	FirstMbInSlice              uint      // ue(v)
	SliceType                   SliceType // ue(v)
	PicParameterSetID           uint      // ue(v)
	ColourPlaneID               uint      // u(2)
	FrameNum                    uint      // u(v) - uses Log2MaxFrameNumMinus4
	FieldPicFlag                bool      // u(1)
	BottomFieldFlag             bool      // u(1)
	IDRPicID                    uint      // ue(v)
	PicOrderCntLSB              uint      // u(v) - TODO: what does 'v' depend on?
	DeltaPicOrderCntBottom      int       // se(v)
	DeltaPicOrderCnt            [2]int    // se(v)
	RedundantPicCnt             uint      // ue(v)
	DirectSpatialMVPredFlag     bool      // u(1)
	NumRefIdxActiveOverrideFlag bool      // u(1)
	NumRefIdxL0ActiveMinus1     uint      // ue(v)
	NumRefIdxL1ActiveMinus1     uint      // ue(v)
	// Ref Pic List Modification MVC not implmented
	RefPicListModification     *RefPicListModification
	PredWeightTable            *PredWeightTable
	DecRefPicMarking           *DecRefPicMarking
	CabacInitIDC               uint // ue(v)
	SliceQPDelta               int  // se(v)
	SPForSwitchFlag            bool // u(1)
	SliceQSDelta               int  // se(v)
	DisableDeblockingFilterIDC uint // ue(v)
	SliceAlphaC0OffsetDev2     int  // se(v)
	SliceBetaOffsetDev2        int  // se(v)
	SliceGroupChangeCycle      uint // u(v) - TODO: what does 'v' depend on?
}

// RefPicListModification - AVC Ref Pic list modification at slice level
type RefPicListModification struct {
	RefPicListModificationFlagL0 bool // u(1)
	RefPicListModificationFlagL1 bool // u(1)
	PerEntryParams               []PerEntryRefPicListModParams
}

// PerEntryRefPicListModParams - Ref Pic List Mod params per each entry
type PerEntryRefPicListModParams struct {
	ModificationOfPicNumsIDC uint // ue(v)
	AbsDiffPicNumMinus1      uint // ue(v)
	LongTermPicNum           uint // ue(v)
}

// PredWeightTable - AVC Prediction Weight Table in slice header
type PredWeightTable struct {
	LumaLog2WeightDenom   uint    // ue(v)
	ChromaLog2WeightDenom uint    // ue(v)
	LumaWeightL0Flag      bool    // u(1)
	LumaWeightL0          []int   // se(v)
	LumaOffsetL0          []int   // se(v)
	ChromaWeightL0Flag    bool    // u(1)
	ChromaWeightL0        [][]int // se(v)
	ChromaOffsetL0        [][]int // se(v)
	LumaWeightL1Flag      bool    // u(1)
	LumaWeightL1          []int   // se(v)
	LumaOffsetL1          []int   // se(v)
	ChromaWeightL1Flag    bool    // u(1)
	ChromaWeightL1        [][]int // se(v)
	ChromaOffsetL1        [][]int // se(v)
}

// DecRefPicMarking - Decoded Reference Picture Marking Syntax
type DecRefPicMarking struct {
	NoOutputOfPriorPicsFlag       bool // u(1)
	LongTermReferenceFlag         bool // u(1)
	AdaptiveRefPicMarkingModeFlag bool // u(1)
	AdaptiveRefPicMarkingParams   []AdaptiveMemCtrlDecRefPicMarkingParams
}

// AdaptiveMemCtrlDecRefPicMarkingParams - Used as explained in 8.2.5.4
type AdaptiveMemCtrlDecRefPicMarkingParams struct {
	MemoryManagementControlOperation uint // ue(v)
	DifferenceOfPicNumsMinus1        uint // ue(v)
	LongTermPicNum                   uint // ue(v)
	LongTermFrameIdx                 uint // ue(v)
	MaxLongTermFrameIdxPlus1         uint // ue(V)
}

// GetSliceTypeFromNALU - parse slice header to get slice type in interval 0 to 4
// This function is no longer necessary after the ParseSliceHeader is added
func GetSliceTypeFromNALU(data []byte) (sliceType SliceType, err error) {

	if len(data) <= 1 {
		err = ErrTooFewBytesToParse
		return
	}

	naluType := GetNaluType(data[0])
	switch naluType {
	case 1, 2, 5, 19:
		// slice_layer_without_partitioning_rbsp
		// slice_data_partition_a_layer_rbsp

	default:
		err = ErrNoSliceHeader
		return
	}
	r := bits.NewEBSPReader(bytes.NewReader((data[1:])))

	// first_mb_in_slice
	if _, err = r.ReadExpGolomb(); err != nil {
		return
	}

	// slice_type
	var st uint
	if st, err = r.ReadExpGolomb(); err != nil {
		return
	}
	sliceType = SliceType(st)
	if sliceType > 9 {
		err = ErrInvalidSliceType
		return
	}

	if sliceType >= 5 {
		sliceType -= 5 // The same type is repeated twice to tell if all slices in picture are the same
	}
	return
}

// ParseSliceHeader - Parse AVC Slice Header starting with NAL header
func ParseSliceHeader(data []byte, sps *SPS, pps *PPS) (*SliceHeader, int, error) {
	avcsh := &SliceHeader{}
	var err error

	rd := bytes.NewReader(data)
	r := bits.NewEBSPReader(rd)

	// Note! First byte is NAL Header
	nalHdr, err := r.Read(8)
	if err != nil {
		return nil, r.NrBytesRead(), err
	}
	nalType := GetNaluType(byte(nalHdr))
	if !sliceHeaderExpected(nalType) {
		return nil, r.NrBytesRead(), ErrNoSliceHeader
	}

	avcsh.FirstMbInSlice, err = r.ReadExpGolomb()
	if err != nil {
		return nil, r.NrBytesRead(), err
	}
	sliceType, err := r.ReadExpGolomb()
	if err != nil {
		return nil, r.NrBytesRead(), err
	}
	avcsh.SliceType, err = setSliceType(sliceType)
	if err != nil {
		return nil, r.NrBytesRead(), err
	}

	avcsh.PicParameterSetID, err = r.ReadExpGolomb()
	if err != nil {
		return nil, r.NrBytesRead(), err
	}
	if sps.SeparateColourPlaneFlag {
		avcsh.ColourPlaneID, err = r.Read(2)
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
	}
	avcsh.FrameNum, err = r.Read(int(sps.Log2MaxFrameNumMinus4 + 4))
	if err != nil {
		return nil, r.NrBytesRead(), err
	}
	if !sps.FrameMbsOnlyFlag {
		avcsh.FieldPicFlag, err = r.ReadFlag()
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
		if avcsh.FieldPicFlag {
			avcsh.BottomFieldFlag, err = r.ReadFlag()
			if err != nil {
				return nil, r.NrBytesRead(), err
			}
		}
	}
	if getIDRPicFlag(nalType) {
		avcsh.IDRPicID, err = r.ReadExpGolomb()
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
	}

	if sps.PicOrderCntType == 0 {
		avcsh.PicOrderCntLSB, err = r.Read(int(sps.Log2MaxPicOrderCntLsbMinus4 + 4))
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
		if pps.BottomFieldPicOrderInFramePresentFlag && !avcsh.FieldPicFlag {
			avcsh.DeltaPicOrderCntBottom, err = r.ReadSignedGolomb()
			if err != nil {
				return nil, r.NrBytesRead(), err
			}
		}
	}
	if sps.PicOrderCntType == 1 && !sps.DeltaPicOrderAlwaysZeroFlag {
		avcsh.DeltaPicOrderCnt[0], err = r.ReadSignedGolomb()
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
		if pps.BottomFieldPicOrderInFramePresentFlag && !avcsh.FieldPicFlag {
			avcsh.DeltaPicOrderCnt[1], err = r.ReadSignedGolomb()
			if err != nil {
				return nil, r.NrBytesRead(), err
			}
		}
	}
	if pps.RedundantPicCntPresentFlag {
		avcsh.RedundantPicCnt, err = r.ReadExpGolomb()
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
	}

	if avcsh.SliceType == SLICE_B {
		avcsh.DirectSpatialMVPredFlag, err = r.ReadFlag()
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
	}
	if avcsh.SliceType == SLICE_P || avcsh.SliceType == SLICE_SP || avcsh.SliceType == SLICE_B {
		avcsh.NumRefIdxActiveOverrideFlag, err = r.ReadFlag()
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
	}
	if avcsh.NumRefIdxActiveOverrideFlag {
		avcsh.NumRefIdxL0ActiveMinus1, err = r.ReadExpGolomb()
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
		if avcsh.SliceType == SLICE_B {
			avcsh.NumRefIdxL1ActiveMinus1, err = r.ReadExpGolomb()
			if err != nil {
				return nil, r.NrBytesRead(), err
			}
		}
	}

	if nalType == NaluType(20) || nalType == NaluType(21) {
		// MVC not implemented
		return nil, r.NrBytesRead(), ErrParserNotImplemented
	} else {
		avcsh.RefPicListModification, err = ParseRefPicListModification(r, avcsh)
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
	}
	if (pps.WeightedPredFlag && (avcsh.SliceType == SLICE_P || avcsh.SliceType == SLICE_SP)) ||
		(pps.WeightedBipredIDC == 1 && avcsh.SliceType == SLICE_B) {
		avcsh.PredWeightTable, err = ParsePredWeightTable(r, sps, avcsh)
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
	}
	if GetNalRefIDC(byte(nalHdr)) != 0 {
		avcsh.DecRefPicMarking, err = ParseDecRefPicMarking(r, nalType)
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
	}

	if pps.EntropyCodingModeFlag && avcsh.SliceType != SLICE_I && avcsh.SliceType != SLICE_SI {
		avcsh.CabacInitIDC, err = r.ReadExpGolomb()
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
	}

	avcsh.SliceQPDelta, err = r.ReadSignedGolomb()
	if err != nil {
		return nil, r.NrBytesRead(), err
	}
	if avcsh.SliceType == SLICE_SP || avcsh.SliceType == SLICE_SI {
		if avcsh.SliceType == SLICE_SP {
			avcsh.SPForSwitchFlag, err = r.ReadFlag()
			if err != nil {
				return nil, r.NrBytesRead(), err
			}
		}
		avcsh.SliceQSDelta, err = r.ReadSignedGolomb()
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
	}

	if pps.DeblockingFilterControlPresentFlag {
		avcsh.DisableDeblockingFilterIDC, err = r.ReadExpGolomb()
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
		if avcsh.DisableDeblockingFilterIDC != 1 {
			avcsh.SliceAlphaC0OffsetDev2, err = r.ReadSignedGolomb()
			if err != nil {
				return nil, r.NrBytesRead(), err
			}
			avcsh.SliceBetaOffsetDev2, err = r.ReadSignedGolomb()
			if err != nil {
				return nil, r.NrBytesRead(), err
			}
		}
	}

	if pps.NumSliceGroupsMinus1 > 0 && pps.SliceGroupMapType >= 3 && pps.SliceGroupMapType <= 5 {
		// based on equation 7-35 H.264 spec
		sgccNumBits := math.Ceil(math.Log2(float64((pps.PicSizeInMapUnitsMinus1+1)/(pps.SliceGroupChangeRateMinus1+1) + 1)))
		avcsh.SliceGroupChangeCycle, err = r.Read(int(sgccNumBits))
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
	}

	return avcsh, r.NrBytesRead(), nil
}

// ParseRefPicListModification - AVC Ref Pic list modification parser using bits r
func ParseRefPicListModification(r *bits.EBSPReader, avcsh *SliceHeader) (*RefPicListModification, error) {
	rplm := &RefPicListModification{}
	var err error

	if avcsh.SliceType%5 != 2 && avcsh.SliceType%5 != 4 {
		rplm.RefPicListModificationFlagL0, err = r.ReadFlag()
		if err != nil {
			return nil, err
		}

		if rplm.RefPicListModificationFlagL0 {
			for mopni := 0; mopni != 3; {
				rplmEntry := PerEntryRefPicListModParams{}
				rplmEntry.ModificationOfPicNumsIDC, err = r.ReadExpGolomb()
				if err != nil {
					return nil, err
				}
				if rplmEntry.ModificationOfPicNumsIDC == 0 || rplmEntry.ModificationOfPicNumsIDC == 1 {
					rplmEntry.AbsDiffPicNumMinus1, err = r.ReadExpGolomb()
					if err != nil {
						return nil, err
					}
				} else if rplmEntry.ModificationOfPicNumsIDC == 2 {
					rplmEntry.LongTermPicNum, err = r.ReadExpGolomb()
					if err != nil {
						return nil, err
					}
				}
				rplm.PerEntryParams = append(rplm.PerEntryParams, rplmEntry)
			}
		}
	}

	if avcsh.SliceType%5 == 1 {
		rplm.RefPicListModificationFlagL1, err = r.ReadFlag()
		if err != nil {
			return nil, err
		}

		if rplm.RefPicListModificationFlagL0 {
			for mopni := 0; mopni != 3; {
				rplmEntry := PerEntryRefPicListModParams{}
				rplmEntry.ModificationOfPicNumsIDC, err = r.ReadExpGolomb()
				if err != nil {
					return nil, err
				}
				if rplmEntry.ModificationOfPicNumsIDC == 0 || rplmEntry.ModificationOfPicNumsIDC == 1 {
					rplmEntry.AbsDiffPicNumMinus1, err = r.ReadExpGolomb()
					if err != nil {
						return nil, err
					}
				} else if rplmEntry.ModificationOfPicNumsIDC == 2 {
					rplmEntry.LongTermPicNum, err = r.ReadExpGolomb()
					if err != nil {
						return nil, err
					}
				}
				rplm.PerEntryParams = append(rplm.PerEntryParams, rplmEntry)
			}
		}
	}

	return rplm, nil
}

// ParsePredWeightTable - AVC Slice Prediction Weight Table parser using bits r
func ParsePredWeightTable(r *bits.EBSPReader, sps *SPS, avcsh *SliceHeader) (*PredWeightTable, error) {
	pwt := &PredWeightTable{}
	var err error

	pwt.LumaLog2WeightDenom, err = r.ReadExpGolomb()
	if err != nil {
		return nil, err
	}

	if getChromaArrayType(sps) != 0 {
		pwt.ChromaLog2WeightDenom, err = r.ReadExpGolomb()
		if err != nil {
			return nil, err
		}
	}

	for i := uint(0); i <= avcsh.NumRefIdxL0ActiveMinus1; i++ {
		pwt.LumaWeightL0Flag, err = r.ReadFlag()
		if err != nil {
			return nil, err
		}
		if pwt.LumaWeightL0Flag {
			lumaWeightL0, err := r.ReadSignedGolomb()
			if err != nil {
				return nil, err
			}
			lumaOffsetL0, err := r.ReadSignedGolomb()
			if err != nil {
				return nil, err
			}
			pwt.LumaWeightL0 = append(pwt.LumaWeightL0, lumaWeightL0)
			pwt.LumaOffsetL0 = append(pwt.LumaWeightL0, lumaOffsetL0)
		}

		if getChromaArrayType(sps) != 0 {
			pwt.ChromaWeightL0Flag, err = r.ReadFlag()
			if err != nil {
				return nil, err
			}
			if pwt.ChromaWeightL0Flag {
				var chromaWeightL0, chromaOffsetL0 []int
				for j := 0; j < 2; j++ {
					chromaWeight, err := r.ReadSignedGolomb()
					if err != nil {
						return nil, err
					}
					chromaOffset, err := r.ReadSignedGolomb()
					if err != nil {
						return nil, err
					}
					chromaWeightL0 = append(chromaWeightL0, chromaWeight)
					chromaOffsetL0 = append(chromaOffsetL0, chromaOffset)
				}
				pwt.ChromaWeightL0 = append(pwt.ChromaWeightL0, chromaWeightL0)
				pwt.ChromaOffsetL0 = append(pwt.ChromaWeightL0, chromaOffsetL0)
			}
		}
	}

	if avcsh.SliceType%5 == 1 {
		for i := uint(0); i <= avcsh.NumRefIdxL1ActiveMinus1; i++ {
			pwt.LumaWeightL1Flag, err = r.ReadFlag()
			if err != nil {
				return nil, err
			}
			if pwt.LumaWeightL1Flag {
				lumaWeightL1, err := r.ReadSignedGolomb()
				if err != nil {
					return nil, err
				}
				lumaOffsetL1, err := r.ReadSignedGolomb()
				if err != nil {
					return nil, err
				}
				pwt.LumaWeightL1 = append(pwt.LumaWeightL1, lumaWeightL1)
				pwt.LumaOffsetL1 = append(pwt.LumaWeightL1, lumaOffsetL1)
			}

			if getChromaArrayType(sps) != 0 {
				pwt.ChromaWeightL1Flag, err = r.ReadFlag()
				if err != nil {
					return nil, err
				}
				if pwt.ChromaWeightL1Flag {
					var chromaWeightL1, chromaOffsetL1 []int
					for j := 0; j < 2; j++ {
						chromaWeight, err := r.ReadSignedGolomb()
						if err != nil {
							return nil, err
						}
						chromaOffset, err := r.ReadSignedGolomb()
						if err != nil {
							return nil, err
						}
						chromaWeightL1 = append(chromaWeightL1, chromaWeight)
						chromaOffsetL1 = append(chromaOffsetL1, chromaOffset)
					}
					pwt.ChromaWeightL1 = append(pwt.ChromaWeightL1, chromaWeightL1)
					pwt.ChromaOffsetL1 = append(pwt.ChromaWeightL1, chromaOffsetL1)
				}
			}
		}
	}

	return pwt, nil
}

// ParseDecRefPicMarking - AVC Slice Decoded Reference Picture Marking parser using bits r
func ParseDecRefPicMarking(r *bits.EBSPReader, naluType NaluType) (*DecRefPicMarking, error) {
	rpm := &DecRefPicMarking{}
	var err error

	if getIDRPicFlag(naluType) {
		rpm.NoOutputOfPriorPicsFlag, err = r.ReadFlag()
		if err != nil {
			return nil, err
		}
		rpm.LongTermReferenceFlag, err = r.ReadFlag()
		if err != nil {
			return nil, err
		}
	} else {
		rpm.AdaptiveRefPicMarkingModeFlag, err = r.ReadFlag()
		if err != nil {
			return nil, err
		}

		if rpm.AdaptiveRefPicMarkingModeFlag {
			for mmco := 1; mmco != 0; {
				arpmParams := AdaptiveMemCtrlDecRefPicMarkingParams{}
				arpmParams.MemoryManagementControlOperation, err = r.ReadExpGolomb()
				if err != nil {
					return nil, err
				}
				mmco := arpmParams.MemoryManagementControlOperation

				if mmco == 1 || mmco == 3 {
					arpmParams.DifferenceOfPicNumsMinus1, err = r.ReadExpGolomb()
					if err != nil {
						return nil, err
					}
				}
				if mmco == 2 {
					arpmParams.LongTermPicNum, err = r.ReadExpGolomb()
					if err != nil {
						return nil, err
					}
				}
				if mmco == 3 || mmco == 6 {
					arpmParams.LongTermFrameIdx, err = r.ReadExpGolomb()
					if err != nil {
						return nil, err
					}
				}
				if mmco == 4 {
					arpmParams.MaxLongTermFrameIdxPlus1, err = r.ReadExpGolomb()
					if err != nil {
						return nil, err
					}
				}

				rpm.AdaptiveRefPicMarkingParams = append(rpm.AdaptiveRefPicMarkingParams, arpmParams)
			}
		}

	}

	return rpm, nil
}

// getIDRPicFlag - Sets IDR Pic flag based on H.264 Spec equation 7-1
// The equation sets value as 1/0, but uses like a bool in the tabular syntax
func getIDRPicFlag(naluType NaluType) bool {
	if naluType == NALU_IDR {
		return true
	}
	return false
}

// getChromaArrayType - Sets ChromaArrayType based on the SeparateColourPlaneFlag from SPS
// this is based on H.264 spec 7.4.2.1.1
func getChromaArrayType(sps *SPS) uint {
	if sps.SeparateColourPlaneFlag {
		return 0
	}
	return sps.ChromaFormatIDC
}

// sliceHeaderExpected - Tells if a slice header should be expected: Nal types: 1,2,5,19
func sliceHeaderExpected(naluType NaluType) bool {
	switch naluType {
	case 1, 2, 5, 19:
		// slice_layer_without_partitioning_rbsp
		// slice_data_partition_a_layer_rbsp
	default:
		err := ErrNoSliceHeader
		return nil, err
	}
	nalRefIDC := (nalHdr >> 5) & 0x3
	sh.FirstMBInSlice = uint32(r.ReadExpGolomb())
	sh.SliceType = SliceType(r.ReadExpGolomb())
	sh.PicParamID = uint32(r.ReadExpGolomb())
	pps, ok := ppsMap[sh.PicParamID]
	if !ok {
		return nil, fmt.Errorf("pps ID %d unknown", sh.PicParamID)
	}
	spsID := pps.PicParameterSetID
	sps, ok := spsMap[uint32(spsID)]
	if !ok {
		return nil, fmt.Errorf("sps ID %d unknown", spsID)
	}
	if sps.SeparateColourPlaneFlag {
		sh.ColorPlaneID = uint32(r.Read(2))
	}
	sh.FrameNum = uint32(r.Read(int(sps.Log2MaxFrameNumMinus4 + 4)))
	if !sps.FrameMbsOnlyFlag {
		sh.FieldPicFlag = r.ReadFlag()
		if sh.FieldPicFlag {
			sh.BottomFieldFlag = r.ReadFlag()
		}
	}
	if naluType == NALU_IDR {
		sh.IDRPicID = uint32(r.ReadExpGolomb())
	}
	if sps.PicOrderCntType == 0 {
		sh.PicOrderCntLsb = uint32(r.Read(int(sps.Log2MaxPicOrderCntLsbMinus4 + 4)))
		if pps.BottomFieldPicOrderInFramePresentFlag && !sh.FieldPicFlag {
			sh.DeltaPicOrderCntBottom = int32(r.ReadSignedGolomb())
		}
	} else if sps.PicOrderCntType == 1 && !sps.DeltaPicOrderAlwaysZeroFlag {
		sh.DeltaPicOrderCnt[0] = int32(r.ReadSignedGolomb())
		if pps.BottomFieldPicOrderInFramePresentFlag && !sh.FieldPicFlag {
			sh.DeltaPicOrderCnt[1] = int32(r.ReadSignedGolomb())
		}
	}
	if pps.RedundantPicCntPresentFlag {
		sh.RedundantPicCnt = uint32(r.ReadExpGolomb())
	}

	sliceType := SliceType(sh.SliceType % 5)
	if sliceType == SLICE_B {
		sh.DirectSpatialMvPredFlag = r.ReadFlag()
	}

	switch sliceType {
	case SLICE_P, SLICE_SP, SLICE_B:
		sh.NumRefIdxActiveOverrideFlag = r.ReadFlag()

		if sh.NumRefIdxActiveOverrideFlag {
			sh.NumRefIdxL0ActiveMinus1 = uint32(r.ReadExpGolomb())
			if sliceType == SLICE_B {
				sh.NumRefIdxL1ActiveMinus1 = uint32(r.ReadExpGolomb())
			}
		} else {
			sh.NumRefIdxL0ActiveMinus1 = uint32(pps.NumRefIdxI0DefaultActiveMinus1)
			sh.NumRefIdxL1ActiveMinus1 = uint32(pps.NumRefIdxI1DefaultActiveMinus1)
		}
	}

	// ref_pic_list_modification (nal unit type != 20)
	if sliceType != SLICE_I && sliceType != SLICE_SI {
		sh.RefPicListModificationL0Flag = r.ReadFlag()
		if sh.RefPicListModificationL0Flag {
		refPicListL0Loop:
			for {
				sh.ModificationOfPicNumsIDC = uint32(r.ReadExpGolomb())
				switch sh.ModificationOfPicNumsIDC {
				case 0, 1:
					sh.AbsDiffPicNumMinus1 = uint32(r.ReadExpGolomb())
				case 2:
					sh.LongTermPicNum = uint32(r.ReadExpGolomb())
				case 3:
					break refPicListL0Loop
				}
				if r.AccError() != nil {
					break refPicListL0Loop
				}
			}
		}
	}
	if sliceType == SLICE_B {
		sh.RefPicListModificationL1Flag = r.ReadFlag()
		if sh.RefPicListModificationL1Flag {
		refPicListL1Loop:
			for {
				sh.ModificationOfPicNumsIDC = uint32(r.ReadExpGolomb())
				switch sh.ModificationOfPicNumsIDC {
				case 0, 1:
					sh.AbsDiffPicNumMinus1 = uint32(r.ReadExpGolomb())
				case 2:
					sh.LongTermPicNum = uint32(r.ReadExpGolomb())
				case 3:
					break refPicListL1Loop
				}
				if r.AccError() != nil {
					break refPicListL1Loop
				}
			}
		}
	}
	// end ref_pic_list_modification

	if pps.WeightedPredFlag && (sliceType == SLICE_P || sliceType == SLICE_SP) ||
		(pps.WeightedBipredIDC == 1 && sliceType == SLICE_B) {
		// pred_weight_table
		sh.LumaLog2WeightDenom = uint32(r.ReadExpGolomb())
		if sps.ChromaArrayType() != 0 {
			sh.ChromaLog2WeightDenom = uint32(r.ReadExpGolomb())
		}

		for i := uint32(0); i <= sh.NumRefIdxL0ActiveMinus1; i++ {
			lumaWeightL0Flag := r.ReadFlag()
			if lumaWeightL0Flag {
				// Just parse, don't store this
				_ = r.ReadExpGolomb() // luma_weight_l0[i] = SignedGolomb()
				_ = r.ReadExpGolomb() // luma_offset_l0[i] = SignedGolomb()
			}
			if sps.ChromaArrayType() != 0 {
				chromaWeightL0Flag := r.ReadFlag()
				if chromaWeightL0Flag {
					for j := 0; j < 2; j++ {
						// Just parse, don't store this
						_ = r.ReadExpGolomb() // chroma_weight_l0[i][j] = SignedGolomb()
						_ = r.ReadExpGolomb() // chroma_offset_l0[i][j] = SignedGolomb()
					}
				}
			}
		}
		if sliceType == SLICE_B {
			for i := uint32(0); i <= sh.NumRefIdxL1ActiveMinus1; i++ {
				lumaWeightL1Flag := r.ReadFlag()
				if lumaWeightL1Flag {
					// Just parse, don't store this
					_ = r.ReadExpGolomb() // luma_weight_l1[i] = SignedGolomb()
					_ = r.ReadExpGolomb() // luma_offset_l1[i] = SignedGolomb()
				}
				if sps.ChromaFormatIDC != 0 {
					chromaWeightL0Flag := r.ReadFlag()
					if chromaWeightL0Flag {
						// Just parse, don't store this
						for j := 0; j < 2; j++ {
							_ = r.ReadSignedGolomb() // chroma_weight_l1[i][j] = SignedGolomb()
							_ = r.ReadSignedGolomb() // chroma_offset_l1[i][j] = SignedGolomb()
						}
					}
				}
			}
		}
		// end pred_weight_table
	}

	if nalRefIDC != 0 {
		// dec_ref_pic_marking
		if naluType == NALU_IDR {
			sh.NoOutputOfPriorPicsFlag = r.ReadFlag()
			sh.LongTermReferenceFlag = r.ReadFlag()
		} else {
			sh.AdaptiveRefPicMarkingModeFlag = r.ReadFlag()
			if sh.AdaptiveRefPicMarkingModeFlag {
			adaptiveRefPicLoop:
				for {
					memoryManagementControlOperation := r.ReadExpGolomb()
					switch memoryManagementControlOperation {
					case 1, 3:
						sh.DifferenceOfPicNumsMinus1 = uint32(r.ReadExpGolomb())
					case 2:
						sh.LongTermPicNum = uint32(r.ReadExpGolomb())
					}
					switch memoryManagementControlOperation {
					case 3, 6:
						sh.LongTermFramIdx = uint32(r.ReadExpGolomb())
					case 4:
						sh.MaxLongTermFrameIdxPlus1 = uint32(r.ReadExpGolomb())
					case 0:
						break adaptiveRefPicLoop
					}
					if r.AccError() != nil {
						break adaptiveRefPicLoop
					}
				}
			}
		}
		// end dec_ref_pic_marking
	}
	if pps.EntropyCodingModeFlag && sliceType != SLICE_I && sliceType != SLICE_SI {
		sh.CabacInitIDC = uint32(r.ReadExpGolomb())
	}
	sh.SliceQPDelta = int32(r.ReadSignedGolomb())
	if sliceType == SLICE_SP || sliceType == SLICE_SI {
		if sliceType == SLICE_SP {
			sh.SPForSwitchFlag = r.ReadFlag()
		}
		sh.SliceQSDelta = int32(r.ReadSignedGolomb())
	}
	if pps.DeblockingFilterControlPresentFlag {
		sh.DisableDeblockingFilterIDC = uint32(r.ReadExpGolomb())
		if sh.DisableDeblockingFilterIDC != 1 {
			sh.SliceAlphaC0OffsetDiv2 = int32(r.ReadSignedGolomb())
			sh.SliceBetaOffsetDiv2 = int32(r.ReadSignedGolomb())
		}
	}
	if pps.NumSliceGroupsMinus1 > 0 &&
		pps.SliceGroupMapType >= 3 &&
		pps.SliceGroupMapType <= 5 {
		picSizeInMapUnits := pps.PicSizeInMapUnitsMinus1 + 1
		sliceGroupChangeRage := pps.SliceGroupChangeRateMinus1 + 1
		nrBits := int(math.Ceil(math.Log2(float64(picSizeInMapUnits/sliceGroupChangeRage + 1))))
		sh.SliceGroupChangeCycle = uint32(r.Read(nrBits))
	}

	// compute the size in bytes. The last byte may not be fully read
	sh.Size = uint32(r.NrBytesRead())
	return &sh, nil
}
