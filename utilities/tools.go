package utilities

import (
	"crypto/rand"
	"crypto/sha256"
	"strconv"
	"time"
)

func GenerateRandomBytesWithSHA(count int) []byte {
	randomBytes := make([]byte, 0)
	for len(randomBytes) < count {
		r := make([]byte, 32)
		rand.Read(r)
		r = append(r, []byte(strconv.FormatInt(time.Now().UnixNano(), 10))...)
		sha := sha256.Sum256(r)
		randomBytes = append(randomBytes, sha[:]...)
	}
	return randomBytes[:count]
}
