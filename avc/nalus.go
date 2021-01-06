package avc

import (
	"encoding/binary"
	"fmt"
)

// GetNalusFromSample - get nalus by following 4 byte length fields
func GetNalusFromSample(sample []byte) ([][]byte, error) {
	nalList := make([][]byte, 0)
	length := len(sample)
	if length < 4 {
		return nalList, fmt.Errorf("Less than 4 bytes, No NALUs")
	}
	var pos uint32 = 0
	for pos < uint32(length-4) {
		nalLength := binary.BigEndian.Uint32(sample[pos : pos+4])
		pos += 4
		if int(pos+nalLength) > len(sample) {
			return nil, fmt.Errorf("NAL length fields are bad. Not video?")
		}
		nalList = append(nalList, sample[pos:pos+nalLength])
		pos += nalLength
	}
	return nalList, nil
}
