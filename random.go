package suk

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
)

const (
	// stringBuffer contains all characters used to randomly generate keys.
	stringBuffer = " !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

// Code taken from https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func init() {
	assertAvailablePRNG()
}

func assertAvailablePRNG() {
	// Assert that a cryptographically secure PRNG is available.
	// Panic otherwise.
	buf := make([]byte, 1)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(fmt.Sprintf("crypto/rand is unavailable: Read() failed with %#v", err))
	}
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func randomID(n uint64) (string, error) {
	ret := make([]byte, n)

	var i uint64
	for i = 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(stringBuffer))))
		if err != nil {
			return "", err
		}

		ret[i] = stringBuffer[num.Int64()]
	}

	return string(ret), nil
}
