package sei_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/Eyevinn/mp4ff/sei"
	"github.com/go-test/deep"
)

func TestSEIStrings(t *testing.T) {
	cases := []struct {
		seiID     int
		seiString string
	}{
		{1, "SEIPicTimingType (1)"},
		{137, "SEIMasteringDisplayColourVolumeType (137)"},
		{144, "SEIContentLightLevelInformationType (144)"},
	}

	for _, tc := range cases {
		got := sei.SEIType(tc.seiID).String()
		if got != tc.seiString {
			t.Errorf("got %s, wanted %s", got, tc.seiString)
		}
	}
}

func TestMasteringDisplayColourVolumeSEI(t *testing.T) {
	hex137 := "11223344556677889900aabbccddeeff0011223344556677"
	pl, err := hex.DecodeString(hex137)
	if err != nil {
		t.Error(err)
	}
	seiData := sei.NewSEIData(sei.SEIMasteringDisplayColourVolumeType, pl)
	msg, err := sei.DecodeMasteringDisplayColourVolumeSEI(seiData)
	if err != nil {
		t.Error(err)
	}
	if msg.Type() != sei.SEIMasteringDisplayColourVolumeType {
		t.Errorf("got SEI type %d, wanted %d", msg.Type(), sei.SEIMasteringDisplayColourVolumeType)
	}
	decPl := msg.Payload()
	if !bytes.Equal(decPl, pl) {
		t.Errorf("decoded payload differs from expected")
	}
}

func TestContentLightLevelInformationSEI(t *testing.T) {
	hex144 := "11223344"
	pl, err := hex.DecodeString(hex144)
	if err != nil {
		t.Error(err)
	}
	seiData := sei.NewSEIData(sei.SEIContentLightLevelInformationType, pl)
	msg, err := sei.DecodeContentLightLevelInformationSEI(seiData)
	if err != nil {
		t.Error(err)
	}
	if msg.Type() != sei.SEIContentLightLevelInformationType {
		t.Errorf("got SEI type %d, wanted %d", msg.Type(), sei.SEIContentLightLevelInformationType)
	}
	decPl := msg.Payload()
	if !bytes.Equal(decPl, pl) {
		t.Errorf("decoded payload differs from expected")
	}
}

func TestPicTimingAvcSEI(t *testing.T) {
	testCases := []struct {
		seiPayload string // after SEI type byte 01 and length byte(s)
		wantedSEI  sei.PicTimingAvcSEI
	}{
		{
			"0904078c1080",
			sei.PicTimingAvcSEI{
				CbpDbpDelay:      nil,
				TimeOffsetLength: 0,
				PictStruct:       0,
				Clocks: []sei.ClockTSAvc{
					{
						ClockTimeStampFlag: true,
						CtType:             0,
						NuitFieldBasedFlag: true,
						CountingType:       0,
						NFrames:            7,
						Hours:              1,
						Minutes:            1,
						Seconds:            35,
						FullTimeStampFlag:  true,
						CntDroppedFlag:     false,
					},
				},
			},
		},
		{
			"1b0509b80000",
			sei.PicTimingAvcSEI{
				CbpDbpDelay:      nil,
				TimeOffsetLength: 0,
				PictStruct:       1,
				Clocks: []sei.ClockTSAvc{
					{
						ClockTimeStampFlag: true,
						CtType:             1,
						NuitFieldBasedFlag: true,
						CountingType:       0,
						NFrames:            9,
						Hours:              0,
						Minutes:            0,
						Seconds:            46,
						FullTimeStampFlag:  true,
						CntDroppedFlag:     true,
					},
				},
			},
		},
		{
			"2b0509b80000",
			sei.PicTimingAvcSEI{
				CbpDbpDelay:      nil,
				TimeOffsetLength: 0,
				PictStruct:       2,
				Clocks: []sei.ClockTSAvc{
					{
						ClockTimeStampFlag: true,
						CtType:             1,
						NuitFieldBasedFlag: true,
						CountingType:       0,
						NFrames:            9,
						Hours:              0,
						Minutes:            0,
						Seconds:            46,
						FullTimeStampFlag:  true,
						CntDroppedFlag:     true,
					},
				},
			},
		},
		{
			"00000000000000021208114de1", // with HRD parameters
			sei.PicTimingAvcSEI{
				CbpDbpDelay: &sei.CbpDbpDelay{
					CpbRemovalDelay:                    0,
					DpbOutputDelay:                     1,
					InitialCpbRemovalDelayLengthMinus1: 26,
					CpbRemovalDelayLengthMinus1:        30,
					DpbOutputDelayLengthMinus1:         31,
				},
				TimeOffsetLength: 0,
				PictStruct:       0,
				Clocks: []sei.ClockTSAvc{
					{
						ClockTimeStampFlag: true,
						CtType:             0,
						NuitFieldBasedFlag: true,
						CountingType:       0,
						NFrames:            8,
						Hours:              1,
						Minutes:            47,
						Seconds:            41,
						FullTimeStampFlag:  true,
						CntDroppedFlag:     false,
					},
				},
			},
		},
		{
			"00000008000000021208313de1", // with HRD parameters
			sei.PicTimingAvcSEI{
				CbpDbpDelay: &sei.CbpDbpDelay{
					CpbRemovalDelay:                    4,
					DpbOutputDelay:                     1,
					InitialCpbRemovalDelayLengthMinus1: 26,
					CpbRemovalDelayLengthMinus1:        30,
					DpbOutputDelayLengthMinus1:         31,
				},
				TimeOffsetLength: 0,
				PictStruct:       0,
				Clocks: []sei.ClockTSAvc{
					{
						ClockTimeStampFlag: true,
						CtType:             0,
						NuitFieldBasedFlag: true,
						CountingType:       0,
						NFrames:            24,
						Hours:              1,
						Minutes:            47,
						Seconds:            39,
						FullTimeStampFlag:  true,
						CntDroppedFlag:     false,
					},
				},
			},
		},
		{
			"0000000c000000021208313de1", // with HRD parameters
			sei.PicTimingAvcSEI{
				CbpDbpDelay: &sei.CbpDbpDelay{
					CpbRemovalDelay:                    6,
					DpbOutputDelay:                     1,
					InitialCpbRemovalDelayLengthMinus1: 26,
					CpbRemovalDelayLengthMinus1:        30,
					DpbOutputDelayLengthMinus1:         31,
				},
				TimeOffsetLength: 0,
				PictStruct:       0,
				Clocks: []sei.ClockTSAvc{
					{
						ClockTimeStampFlag: true,
						CtType:             0,
						NuitFieldBasedFlag: true,
						CountingType:       0,
						NFrames:            24,
						Hours:              1,
						Minutes:            47,
						Seconds:            39,
						FullTimeStampFlag:  true,
						CntDroppedFlag:     false,
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			pl, err := hex.DecodeString(tc.seiPayload)
			if err != nil {
				t.Fail()
			}
			seiData := sei.NewSEIData(sei.SEIPicTimingType, pl)
			msg, err := sei.DecodePicTimingAvcSEIHRD(seiData, tc.wantedSEI.CbpDbpDelay, tc.wantedSEI.TimeOffsetLength)
			if err != nil {
				t.Error(err)
			}
			if msg.Type() != sei.SEIPicTimingType {
				t.Errorf("got SEI type %d, wanted %d", msg.Type(), sei.SEIPicTimingType)
			}
			picTiming := msg.(*sei.PicTimingAvcSEI)
			if len(picTiming.Clocks) != len(tc.wantedSEI.Clocks) {
				t.Errorf("got %d clocks, wanted %d", len(picTiming.Clocks), len(tc.wantedSEI.Clocks))
			}
			diff := deep.Equal(picTiming, &tc.wantedSEI)
			if diff != nil {
				t.Errorf("case %d: %s", i, diff)
			}
			decPl := msg.Payload()
			if !bytes.Equal(decPl, pl) {
				t.Errorf("decoded payload differs from expected")
				fmt.Printf("decPl: %s\n", hex.EncodeToString(decPl))
				fmt.Printf("pl:    %s\n", hex.EncodeToString(pl))
			}
		})
	}
}

func TestTimeCodeSEI(t *testing.T) {
	seiHex := "60404198b410"
	pl, err := hex.DecodeString(seiHex)
	if err != nil {
		t.Error(err)
	}
	seiData := sei.NewSEIData(sei.SEITimeCodeType, pl)
	msg, err := sei.DecodeTimeCodeSEI(seiData)
	if err != nil {
		t.Error(err)
	}
	if msg.Type() != sei.SEITimeCodeType {
		t.Errorf("got SEI type %d, wanted %d", msg.Type(), sei.SEITimeCodeType)
	}
	decPl := msg.Payload()
	if !bytes.Equal(decPl, pl) {
		t.Errorf("decoded payload differs from expected")
	}
}

const (
	// The following examples are without NAL Unit header
	sei0Hex      = "0007810f1c0050744080"
	seiCEA608Hex = "0434b500314741393403cefffc9420fc94aefc9162fce56efc67bafc91b9" +
		"fcb0b0fcbab0fcb0bafcb031fcbab0fcb080fc942cfc942f80"
	seiAVCMulti             = "0001c001061b0509b8000080"
	missingRbspTrailingBits = "01061b0509b80000"
	seiHEVCMulti            = "000a8000000300403dc017a6900105040000be05880660404198b41080"
	seiHEVCHDR              = "891800000300000300000300000300000300000300000300000300000300000300000300009004000003000080"
)

func TestParseSEI(t *testing.T) {

	testCases := []struct {
		name           string
		codec          sei.Codec
		naluHex        string
		wantedTypes    []uint
		wantedStrings  []string
		expNonFatalErr error
	}{
		{"AVC multi", sei.AVC, seiAVCMulti, []uint{0, 1},
			[]string{
				`SEIBufferingPeriodType (0), size=1, "c0"`,
				`SEIPicTimingType (1), size=6, time=00:00:46:09 offset=0`,
			},
			nil,
		},
		{"Missing RBSP Trailing Bits", sei.AVC, missingRbspTrailingBits, []uint{1},
			[]string{
				`SEIPicTimingType (1), size=6, time=00:00:46:09 offset=0`,
			},
			sei.ErrRbspTrailingBitsMissing,
		},
		{"Type 0", sei.AVC, sei0Hex, []uint{0}, []string{`SEIBufferingPeriodType (0), size=7, "810f1c00507440"`}, nil},
		{"CEA-608", sei.AVC, seiCEA608Hex, []uint{4},
			[]string{`SEI type 4 CEA-608, size=52, field1: "942094ae9162e56e67ba91b9b0b0bab0b0bab031bab0b080942c942f", field2: ""`}, nil},
		{"HEVC multi", sei.HEVC, seiHEVCMulti, []uint{0, 1, 136},
			[]string{
				`SEIBufferingPeriodType (0), size=10, "80000000403dc017a690"`,
				`SEIPicTimingType (1), size=5, "040000be05"`,
				`SEITimeCodeType (136), size=6, time=13:49:12:08 offset=0`,
			},
			nil,
		},
		{"Type HDR HEVC", sei.HEVC, seiHEVCHDR, []uint{137, 144},
			[]string{
				"SEIMasteringDisplayColourVolumeType (137) 24B: primaries=(0, 0) (0, 0) (0, 0)," +
					" whitePoint=(0, 0), maxLum=0, minLum=0",
				"SEIContentLightLevelInformationType (144) 4B: maxContentLightLevel=0, maxPicAverageLightLevel=0",
			},
			nil,
		},
	}

	for _, tc := range testCases {
		seiNALU, _ := hex.DecodeString(tc.naluHex)

		rs := bytes.NewReader(seiNALU)

		seis, err := sei.ExtractSEIData(rs)
		if err != nil && err != tc.expNonFatalErr {
			t.Error(err)
		}
		if len(seis) != len(tc.wantedStrings) {
			t.Errorf("%s: Not %d but %d sei messages found", tc.name, len(tc.wantedStrings), len(seis))
		}
		for i := range seis {
			seiMessage, err := sei.DecodeSEIMessage(&seis[i], tc.codec)
			if err != nil {
				t.Error(err)
			}
			if seiMessage.Type() != tc.wantedTypes[i] {
				t.Errorf("%s: got SEI type %d instead of %d", tc.name, seiMessage.Type(), tc.wantedTypes[i])
			}
			if seiMessage.String() != tc.wantedStrings[i] {
				t.Errorf("%s: got %q instead of %q", tc.name, seiMessage.String(), tc.wantedStrings[i])
			}
		}
	}
}

func TestWriteSEI(t *testing.T) {

	cases := []struct {
		name  string
		codec sei.Codec
		hex   string
	}{
		{"seiHEVCHDR", sei.HEVC, seiHEVCHDR},
	}
	for _, tc := range cases {
		seiNALU, _ := hex.DecodeString(tc.hex)
		rs := bytes.NewReader(seiNALU)
		seis, err := sei.ExtractSEIData(rs)
		if err != nil {
			t.Error(err)
		}
		var seiMessages []sei.SEIMessage
		for i := range seis {
			seiMessage, err := sei.DecodeSEIMessage(&seis[i], tc.codec)
			if err != nil {
				t.Error(err)
			}
			seiMessages = append(seiMessages, seiMessage)
		}
		buf := bytes.Buffer{}
		err = sei.WriteSEIMessages(&buf, seiMessages)
		if err != nil {
			t.Error(err)
		}
		output := buf.Bytes()
		outputHex := hex.EncodeToString(output)
		if outputHex != tc.hex {
			t.Errorf("%s: wanted %s but got %s", tc.name, tc.hex, outputHex)
		}
	}

}

func TestParseAVCPicTimingWithHRD(t *testing.T) {
	sei1AVCEbsp := "010d00000300000300000300021208114de180"
	cbpDelay := sei.CbpDbpDelay{
		CpbRemovalDelay:                    0,
		DpbOutputDelay:                     0,
		InitialCpbRemovalDelayLengthMinus1: 26,
		CpbRemovalDelayLengthMinus1:        30,
		DpbOutputDelayLengthMinus1:         31,
	}
	timeOffsetLen := byte(0)

	testCases := []struct {
		name           string
		codec          sei.Codec
		naluHex        string
		wantedTypes    []uint
		wantedStrings  []string
		expNonFatalErr error
	}{
		{"PicTimingWithHRD", sei.AVC, sei1AVCEbsp, []uint{1},
			[]string{
				`SEIPicTimingType (1), size=13, time=01:47:41:08 offset=0`,
			},
			sei.ErrRbspTrailingBitsMissing,
		},
	}

	for _, tc := range testCases {
		seiNALU, _ := hex.DecodeString(tc.naluHex)

		rs := bytes.NewReader(seiNALU)

		seis, err := sei.ExtractSEIData(rs)
		if err != nil && err != tc.expNonFatalErr {
			t.Error(err)
		}
		if len(seis) != len(tc.wantedStrings) {
			t.Errorf("%s: Not %d but %d sei messages found", tc.name, len(tc.wantedStrings), len(seis))
		}
		for i := range seis {
			seiMessage, err := sei.DecodePicTimingAvcSEIHRD(&seis[i], &cbpDelay, timeOffsetLen)
			if err != nil {
				t.Error(err)
			}
			if seiMessage.Type() != tc.wantedTypes[i] {
				t.Errorf("%s: got SEI type %d instead of %d", tc.name, seiMessage.Type(), tc.wantedTypes[i])
			}
			if seiMessage.String() != tc.wantedStrings[i] {
				t.Errorf("%s: got %q instead of %q", tc.name, seiMessage.String(), tc.wantedStrings[i])
			}
		}
	}
}
