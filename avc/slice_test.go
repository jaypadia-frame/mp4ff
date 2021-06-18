package avc

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/go-test/deep"
)

const (
	// Slice Type Test Data
	videoNaluStart = "25888040ffde08e47a7bff05ab"
	// IDR Test Data
	videoSliceDataIDR = "6588840B5B07C341"
	SPSIDRTest        = "674d4028d900780227e59a808080a000000300c0000023c1e30649"
	PPSIDRTest        = "68ebc08cf2"
	// P Frame Test Data
	videoSliceDataPFrame = "419A384603FA42D6FFB5F01137F156003C"
	SPSPFrameTest        = "674d4028d900780227e59a808080a000000300c0000023c1e30649"
	PPSPFrameTest        = "68ebc08cf2"
	// P Frame Encrypted Slice Data Test - slice data (after slice header is encrypted)
	videoSliceDataPFrameEnc = "419A384603FA42D6FF62ADEB"
	// Dec Ref pic marking present
	SPSRefPicMod             = "674d4028d900780227e59a808080a000000300c0000023c1e30649"
	PPSRefPicMod             = "68ebc08cf2"
	videoSliceDataRefPicMod  = "419ab27843c994c08eb70001ae9cc514978189bd51a8bce3a781b4a2b6c16a4b24ae3d95e7eed7f88500"
	videoSliceDataRefPicMod2 = "419a727843c994c08bb7919c804eb2def5b6c63acc15eedcb66"
	// frame io tests
	videoSliceDecRefPicMarking = "419a4f0864ca611f6ffe9e213ed705ab96e200580cf45006ba6fac874bbc96c4b96eccc36853d6537ef172c01f82"
)

func TestSliceTypeParser(t *testing.T) {
	byteData, _ := hex.DecodeString(videoNaluStart)
	want := SLICE_I
	got, err := GetSliceTypeFromNALU(byteData)
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestSliceHeaderParserIDR(t *testing.T) {
	// SPS needed to parse PPS and Slice Header
	spsData, _ := hex.DecodeString(SPSIDRTest)
	sps, err := ParseSPSNALUnit(spsData, false)
	if err != nil {
		t.Errorf("Parse IDR Failed to parse SPS")
	}
	// PPS needed to Parse Slice Header
	ppsData, _ := hex.DecodeString(PPSIDRTest)
	pps, err := ParsePPSNALUnit(ppsData, sps)
	if err != nil {
		t.Errorf("Parse IDR Failed to parse PPS")
	}

	byteData, _ := hex.DecodeString(videoSliceDataIDR) // Actual slice header data
	_, _, err = ParseSliceHeader(byteData, sps, pps)
	if err != nil {
		t.Error(err)
	}
}

func TestSliceHeaderParserPFrame(t *testing.T) {
	// SPS needed to parse PPS and Slice Header
	spsData, _ := hex.DecodeString(SPSPFrameTest)
	sps, err := ParseSPSNALUnit(spsData, false)
	if err != nil {
		t.Errorf("Parse PFrame Failed to parse SPS")
	}
	// PPS needed to Parse Slice Header
	ppsData, _ := hex.DecodeString(PPSPFrameTest)
	pps, err := ParsePPSNALUnit(ppsData, sps)
	if err != nil {
		t.Errorf("Parse PFrame Failed to parse PPS")
	}
	// Actual slice header data plus unencrypted slice data
	byteData, _ := hex.DecodeString(videoSliceDataPFrame)
	_, _, err = ParseSliceHeader(byteData, sps, pps)
	if err != nil {
		t.Error(err)
	}
}

func TestSliceHeaderParserPFrameEnc(t *testing.T) {
	// SPS needed to parse PPS and Slice Header
	spsData, _ := hex.DecodeString(SPSPFrameTest)
	sps, err := ParseSPSNALUnit(spsData, false)
	if err != nil {
		t.Errorf("Parse PFrame Failed to parse SPS")
	}
	// PPS needed to Parse Slice Header
	ppsData, _ := hex.DecodeString(PPSPFrameTest)
	pps, err := ParsePPSNALUnit(ppsData, sps)
	if err != nil {
		t.Errorf("Parse PFrame Failed to parse PPS")
	}
	// Actual slice header plus encrypted slice data
	byteData, _ := hex.DecodeString(videoSliceDataPFrameEnc)
	_, _, err = ParseSliceHeader(byteData, sps, pps)
	if err != nil {
		t.Error(err)
	}
}

func TestSliceHeaderParserRefPicMod(t *testing.T) {
	// SPS needed to parse PPS and Slice Header
	spsData, _ := hex.DecodeString(SPSRefPicMod)
	sps, err := ParseSPSNALUnit(spsData, false)
	if err != nil {
		t.Errorf("Parse PFrame Failed to parse SPS")
	}
	// PPS needed to Parse Slice Header
	ppsData, _ := hex.DecodeString(PPSRefPicMod)
	pps, err := ParsePPSNALUnit(ppsData, sps)
	if err != nil {
		t.Errorf("Parse PFrame Failed to parse PPS")
	}
	// Actual slice header plus encrypted slice data
	byteData, _ := hex.DecodeString(videoSliceDataRefPicMod)
	_, sz, err := ParseSliceHeader(byteData, sps, pps)
	if err != nil {
		t.Error(err)
	}
	if sz != 10 {
		t.Errorf("AVC Slice header size not parsed correctly")
	}
}

func TestSliceHeaderParserRefPicMod2(t *testing.T) {
	// SPS needed to parse PPS and Slice Header
	spsData, _ := hex.DecodeString(SPSRefPicMod)
	sps, err := ParseSPSNALUnit(spsData, false)
	if err != nil {
		t.Errorf("Parse PFrame Failed to parse SPS")
	}
	// PPS needed to Parse Slice Header
	ppsData, _ := hex.DecodeString(PPSRefPicMod)
	pps, err := ParsePPSNALUnit(ppsData, sps)
	if err != nil {
		t.Errorf("Parse PFrame Failed to parse PPS")
	}
	// Actual slice header plus encrypted slice data
	byteData, _ := hex.DecodeString(videoSliceDataRefPicMod2)
	_, sz, err := ParseSliceHeader(byteData, sps, pps)
	if err != nil {
		t.Error(err)
	}
	if sz != 10 {
		t.Errorf("AVC Slice header size not parsed correctly")
	}
}

func TestSliceHeaderParserDecRefPicMarking(t *testing.T) {
	// SPS needed to parse PPS and Slice Header
	spsData, _ := hex.DecodeString(SPSRefPicMod)
	sps, err := ParseSPSNALUnit(spsData, false)
	if err != nil {
		t.Errorf("Parse PFrame Failed to parse SPS")
	}
	// PPS needed to Parse Slice Header
	ppsData, _ := hex.DecodeString(PPSRefPicMod)
	pps, err := ParsePPSNALUnit(ppsData, sps)
	if err != nil {
		t.Errorf("Parse PFrame Failed to parse PPS")
	}
	// Actual slice header plus slice data
	byteData, _ := hex.DecodeString(videoSliceDecRefPicMarking)
	_, sz, err := ParseSliceHeader(byteData, sps, pps)
	if err != nil {
		t.Error(err)
	}
	if sz != 9 {
		t.Errorf("AVC Slice header size not parsed correctly.")
	}
}

// Test coverage that needs to be added - B slice, SP & SI slices
