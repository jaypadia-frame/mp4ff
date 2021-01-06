// mp4ff-nallister lists NAL units and slice types of AVC tracks of an mp4 (ISOBMFF) file.
// Takes first video track in a progressive file and the first track in a fragmented file.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/edgeware/mp4ff/avc"
	"github.com/edgeware/mp4ff/mp4"
)

var usg = `Usage of mp4ff-nallister:

mp4ff-nallister lists NAL units and slice types of AVC tracks of an mp4 (ISOBMFF) file.
Takes first video track in a progressive file and the first track in
a fragmented file.
`

var Usage = func() {
	parts := strings.Split(os.Args[0], "/")
	name := parts[len(parts)-1]
	fmt.Fprintln(os.Stderr, usg)
	fmt.Fprintf(os.Stderr, "%s [-m <max>] <mp4File>\n", name)
	flag.PrintDefaults()
}

func main() {
	maxNrSamples := flag.Int("m", -1, "Max nr of samples to parse")

	flag.Parse()

	var inFilePath = flag.Arg(0)
	if inFilePath == "" {
		Usage()
		os.Exit(1)
	}

	ifd, err := os.Open(inFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer ifd.Close()
	parsedMp4, err := mp4.DecodeFile(ifd)
	if err != nil {
		log.Fatal(err)
	}

	// Need to handle progressive files as well as fragmented files

	if !parsedMp4.IsFragmented() {
		err = parseProgressiveMp4(parsedMp4, *maxNrSamples)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		return
	}
	err = parseFragmentedMp4(parsedMp4, *maxNrSamples)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func parseProgressiveMp4(f *mp4.File, maxNrSamples int) error {
	var videoTrak *mp4.TrakBox
	for _, inTrak := range f.Moov.Traks {
		hdlrType := inTrak.Mdia.Hdlr.HandlerType
		if hdlrType != "vide" {
			continue
		}
		videoTrak = inTrak
		break
	}
	if videoTrak == nil {
		return fmt.Errorf("New video track found")
	}

	stbl := videoTrak.Mdia.Minf.Stbl
	nrSamples := stbl.Stsz.SampleNumber
	mdat := f.Mdat
	mdatPayloadStart := mdat.PayloadAbsoluteOffset()
	for sampleNr := 1; sampleNr <= int(nrSamples); sampleNr++ {
		chunkNr, sampleNrAtChunkStart, err := stbl.Stsc.ChunkNrFromSampleNr(sampleNr)
		if err != nil {
			return err
		}
		offset := int64(stbl.Stco.ChunkOffset[chunkNr-1])
		for sNr := sampleNrAtChunkStart; sNr < sampleNr; sNr++ {
			offset += int64(stbl.Stsz.SampleSize[sNr-1])
		}
		size := stbl.Stsz.GetSampleSize(sampleNr)
		decTime, _ := stbl.Stts.GetDecodeTime(uint32(sampleNr))
		var cto int32 = 0
		if stbl.Ctts != nil {
			cto = stbl.Ctts.GetCompositionTimeOffset(uint32(sampleNr))
		}
		// Next find sample bytes as slice in mdat
		offsetInMdatData := uint64(offset) - mdatPayloadStart
		sample := mdat.Data[offsetInMdatData : offsetInMdatData+uint64(size)]
		err = printAVCNalus(sample, sampleNr, decTime+uint64(cto))
		if err != nil {
			return err
		}
		if sampleNr == maxNrSamples {
			break
		}
	}
	return nil
}

func parseFragmentedMp4(f *mp4.File, maxNrSamples int) error {
	iSamples := make([]*mp4.FullSample, 0)
	for _, iSeg := range f.Segments {
		for _, iFrag := range iSeg.Fragments {
			fSamples, err := iFrag.GetFullSamples(nil)
			if err != nil {
				return err
			}
			iSamples = append(iSamples, fSamples...)
		}
	}
	for i, s := range iSamples {
		err := printAVCNalus(s.Data, i+1, s.PresentationTime())
		if err != nil {
			return err
		}
		if i+1 == maxNrSamples {
			break
		}
	}
	return nil
}

func printAVCNalus(sample []byte, nr int, pts uint64) error {
	nalus, err := avc.GetNalusFromSample(sample)
	if err != nil {
		return err
	}
	msg := ""
	for i, nalu := range nalus {
		if i > 0 {
			msg += ","
		}
		nalType := avc.GetNalType(nalu[0])
		imgType := ""
		switch nalType {
		case avc.NALU_NON_IDR, avc.NALU_IDR:
			sliceType, err := avc.GetSliceTypeFromNAL(nalu)
			if err == nil {
				imgType = fmt.Sprintf("[%s] ", sliceType)
			}
		}
		msg += fmt.Sprintf(" %s %s(%dB)", nalType, imgType, len(nalu))
	}
	fmt.Printf("Sample %d, pts=%d (%dB):%s\n", nr, pts, len(sample), msg)
	return nil
}
