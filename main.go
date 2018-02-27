// main
package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	fname := `MoP - Rhythm L`
	extIn := `.syx`
	extBin := `.bin`
	//extRaw := `.raw`

	f, err := os.Open(fname + extIn)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	rawBytes := make([]byte, 0, 10240)
	ob := make([]byte, 1)
	msg := make([]byte, 0, 4+0xA0+1)

	hdr := make([]byte, 4)

	msgs := 0
	// Read SysEx blocks:
reading:
	for {
		// read SysEx header F0 00 01 74
		f.Read(hdr)
		if hdr[0] != 0xF0 || hdr[1] != 0x00 || hdr[2] != 0x01 || hdr[3] != 0x74 {
			break
		}

		// Read until 0xF7:
		msg = msg[0:0]
		ob[0] = byte(0xF0)
		for i := 0; ob[0] != 0xF7; i++ {
			n, err := f.Read(ob)
			if n <= 0 || err != nil {
				break reading
			}
			if ob[0] == 0xF7 {
				break
			}

			msg = append(msg, ob[0])
		}

		// 03 7B 20 00 is a data packet for IR:
		if msg[0] == 0x03 && msg[1] == 0x7B {
			msgs++
			data7 := msg[4 : len(msg)-1]
			data := make([]byte, len(data7)*4/5)
			j := 0
			for i := 0; i < len(data7); i += 5 {
				if data7[i+4]&^0x0F != 0 {
					panic("Unexpected extra bits in final byte of 5-byte string")
				}
				b := sysexToRaw(data7[i : i+5])
				binary.LittleEndian.PutUint32(data[j:j+4], b)
				j += 4
			}
			rawBytes = append(rawBytes, data...)
		}

		// Final message:
		if msg[0] == 0x03 && msg[1] == 0x7C {
			b := sysexToRaw(msg[2 : 2+5])
			data := make([]byte, 4)
			binary.LittleEndian.PutUint32(data[0:4], b)
			fmt.Println(data)
		}
	}

	ioutil.WriteFile(fname+extBin, rawBytes, 0644)

	// fmt.Printf("%v\n", msgs)
	// fmt.Printf("%v\n", len(rawBytes))

	// Bytes 0x00 to 0x1E bytes are the name of the cab
	// Byte 0x1F being 0xA0 probably indicates UltraRes IR?

	// UltraRes mode contains exactly 170ms of sample data in 0x1FE0 bytes
	// means 1,000 ms / 48,000 Hz = 0.02083 ms / 1 sample * 8,160 samples = 170 ms
}

func sysexToRaw(data7 []byte) uint32 {
	return uint32(data7[0]&0x7F) | uint32(data7[1]&0x7F)<<7 | uint32(data7[2]&0x7F)<<14 | uint32(data7[3]&0x7F)<<21 | uint32(data7[4]&0x0F)<<28
}
