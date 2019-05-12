package data

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Transaction struct{
	Source 		string 	`json:"source"`
	Destination string 	`json:"destination"`
	Amount 		float64	`json:"amount"`
	Fee			float64 `json:"fee"`
	TimeToLive	int		`json:"timeToLive"`
	Timestamp	string	`json:"timestamp"` //note: need to put in something to prevent double spending. Probably a timestamp
	Signature 	string 	`json:"signature"`//note: do we need to put signatures in here eventually? if so the json will get much more complicated

}

func NewTransaction(src string, dst string, amt float64, fee float64, initTimestamp int64) Transaction{
	res := Transaction{
		Source: src,
		Destination: dst,
		Amount: amt,
		Fee: fee,
		TimeToLive: 3, //default ttl is hardcoded, maybe include as variable in handlers.go instead?
		Timestamp: fmt.Sprint(initTimestamp),

	}
	
	return res
}

func (tAction *Transaction) SignTransaction(privateKey *rsa.PrivateKey) error {
	//note: only signs the hash of the transaction

	hash := tAction.HashTransaction()
	fmt.Printf("  result hash: %s\n", hash)
	hashBytes, err :=base64.URLEncoding.DecodeString( hash)

	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashBytes[:])

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from signing: %s\n", err)
		return err
	} else {
		tAction.Signature = base64.URLEncoding.EncodeToString(sig)
		return nil
	}
}

func (tAction *Transaction) HashTransaction() string {
	str := ""
	str += tAction.Source +":"
	str += tAction.Destination +":"
	str += fmt.Sprint(tAction.Amount) +":"
	str += fmt.Sprint(tAction.Fee) +":"
	str += fmt.Sprint(tAction.TimeToLive) +":"
	str += tAction.Timestamp

	sum := sha256.Sum256([]byte(str))
	fmt.Printf("original hash: %s\n", base64.URLEncoding.EncodeToString(sum[:]))
	hash :=  base64.URLEncoding.EncodeToString(sum[:]) //not the best way to convert, may try something else later

	return hash

}

func (tAction *Transaction) ShowTransaction() string {
	var res string

	res += "\n"
	res += "Source: " + tAction.Source + "\n"
	res += "Destination: " + tAction.Destination + "\n"
	res += "Amount: " + fmt.Sprint(tAction.Amount) + "\n"
	res += "Fee: " + fmt.Sprint(tAction.Fee) + "\n"
	res += "TTL: " + fmt.Sprint(tAction.TimeToLive) + "\n"
	res += "Timestamp: " + tAction.Timestamp + "\n"
	res += "Signature: " + tAction.Signature + "\n"

	return res

}

func (tAction *Transaction) VerifyTransaction() bool {
	//verifies that the transaction is actually valid
	//steps:
		//Check that fee is actually 5%
		//checks signature to validate
		//TODO: should it check that destination is a valid address?

	var validFee bool
	var validSignature bool

	//validate fee
	validFee = tAction.Fee == .05 * tAction.Amount

	//validate signature
	if tAction.Signature == ""{
		validSignature = false
	} else {
		pubKey, err1 := stringToKey(tAction.Source)

		if err1 != nil {
			fmt.Fprintf(os.Stderr, "Error getting key in validation: %s\n", err1)
			return false
		}

		hash := tAction.HashTransaction()
		hashBytes, err2 :=base64.URLEncoding.DecodeString(hash)
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "Error decoding hash in validation: %s\n", err1)
			return false
		}

		sig, err3 := base64.URLEncoding.DecodeString(tAction.Signature)

		if err3 != nil {
			fmt.Fprintf(os.Stderr, "Error decoding signature in validation: %s\n", err1)
			return false
		}

		err4 := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashBytes, sig)

		if err4 != nil{
			fmt.Fprintf(os.Stderr, "Error from validation: %s\n", err1)
			validSignature = false
		} else {
			validSignature = true
		}
	}

	return validFee==true && validSignature==true

}

func (tAction *Transaction) TransactionToJson() (string, error){
	res, err := json.Marshal(tAction)

	if err != nil{
		return "", errors.New("unable_to_convert_transaction")
	}

	return string(res), nil
}

func DecodeTransactionFromJson(jsonStr string) (Transaction, error){
	var res Transaction

	err := json.Unmarshal([]byte(jsonStr), &res)

	if err != nil{
		return res, errors.New("unable_to_decode_json")
	}

	return res, nil

}



