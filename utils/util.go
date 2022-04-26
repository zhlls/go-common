package utils

import (
	"compress/gzip"
	"encoding/hex"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

func NewLongID() string {
	return uuid.New().String()
}
func NewID() string {
	return hex.EncodeToString(uuid.New().NodeID())
}

const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func RemoveDuplicatesAndEmpty(a []string) (ret []string) {
	for i, n := 0, len(a); i < n; i++ {
		if (i > 0 && a[i-1] == a[i]) || len(a[i]) == 0 {
			continue
		}
		ret = append(ret, a[i])
	}
	return
}

func InArray(s string, ss []string) bool {
	for i, n := 0, len(ss); i < n; i++ {
		if ss[i] == s {
			return true
		}
	}
	return false
}

type CoWrapFunc func()

func coWrap(f CoWrapFunc, wg *sync.WaitGroup) {
	f()
	wg.Done()
}

func CoWrap(f CoWrapFunc) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go coWrap(f, &wg)
	wg.Wait()
}

// CompressWithGzip takes an io.Reader as input and pipes
// it through a gzip.Writer returning an io.Reader containing
// the gzipped data.
// An error is returned if passing data to the gzip.Writer fails
func CompressWithGzip(data io.Reader) (io.Reader, error) {
	pipeReader, pipeWriter := io.Pipe()
	gzipWriter := gzip.NewWriter(pipeWriter)

	var err error
	go func() {
		_, err = io.Copy(gzipWriter, data)
		gzipWriter.Close()
		// subsequent reads from the read half of the pipe will
		// return no bytes and the error err, or EOF if err is nil.
		pipeWriter.CloseWithError(err)
	}()

	return pipeReader, err
}
