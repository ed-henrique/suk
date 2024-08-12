package suk

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
)

const (
	// stringBuffer contains all characters used to randomly generate keys.
	stringBuffer = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890.-_"
)

// Most of this code was taken from
// https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func init() {
	assertAvailablePRNG()
}

// assertAvailablePRNG asserts that there is an available pseudorandom number
// generator, which is used to create random IDs via randomID.
func assertAvailablePRNG() {
	buf := make([]byte, 1)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(
			fmt.Sprintf("crypto/rand is unavailable: Read() failed with %#v", err),
		)
	}
}

// defaultRandomKeyGenerator returns a securely generated random string. It will
// return an error if the system's secure random number generator fails to
// function correctly, in which case the caller should not continue.
func defaultRandomKeyGenerator(n uint64) (string, error) {
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
