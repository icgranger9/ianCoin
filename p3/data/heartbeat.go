package data

type HeartBeatData struct {
	IfNewBlock  bool   `json:"ifNewBlock"`
	Id          int32  `json:"id"`
	BlockJson   string `json:"blockJson"`
	PeerMapJson string `json:"peerMapJson"`
	Addr        string `json:"addr"`
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
	blockJson, _ := sbc.bc.EncodeChainToJSON()
	return NewHeartBeatData(false, selfId, blockJson, peerMapJson, addr)

}
