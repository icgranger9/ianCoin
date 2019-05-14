package data

import (
	"math/rand"
)

func GenerateNonce() string {

	hexChars := [16]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}
	nonce := ""
	nonceLen := 16

	for x := 0; x < nonceLen; x++ {
		charInd := rand.Intn(len(hexChars) - 1)
		nonce += hexChars[charInd]
	}

	return nonce
}
