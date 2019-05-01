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
	Size       int32
}

type Block struct {
	Header Header
	Value  p1.MerklePatriciaTrie
}

func (b *Block) Initial(initHeight int32, initParentHash string, initNonce string, initValue p1.MerklePatriciaTrie) {

	//Assign header variables
	b.Header.Height = initHeight
	b.Header.Timestamp = time.Now().Unix()
	b.Header.ParentHash = initParentHash
	b.Header.Nonce = initNonce
	b.Header.Size = int32(len([]byte(initValue.String()))) //converts mpt to string, changers that to byte array, gets length of array, and sets it to int32

	//assign value
	b.Value = initValue

	//get the hash, and assign it
	hashStr := string(b.Header.Height) + string(b.Header.Timestamp) + b.Header.ParentHash + b.Value.Root + string(b.Header.Size)
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
	sizeStr := `"size":` + fmt.Sprint(b.Header.Size) + `,`

	//mpt harder to convert
	mptStr := `"mpt":{`

	//add each (key, value) to mptStr
	//remove last comma, and add closing }
	keyValue := b.Value.KeyVal
	for key, value := range keyValue {
		mptStr += `"` + key + `":"` + value + `", `
	}

	mptStr = strings.TrimSuffix(mptStr, ", ")
	mptStr += `}`

	//add everything together, and return
	res := "{" + hashStr + timeStr + heightStr + parentStr + nonceStr + sizeStr + mptStr + "}"
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

	//create the mpt, and insert all values from map
	mptInterface := jsonInterface["mpt"]
	mptMap, success := mptInterface.(map[string]interface{})

	if !success {
		fmt.Println(mptInterface)
		return blck, errors.New("failed_to_decode_mpt")
	}

	var mpt p1.MerklePatriciaTrie
	mpt.Initial()
	for key, value := range mptMap {
		mpt.Insert(key, value.(string))
	}

	//assign all values to block
	//note: theoretically we should check for failure on every type cast, but it shouldn't be necessary
	blck.Header.Size = int32(jsonInterface["size"].(float64))
	blck.Header.Height = int32(jsonInterface["height"].(float64))
	blck.Header.Timestamp = int64(jsonInterface["timeStamp"].(float64))
	blck.Header.Hash = jsonInterface["hash"].(string)
	blck.Header.ParentHash = jsonInterface["parentHash"].(string)
	blck.Header.Nonce = jsonInterface["nonce"].(string)
	blck.Value = mpt

	//fmt.Println(blck)

	return blck, nil
}

// ---------------- Added for p3 ----------------

func GenBlock(initHeight int32, initParentHash string, initNonce string, initValue p1.MerklePatriciaTrie) Block {
	newBlock := Block{}
	newBlock.Initial(initHeight, initParentHash, initNonce, initValue)

	return newBlock
}

func (b *Block) Show() string {
	res := ""

	height := "height=" + fmt.Sprint(b.Header.Height) + ", "
	timestamp := "timestamp=" + fmt.Sprint(b.Header.Timestamp) + ", "
	hash := "hash=" + b.Header.Hash + ", "
	parentHash := "parentHash=" + b.Header.ParentHash + ", "
	size := "size=" + fmt.Sprint(b.Header.Size) + "\n"

	res = height + timestamp + hash + parentHash + size

	return res
}
