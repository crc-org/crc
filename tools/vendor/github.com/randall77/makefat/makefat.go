package main

// makefat <output file> <input file 1> <input file 2> ...

import (
	"debug/macho"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	MagicFat64 = macho.MagicFat + 1 // TODO: add to stdlib (...when it works)

	// Alignment wanted for each sub-file.
	// amd64 needs 12 bits, arm64 needs 14. We choose the max of all requirements here.
	alignBits = 14
	align     = 1 << alignBits
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <output file> <input file 1> <input file 2> ...\n", os.Args[0])
		os.Exit(2)
	}

	// Read input files.
	type input struct {
		data   []byte
		cpu    uint32
		subcpu uint32
		offset int64
	}
	var inputs []input
	offset := int64(align)
	for _, i := range os.Args[2:] {
		data, err := ioutil.ReadFile(i)
		if err != nil {
			panic(err)
		}
		if len(data) < 12 {
			panic(fmt.Sprintf("file %s too small", i))
		}
		// All currently supported mac archs (386,amd64,arm,arm64) are little endian.
		magic := binary.LittleEndian.Uint32(data[0:4])
		if magic != macho.Magic32 && magic != macho.Magic64 {
			panic(fmt.Sprintf("input %s is not a macho file, magic=%x", i, magic))
		}
		cpu := binary.LittleEndian.Uint32(data[4:8])
		subcpu := binary.LittleEndian.Uint32(data[8:12])
		inputs = append(inputs, input{data: data, cpu: cpu, subcpu: subcpu, offset: offset})
		offset += int64(len(data))
		offset = (offset + align - 1) / align * align
	}

	// Decide on whether we're doing fat32 or fat64.
	sixtyfour := false
	if inputs[len(inputs)-1].offset >= 1<<32 || len(inputs[len(inputs)-1].data) >= 1<<32 {
		sixtyfour = true
		// fat64 doesn't seem to work:
		//   - the resulting binary won't run.
		//   - the resulting binary is parseable by lipo, but reports that the contained files are "hidden".
		//   - the native OSX lipo can't make a fat64.
		panic("files too large to fit into a fat binary")
	}

	// Make output file.
	out, err := os.Create(os.Args[1])
	if err != nil {
		panic(err)
	}
	err = out.Chmod(0755)
	if err != nil {
		panic(err)
	}

	// Build a fat_header.
	var hdr []uint32
	if sixtyfour {
		hdr = append(hdr, MagicFat64)
	} else {
		hdr = append(hdr, macho.MagicFat)
	}
	hdr = append(hdr, uint32(len(inputs)))

	// Build a fat_arch for each input file.
	for _, i := range inputs {
		hdr = append(hdr, i.cpu)
		hdr = append(hdr, i.subcpu)
		if sixtyfour {
			hdr = append(hdr, uint32(i.offset>>32)) // big endian
		}
		hdr = append(hdr, uint32(i.offset))
		if sixtyfour {
			hdr = append(hdr, uint32(len(i.data)>>32)) // big endian
		}
		hdr = append(hdr, uint32(len(i.data)))
		hdr = append(hdr, alignBits)
		if sixtyfour {
			hdr = append(hdr, 0) // reserved
		}
	}

	// Write header.
	// Note that the fat binary header is big-endian, regardless of the
	// endianness of the contained files.
	err = binary.Write(out, binary.BigEndian, hdr)
	if err != nil {
		panic(err)
	}
	offset = int64(4 * len(hdr))

	// Write each contained file.
	for _, i := range inputs {
		if offset < i.offset {
			_, err = out.Write(make([]byte, i.offset-offset))
			if err != nil {
				panic(err)
			}
			offset = i.offset
		}
		_, err := out.Write(i.data)
		if err != nil {
			panic(err)
		}
		offset += int64(len(i.data))
	}
	err = out.Close()
	if err != nil {
		panic(err)
	}
}
