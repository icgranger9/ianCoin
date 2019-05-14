package tests

import (
	"../p1"
	"../p2"
	"../p3"
	crand "crypto/rand"
	"crypto/rsa"
	"testing"
)

var SELF = ""

func TestGenerateNextMpt(t *testing.T){
	//adding keys
	priv, _ := rsa.GenerateKey(crand.Reader, 2048)
	pub := &priv.PublicKey

	var transactions p1.MerklePatriciaTrie
	var balances p1.MerklePatriciaTrie
	blk := p2.GenBlock(0, "", "", transactions, balances)
	newTX, newBal, err := p3.GenerateNextMpt(blk)


	if err != nil {
		t.Log(err)
		t.Fail()
	} else {
		t.Log(newTX.Order_nodes())
		t.Log(newBal.Order_nodes())
	}
}
