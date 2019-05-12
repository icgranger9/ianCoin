package p3

import (
	"../p1"
	"../p2"
	"./data"
	"bytes"
	crand "crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)


var TRUSTED_SERVER = "http://localhost:9901"
var SELF_PUBLIC *rsa.PublicKey
var SELF_PRIVATE *rsa.PrivateKey
var SELF_ADDR = "http://localhost:" + os.Args[1]
var NUM_0s = "000000"

var SBC data.SyncBlockChain
var TRANSACTION_POOL []data.Transaction
var Peers data.PeerList
var ifStarted bool //is this really needed as a global variable


func init() {
	// This function will be executed before everything else.
	// Do some initialization here.
	SBC = data.NewBlockChain()
	Peers = data.NewPeerList(-1, 32) //ID set to flag val, will be changed in Register()
}

// Register ID, download BlockChain, start HeartBeat
func Start(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Running Start")

	//generate keys
		//note: a real mine wouldn't need new keys every time, should have option to read in keys from file
	var err error
	SELF_PRIVATE, err = rsa.GenerateKey(crand.Reader, 2048)
	if err != nil{
		fmt.Println(err)
		return
	} else {
		SELF_PUBLIC = &SELF_PRIVATE.PublicKey
	}

	//add self to peerlist, and download BC
	Register()
	Download()

	go StartHeartBeat()
	go StartTryingNonces()
}

// Display peerList and sbc
func Show(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())
}

// Register to Trusted server
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

	if TRUSTED_SERVER == SELF_ADDR {
		//if this is the first node, just create the blockchain
		fmt.Println("\tAm node1, creating blockchain")

		//create simple / random MPT
		var transactions p1.MerklePatriciaTrie
		var balances p1.MerklePatriciaTrie

		transactions.Initial()
		balances.Initial()

		balances.Insert(data.KeyToString(SELF_PUBLIC), "100") //gives initial node 100 ianCoins to start

		//fmt.Println("Balances", balances.Order_nodes())

		var newBlock p2.Block
		newBlock.Initial(0, "", "", transactions, balances)

		SBC.Insert(newBlock)

	} else {
		fmt.Println("\tGetting blockchain from node1")
		//otherwise, download it from node 1


		//create URL, with params
		baseUrl, err := url.Parse(TRUSTED_SERVER)
		if err != nil {
			fmt.Println(err)
			return
		}

		baseUrl.Path += "upload"

		params := url.Values{}
		params.Add("address", SELF_ADDR)
		params.Add("id", fmt.Sprint(Peers.GetSelfId()))

		pubKey := data.KeyToString(SELF_PUBLIC)
		params.Add("key", pubKey)

		baseUrl.RawQuery = params.Encode()

		resp, err := http.Get(baseUrl.String())

		if err != nil {
			fmt.Println(err)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		//fmt.Println("body: ", string(body))

		var jsonInterface map[string]interface{}

		//converts the json string into an interface
		err = json.Unmarshal(body, &jsonInterface)

		//checks that it worked
		if err != nil {
			fmt.Println(body)
			panic(err)
		}

		//gets blockchain from interface
		bcInterface := jsonInterface["blockchain"]
		bcJson, err := json.Marshal(bcInterface)

		if err != nil {
			fmt.Println("invalid blockchain:",bcInterface)
			return
		}

		//gets blockchain from interface
		peersInterface := jsonInterface["peers"]
		peersJson, err := json.Marshal(peersInterface)

		if err != nil {
			fmt.Println(peersInterface)
			return
		}

		//update if everything is successful

		SBC.UpdateEntireBlockChain(string(bcJson))
		Peers.InjectPeerMapJson(string(peersJson), SELF_ADDR)

	}

	//fmt.Println(SBC.BlockChainToJson())

}

// Upload blockchain to whoever called this method, return jsonStr
	//updated to read address and id from URL parameters
func Upload(w http.ResponseWriter, r *http.Request) {

	//handles adding new node to peerList

	//fmt.Println("url:", r.URL.String())
	query := r.URL.Query()
	address := query.Get("address")
	id := query.Get("id")
	pubKey := query.Get("key")

	if id == "" || address == "" || pubKey==""{
		fmt.Println("invalid address or id from upload parameters")
		return
	} else {
		idInt32, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			fmt.Println("id could not be converted from string to int32")
			return
		} else {
			fmt.Println("adding address: ", address, " and id: ", idInt32, "with peer key")
			Peers.Add(address, int32(idInt32), pubKey)
		}
	}

	//returns blockchain, and peerlist, so nodes can add trusted_node to their peermaps
	blockChainJson, err := SBC.BlockChainToJson()
	if err != nil {
		fmt.Println(err)
		return
	}

	peerListJson, err := Peers.PeerMapToJson()
	if err != nil {
		fmt.Println(err)
		return
	}

	res := ""
	res += `"blockchain": ` + blockChainJson + `,`
	res += `"peers": ` + peerListJson
	fmt.Fprint(w, "{"+res+"}") //note, should we handle error?
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
	err =json.Unmarshal(jsonBody, &hBeat)
	if err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Println("got heartbeat from:", hBeat.Addr)
	//fmt.Println("with key:", hBeat.PublicKey)

	//add the node that we get the heartbeat from, and it's peers
	Peers.Add(hBeat.Addr, hBeat.Id, hBeat.PublicKey)
	Peers.InjectPeerMapJson(hBeat.PeerMapJson, SELF_ADDR)

	//no new block, do nothing else
	if hBeat.IfNewBlock == false {
		return
	}

	//verify heartbeat
	verified := false

	newBlock, err := p2.DecodeFromJson(hBeat.BlockJson)
	if err != nil {
		fmt.Println(hBeat.BlockJson)
		fmt.Println("err2:", err)
		return
	}

	//get all variables needed to compute the hash
	parentHash := newBlock.Header.ParentHash
	nonce := newBlock.Header.Nonce
	transactionsHash := newBlock.Transactions.Root
	concatInfo := parentHash + nonce + transactionsHash

	//use that hash to check the proof of work
	proofOfWork := sha3.Sum256([]byte(concatInfo))
	powString := hex.EncodeToString(proofOfWork[:])

	verified = strings.HasPrefix(powString, NUM_0s) //difficult here, because they both need to agree on the number of 0's

	if verified {
		//add the node that we get the heartbeat from, and it's peers
		Peers.Add(hBeat.Addr, hBeat.Id, hBeat.PublicKey)
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

		body, err := ioutil.ReadAll(resp.Body) // occasionally gives error, look into it
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

//not need to reduce life on hop
func ForwardHeartBeat(heartBeatData data.HeartBeatData) {
	url := "/heartbeat/receive"
	httpType := "application/json"
	hBeatJson, _ := json.Marshal(heartBeatData)

	peerMap := Peers.Copy()

	for keyAddr := range peerMap {

		resp, err := http.Post(keyAddr+url, httpType, bytes.NewBuffer(hBeatJson))

		//not really needed, since the response doesn't matter
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error did not get response in forwardHearBeat: %v\n", err)
		}

		body, err2 := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error could not read response in forwardHearBeat: %v\n", err2)
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
			heartBeatData := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), PeersJson, SELF_ADDR, data.KeyToString(SELF_PUBLIC))

			urlAddress := "/heartbeat/receive"
			httpType := "application/json"
			hBeatJson, _ := json.Marshal(heartBeatData)

			peerMap := Peers.Copy()

			for keyAddr := range peerMap {

				fmt.Println("sent heartbeat to:", keyAddr+urlAddress)
				_, err := http.Post(keyAddr+urlAddress, httpType, bytes.NewBuffer(hBeatJson))
				//not really needed, since the response doesn't matter
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error in response in startTransaction: %v\n", err)
				}

				//body, err := ioutil.ReadAll(resp.Body)
				//if err != nil {
				//	fmt.Println("Got an error in reading the body")
				//	fmt.Println(err)
				//} else {
				//	fmt.Println("\tPrinting the response body: ", string(body))
				//}

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

	//generates mpt. Should be replaced getting transactoins from pool, and adding those to the previous mpt
	transactions := data.GenerateMPT()
	balances := data.GenerateMPT()

	for calculateNonce {

		//get block
		currLatest := SBC.GetLatestBlocks()
		var currHead p2.Block
		if currLatest == nil {
			fmt.Fprintf(os.Stderr, "Error getting latest block in TryingNonces\n")
			return
		} else {
			currHead = currLatest[0]
		}

		//generate nonce
		nonce := data.GenerateNonce()

		//test nonce
		validNonce := false

		concatInfo := currHead.Header.Hash + nonce + transactions.Root + balances.Root
		proofOfWork := sha3.Sum256([]byte(concatInfo))
		powString := hex.EncodeToString(proofOfWork[:])

		validNonce = strings.HasPrefix(powString, NUM_0s)

		if validNonce {
			//fmt.Println("hashing:", currHead.Header.Hash, "\n", nonce, "\n",  mpt.Root)

			newBlock := SBC.GenBlock(transactions, balances, nonce)
			blockJson, _ := p2.EncodeToJSON(newBlock)
			peersJson, _ := Peers.PeerMapToJson()

			hBeat := data.NewHeartBeatData(true, Peers.GetSelfId(), blockJson, peersJson, SELF_ADDR, data.KeyToString(SELF_PUBLIC))

			ForwardHeartBeat(hBeat)
			fmt.Println("found valid nonce")

		}

	}
}

// ---------------- Added for p5 ----------------

func ReceiveTransaction(w http.ResponseWriter, r *http.Request){
	jsonBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error with json in recieveTransaction: %v\n", err)
		return
	}

	tAction, err2 := data.DecodeTransactionFromJson(string(jsonBody))
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "Error decodingJson in recieveTransaction: %v\n", err2)
		return
	}

	//verify transaction
	verified := tAction.VerifyTransaction()

	if verified==false{
		fmt.Fprintf(os.Stderr, "Error With invalid transaction in recieveTransaction\n" )
		return
	}

	//add to pool
	TRANSACTION_POOL = append(TRANSACTION_POOL, tAction)

	//reduce ttl
	tAction.TimeToLive = tAction.TimeToLive-1
	if tAction.TimeToLive <= 0{
		return
	} else {
		//forward to peers
		ForwardTransaction(tAction)
	}

}


func CreateTransaction(w http.ResponseWriter, r *http.Request){
	//what should this look like? Who creates the transaction, do they need to know all keys, or dest, or anything else?
}

func ForwardTransaction(tAction data.Transaction){
	url := "/transaction/receive"
	httpType := "application/json"
	tActionJson, _ := tAction.TransactionToJson()

	peerMap := Peers.Copy()

	for keyAddr := range peerMap {

		resp, err := http.Post(keyAddr+url, httpType, bytes.NewBuffer([]byte(tActionJson)))

		//not really needed, since the response doesn't matter
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error did not get response in forwardHearBeat: %v\n", err)
		}

		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "Error could not read response in forwardHearBeat: %v\n", err)
		} else {
			fmt.Println(string(body))
		}

	}
}