package p3

import (
	"../p1"
	"../p2"
	"./data"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)


var TRUSTED_SERVER = "http://localhost:9901"
var SELF_PUBLIC = ""
var SELF_PRIVATE = ""
var SELF_ADDR = "http://localhost:" + os.Args[1]
var NUM_0s = "000000"

var SBC data.SyncBlockChain
var Peers data.PeerList
var ifStarted bool

func init() {
	// This function will be executed before everything else.
	// Do some initialization here.
	SBC = data.NewBlockChain()
	Peers = data.NewPeerList(-1, 32) //ID set to flag val, will be changed in Register()
}

// Register ID, download BlockChain, start HeartBeat
func Start(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Running start")
	Register()
	Download()

	go StartHeartBeat()
	go StartTryingNonces()
}

// Display peerList and sbc
func Show(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())
}

// Register to TA's server, get an ID
func Register() {

	//Should connect to TA's server to get unique ID
	//Instead, hardcoded while the TA's server is down, to read in from command line
	if len(os.Args) > 2 {
		id, _ := strconv.ParseInt(os.Args[2], 10, 32)
		Peers.Register(int32(id))
	} else {
		Peers.Register(int32(9999))
	}
}

// Download blockchain from TA server
func Download() {

	if Peers.GetSelfId() == 1 {
		//if this is the first node, just create the blockchain
		fmt.Println("\tAm node1, creating blockchain")

		//create simple / random MPT
		//same as in PrepareHeartBeatData()
		mpt := p1.MerklePatriciaTrie{}
		mpt.Initial()

		dict := []string{"this", "is", "simple", "dictionary", "of", "words", "that", "can", "be", "added", "into", "our", "mtp", "I", "will", "now", "add", "many", "other", "options", "just", "to", "make", "it", "more", "random", "sound", "good?", "hello", "world", "golang", "USF", "computer", "science", "san", "francisco", "california", "america", "golden", "state", "warriors", "hopefully", "thats", "enough"}
		mptSize := rand.Intn(len(dict)-1) + 1
		fmt.Println("mpt size:", mptSize)

		for wordsAdded := 0; wordsAdded < mptSize; wordsAdded++ {
			randNum := rand.Intn(len(dict) - 1)
			word := dict[randNum]

			mpt.Insert(word, fmt.Sprint("word num:", wordsAdded))
		}

		//fmt.Println(mpt.String())

		var newBlock p2.Block

		newBlock.Initial(0, "", "", mpt)

		SBC.Insert(newBlock)

	} else {
		fmt.Println("\tGetting blockchain from node1")
		//otherwise, download it from node 1

		//hardcoded node1's url, not great
		node1 := TRUSTED_SERVER
		request := "/upload"
		parameters := "?address=" + SELF_ADDR + "&id=" + fmt.Sprint(Peers.GetSelfId())

		resp, err := http.Get(node1 + request + parameters)

		if err != nil {
			log.Fatalln(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		SBC.UpdateEntireBlockChain(string(body))

	}

	fmt.Println(SBC.BlockChainToJson())

}

// Upload blockchain to whoever called this method, return jsonStr
//updated to read address and id from URL parameters
func Upload(w http.ResponseWriter, r *http.Request) {

	//handles adding new node to peerList
	query := r.URL.Query()
	address := query.Get("address")
	id := query.Get("id")

	if id == "" || address == "" {
		fmt.Println("invalid address or id from upload parameters")
		return
	} else {
		idInt32, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			fmt.Println("id could not be converted from string to int32")
			return
		} else {
			fmt.Println("adding address: ", address, " and id: ", idInt32, " to peerList")
			Peers.Add(address, int32(idInt32))
		}
	}

	//returns blockchain
	blockChainJson, err := SBC.BlockChainToJson()
	if err != nil {
		fmt.Println(err)
		return
		//data.PrintError(err, "Upload") //not sure if this is a function we were supposed to add, or what
	}
	fmt.Fprint(w, blockChainJson) //note, should we handle error?
}

// Upload a block to whoever called this method, return jsonStr
func UploadBlock(w http.ResponseWriter, r *http.Request) {
	uri := r.RequestURI
	uriSplit := strings.Split(uri, "/")

	height, err := strconv.ParseInt(uriSplit[1], 10, 32)
	if err != nil {
		fmt.Print("id could not be converted from string to int32")
		return
	}

	block, found := SBC.GetBlock(int32(height), uriSplit[2])

	if found {
		blockJson, _ := p2.EncodeToJSON(block)
		fmt.Fprint(w, blockJson)
	}
}

// Received a heartbeat
func HeartBeatReceive(w http.ResponseWriter, r *http.Request) {
	jsonBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Print(err)
		return
	}
	var hBeat data.HeartBeatData
	json.Unmarshal(jsonBody, &hBeat)

	fmt.Println("got heartbeat from:", hBeat.Addr)

	//verify heartbeat
	verified := false

	newBlock, _ := p2.DecodeFromJson(hBeat.BlockJson)

	parentHash := newBlock.Header.ParentHash
	nonce := newBlock.Header.Nonce
	mptHash := newBlock.Value.Root

	concatInfo := parentHash + nonce + mptHash
	//fmt.Println("hashing:", parentHash, "\n", nonce, "\n", mptHash)

	proofOfWork := sha3.Sum256([]byte(concatInfo))
	powString := hex.EncodeToString(proofOfWork[:])

	verified = strings.HasPrefix(powString, NUM_0s)

	if verified {
		//add the node that we get the heartbeat from, and it's peers
		Peers.Add(hBeat.Addr, hBeat.Id)
		Peers.InjectPeerMapJson(hBeat.PeerMapJson, SELF_ADDR)

		//no new block, do nothing else
		if hBeat.IfNewBlock == false {
			return
		}

		block, err := p2.DecodeFromJson(hBeat.BlockJson)
		if err != nil {
			fmt.Println(err)
			return
		}

		//1
		//note: Why are we only checking it's direct parent, shouldn't we check it's entire ancestary, to get any missing links
		parentExists := SBC.CheckParentHash(block)
		if parentExists == false {
			fmt.Println("Parent doesn't exist")
			AskForBlock(block.Header.Height-1, block.Header.ParentHash)
		}

		//2
		SBC.Insert(block)

		//3
		hBeat.Hops = hBeat.Hops - 1
		if hBeat.Hops > 0 {
			ForwardHeartBeat(hBeat)
		}

		//write simple response
		w.Write([]byte("\tBlock received by: " + SELF_ADDR))

		fmt.Println("Finished Receive heart beat")
	}

}

// Ask another server to return a block of certain height and hash
func AskForBlock(height int32, hash string) {
	fmt.Println("Asking for block")
	url := "/block/" + fmt.Sprint(height) + "/" + hash

	peerMap := Peers.Copy()

	for keyAddr := range peerMap {

		resp, err := http.Get("http://" + keyAddr + url)

		if err != nil {
			fmt.Println(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		} else {
			block, err := p2.DecodeFromJson(string(body))
			if err == nil {
				SBC.Insert(block)

				//found represents the error returned when whe parent is not actually
				_, found := SBC.GetParentBlock(block)
				if found != nil {
					parentHash := block.Header.ParentHash
					parentHeight := block.Header.Height - 1

					AskForBlock(parentHeight, parentHash)
				}
			}
			break
		}

	}
}

func ForwardHeartBeat(heartBeatData data.HeartBeatData) {
	url := "/heartbeat/receive"
	httpType := "application/json"
	hBeatJson, _ := json.Marshal(heartBeatData)

	peerMap := Peers.Copy()

	for keyAddr := range peerMap {

		resp, err := http.Post(keyAddr+url, httpType, bytes.NewBuffer(hBeatJson))

		//not really needed, since the response doesn't matter
		if err != nil {
			fmt.Println(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(body))
		}

	}
}

func StartHeartBeat() {

	ifStarted = true

	for ifStarted {

		fmt.Println("In start heartbeat")

		timeToSleep := rand.Intn(5) + 5
		time.Sleep(time.Duration(timeToSleep) * time.Second)

		PeersJson, err := Peers.PeerMapToJson()
		if err != nil {
			fmt.Println(err)
		} else {
			heartBeatData := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), PeersJson, SELF_ADDR)

			url := "/heartbeat/receive"
			httpType := "application/json"
			hBeatJson, _ := json.Marshal(heartBeatData)

			peerMap := Peers.Copy()

			for keyAddr := range peerMap {

				fmt.Println("sent heartbeat to:", keyAddr+url)
				resp, err := http.Post(keyAddr+url, httpType, bytes.NewBuffer(hBeatJson))

				//not really needed, since the response doesn't matter
				if err != nil {
					fmt.Println("Got an error in the response")
					fmt.Println(err)
				}

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println("Got an error in reading the body")
					fmt.Println(err)
				} else {
					fmt.Println("\tPrinting the response body: ", string(body))
				}

			}
		}

	}
}

// ---------------- Added for p4 ----------------

func Canonical(w http.ResponseWriter, r *http.Request) {
	currChains := SBC.GetLatestBlocks()

	res := ""

	for ind, block := range currChains {
		res += "Chain #" + fmt.Sprint(ind+1) + ":\n"

		for block.Header.ParentHash != "" {
			res += block.Show()
			block, _ = SBC.GetParentBlock(block)

		}

		res += block.Show()

		w.Write([]byte(res))
	}
}

func StartTryingNonces() {
	fmt.Println("In start trying nonces")

	calculateNonce := true

	//generate mpt
	mpt := data.GenerateMPT()

	for calculateNonce {

		//get block
		currHead := SBC.GetLatestBlocks()[0]

		//generate nonce
		nonce := data.GenerateNonce()

		//test nonce
		validNonce := false

		concatInfo := currHead.Header.Hash + nonce + mpt.Root
		proofOfWork := sha3.Sum256([]byte(concatInfo))
		powString := hex.EncodeToString(proofOfWork[:])

		validNonce = strings.HasPrefix(powString, NUM_0s)

		if validNonce {
			//fmt.Println("hashing:", currHead.Header.Hash, "\n", nonce, "\n",  mpt.Root)

			newBlock := SBC.GenBlock(mpt, nonce)
			blockJson, _ := p2.EncodeToJSON(newBlock)
			peersJson, _ := Peers.PeerMapToJson()

			hBeat := data.NewHeartBeatData(true, Peers.GetSelfId(), blockJson, peersJson, SELF_ADDR)

			ForwardHeartBeat(hBeat)
			fmt.Println("found valid nonce")

		}

	}
}
