package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var (
	consumersCount = 8
	sleepTimeout   = 10 * time.Millisecond
	endTimeout     = 100 * time.Second
)

type predicate func([]byte) bool

func preparePreimage(number uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, number)
	hash := sha256.New()
	hash.Write(bytes)
	return hash.Sum(nil)
}

func findHash(match predicate, numbers <-chan uint64, wait *sync.WaitGroup) {

	defer wait.Done()
	for {
		select {
		case number, ok := <-numbers:
			if ok {
				preimage := preparePreimage(number)
				if match(preimage) {
					fmt.Printf("%d: %s\n", number, hex.EncodeToString(preimage))
				}
			} else {
				return
			}
		}
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func produceData(numbers chan<- uint64, timer <-chan time.Time, wait *sync.WaitGroup) {
	defer wait.Done()
	defer close(numbers)
	for {
		select {
		case <-timer:
			return
		default:
			numbers <- rand.Uint64()
			time.Sleep(sleepTimeout)
		}
	}
}

func main() {

	fmt.Println("starting...")
	defer fmt.Println("ending...")

	var (
		numbers = make(chan uint64, 100)
		matcher = func(value []byte) bool {
			return bytes.Equal(value[0:3], []byte{0, 0, 0})
		}
		timer     = time.After(endTimeout)
		waitGroup sync.WaitGroup
	)
	waitGroup.Add(1)
	go produceData(numbers, timer, &waitGroup)

	waitGroup.Add(consumersCount)
	for i := 0; i < consumersCount; i++ {
		go findHash(matcher, numbers, &waitGroup)
	}

	waitGroup.Wait()
}
