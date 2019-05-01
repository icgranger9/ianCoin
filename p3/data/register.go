package data

import "encoding/json"

type RegisterData struct {
	AssignedId  int32  `json:"assignedId"`
	PeerMapJson string `json:"peerMapJson"`
}

func NewRegisterData(id int32, peerMapJson string) RegisterData {
	regData := RegisterData{}
	regData.AssignedId = id
	regData.PeerMapJson = peerMapJson

	return regData
}

func (data *RegisterData) EncodeToJson() (string, error) {
	res, err := json.Marshal(data)
	return string(res), err
}
