package main

import (
	"fmt"
	"hash"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"crypto/rand"

	"crypto/sha256"

	"golang.org/x/crypto/sha3"
)

var hashvalue []byte
var blocksize int64 = 1 << 17
var buf = make([]byte, blocksize)

func init() {
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
}

func must(hash hash.Hash, err error) hash.Hash {
	if err != nil {
		panic(err)
	}
	return hash
}

func wrap(hasher hash.Hash) func(*testing.B) {
	return func(b *testing.B) {
		b.SetBytes(blocksize)
		for i := 0; i < b.N; i++ {
			hasher.Write(buf)
			hashvalue = hasher.Sum(nil)
			hasher.Reset()
		}
	}
}

func main() {
	if len(os.Args) == 1 {
		cmd := exec.Command(os.Args[0], "-test.bench=.*", "-test.benchmem=true", "-test.benchtime=2s")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return
	}

	fmt.Printf("Build: %s %s-%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	hashers := []struct {
		name string
		hash hash.Hash
	}{
		{"SHA256", sha256.New()},
		{"SHAKE256", shakeAdapter{sha3.NewShake256(), 64}},
	}

	benchmarks := []testing.InternalBenchmark{}
	for _, h := range hashers {
		benchmarks = append(benchmarks, testing.InternalBenchmark{
			Name: h.name,
			F:    wrap(h.hash),
		})
	}
	testing.Main(func(pat, str string) (bool, error) {
		return true, nil
	}, nil, benchmarks, nil)
}

type shakeAdapter struct {
	sha3.ShakeHash
	length int
}

func (a shakeAdapter) Sum(_ []byte) []byte {
	bs := make([]byte, a.length)
	a.Read(bs)
	return bs
}

func (a shakeAdapter) Size() int {
	return a.length
}
func (a shakeAdapter) BlockSize() int {
	return 1
}
