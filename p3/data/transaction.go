package data

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
)

type Transaction struct{
	Source 		string 	`json:"source"`
	Destination string 	`json:"destination"`
	Amount 		float64	`json:"amount"`
	Fee			float64 `json:"fee"`
	//note: do we need to put signatures in here eventually? if so the json will get much more complicated

}

func NewTransaction(src string, dst string, amt float64, fee float64) Transaction{
	res := Transaction{
		Source: src,
		Destination: dst,
		Amount: amt,
		Fee: fee,
	}

	return res
}

func (tAction *Transaction) HashTransaction() string {
	str := ""
	str += tAction.Source +":"
	str += tAction.Destination +":"
	str += fmt.Sprint(tAction.Amount) +":"
	str += fmt.Sprint(tAction.Fee)

	sum := sha3.Sum256([]byte(str))
	hash :=  hex.EncodeToString(sum[:])

	return hash

}

func (tAction *Transaction) TransactionToJson() (string, error){
	res, err := json.Marshal(tAction)

	if err != nil{
		return "", errors.New("unable_to_convert_transaction")
	}

	return string(res), nil
}

func DecodeTransactionFromJsom(jsonStr string) (Transaction, error){
	var res Transaction

	err := json.Unmarshal([]byte(jsonStr), &res)

	if err != nil{
		return res, errors.New("unable_to_decode_json")
	}

	return res, nil

}



