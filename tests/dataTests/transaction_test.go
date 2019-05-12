package dataTests

import (
	"../../p3/data"
	crand "crypto/rand"
	"crypto/rsa"
	"testing"
	"time"
)

func TestNewTransaction(t *testing.T){
	var emptyT data.Transaction
	newT := data.NewTransaction("src_address", "dts_address", 1.00, .05, time.Now().UnixNano())

	if newT == emptyT {
		t.Fail()
	} else {
		t.Log(newT.ShowTransaction())
	}
}

func TestValidTransaction(t *testing.T){
	//adding keys
	priv, _ := rsa.GenerateKey(crand.Reader, 2048)
	pub := &priv.PublicKey

	newT := data.NewTransaction(data.KeyToString(pub), "dts_address", 1.00, .05, time.Now().UnixNano())

	err := newT.SignTransaction(priv)


	if !newT.VerifyTransaction() || err != nil{
		t.Fail()
	} else {
		t.Log(newT.ShowTransaction())
	}
}

func TestInvalidTransaction(t *testing.T){
	//adding keys
	priv, _ := rsa.GenerateKey(crand.Reader, 2048)
	pub := &priv.PublicKey

	newT := data.NewTransaction(data.KeyToString(pub), "dts_address", 1.00, .06, time.Now().UnixNano())

	err := newT.SignTransaction(priv)


	if newT.VerifyTransaction() || err != nil{
		t.Log("Said valid when not true")
		t.Fail()
	} else {
		t.Log(newT.ShowTransaction())
	}
}

func TestEncodeTransaction(t *testing.T){
	//adding keys
	priv, _ := rsa.GenerateKey(crand.Reader, 2048)
	pub := &priv.PublicKey

	newT := data.NewTransaction(data.KeyToString(pub), "dts_address", 1.00, .05, time.Now().UnixNano())

	err := newT.SignTransaction(priv)

	json, err := newT.TransactionToJson()

	if err != nil {
		t.Log(err)
		t.Fail()
	} else {
		t.Log(json)
	}
}

func TestDecodeTransaction(t *testing.T){
	json := ` {"source":"src_address","destination":"dts_address","amount":1,"fee":0.06,"timestamp":"1557617325250947000"} `

	tAction, err  := data.DecodeTransactionFromJson(json)

	if err != nil {
		t.Log(err)
		t.Fail()
	} else {
		t.Log(tAction.ShowTransaction())
	}
}

func TestSignatureTransaction(t *testing.T){
	//adding keys
	priv, _ := rsa.GenerateKey(crand.Reader, 2048)
	pub := &priv.PublicKey

	newT := data.NewTransaction(data.KeyToString(pub), "dts_address", 1.00, .05, time.Now().UnixNano())

	err := newT.SignTransaction(priv)


	if err != nil {
		t.Log(err)
		t.Fail()
	} else {
		t.Log(newT.ShowTransaction())
	}
}
