package hevc

import (
	"math"
)

// SliceSegmentHeader - 7.3.6.1 General HEVC Slice Segment Header syntax
type SliceSegmentHeader struct {
	FirstSliceSegmentInPicFlag              bool   // u(1)
	NoOutputOfPriorPicsFlag                 bool   // u(1)
	SlicePicParameterSetID                  uint   // ue(v)
	DependentSliceSegmentFlag               bool   // u(1)
	SliceSegmentAddress                     uint   // u(v)
	SliceReservedFlag                       []bool // []u(1)
	SliceType                               uint   // ue(v) -- add an enum?
	PicOutputFlag                           bool   // u(1)
	ColourPlaneID                           uint   // u(2)
	SlicePicOrderCntLsb                     uint   // u(v)
	ShortTermRefPicSetSPSFlag               bool   // u(1)
	ShortTermRefPicSet                      *STRefPicSet
	ShortTermRefPicSetIdx                   uint   // u(v)
	//CurrRPSIdx 								uint   // calculated from STRefPicFlag and / or Idx
	NumLongTermSPS                          uint   // ue(v)
	NumLongTermPics                         uint   // ue(v)
	LtIdxSPS                                []uint // u(v)
	POCLSBLt                                []uint // u(v)
	UsedByCurrPicLtFlag                     []bool // u(1)
	DeltaPOCMSBPresentFlag                  []bool // u(1)
	DeltaPOCMSBCycleLt                      []uint // ue(v)
	SliceTemporalMVPEnabledFlag             bool   // u(1)
	SliceSAOLumaFlag                        bool   // u(1)
	SliceSAOChromaFlag                      bool   // u(1)
	NumRefIdxActiveOverrideFlag             bool   // u(1)
	NumRefIdxL0ActiveMinus1                 uint   // ue(v)
	NumRefIdxL1ActiveMinus1                 uint   // ue(v)
	RefPicListModification                  *RefPicListModification
	MVDL1ZeroFlag                           bool // u(1)
	CabacInitFlag                           bool // u(1)
	CollocatedFromL0Flag                    bool // u(1)
	CollocatedRefIdx                        uint // ue(v)
	PredWeightTable                         *PredWeightTable
	FiveMinusMaxNumMergeCand                uint   // ue(v)
	UseIntegerMVFlag                        bool   // u(1)
	SliceQPDelta                            int    // se(v)
	SliceCbQPOffset                         int    // se(v)
	SliceCrQPOffset                         int    // se(v)
	SliceActYQPOffset                       int    // se(v)
	SliceActCbQPOffset                      int    // se(v)
	SliceActCrQPOffset                      int    // se(v)
	CuChromaQPOffsetEnabledFlag             bool   // u(1)
	DeblockingFilterOverrideFlag            bool   // u(1)
	SliceDeblockingFilterDisabledFlag       bool   // u(1)
	SliceBetaOffsetDiv2                     int    // se(v)
	SliceTcOffsetDiv2                       int    // se(v)
	SliceLoopFilterAccrossSlicesEnabledFlag bool   // u(1)
	NumEntryPointOffsets                    uint   // ue(V)
	OffsetLenMinus1                         uint   // ue(v)
	EntryPointOffsetMinus1                  []uint // u(v)
	SliceSegmentHeaderExtensionLength       uint   // ue(v)
	SliceSegmentHeaderExtensionDataByte     []uint // u(8)
}

// RefPicListsModification - 7.3.6.2 Reference picture list modification syntax
type RefPicListModification struct {
	RefPicListModificationFlagL0 bool // u(1)
	ListEntryL0                  uint // u(v)
	RefPicListModificationFlagL1 bool // u(1)
	ListEntryL1                  uint // u(v)
}

// PredWeightTable - 7.3.6.3 Weighted prediction parameters syntax
type PredWeightTable struct {
	LumaLog2WeightDenom        uint    // ue(v)
	DeltaChromaLog2WeightDenom int     // se(v)
	LumaWeightL0Flag           []bool  // u(1)
	ChromaWeightL0Flag         []bool  // u(1)
	DeltaLumaWeightL0          []int   // se(v)
	LumaOffsetL0               []int   // se(v)
	DeltaChromaWeightL0        [][]int // se(v)
	DeltaChromaOffsetL0        [][]int // se(v)
	LumaWeightL1Flag           []bool  // u(1)
	ChromaWeightL1Flag         []bool  // u(1)
	DeltaLumaWeightL1          []int   // se(v)
	LumaOffsetL1               []int   // se(v)
	DeltaChromaWeightL1        [][]int // se(v)
	DeltaChromaOffsetL1        [][]int // se(v)
}

// STRefPicSet - 7.3.7 Short-term reference picture set syntax
type STRefPicSet struct {
	InterRefPicSetPredictionFlag bool   // u(1)
	DeltaIdxMinus1               uint   // ue(v)
	DeltaRPSSign                 bool   // u(1)
	AbsDeltaRPSMinus1            uint   // ue(v)
	UsedByCurrPicLtFlag          []bool // u(1)
	UseDeltaFlag                 []bool // u(1)
	NumNegativePics              uint   // ue(v)
	NumPositivePics              uint   // ue(v)
	DeltaPOCS0Minus1             []uint // ue(v)
	UsedByCurrPicS0Flag          []bool // u(1)
	DeltaPOCS1Minus1             []uint // ue(v)
	UsedByCurrPicS1Flag          []bool // u(1)
}

// getNumPicTotalCurr - derived as per eq: 7.5.7 from section 7.4.7.2
func getNumPicTotalCurr(uint currRpsIdx) () {
	numPicTotalCurr = 0
	for i = 0; i < NumNegativePics[currRpsIdx]; i++ {
		if UsedByCurrPicS0[currRpsIdx][i] {
			NumPicTotalCurr++
		}
	}
	for i = 0; i < NumPositivePics[currRpsIdx]; i++ {
		if UsedByCurrPicS1[currRpsIdx][i] {
			NumPicTotalCurr++
		}
	}
	for i = 0; i < num_long_term_sps + num_long_term_pics; i++ ) {
		if UsedByCurrPicLt[i] {
			NumPicTotalCurr++
		}
		if pps_curr_pic_ref_enabled_flag {
			NumPicTotalCurr++
		}
	}
}

// ParseRefPicListModification - Parse Ref Pic List Mod as per 7.3.6.2
func ParseRefPicListModification(r *bits.EBSPReader, sh *SliceHeader) (*RefPicListModification, error) {
	rplm := &RefPicListModification{}
	var err error

	rplm.RefPicListModificationFlagL0, err = r.ReadFlag()
	if err != nil {
		return nil, err
	}
	listEntryBits := int(math.Ceil(math.Log2(float64(NumPicTotalCurr))))
	if rplm.RefPicListModificationFlagL0 {
		for uint i := 0; i <= sh.NumRefIdxL0ActiveMinus1; i++ {
			rplm.ListEntryL0[i] = r.Read(listEntryBits)
		}
	}
	if sh.sliceType == SLICE_B {
		rplm.RefPicListModificationFlagL1, err = r.ReadFlag()
		if err != nil {
			return nil, err
		}
		if rplm.RefPicListModificationFlagL1 {
			for uint i := 0; i <= sh.NumRefIdxL1ActiveMinus1; i++ {
				rplm.ListEntryL1[i] = r.Read(listEntryBits)
			}
		}
	}
} 

// The variable ChromaArrayType is derived as equal to 0 when separate_colour_plane_flag is equal to 1 and chroma_format_idc is equal to 3. In the decoding process, the value of this variable is evaluated resulting in operations identical to that of monochrome pictures (when chroma_format_idc is equal to 0).

// picLayerId - value of the nuh_layer_id of the VCL NAL units in the picture picX.
func picLayerId()

// ParsePredWeightTable - Parse the 7.3.6.3 Weighted prediction parameters syntax
func ParsePredWeightTable(r *bits.EBSPReader, sh *SliceHeader) (*PredWeightTable, error) {
	pwm := &PredWeightTable{}
	var err error

	pwm.LumaLog2WeightDenom, err = r.ReadExpGolomb()
	if err != nil {
		return nil, err
	}s
	if getChromaArrayType() != 0 {
		pwm.DeltaChromaLog2WeightDenom, err = r.ReadSignedGolomb()
		if err != nil {
			return nil, err
		}
	}
	for uint i = 0; i <= sh.NumRefIdxL0ActiveMinus1; i++ {
		if 
	}
}

// need to name function
func abc() () {
	if inter_ref_pic_set_prediction_flag == 1 {
		i =0
		for int j := NumPositivePics[RefRpsIdx]−1; j >= 0; j−− {
			dPoc = DeltaPocS1[RefRpsIdx][j] + deltaRps 
			if dPoc < 0 && use_delta_flag[NumNegativePics[RefRpsIdx]+j] {
				DeltaPocS0[stRpsIdx][i] = dPoc
				UsedByCurrPicS0[stRpsIdx][i++] = used_by_curr_pic_flag[NumNegativePics[RefRpsIdx]+j] 
			}
		}
		if deltaRps < 0 && use_delta_flag[NumDeltaPocs[RefRpsIdx]] {
			DeltaPocS0[stRpsIdx][i] = deltaRps
			UsedByCurrPicS0[stRpsIdx][i++] = used_by_curr_pic_flag[NumDeltaPocs[RefRpsIdx]] 
		}
		for j = 0; j < NumNegativePics[RefRpsIdx]; j++ {
			dPoc = DeltaPocS0[RefRpsIdx][j] + deltaRps 
			if dPoc < 0 && use_delta_flag[j] {
				DeltaPocS0[stRpsIdx][i] = dPoc
				UsedByCurrPicS0[stRpsIdx][i++] = used_by_curr_pic_flag[j] 
			}
		}
		NumNegativePics[stRpsIdx] = i
	
		i =0
		for j = NumNegativePics[RefRpsIdx]−1; j >= 0; j−− {
			dPoc = DeltaPocS0[RefRpsIdx][j] + deltaRps 
			if dPoc > 0 && use_delta_flag[j] {
				DeltaPocS1[stRpsIdx][i] = dPoc
				UsedByCurrPicS1[stRpsIdx][i++] = used_by_curr_pic_flag[j] 
			}
		}
		if deltaRps > 0 && use_delta_flag[NumDeltaPocs[RefRpsIdx]] {
			DeltaPocS1[stRpsIdx][i] = deltaRps
			UsedByCurrPicS1[stRpsIdx][i++] = used_by_curr_pic_flag[NumDeltaPocs[RefRpsIdx]] 
		}
		for j = 0; j < NumPositivePics[RefRpsIdx]; j++ {
			dPoc = DeltaPocS1[RefRpsIdx][j] + deltaRps
			if dPoc > 0 && use_delta_flag[NumNegativePics[RefRpsIdx] + j] {
				DeltaPocS1[stRpsIdx][i] = dPoc
				UsedByCurrPicS1[stRpsIdx][i++] = used_by_curr_pic_flag[NumNegativePics[RefRpsIdx]+j]
			} 
		}
		NumPositivePics[stRpsIdx] = i
	} else {
		NumNegativePics[stRpsIdx] = num_negative_pics
		NumPositivePics[stRpsIdx] = num_positive_pics
		UsedByCurrPicS0[stRpsIdx][i] = used_by_curr_pic_s0_flag[i] 
		UsedByCurrPicS1[stRpsIdx][i] = used_by_curr_pic_s1_flag[i]
		if i == 0 {			
			DeltaPocS0[stRpsIdx][i] = −(delta_poc_s0_minus1[i]+1)
			DeltaPocS1[stRpsIdx][i] = delta_poc_s1_minus1[i] + 1
		} else {
			DeltaPocS0[stRpsIdx][i] = DeltaPocS0[stRpsIdx][i−1] − (delta_poc_s0_minus1[i] + 1)
			DeltaPocS1[stRpsIdx][i] = DeltaPocS1[stRpsIdx][i−1] + (delta_poc_s1_minus1[i] + 1)
		}
		NumDeltaPocs[stRpsIdx] = NumNegativePics[stRpsIdx] + NumPositivePics[stRpsIdx]
	}
}

// sliceHeaderExpected - Tells if a slice header should be expected: Nal types: 1,2,5,19
func sliceHeaderExpected(naluType NaluType) bool {
	// slice_segment_layer_rbsp() is present in the following Nal Unit Types
	// TRAIL_N / TRAIL_R - 0 / 1
	// TSA_N / TSA_R     - 2 / 3
	// STSA_N / STSA_R   - 4 / 5
	// RADL_N / RADL_R   - 6 / 7
	// RASL_N / RASL_R   - 8 / 9
	// BLA_W_LP / BLA_W_RADL / BLA_N_LP - 16 / 17 / 18
	// IDR_W_RADL / IDR_W_RADL - 19 / 20
	// CRA_NUT - 21
	if naluType >= 0 && naluType <= 9 || naluType >= 16 && naluType <= 21 {
		return true
	} else {
		return false
	}
}

// ParseSliceSegmentHeader - Parse the Slice segment header from VCL NAL unit
func ParseSliceSegmentHeader(data []byte, sps *SPS, pps *PPS) (*SliceHeader, int, error) {
	hevcssh := &SliceSegmentHeader{}
	var err error

	rd := bytes.NewReader(data)
	r := bits.NewEBSPReader(rd)
	// Note! First two bytes are NALU Header

	naluHdrBits := r.Read(16)
	naluType := GetNaluType(byte(naluHdrBits >> 8))
	if !sliceHeaderExpected(naluType) {
		return nil, fmt.Errorf("No slice header")
	}

	hevcssh.FirstSliceSegmentInPicFlag, err = r.ReadFlag()
	if err != nil {
		return nil, r.NrBytesRead(), err
	}

	if naluType >= BLA_W_LP && naluType <= NALU_RSV_IRAP_VCL23 {
		hevcssh.NoOutputOfPriorPicsFlag, err = r.ReadFlag()
		if err != nil {
			return nil, r.NrBytesRead(), err
		}
	}

	if !hevcssh.FirstSliceSegmentInPicFlag {
		if pps.DependentSliceSegmentsEnabledFlag
	}

}