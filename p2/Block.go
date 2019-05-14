package p2

import (
	"../p1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"strings"
	"time"
)

type Header struct {
	Height     int32
	Timestamp  int64
	Hash       string
	ParentHash string
	Nonce      string
	SizeTransactions	int32
	SizeBalances      	int32
}

type Block struct {
	Header Header
	Transactions  p1.MerklePatriciaTrie
	Balances p1.MerklePatriciaTrie
}

func (b *Block) Initial(initHeight int32, initParentHash string, initNonce string, initTransactions p1.MerklePatriciaTrie, initBalances p1.MerklePatriciaTrie) {

	//Assign header variables
	b.Header.Height = initHeight
	b.Header.Timestamp = time.Now().Unix()
	b.Header.ParentHash = initParentHash
	b.Header.Nonce = initNonce
	b.Header.SizeTransactions = int32(len([]byte(initTransactions.String())))  //converts mpt to string, changers that to byte array, gets length of array, and sets it to int32
	b.Header.SizeBalances = int32(len([]byte(initBalances.String())))  //converts mpt to string, changers that to byte array, gets length of array, and sets it to int32


	//assign value
	b.Transactions = initTransactions
	b.Balances = initBalances

	//get the hash, and assign it
	hashStr := string(b.Header.Height) + string(b.Header.Timestamp) + b.Header.ParentHash + b.Transactions.Root + string(b.Header.SizeTransactions) + b.Balances.Root + string(b.Header.SizeBalances)
	hashSum := sha3.Sum256([]byte(hashStr))
	b.Header.Hash = hex.EncodeToString(hashSum[:])

	//Note: should this return an error if it fails?
}

//Note: probably a cleaner way to do this, with json.marshal()
func EncodeToJSON(b Block) (string, error) {
	//all the easy values to convert to json
	hashStr := `"hash":"` + b.Header.Hash + `",`
	timeStr := `"timeStamp":` + fmt.Sprint(b.Header.Timestamp) + `,`
	heightStr := `"height":` + fmt.Sprint(b.Header.Height) + `,`
	parentStr := `"parentHash":"` + b.Header.ParentHash + `",`
	nonceStr := `"nonce":"` + b.Header.Nonce + `",`
	sizeTrStr := `"sizeTransactions":` + fmt.Sprint(b.Header.SizeTransactions) + `,`
	sizeAccStr :=`"sizeBalances":` + fmt.Sprint(b.Header.SizeBalances) + `,`
	//mpts harder to convert
	transactionStr := `"transactionMpt":{`

	//add each (key, value) to transactionStr
	//remove last comma, and add closing }
	transactionsValue := b.Transactions.KeyVal
	for key, value := range transactionsValue {
		transactionStr += `"` + key + `":"` + value + `", `
	}

	transactionStr = strings.TrimSuffix(transactionStr, ", ")
	transactionStr += `},`

	BalanceStr := `"balancesMpt":{`

	//add each (key, value) to transactionStr
	//remove last comma, and add closing }
	BalancesValue := b.Balances.KeyVal
	for key, value := range BalancesValue {
		BalanceStr += `"` + key + `":"` + value + `", `
	}

	BalanceStr = strings.TrimSuffix(BalanceStr, ", ")
	BalanceStr += `}`



	//add everything together, and return
	res := "{" + hashStr + timeStr + heightStr + parentStr + nonceStr + sizeTrStr + transactionStr + sizeAccStr+ BalanceStr+ "}"

	return res, nil
}

func DecodeFromJson(jsonStr string) (Block, error) {

	var blck Block
	var jsonInterface map[string]interface{}

	//converts the json string into an interface
	err := json.Unmarshal([]byte(jsonStr), &jsonInterface)

	//checks that it worked
	if err != nil {
		return blck, err
	}

	//create the transaction mpt, and insert all values from map
	transactionInterface := jsonInterface["transactionMpt"]
	transactionMap, success := transactionInterface.(map[string]interface{})

	if !success {
		return blck, errors.New("failed_to_decode_TransactionMpt")
	}

	var transactionMpt p1.MerklePatriciaTrie
	transactionMpt.Initial()
	for key, value := range transactionMap {
		transactionMpt.Insert(key, value.(string))
	}

	//create the balances mpt, and insert all values from map
	balancesInterface := jsonInterface["balancesMpt"]
	balancesMap, success := balancesInterface.(map[string]interface{})

	if !success {
		return blck, errors.New("failed_to_decode_balancesMpt")
	}

	var balancesMpt p1.MerklePatriciaTrie
	balancesMpt.Initial()
	for key, value := range balancesMap {
		balancesMpt.Insert(key, value.(string))
	}

	//assign all values to block
	//note: theoretically we should check for failure on every type cast, but it shouldn't be necessary
	blck.Header.SizeTransactions = int32(jsonInterface["sizeTransactions"].(float64))
	blck.Header.SizeBalances = int32(jsonInterface["sizeBalances"].(float64))
	blck.Header.Height = int32(jsonInterface["height"].(float64))
	blck.Header.Timestamp = int64(jsonInterface["timeStamp"].(float64))
	blck.Header.Hash = jsonInterface["hash"].(string)
	blck.Header.ParentHash = jsonInterface["parentHash"].(string)
	blck.Header.Nonce = jsonInterface["nonce"].(string)
	blck.Transactions = transactionMpt
	blck.Balances = balancesMpt

	return blck, nil
}

// ---------------- Added for p3 ----------------

func GenBlock(initHeight int32, initParentHash string, initNonce string, initTransactions p1.MerklePatriciaTrie, initBalances p1.MerklePatriciaTrie) Block {
	newBlock := Block{}
	newBlock.Initial(initHeight, initParentHash, initNonce, initTransactions, initBalances)

	return newBlock
}

func (b *Block) Show() string {
	res := ""

	height := "height=" + fmt.Sprint(b.Header.Height) + ", "
	timestamp := "timestamp=" + fmt.Sprint(b.Header.Timestamp) + ", "
	hash := "hash=" + b.Header.Hash + ", "
	parentHash := "parentHash=" + b.Header.ParentHash + "\n"
	//sizeTrn := "Transactions Size=" + fmt.Sprint(b.Header.SizeTransactions) + ","
	//sizeBal := "Balances Size=" + fmt.Sprint(b.Header.SizeBalances) + "\n"

	res = height + timestamp + hash + parentHash// + sizeTrn + sizeBal

	return res
}
