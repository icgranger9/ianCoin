package data

import (
	"../../p1"
	"fmt"
	"math/rand"
)

func GenerateMPT() p1.MerklePatriciaTrie {
	mpt := p1.MerklePatriciaTrie{}
	mpt.Initial()

	dict := []string{"this", "is", "one", "simple", "dictionary", "of", "words", "that", "can", "be", "added", "into", "our", "mtp", "I", "will", "now", "add", "many", "other", "options", "just", "to", "make", "it", "more", "random", "sound", "good?", "hello", "world", "golang", "USF", "computer", "science", "san", "francisco", "california", "america", "golden", "state", "warriors", "hopefully", "thats", "enough"}
	mptSize := rand.Intn(len(dict)-1) + 1

	for wordsAdded := 0; wordsAdded < mptSize; wordsAdded++ {

		randNum := rand.Intn(len(dict) - 1)
		word := dict[randNum]

		//fmt.Println("\t\t\tInserting into mpt: ", word, " --> ", fmt.Sprint("word num: ", wordsAdded))
		mpt.Insert(word, fmt.Sprint("word num: ", wordsAdded))
	}
	return mpt
}

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
