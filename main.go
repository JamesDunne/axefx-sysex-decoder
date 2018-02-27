// main
package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"os"
)

func main() {
	fname := `MoP - Rhythm L`
	extIn := `.syx`
	extBin := `.bin`
	extRaw := `.raw`

	f, err := os.Open(fname + extIn)
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

	ioutil.WriteFile(fname+extBin, rawBytes, 0644)

	// Bytes 0x00 to 0x1E bytes are the name of the cab
	// Byte 0x1F being 0xA0 probably indicates UltraRes IR?

	// UltraRes mode contains exactly 170 samples
	// Total binary size is 0x2000, subtract 0x20 for name+indicator leaves 0x1FE0 bytes for samples
	// 0x1FE0 (8,160) divided by 170 is exactly 48 bytes per sample
	// Perhaps some 4- or 8-byte offset into those 48 bytes contains an IEEE float32 or float64
	// that will yield our IR sample for WAV export?

	const floatSize = 4
nextoffset:
	for offs := 0; offs < 48-floatSize; offs += 1 {
		rawName := fname + fmt.Sprintf("%02d", offs) + extRaw

		fmt.Print("::: ")
		fmt.Println(offs)
		samples := make([]byte, 0, 170*floatSize)
		x := 0x20 + offs
		for i := 0; i < 170; i++ {
			sample := rawBytes[x : x+floatSize]
			f32 := math.Float32frombits(binary.LittleEndian.Uint32(sample))
			//f64 := math.Float64frombits(binary.BigEndian.Uint64(sample))
			if math.IsNaN(float64(f32)) {
				os.Remove(rawName)
				continue nextoffset
			}

			samples = append(samples, sample...)
			fmt.Printf("%v\n", f32)
			x += 48
		}

		ioutil.WriteFile(rawName, samples, 0644)
	}
}
