package data

import (
	"../../p1"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

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
	tx := Transaction{
		Source: src,
		Destination: dst,
		Amount: amt,
		Fee: fee,
		TimeToLive: 3, //default ttl is hardcoded, maybe include as variable in handlers.go instead?
		Timestamp: fmt.Sprint(initTimestamp),

	}
	
	return tx
}

func (tx *Transaction) SignTransaction(privateKey *rsa.PrivateKey) error {
	//note: only signs the hash of the transaction

	hash := tx.HashTransaction()
	hashBytes, err :=base64.URLEncoding.DecodeString( hash)

	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashBytes[:])

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from signing: %v\n", err)
		return err
	} else {
		tx.Signature = base64.URLEncoding.EncodeToString(sig)
		return nil
	}
}

func (tx *Transaction) HashTransaction() string {
	str := ""
	str += tx.Source +":"
	str += tx.Destination +":"
	str += fmt.Sprint(tx.Amount) +":"
	str += fmt.Sprint(tx.Fee) +":"
	//str += fmt.Sprint(tx.TimeToLive) +":" //note: should ttl really be part of the hash?
	str += tx.Timestamp

	sum := sha256.Sum256([]byte(str))
	hash :=  base64.URLEncoding.EncodeToString(sum[:]) //not the best way to convert, may try something else later

	return hash

}

func (tx *Transaction) ShowTransaction() string {
	var res string

	res += "\n"
	res += "Source: " + tx.Source + "\n"
	res += "Destination: " + tx.Destination + "\n"
	res += "Amount: " + fmt.Sprint(tx.Amount) + "\n"
	res += "Fee: " + fmt.Sprint(tx.Fee) + "\n"
	res += "TTL: " + fmt.Sprint(tx.TimeToLive) + "\n"
	res += "Timestamp: " + tx.Timestamp + "\n"
	res += "Signature: " + tx.Signature + "\n"

	return res

}

func (tx *Transaction) VerifyTransaction() bool {
	//verifies that the transaction is actually valid
	//steps:
		//Check that fee is actually 5%
		//checks signature to validate
		//TODO: should it check that destination is a valid address?

	var validFee bool
	var validSignature bool

	//validate fee
	validFee = tx.Fee == .05 * tx.Amount

	//validate signature
	if tx.Signature == ""{
		validSignature = false
	} else {
		pubKey, err1 := stringToKey(tx.Source)

		if err1 != nil {
			fmt.Fprintf(os.Stderr, "Error getting key in validation: %v\n", err1)
			return false
		}

		hash := tx.HashTransaction()
		hashBytes, err2 :=base64.URLEncoding.DecodeString(hash)
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "Error decoding hash in validation: %v\n", err2)
			return false
		}

		sig, err3 := base64.URLEncoding.DecodeString(tx.Signature)

		if err3 != nil {
			fmt.Fprintf(os.Stderr, "Error decoding signature in validation: %v\n", err3)
			return false
		}

		err4 := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashBytes, sig)

		if err4 != nil{
			fmt.Fprintf(os.Stderr, "Error from validation: %v\n", err4)
			validSignature = false
		} else {
			validSignature = true
		}
	}

	return validFee==true && validSignature==true

}

func (tx *Transaction) AddTransactionToMPTs(balances p1.MerklePatriciaTrie, transactions p1.MerklePatriciaTrie) (p1.MerklePatriciaTrie, p1.MerklePatriciaTrie, error){
	fmt.Println("---------------- Adding TX to MPTs ----------------")
	// Function to add tx to both MPTs, if it can co through
		// Only errors we are supposed to throw are: "transaction_is_double_spending" and "src_cannot_afford_tx"
		// anything else means we have a serious problem

		// note: does this need to be atomic? I think not, because this isn't done in parallel


	//first check if transaction is already in transactions, to prevent double spending
	if transactions.Root == ""{
		//if root of transactions is empty, then there are no previous TXs, and obviously no double spending

	} else {
		_, err := transactions.Get(tx.HashTransaction()) //what does this return if the transaction is not in the MPT?
		if err != nil {
			//means either no double spending, orr there's an error in my mpt

			if err.Error() != "reached_invalid_leaf" {
				//reached unusual err. log it, but do not return
				fmt.Fprintf(os.Stderr, "Error checking for double spending in AddTransactionToMPTs: %v\n", err)
			}

		} else {
			//definitely double spending
			return p1.MerklePatriciaTrie{}, p1.MerklePatriciaTrie{}, errors.New("transaction_is_double_spending")
		}
	}


	//second, check balance, to see if there are enough funds
	srcCurrBal, err2 := balances.Get(tx.Source)
	var srcBalFloat float64 = 0
	if err2 != nil{
		return p1.MerklePatriciaTrie{}, p1.MerklePatriciaTrie{}, errors.New("could_not_get_src_balance")
	} else {
		//converts string to float
		srcBalFloat, err3 := strconv.ParseFloat(srcCurrBal, 64)

		if err3 != nil{
			return p1.MerklePatriciaTrie{}, p1.MerklePatriciaTrie{}, errors.New("could_not_convert_src_balance")
		} else if srcBalFloat < (tx.Fee+tx.Amount) {
			//finally an error we are likely / expected to hit
			return p1.MerklePatriciaTrie{}, p1.MerklePatriciaTrie{}, errors.New("src_cannot_afford_tx")
		}
	}

	//third, update balance of both accounts

	dstCurrBal, err3 := balances.Get(tx.Destination)
	if err3 != nil {
		if err3.Error() == "reached_invalid_leaf" {
			//must add dst to balances mpt
			balances.Insert(tx.Source, fmt.Sprint(srcBalFloat - (tx.Fee+tx.Amount)) ) //reduce source balance
			balances.Insert(tx.Destination, fmt.Sprint(tx.Amount) ) //create dst and set its balance

		} else {
				//if mpt.get has unexpected error
				fmt.Fprintf(os.Stderr, "Error checkingdstBalance in AddTransactionToMPTs: %v\n", err2) //is this needed?
				return p1.MerklePatriciaTrie{}, p1.MerklePatriciaTrie{}, errors.New("dst_breaks_balance_mpt")
		}
	} else {

		//must update balance of dst in MPT
		dstBalFloat, err4 := strconv.ParseFloat(dstCurrBal, 64)
		if err4 != nil{
			return p1.MerklePatriciaTrie{}, p1.MerklePatriciaTrie{}, errors.New("could_not_convert_dts_balance")
		} else {
			//dst exists, must update both balances
			balances.Insert(tx.Source, fmt.Sprint( srcBalFloat - (tx.Fee+tx.Amount)) ) 		//reduce source balance
			balances.Insert(tx.Destination, fmt.Sprint( dstBalFloat +tx.Amount) )	//create dst and set its balance
		}
	}

	//add transaction to transactions mpt
	txJson, err5 := tx.TransactionToJson()
	if err5 != nil{
		return p1.MerklePatriciaTrie{}, p1.MerklePatriciaTrie{}, errors.New("could_not_convert_tx_to_json")
	}

	transactions.Insert(tx.HashTransaction(), txJson)

	//fmt.Println("MPTs after TXs added:","\n\tBalances:", balances.Order_nodes(), "\n\tTransactions:", transactions.Order_nodes())

	fmt.Println("---------------- Finished Adding TX to MPTs ----------------")
	//if everything above works, return updated MPTs with no error
	return balances, transactions, nil



}

func (tx *Transaction) TransactionToJson() (string, error){
	res, err := json.Marshal(tx)

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





