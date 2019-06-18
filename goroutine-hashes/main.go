package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

func logging(s []byte) {
	ok, hash := encode(s)
	if ok {
		fmt.Println(string(s) + ": " + hex.EncodeToString(hash))
	}
}

func encode(data []byte) (bool, []byte) {
	hash := sha256.New()
	hash.Write(data)
	value := hash.Sum(nil)
	return bytes.Equal(value[0:3], []byte{0, 0, 0}), value
}

func findHash(ch <-chan uint64, timer <-chan time.Time, done chan<- bool) {
	for {
		select {
		case <-timer:
			done <- true
			return
		default:
			rnd := <-ch
			bs := make([]byte, 8)
			binary.LittleEndian.PutUint64(bs, rnd)
			logging(bs)
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("starting...")
	ch := make(chan uint64, 100)
	timer := time.After(10 * time.Second)
	done := make(chan bool)
	for i := 0; i < 8; i++ {
		go findHash(ch, timer, done)
	}

	for finished := false; finished != true; {
		select {
		case <-done:
			finished = true
		default:
			ch <- rand.Uint64()
		}
	}
	close(ch)
	close(done)
}
