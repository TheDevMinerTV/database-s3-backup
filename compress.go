package main

import (
	"fmt"
	"io"
	"os"

	"github.com/klauspost/compress/zstd"
)

func CompressFile(inFile string) (string, error) {
	in, err := os.Open(inFile)
	if err != nil {
		return "", err
	}
	defer in.Close()

	outFile := fmt.Sprintf("%s.zst", inFile)
	out, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}
	defer out.Close()

	writer, err := zstd.NewWriter(out, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return "", err
	}
	defer writer.Close()

	if _, err = io.Copy(writer, in); err != nil {
		return "", err
	}
	return outFile, nil
}
