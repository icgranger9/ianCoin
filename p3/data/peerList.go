package data

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

type PeerList struct {
	selfId    int32
	peerMap   map[string]int32
	maxLength int32
	mux       sync.Mutex
}

func NewPeerList(id int32, maxLength int32) PeerList {

	pList := PeerList{}

	pList.peerMap = make(map[string]int32)
	pList.selfId = id
	pList.maxLength = maxLength

	return pList
}

//not sure if lock is needed, but used to be safe
func (peers *PeerList) Add(addr string, id int32) {
	//note, don't need to rebalance here

	peers.mux.Lock()
	defer peers.mux.Unlock()

	//do nothing if id matches this node's Id
	//double check if true
	if id == peers.selfId {
		return
	}

	//add the new address to the peerMap
	peers.peerMap[addr] = id

}

//not sure if lock is needed, but used to be safe
func (peers *PeerList) Delete(addr string) {
	peers.mux.Lock()
	defer peers.mux.Unlock()

	delete(peers.peerMap, addr)

}

//not sure if lock is needed, but used to be safe
func (peers *PeerList) Rebalance() {
	//if the list is not longer that max (32), we do not need to rebalance
	if len(peers.peerMap) < int(peers.maxLength) {
		return
	}

	//grabs lock, so no one else can change/read the list
	peers.mux.Lock()
	defer peers.mux.Unlock()

	newMap := map[string]int32{}

	//create a slice, from the Id's of all peers, plus the Id of this node
	peerSlice := make([]int32, len(peers.peerMap))
	for _, value := range peers.peerMap {
		peerSlice = append(peerSlice, value)
	}
	peerSlice = append(peerSlice, peers.selfId)

	//sort the slice
	//note, implements inline comparison function, because they're int32's
	sort.Slice(peerSlice, func(i, j int) bool { return peerSlice[i] < peerSlice[j] })

	//find index of our Id
	var ourInd int
	for ind := 0; ind < len(peerSlice); ind++ {
		if peerSlice[ind] == peers.selfId {
			ourInd = ind
			break
		}
	}

	//create variables, dist below and dist above
	indBelow := ourInd - 1
	indAbove := ourInd + 1

	//while we can add new Id's
	for len(newMap) < int(peers.maxLength) {
		fmt.Println("Inside loop, len:", len(newMap))
		//find the closest Id, and add to new map

		//used to avoid repeated code. This is the index if the id that will be added to the map
		indToAdd := -1

		if indBelow < 0 && indAbove > len(peerSlice) {
			//if both our of range, just break
			//note, shouldn't be possible
			break

		} else if indBelow < 0 {
			//if below out of range
			indToAdd = indAbove
			indAbove++

		} else if indAbove > len(peerSlice) {
			//if above out of range
			indToAdd = indBelow
			indBelow--

		} else {
			//if neither out of range, must compare

			difBelow := peerSlice[ourInd] - peerSlice[indBelow]
			difAbove := peerSlice[indAbove] - peerSlice[ourInd]

			if difBelow < difAbove {
				//if below is closer
				indToAdd = indBelow
				indBelow--

			} else {
				//if above is closer, or equal
				indToAdd = indAbove
				indAbove++
			}
		}

		//adds to the new map
		for key, value := range peers.peerMap {
			if indToAdd < 0 {
				//theoretically should never happen, included just to be safe

				break
			}
			if value == peerSlice[indToAdd] {
				fmt.Println("\tAdding val to newMap:", key, value)
				newMap[key] = value
				break
			}
		}

	}

	//finally update the old map to equal the new map
	peers.peerMap = newMap

}

func (peers *PeerList) Show() string {
	res := "This is PeerMap:\n"

	for key, value := range peers.peerMap {
		valueStr := fmt.Sprint(value) //converts int32 to string
		res += "\taddr=" + key + ", id=" + valueStr + "\n"
	}

	return res
}

func (peers *PeerList) Register(id int32) {
	peers.selfId = id
	fmt.Printf("\tSelfId=%v\n", id)
}

func (peers *PeerList) Copy() map[string]int32 {
	return peers.peerMap
}

func (peers *PeerList) GetSelfId() int32 {
	return peers.selfId
}

//need to decide format of json based on piazza, will come back to these two later
func (peers *PeerList) PeerMapToJson() (string, error) {
	mapJson, err := json.Marshal(peers.peerMap)

	return string(mapJson), err
}

func (peers *PeerList) InjectPeerMapJson(peerMapJsonStr string, selfAddr string) {
	var newMap map[string]int32
	err := json.Unmarshal([]byte(peerMapJsonStr), &newMap)

	if err != nil {
		panic(err)
	}

	for key, value := range newMap {
		//don't need to add our own address
		if key != selfAddr {
			peers.Add(key, value)
		}
	}
}

func TestPeerListRebalance() {
	peers := NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected := NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	expected.Add("-1-1", -1)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 2)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected = NewPeerList(5, 2)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("7777", 7)
	peers.Add("9999", 9)
	peers.Add("11111111", 11)
	peers.Add("2020", 20)
	peers.Rebalance()
	expected = NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("7777", 7)
	expected.Add("9999", 9)
	expected.Add("2020", 20)
	fmt.Println(reflect.DeepEqual(peers, expected))
}
