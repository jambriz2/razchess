package razchess

import (
	"crypto/rand"
	"io"
)

const (
	charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	csLen   = byte(len(charset))
)

func GenerateID(length int) string {
	if length == 0 {
		return ""
	}
	output := make([]byte, 0, length)
	batchSize := length + length/4
	buf := make([]byte, batchSize)
	for {
		if _, err := io.ReadFull(rand.Reader, buf); err != nil {
			panic(err)
		}
		for _, b := range buf {
			if b < (csLen * 4) {
				output = append(output, charset[b%csLen])
				if len(output) == length {
					return string(output)
				}
			}
		}
	}
}
