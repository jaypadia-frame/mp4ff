package mp4

import (
	"errors"
	"fmt"
	"io"
)

// TrafBox - Track Fragment Box (traf)
//
// Contained in : Movie Fragment Box (moof)
//
type TrafBox struct {
	Tfhd     *TfhdBox
	Tfdt     *TfdtBox
	Saiz     *SaizBox
	Saio     *SaioBox
	Sbgp     *SbgpBox
	Sgpd     *SgpdBox
	Senc     *SencBox
	Trun     *TrunBox // The first TrunBox
	Truns    []*TrunBox
	Children []Box
}

// DecodeTraf - box-specific decode
func DecodeTraf(hdr boxHeader, startPos uint64, r io.Reader) (Box, error) {
	children, err := DecodeContainerChildren(hdr, startPos+8, startPos+hdr.size, r)
	if err != nil {
		return nil, err
	}
	t := &TrafBox{Children: make([]Box, 0, len(children))}
	for _, child := range children {
		err := t.AddChild(child)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

// ContainsSencBox - is there a senc box in traf and is it parsed
// If not parsed, call ParseReadSenc to parse it
func (t *TrafBox) ContainsSencBox() (ok, parsed bool) {
	if t.Senc != nil {
		return true, !t.Senc.readButNotParsed
	}
	return false, false
}

func (t *TrafBox) ParseReadSenc(defaultIVSize byte, moofStartPos uint64) error {
	if t.Senc == nil {
		return fmt.Errorf("no senc box")
	}
	if t.Saio != nil {
		// saio should be present, but we try without it, if it doesn't exist
		posFromSaio := t.Saio.Offset[0] + int64(moofStartPos)
		if uint64(posFromSaio) != t.Senc.StartPos+16 {
			return fmt.Errorf("offset from saio (%d) and moof differs from senc data start %d", posFromSaio, t.Senc.StartPos+16)
		}
	}
	perSampleIVSize := defaultIVSize
	if t.Sbgp != nil && t.Sgpd != nil {
		sbgp, sgpd := t.Sbgp, t.Sgpd
		if sbgp.GroupingType != "seig" {
			return fmt.Errorf("sbgp grouping type %s not supported", sbgp.GroupingType)
		}
		nrSbgpEntries := len(sbgp.SampleCounts)
		if nrSbgpEntries != 1 {
			return fmt.Errorf("sbgp entries = %d, only 1 supported for now", nrSbgpEntries)
		}
		sgpdEntryNr := sbgp.GroupDescriptionIndices[0]
		if sgpdEntryNr != sbgpInsideOffset+1 {
			return fmt.Errorf("sgpd entry number must be first inside = 65536 + 1")
		}
		if sgpd.GroupingType != "seig" {
			return fmt.Errorf("sgpd grouping type %s not supported", sgpd.GroupingType)
		}

		sgpdEntry := sgpd.SampleGroupEntries[sgpdEntryNr-sbgpInsideOffset-1]
		if sgpdEntry.Type() != "seig" {
			return fmt.Errorf("expected sgpd entry type seig but found %q", sgpdEntry.Type())
		}
		seigEntry := sgpdEntry.(*SeigSampleGroupEntry)
		perSampleIVSize = seigEntry.PerSampleIVSize
	}
	err := t.Senc.ParseReadBox(perSampleIVSize, t.Saiz)
	if err != nil {
		return err
	}
	return nil
}

// AddChild - add child box
func (t *TrafBox) AddChild(b Box) error {
	switch b.Type() {
	case "tfhd":
		t.Tfhd = b.(*TfhdBox)
	case "tfdt":
		t.Tfdt = b.(*TfdtBox)
	case "saiz":
		t.Saiz = b.(*SaizBox)
	case "saio":
		t.Saio = b.(*SaioBox)
	case "sbgp":
		t.Sbgp = b.(*SbgpBox)
	case "sgpd":
		t.Sgpd = b.(*SgpdBox)
	case "senc":
		t.Senc = b.(*SencBox)
	case "trun":
		if t.Trun == nil {
			t.Trun = b.(*TrunBox)
		}
		t.Truns = append(t.Truns, b.(*TrunBox))
	default:
	}
	t.Children = append(t.Children, b)
	return nil
}

// Type - return box type
func (t *TrafBox) Type() string {
	return "traf"
}

// Size - return calculated size
func (t *TrafBox) Size() uint64 {
	return containerSize(t.Children)
}

// GetChildren - list of child boxes
func (t *TrafBox) GetChildren() []Box {
	return t.Children
}

// Encode - write box to w
func (t *TrafBox) Encode(w io.Writer) error {
	return EncodeContainer(t, w)
}

// Info - write box-specific information
func (t *TrafBox) Info(w io.Writer, specificBoxLevels, indent, indentStep string) error {
	return ContainerInfo(t, w, specificBoxLevels, indent, indentStep)
}

// OptimizeTfhdTrun - optimize trun by default values in tfhd box
// Only look at first trun, even if there is more than one
// Don't optimize again, if already done so that no data is present
func (t *TrafBox) OptimizeTfhdTrun() error {
	tfhd := t.Tfhd
	trun := t.Trun
	if len(trun.Samples) == 0 {
		return errors.New("No samples in trun")
	}
	if len(trun.Samples) == 1 {
		return nil // No need to optimize
	}

	if trun.HasSampleDuration() {
		hasCommonDur := true
		commonDur := trun.Samples[0].Dur
		for _, s := range trun.Samples {
			if s.Dur != commonDur {
				hasCommonDur = false
				break
			}
		}
		if hasCommonDur {
			// Set defaultSampleDuration in tfhd and remove from trun
			tfhd.Flags = tfhd.Flags | defaultSampleDurationPresent
			tfhd.DefaultSampleDuration = commonDur
			trun.flags = trun.flags & ^sampleDurationPresentFlag
		}
	}

	if trun.HasSampleSize() {
		hasCommonSize := true
		commonSize := trun.Samples[0].Size
		for _, s := range trun.Samples {
			if s.Size != commonSize {
				hasCommonSize = false
				break
			}
		}
		if hasCommonSize {
			// Set defaultSampleSize in tfhd and remove from trun
			tfhd.Flags = tfhd.Flags | defaultSampleSizePresent
			tfhd.DefaultSampleSize = commonSize
			trun.flags = trun.flags & ^sampleSizePresentFlag
		}
	}

	if trun.HasSampleFlags() {
		firstSampleFlags := trun.Samples[0].Flags
		hasCommonFlags := true
		commonSampleFlags := trun.Samples[1].Flags
		for i, s := range trun.Samples {
			if i >= 1 {
				if s.Flags != commonSampleFlags {
					hasCommonFlags = false
					break
				}
			}
		}
		if hasCommonFlags {
			if firstSampleFlags != commonSampleFlags {
				trun.SetFirstSampleFlags(firstSampleFlags)
			}
			tfhd.Flags = tfhd.Flags | defaultSampleFlagsPresent
			tfhd.DefaultSampleFlags = commonSampleFlags
			trun.flags = trun.flags & ^sampleFlagsPresentFlag
		}
	}

	if trun.HasSampleCompositionTimeOffset() {
		allZeroCTO := true
		for _, s := range trun.Samples {
			if s.CompositionTimeOffset != 0 {
				allZeroCTO = false
				break
			}
		}
		if allZeroCTO {
			trun.flags = trun.flags & ^sampleCompositionTimeOffsetPresentFlag
		}
	}
	return nil
}

//RemoveEncryptionBoxes - remove encryption boxes and return number of bytes removed
func (t *TrafBox) RemoveEncryptionBoxes() uint64 {
	remainingChildren := make([]Box, 0, len(t.Children))
	var nrBytesRemoved uint64 = 0
	for _, ch := range t.Children {
		switch ch.Type() {
		case "saiz":
			nrBytesRemoved += ch.Size()
			t.Saiz = nil
		case "saio":
			nrBytesRemoved += ch.Size()
			t.Saio = nil
		case "senc":
			nrBytesRemoved += ch.Size()
			t.Senc = nil
		default:
			remainingChildren = append(remainingChildren, ch)
		}
	}
	t.Children = remainingChildren
	return nrBytesRemoved
}
