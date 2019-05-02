package data

import (
	"../../p2"
	"fmt"
)

type HeartBeatData struct {
	IfNewBlock  bool   `json:"ifNewBlock"`
	Id          int32  `json:"id"`
	BlockJson   string `json:"blockJson"`
	PeerMapJson string `json:"peerMapJson"`
	Addr        string `json:"addr"`
	PublicKey   string `json:"pubKey"`
	Hops        int32  `json:"hops"`
}

func NewHeartBeatData(ifNewBlock bool, id int32, blockJson string, peerMapJson string, addr string) HeartBeatData {
	hBeat := HeartBeatData{}

	hBeat.IfNewBlock = ifNewBlock
	hBeat.Id = id
	hBeat.BlockJson = blockJson
	hBeat.PeerMapJson = peerMapJson
	hBeat.Addr = addr
	hBeat.Hops = 3

	return hBeat
}

func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapJson string, addr string) HeartBeatData {
	blockJson, err:= p2.EncodeToJSON(sbc.bc.GetLatestBlocks()[0])
	if err != nil {
		fmt.Println(err)
		return HeartBeatData{}
	}
	return NewHeartBeatData(false, selfId, blockJson, peerMapJson, addr)

}
