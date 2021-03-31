package random

import (
	"math/rand"
	"sync"
	"time"
)

var randomChars = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

//Chars random n chars from [0-9a-zA-z]
func Chars(length int) string {
	if length <= 0 {
		return ""
	}
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = randomChars[Between(0, len(randomChars))]
	}
	return string(result)
}

/// gRander need lock() and defer unlock
var randMutex = sync.Mutex{}
var gRander = rand.New(rand.NewSource(time.Now().UnixNano()))

//Range rand [min,max)
func Between(i, j int) int {
	min := i
	max := j
	if min > max {
		min, max = max, min
	}
	if (max - min) <= 0 {
		panic("invalid argument to randrange max cant equal min")
	}
	randMutex.Lock()
	rand := gRander.Intn(max-min) + min
	randMutex.Unlock()
	return rand
}

//Intn return [0,n)
func Intn(n int) int {
	randMutex.Lock()
	rand := gRander.Intn(n)
	randMutex.Unlock()
	return rand
}

//FloatN return [0.0,1.0)
func FloatN() float64 {
	randMutex.Lock()
	rand := gRander.Float64()
	randMutex.Unlock()
	return rand
}

func BetweenF(a float64, b float64) float64 {
	if b < a {
		return a
	}
	x := b - a
	randMutex.Lock()
	n := gRander.Float64()
	randMutex.Unlock()
	return a + x*n
}

func PowerRand(powerlist []int) int {
	var sum = 0
	for i, power := range powerlist {
		if power < 0 {
			powerlist[i] = 0
		}
		sum += powerlist[i]
	}
	rand := Intn(sum)
	sum = 0
	for i, power := range powerlist {
		sum += power
		if rand < sum {
			return i
		}
	}
	return -1
}

func RandomWeight(groups []int, total int) int {
	if total == 0 {
		for _, v := range groups {
			total += v
		}
	}
	n := Intn(total)

	for i, v := range groups {
		n -= v
		if n < 0 {
			return i
		}
	}
	return len(groups) - 1
}

func RandString(n int) string {
	rnd := rand.New(rand.NewSource(int64(n)))
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(rnd.Intn(256))
	}
	return string(b)
}
