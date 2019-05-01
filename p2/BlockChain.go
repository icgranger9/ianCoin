package p2

import (
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"sort"
	"strings"
)

type BlockChain struct {
	Chain  map[int32][]Block
	Height int32
}

func (bChain *BlockChain) Get(height int32) []Block {
	return bChain.Chain[height] //is it really that simple? What if empty? what if not initialized?
}

func (bChain *BlockChain) Insert(b Block) error {
	//if map has not been created, do so
	//Note: are we sure this is needed
	if bChain.Chain == nil {
		bChain.Chain = make(map[int32][]Block)
		bChain.Height = 0
	}

	bHeight := b.Header.Height
	bHash := b.Header.Hash

	//check if hash is already in position, can't insert same block twice
	blocks := bChain.Chain[bHeight]

	if len(blocks) != 0 {
		for ind := 0; ind < len(blocks); ind++ {
			blockInChain := blocks[ind]
			if blockInChain.Header.Hash == bHash {
				return errors.New("block_already_in_chain")
			}
		}
	}

	//if not, insert
	bChain.Chain[bHeight] = append(bChain.Chain[bHeight], b)

	//update height if needed
	if bHeight > bChain.Height {
		bChain.Height = bHeight
		fmt.Println("updated height")
	}
	return nil
}

// my implementation of EncodeToJSON
func (bChain *BlockChain) EncodeChainToJSON() (string, error) {
	var res string

	//iterate through all values in the map
	for _, blockList := range bChain.Chain {

		//for all of those, iterate through the entire list
		for _, block := range blockList {

			//get the Json for that block
			//add to res
			blockJson, err := EncodeToJSON(block)
			if err != nil {
				return "", errors.New("failed_to_encode_blockchain")
			}

			res += blockJson + ", "
		}
	}

	//remove additional , from the end of it
	res = strings.TrimSuffix(res, ", ")

	return `[` + res + `]`, nil
}

// my implementation of DecodeFromJSON
func DecodeChainFromJson(json string) (*BlockChain, error) {
	var resChain BlockChain

	//remove [ and ]
	json = strings.Trim(json, "[]")

	//split using "Split after"
	jsonList := strings.SplitAfter(json, "}}")

	//iterate over the json for all blocks
	for _, blockJson := range jsonList {

		//make sure it isn't empty
		if blockJson != "" {

			//remove leading ,
			blockJson = strings.Trim(blockJson, ", ")

			block, err := DecodeFromJson(blockJson)
			if err != nil {
				return &resChain, errors.New("failed_to_decode_chain")
			}

			err = resChain.Insert(block)
			if err != nil {
				return &resChain, errors.New("failed_to_insert_decoded_block")
			}
		}
	}

	return &resChain, nil
}

func (bChain *BlockChain) Show() string {
	rs := ""
	var idList []int
	for id := range bChain.Chain {
		idList = append(idList, int(id))
	}
	sort.Ints(idList)
	for _, id := range idList {
		var hashs []string
		for _, block := range bChain.Chain[int32(id)] {
			hashs = append(hashs, block.Header.Hash+"<="+block.Header.ParentHash)
		}
		sort.Strings(hashs)
		rs += fmt.Sprintf("%v: ", id)
		for _, h := range hashs {
			rs += fmt.Sprintf("%s, ", h)
		}
		rs += "\n"
	}
	sum := sha3.Sum256([]byte(rs))
	rs = fmt.Sprintf("This is the BlockChain: %s\n", hex.EncodeToString(sum[:])) + rs
	return rs
}

// ---------------- Added for p3 ----------------

func NewBlockChain() BlockChain {
	var bChain BlockChain

	bChain.Chain = make(map[int32][]Block)
	bChain.Height = 0

	return bChain
}

// ---------------- Added for p4 ----------------

func (bChain *BlockChain) GetLatestBlocks() []Block {
	return bChain.Get(bChain.Height)
}

func (bChain *BlockChain) GetParentBlock(b Block) (Block, error) {
	parentHash := b.Header.ParentHash
	parentHeight := b.Header.Height - 1

	possibleParents := bChain.Get(parentHeight)

	for _, block := range possibleParents {
		if block.Header.Hash == parentHash {
			return block, nil
		}
	}
	//on off chance it isn't found
	var emptyBlock Block
	return emptyBlock, errors.New("parent_not_found")
}
