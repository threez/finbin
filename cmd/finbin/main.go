package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/inhies/go-bytesize"
)

var (
	file, pattern, size, dir string
)

func main() {
	flag.StringVar(&file, "file", "", "path to the file to open")
	flag.StringVar(&pattern, "pattern", "testdata", "regex pattern to find")
	flag.StringVar(&size, "size", "10B", "size to capture before and after")
	flag.StringVar(&dir, "dir", "found", "path to store found records")
	flag.Parse()

	bs, err := bytesize.Parse(size)
	if err != nil {
		log.Fatalf("invalid size %q: %v", size, err)
	}

	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		log.Fatalf("unable to create output dir %q: %v", dir, err)
	}

	s, err := os.Stat(file)
	if err != nil {
		log.Fatalf("unable to stat file %q: %v", file, err)
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

		start := offset + int64(positions[0]) - int64(bs)
		if start < 0 {
			start = 0
		}
		end := offset + int64(positions[1]) + int64(bs)
		if end > s.Size() {
			end = s.Size()
		}
		total := end - start

		outFile := fmt.Sprintf("file-%d", positions[0])

		fmt.Printf("Found %s (%d to %d size %v)\n",
			outFile, start, end, bytesize.ByteSize(total))

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

		_, err = io.CopyN(ofh, fh, total)
		if err != nil {
			log.Fatalf("unable to copy data to file %q: %v", outFilePath, err)
		}

		ofh.Close()

		_, err = fh.Seek(offset+int64(positions[1]), io.SeekStart)
		if err != nil {
			log.Fatalf("unable to seek back to %d in file %q: %v", start, file, err)
		}
	}
}
