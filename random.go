package utilz

import (
	crand "crypto/rand"
	"math/big"
	"math/rand"
	"time"
)

/// <RANDOM STRING GENERATOR>
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

// RandomString returns a randomly-generated string of length = n
func RandomString(n int) string {
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

/// </RANDOM STRING GENERATOR>

// RandomIntRange returns a random integer in the given range;
// Randomness is from "math/rand" package.
func RandomIntRange(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

// CryptoRandomIntRange returns a random integer in the given range.
// Randomness is from "crypto/rand" package.
func CryptoRandomIntRange(min, max int) int {
	jj, err := crand.Int(crand.Reader, big.NewInt(int64(max-min)))
	if err != nil {
		panic(err)
	}
	return int(jj.Int64()) + min
}

// RandomIntRange returns a random integer in the given range
func DeterministicRandomIntRange(min, max int) int {
	return rand.Intn(max-min) + min
}
