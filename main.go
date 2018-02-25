// main
package main

import (
	"encoding/binary"
	"io/ioutil"
	"os"
)

func main() {
	fnameIn := `MoP - Rhythm L.syx`
	fnameOut := `MoP - Rhythm L.bin`
	f, err := os.Open(fnameIn)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	x := make([]byte, 4)
	rawBytes := make([]byte, 0, 2048*4)
	syxBytes := make([]byte, 5)

	f.Read(x[0:1])
	f.Seek(3, os.SEEK_CUR)
	f.Read(x[0:1])
	f.Seek(6, os.SEEK_CUR)
	for k := 0; k < 64; k++ {
		// skip sysex header F0 00 01 74 03 7B 20 00
		f.Seek(8, os.SEEK_CUR)

		for m := 0; m < 32; m++ {
			// Read 5 bytes:
			f.Read(syxBytes)

			// Convert 5x 7-bit values (35 bits) into 4x 8-bit values (32 bits):
			rawUint32 := (uint32(syxBytes[0])&0x7F | (uint32(syxBytes[1])&0x7F)<<7 | (uint32(syxBytes[2])&0x7F)<<14 | (uint32(syxBytes[3])&0x7F)<<21 | (uint32(syxBytes[4])&0x7F)<<28)
			// TODO: are the extra 3 bits being used?

			// Write as 4 bytes:
			binary.LittleEndian.PutUint32(x, rawUint32)
			rawBytes = append(rawBytes, x...)
		}

		// skip sysex checksum:
		f.Seek(2, os.SEEK_CUR)
	}

	ioutil.WriteFile(fnameOut, rawBytes, 0644)
}
