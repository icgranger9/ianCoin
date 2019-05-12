package data

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"sync"
)

type PeerList struct {
	selfId    int32
	peerMap   map[string]int32
	peerKeys  map[string]*rsa.PublicKey
	maxLength int32
	mux       sync.Mutex
}

func NewPeerList(id int32, maxLength int32) PeerList {

	pList := PeerList{}

	pList.peerMap = make(map[string]int32)
	pList.peerKeys = make(map[string]*rsa.PublicKey)
	pList.selfId = id
	pList.maxLength = maxLength

	return pList
}

//not sure if lock is needed, but used to be safe
func (peers *PeerList) Add(addr string, id int32, keyStr string) {
	//note, don't need to re-balance here

	peers.mux.Lock()
	defer peers.mux.Unlock()

	//do nothing if id matches this node's Id
	//double check if true
	if id == peers.selfId {
		return
	}

	//add the new address to the peerMap, and the key  to peerKeys
	peers.peerMap[addr] = id

	key, err := stringToKey(keyStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding key, in pList add: %v\n", err)
		return
	}

	peers.peerKeys[addr] = key


	//fmt.Println("New PeerList: ", peers.Show())

}

//not sure if lock is needed, but used to be safe
func (peers *PeerList) Delete(addr string) {
	peers.mux.Lock()
	defer peers.mux.Unlock()

	delete(peers.peerMap, addr)
	delete(peers.peerKeys, addr)

}

//not sure if lock is needed, but used to be safe
func (peers *PeerList) Rebalance() {
	//if the list is not longer that max (32), we do not need to re-balance
	if len(peers.peerMap) < int(peers.maxLength) {
		return
	}

	//grabs lock, so no one else can change/read the list
	peers.mux.Lock()
	defer peers.mux.Unlock()

	newMap := make(map[string]int32)
	newKeys := make(map[string]*rsa.PublicKey)

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
		//fmt.Println("Inside loop, len:", len(newMap))
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
				//fmt.Println("\tAdding val to newMap:", key, value)
				newMap[key] = value
				newKeys[key] = peers.peerKeys[key]
				break
			}
		}

	}

	//finally update the old map to equal the new map
	peers.peerMap = newMap
	peers.peerKeys = newKeys

}

func (peers *PeerList) Show() string {
	res := "This is PeerMap:\n"

	for key, value := range peers.peerMap {
		valueStr := fmt.Sprint(value) //converts int32 to string
		res += "\taddr=" + key + ", id=" + valueStr + ", public key=" + KeyToString(peers.peerKeys[key]) + "\n"
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

	tmpMap := make(map[string]string)

	for key, value := range peers.peerKeys {
		tmpMap[key] = KeyToString(value)
	}

	//combination := map
	mapJson, err1 := json.Marshal(peers.peerMap)
	keysJson, err2 := json.Marshal(tmpMap)

	if err1 != nil {
		return "", err1
	} else if err2 != nil {
		return "", err2
	}

	res := ""
	res += `"map": ` + string(mapJson) + `,`
	res += `"keys": ` + string(keysJson) + ``
	return "{" + res + "}", nil
}

func (peers *PeerList) InjectPeerMapJson(peerMapJsonStr string, selfAddr string) {

	var jsonInterface map[string]interface{}

	//converts the json string into an interface
	err := json.Unmarshal([]byte(peerMapJsonStr), &jsonInterface)

	//checks that it worked
	if err != nil {
		fmt.Println(peerMapJsonStr)
		panic(err)
	}

	//gets peerMap from interface
	mapInterface := jsonInterface["map"]
	mapMap, success := mapInterface.(map[string]interface{})

	if !success {
		panic(mapInterface)
	}

	var newMap = make(map[string]int32)
	for key, value := range mapMap {
		//fmt.Println("\n\nkey:", key,"\nvalue:", value)
		newMap[key]=int32(value.(float64))
	}

	//gets peerKeys from interface
	keyInterface := jsonInterface["keys"]
	keyMap, success := keyInterface.(map[string]interface{})

	if !success {
		panic(keyInterface)
	}

	//fmt.Println("\n\n", keyMap)

	var newKeys = make(map[string]string)
	for key, value := range keyMap {
		//fmt.Println("\n\nkey:", key,"\nvalue:", value)
		newKeys[key]=value.(string)
	}


	//adds both to PeerList
	for key, value := range newMap {
		//don't need to add our own address
		if key != selfAddr {
			peers.Add(key, value, newKeys[key])
		}
	}

	//fmt.Println("New PeerList: ", peers.Show())
}

func TestPeerListRebalance() {
	peers := NewPeerList(5, 4)
	peers.Add("1111", 1, "11")
	peers.Add("4444", 4, "44")
	peers.Add("-1-1", -1, "-1")
	peers.Add("0000", 0, "00")
	peers.Add("2121", 21, "21")
	peers.Rebalance()
	expected := NewPeerList(5, 4)
	expected.Add("1111", 1, "11")
	expected.Add("4444", 4, "44")
	expected.Add("2121", 21, "21")
	expected.Add("-1-1", -1, "-1")
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 2)
	peers.Add("1111", 1, "11")
	peers.Add("4444", 4, "44")
	peers.Add("-1-1", -1, "-1")
	peers.Add("0000", 0, "00")
	peers.Add("2121", 21, "21")
	peers.Rebalance()
	expected = NewPeerList(5, 2)
	expected.Add("4444", 4, "44")
	expected.Add("2121", 21, "21")
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 4)
	peers.Add("1111", 1, "11")
	peers.Add("7777", 7, "77")
	peers.Add("9999", 9, "99")
	peers.Add("11111111", 11, "1111")
	peers.Add("2020", 20, "20")
	peers.Rebalance()
	expected = NewPeerList(5, 4)
	peers.Add("1111", 1, "11")
	peers.Add("7777", 7, "77")
	peers.Add("9999", 9, "99")
	peers.Add("2020", 20, "20")
	fmt.Println(reflect.DeepEqual(peers, expected))
}

func KeyToString(key *rsa.PublicKey) string {

	bytes := x509.MarshalPKCS1PublicKey(key)
	res := base64.URLEncoding.EncodeToString(bytes)

	return res

	// --- old implementation ---

	//e := fmt.Sprint(key.E)
	//n := key.N.String()
	//
	//return e +":"+n
}

func stringToKey(keyString string) (*rsa.PublicKey, error){

	bytes, err := base64.URLEncoding.DecodeString(keyString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding string in strToKey: %v\n", err)
		return nil, err
	}
	resKey, err2 :=x509.ParsePKCS1PublicKey(bytes)
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "ErrorParsing bytes to key in strToKey: %v\n", err2)
		return nil, err
	}

	return resKey, nil

	// --- old implementation ---

	//slice := strings.Split(keyString, ":")
	//
	//
	//e, err1 := strconv.Atoi(slice[0])
	//n, err2 := new(big.Int).SetString(slice[1], 10)
	//
	//
	//if err1 != nil || err2 == false{
	//	return nil, errors.New("could_not_convert_key")
	//} else {
	//
	//	res := rsa.PublicKey{
	//		N: n,
	//		E: e,
	//	}
	//
	//	return &res, nil
	//}
}
