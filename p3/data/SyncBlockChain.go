package data

import (
	"../../p1"
	"../../p2"
	"fmt"
	"sync"
)

type SyncBlockChain struct {
	bc  p2.BlockChain
	mux sync.Mutex
}

func NewBlockChain() SyncBlockChain {
	return SyncBlockChain{bc: p2.NewBlockChain()}
}

func (sbc *SyncBlockChain) Get(height int32) ([]p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()

	blocks := sbc.bc.Get(height)

	return blocks, len(blocks) > 0
}

func (sbc *SyncBlockChain) GetBlock(height int32, hash string) (p2.Block, bool) {

	blocks := sbc.bc.Get(height)

	sbc.mux.Lock()
	defer sbc.mux.Unlock()

	for _, block := range blocks {
		if block.Header.Hash == hash {
			return block, true
		}
	}

	return p2.Block{}, false

}

func (sbc *SyncBlockChain) Insert(block p2.Block) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()

	sbc.bc.Insert(block)
}

func (sbc *SyncBlockChain) CheckParentHash(insertBlock p2.Block) bool {

	currHeight := insertBlock.Header.Height

	if currHeight == 0 {
		return true
	} else {
		_, exists := sbc.GetBlock(currHeight-1, insertBlock.Header.ParentHash)
		return exists
	}
}

func (sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJson string) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()

	chain, err := p2.DecodeChainFromJson(blockChainJson)

	if err == nil {
		sbc.bc = *chain
	}
}

func (sbc *SyncBlockChain) BlockChainToJson() (string, error) {
	return sbc.bc.EncodeChainToJSON()
}

func (sbc *SyncBlockChain) GenBlock(transactions p1.MerklePatriciaTrie, balances p1.MerklePatriciaTrie,nonce string) p2.Block {

	height := sbc.bc.Height

	parentHash := "" //automatically set parents hash to empty
	if height != -1 {
		// change if it's not the first node
		parents, _ := sbc.Get(height)
		parentHash = parents[0].Header.Hash
	}

	block := p2.GenBlock(height+1, parentHash, nonce, transactions, balances)

	fmt.Println(block.Show())

	return block

}

func (sbc *SyncBlockChain) Show() string {
	return sbc.bc.Show()
}

// ---------------- Added for p4 ----------------

func (sbc *SyncBlockChain) GetLatestBlocks() []p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()

	return sbc.bc.GetLatestBlocks()

}

func (sbc *SyncBlockChain) GetParentBlock(b p2.Block) (p2.Block, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()

	return sbc.bc.GetParentBlock(b)

}
