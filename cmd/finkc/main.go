package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/inhies/go-bytesize"
)

var (
	file, dir, skip string
	maxFile         = int64(50 * 1024 * 1024) // 50 MB
)

const pattern = "kych\x00"

func main() {
	flag.StringVar(&file, "file", "", "path to the file to open")
	flag.StringVar(&dir, "dir", "keychains", "path to store found records")
	flag.StringVar(&skip, "skip", "0", "skip over bytes from the beginning of the file")
	flag.Parse()

	sb, err := strconv.ParseInt(skip, 10, 64)
	if err != nil {
		log.Fatalf("invalid skip %q: %v", skip, err)
	}

	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		log.Fatalf("unable to create output dir %q: %v", dir, err)
	}

	s, err := os.Stat(file)
	if err != nil {
		log.Fatalf("unable to stat file %q: %v", file, err)
	}
	if s.Size() == 0 {
		log.Println("Block device mode.")
	}

	fh, err := os.Open(file)
	if err != nil {
		log.Fatalf("unable to open file %q: %v", file, err)
	}
	defer fh.Close()

	regex, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatalf("unable to compile pattern %q: %v", pattern, err)
	}

	if sb > 0 {
		_, err := fh.Seek(sb, io.SeekStart)
		if err != nil {
			log.Fatalf("unable to skip in %q: %v", pattern, err)
		}
	}

	for {
		offset, err := fh.Seek(0, io.SeekCurrent)
		if err != nil {
			log.Fatalf("unable to get current position in file %q: %v", file, err)
		}

		r := bufio.NewReader(fh)
		positions := regex.FindReaderIndex(r)

		if positions == nil {
			break
		}

		start := offset + int64(positions[0])
		if start < 0 {
			start = 0
		}

		// calculate end by reading file length
		// from the header of the keychain file
		_, err = fh.Seek(start+20, io.SeekStart)
		if err != nil {
			log.Fatalf("unable to seek to header in file %q: %v", file, err)
		}

		var sizeHeader [4]byte
		_, err = fh.Read(sizeHeader[:])
		if err != nil {
			log.Fatalf("unable to read header in file %q: %v", file, err)
		}

		size := binary.BigEndian.Uint32(sizeHeader[:]) + 24
		end := offset + int64(size)
		if end > s.Size() && s.Size() != 0 {
			end = s.Size()
		}
		total := end - start
		if total == 0 || total > maxFile || total < 0 {
			continue
		}

		outFile := fmt.Sprintf("file-%d.keychain", offset+int64(positions[0]))

		fmt.Printf("Found %s - copy section from %d (%v) to %d (%v) size %v\n",
			outFile,
			start, bytesize.ByteSize(start),
			end, bytesize.ByteSize(end),
			bytesize.ByteSize(total),
		)

		outFilePath := filepath.Join(dir, outFile)
		ofh, err := os.Create(outFilePath)
		if err != nil {
			log.Fatalf("unable to create output file %q: %v", outFilePath, err)
		}

		// go to start of section to copy
		_, err = fh.Seek(start, io.SeekStart)
		if err != nil {
			log.Fatalf("unable to seek to %d in file %q: %v", start, file, err)
		}

		n, err := io.CopyN(ofh, fh, total)
		if s.Size() == 0 { // blockdevice mode
			if err != nil {
				if n > 0 && n < total {
					// ignore
				} else {
					log.Fatalf("unable to copy data to file %q: %v", outFilePath, err)
				}
			}
		} else {
			if err != nil {
				log.Fatalf("unable to copy data to file %q: %v", outFilePath, err)
			}
		}

		ofh.Close()

		_, err = fh.Seek(offset+int64(positions[1]), io.SeekStart)
		if err != nil {
			log.Fatalf("unable to seek back to %d in file %q: %v", start, file, err)
		}

		offset += int64(size)
	}
}
