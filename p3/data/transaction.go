package data

import (
	"encoding/json"
	"errors"
)

type Transaction struct{
	Source 		string 	`json:"source"`
	Destination string 	`json:"destination"`
	Amount 		float64	`json:"amount"`
	//note: do we need to put signatures in here eventually? if so the json will get much more complicated

}

func NewTransaction(src string, dst string, amt float64) Transaction{
	res := Transaction{
		Source: src,
		Destination: dst,
		Amount: amt,
	}

	return res
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
