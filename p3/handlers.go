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

//server variables
var TRUSTED_SERVER = "http://localhost:9901"
var SELF_PUBLIC *rsa.PublicKey
var SELF_PRIVATE *rsa.PrivateKey
var SELF_ADDR = "http://localhost:" + os.Args[1]

//variables used in block creation
var NUM_0s = "000000" //best is 6
var BLOCK_REWARD float64 = 250

//blockChain variables
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
		fmt.Fprintf(os.Stderr, "Error in generating key in Start: %v\n", err)
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

		balances.Insert(data.KeyToString(SELF_PUBLIC), fmt.Sprint(BLOCK_REWARD)) //gives initial node a block reward, so we start with some money in the system

		var newBlock p2.Block
		newBlock.Initial(0, "", "", balances, transactions)

		SBC.Insert(newBlock)

	} else {
		fmt.Println("\tGetting blockchain from node1")
		//otherwise, download it from node 1


		//create URL, with params
		baseUrl, err := url.Parse(TRUSTED_SERVER)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing url in  Download: %v\n", err)
			return
		}

		baseUrl.Path += "upload"

		params := url.Values{}
		params.Add("address", SELF_ADDR)
		params.Add("id", fmt.Sprint(Peers.GetSelfId()))

		pubKey := data.KeyToString(SELF_PUBLIC)
		params.Add("key", pubKey)

		baseUrl.RawQuery = params.Encode()

		resp, err2 := http.Get(baseUrl.String())

		if err2 != nil {
			fmt.Fprintf(os.Stderr, "Error converting URL to string in Download: %v\n", err2)

			return
		}

		body, err3 := ioutil.ReadAll(resp.Body)
		if err3 != nil {
			fmt.Fprintf(os.Stderr, "Error in reading resp body in Download: %v\n", err3)
			return
		}

		var jsonInterface map[string]interface{}

		//converts the json string into an interface
		err4 := json.Unmarshal(body, &jsonInterface)

		//checks that it worked
		if err4 != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshaling in Download: %v\n", err4)
			return
		}

		//gets blockchain from interface
		bcInterface := jsonInterface["blockchain"]
		bcJson, err5 := json.Marshal(bcInterface)

		if err5 != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshaling in Download: %v\n", err5)
			return
		}

		//gets blockchain from interface
		peersInterface := jsonInterface["peers"]
		peersJson, err6 := json.Marshal(peersInterface)

		if err6 != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshaling in Download: %v\n", err6)
			return
		}

		//update if everything is successful

		SBC.UpdateEntireBlockChain(string(bcJson))
		Peers.InjectPeerMapJson(string(peersJson), SELF_ADDR)

	}

}

// Upload blockchain to whoever called this method, return jsonStr
	//updated to read address and id from URL parameters
func Upload(w http.ResponseWriter, r *http.Request) {

	//handles adding new node to peerList

	query := r.URL.Query()
	address := query.Get("address")
	id := query.Get("id")
	pubKey := query.Get("key")

	if id == "" || address == "" || pubKey==""{
		fmt.Fprintf(os.Stderr, "Error invalid address or id recieved in Upload \n")
		return
	} else {
		idInt32, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error id could not be converted in Upload: %v\n", err)
			return
		} else {
			fmt.Println("Adding address: ", address, " and id: ", idInt32, "to peers")
			Peers.Add(address, int32(idInt32), pubKey)
		}
	}

	//returns blockchain, and peerlist, so nodes can add trusted_node to their peermaps
	blockChainJson, err2 := SBC.BlockChainToJson()
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "Error bc converted in Upload: %v\n", err2)
		return
	}

	peerListJson, err3 := Peers.PeerMapToJson()
	if err3 != nil {
		fmt.Fprintf(os.Stderr, "Error PeerList could not be converted in Upload: %v\n", err3)
		return
	}

	res := ""
	res += `"blockchain": ` + blockChainJson + `,`
	res += `"peers": ` + peerListJson
	fmt.Fprint(w, "{"+res+"}") //note, should we handle error?
}

// Upload a block to whoever called this method, return jsonStr
func UploadBlock(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Uploading Block")
	uri := r.RequestURI
	uriSplit := strings.Split(uri, "/")

	//note: uriSplit is shifted one because uri begins with /, so uriSplit[0] == ""

	height, err := strconv.ParseInt(uriSplit[2], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Id could not be converted in UploadBlock: %v\n", err)
		return
	}

	block, found := SBC.GetBlock(int32(height), uriSplit[3])

	if found {
		blockJson, _ := p2.EncodeToJSON(block)
		fmt.Println("Found block:", blockJson)
		fmt.Fprint(w, blockJson)
	} else {
		fmt.Fprintf(os.Stderr, "Could not find requested block\n")
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
	err2 :=json.Unmarshal(jsonBody, &hBeat)
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "Error hBeat could not be unmarshaled in hBeatReceive: %v\n", err2)
		return
	}

	fmt.Println("Got heartbeat from:", hBeat.Addr)

	//add the node that we get the heartbeat from, and it's peers
	Peers.Add(hBeat.Addr, hBeat.Id, hBeat.PublicKey)
	Peers.InjectPeerMapJson(hBeat.PeerMapJson, SELF_ADDR)

	//no new block, do nothing else
	if hBeat.IfNewBlock == false {
		return
	} else {
		fmt.Println("Received new block")
	}

	//verify heartbeat
	verified := false

	newBlock, err3 := p2.DecodeFromJson(hBeat.BlockJson)
	if err3 != nil {
		fmt.Fprintf(os.Stderr, "Error newBlock could not be decoded in hBeatReceive: %v\n", err3)
		return
	}

	//get all variables needed to compute the hash
	parentHash := newBlock.Header.ParentHash
	nonce := newBlock.Header.Nonce
	transactionsRoot := newBlock.Transactions.Root
	balancesRoot := newBlock.Balances.Root
	concatInfo := parentHash + nonce +  balancesRoot + transactionsRoot

	//use that hash to check the proof of work
	proofOfWork := sha3.Sum256([]byte(concatInfo))
	powString := hex.EncodeToString(proofOfWork[:])

	verified = strings.HasPrefix(powString, NUM_0s) //difficult here, because they both need to agree on the number of 0's

	if verified {

		//add the node that we get the heartbeat from, and it's peers
		Peers.Add(hBeat.Addr, hBeat.Id, hBeat.PublicKey)
		Peers.InjectPeerMapJson(hBeat.PeerMapJson, SELF_ADDR)

		//1
		parentExists := SBC.CheckParentHash(newBlock)
		if parentExists == false {
			fmt.Println("Parent doesn't exist")
			AskForBlock(newBlock.Header.Height-1, newBlock.Header.ParentHash)
		}

		//2
		SBC.Insert(newBlock)

		//3
		hBeat.Hops = hBeat.Hops - 1
		if hBeat.Hops > 0 {
			ForwardHeartBeat(hBeat)
		}

		//write simple response
		w.Write([]byte("\tBlock received"))

		fmt.Println("Finished Receive heart beat")
	} else {
		fmt.Fprintln(os.Stderr, "Received invalid nonce in HeartBeatReceive")
	}

}

// Ask another server to return a block of certain height and hash
func AskForBlock(height int32, hash string) {
	fmt.Println("Asking for block")

	urlVar := "/block/" + fmt.Sprint(height) + "/" + hash

	peerMap := Peers.Copy()

	for keyAddr := range peerMap {

		resp, err := http.Get(keyAddr + urlVar)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error in get request in AskForBlock: %v\n", err)
			return
		}

		body, err2 := ioutil.ReadAll(resp.Body) // occasionally gives error, look into it
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "Error in body in AskForBlock: %v\n", err2)
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
		} else {

			body, err2 := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error could not read response in forwardHearBeat: %v\n", err2)
			} else {
				fmt.Println("\tForwarded hBeat to ", keyAddr, " response is: ", string(body))
			}
		}

	}
}

func StartHeartBeat() {

	ifStarted = true

	for ifStarted {

		timeToSleep := rand.Intn(5) + 5
		time.Sleep(time.Duration(timeToSleep) * time.Second)

		PeersJson, err := Peers.PeerMapToJson()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting peers in StartHeartBeat: %v\n", err)
		} else {
			heartBeatData := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), PeersJson, SELF_ADDR, data.KeyToString(SELF_PUBLIC))

			urlAddress := "/heartbeat/receive"
			httpType := "application/json"
			hBeatJson, _ := json.Marshal(heartBeatData)

			peerMap := Peers.Copy()

			for keyAddr := range peerMap {

				fmt.Println("Sent heartbeat to:", keyAddr+urlAddress)
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
	fmt.Println("In canonical")

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

	//get latest block
	currLatest := SBC.GetLatestBlocks()
	var currHead p2.Block
	if currLatest == nil {
		fmt.Fprintf(os.Stderr, "Error getting latest block in StartTryingNonces\n")
		return
	} else {
		currHead = currLatest[0]
	}

	//generates mpt, based on that latest block
	balances, transactions, err:= GenerateNextMpt(currHead)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error could not get mpts in StartTryingNonces: %v\n", err)
		return //is this really what we should do?
	}


	for calculateNonce {
		//Note: should switch to latest block if a new one is received


		//generate nonce
		nonce := data.GenerateNonce()

		//test nonce
		validNonce := false

		concatInfo := currHead.Header.Hash + nonce + balances.Root + transactions.Root
		proofOfWork := sha3.Sum256([]byte(concatInfo))
		powString := hex.EncodeToString(proofOfWork[:])

		validNonce = strings.HasPrefix(powString, NUM_0s)

		if validNonce {

			newBlock := SBC.GenBlock(balances, transactions, nonce)
			blockJson, _ := p2.EncodeToJSON(newBlock)
			peersJson, _ := Peers.PeerMapToJson()

			hBeat := data.NewHeartBeatData(true, Peers.GetSelfId(), blockJson, peersJson, SELF_ADDR, data.KeyToString(SELF_PUBLIC))

			ForwardHeartBeat(hBeat)

			//should we add the new block to our own BC?
				//Am I positive this is the way to do it?
			SBC.Insert(newBlock)

			//generates mpt, based on that latest block
			currHead = SBC.GetLatestBlocks()[0] //doesn't directly set it to the block we created, in case someone else made a block faster
			balances, transactions, err = GenerateNextMpt(newBlock)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error could not get mpts a second time in StartTryingNonces: %v\n", err)
				return //is this really what we should do?
			}
		}

	}
}

// ---------------- Added for p5 ----------------

func ReceiveTransaction(w http.ResponseWriter, r *http.Request){
	jsonBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error with json in recieveTransaction: %v\n", err)
		w.Write([]byte("Received invalid transaction"))
		return
	}

	tx, err2 := data.DecodeTransactionFromJson(string(jsonBody))
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "Error decodingJson in recieveTransaction: %v\n", err2)
		w.Write([]byte("Received invalid transaction"))
		return
	}

	//verify transaction
	verified := tx.VerifyTransaction()

	if verified==false{
		fmt.Fprintf(os.Stderr, "Error With invalid transaction in recieveTransaction\n" )
		w.Write([]byte("Received invalid transaction"))
		return
	}

	//Checks if TX is in pool, no point having repeats
	var inPool = false
	for _, txInPool:= range TRANSACTION_POOL{
		if txInPool.HashTransaction() == tx.HashTransaction() {
			//fmt.Fprintf(os.Stderr, "Transaction already in pool in recieveTransaction\n" )
			inPool = true
			break
		}
	}

	//add to pool
	if !inPool {
		TRANSACTION_POOL = append(TRANSACTION_POOL, tx)
	}

	//reduce ttl
	tx.TimeToLive = tx.TimeToLive-1
	if tx.TimeToLive <= 0{
		//do nothing
	} else {
		//forward to peers
		ForwardTransaction(tx)
	}

	//write response
	w.Write([]byte("Received valid transaction"))

}

func ForwardTransaction(tx data.Transaction){
	url := "/transaction/receive"
	httpType := "application/json"
	txJson, _ := tx.TransactionToJson()

	peerMap := Peers.Copy()

	for keyAddr := range peerMap {

		resp, err := http.Post(keyAddr+url, httpType, bytes.NewBuffer([]byte(txJson)))

		//not really needed, since the response doesn't matter
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error did not get response in forwardHearBeat: %v\n", err)
		}

		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "Error could not read response in forwardHearBeat: %v\n", err)
		} else {
			fmt.Println("Forwarded transaction to ", keyAddr ," response is: ", string(body))
		}

	}
}

func GenerateNextMpt(currHead p2.Block) (p1.MerklePatriciaTrie, p1.MerklePatriciaTrie, error){

	fmt.Println("Generating next mpt")

	//helper function to generate next MPTs based off of a given block


	for len(TRANSACTION_POOL) == 0 {
		//busy wait while there are no new transactions. What else can we do? maybe make a new transaction?
			//if there are no transactions in the pool, send 1 ianCoin to all peers.
			// This obviously would not happen in a real world scenario, and is just done to increase simulated traffic

			peerKeys := Peers.CopyKeys()

			for _, peerKey := range peerKeys {
				CreateTransaction(peerKey, 1)
			}

			//sleep for 5 sec, to allow those transactions to be sent
			time.Sleep(time.Duration(2) * time.Second)

	}

	//get previous mpt's
	transactions := currHead.Transactions
	balances := currHead.Balances

	//fmt.Println("Starting MPTs:","\n\tBalances:", balances.Order_nodes(), "\n\tTransactions:", transactions.Order_nodes())

	//fmt.Println("Pool:")
	//for ind, tx := range TRANSACTION_POOL {
	//	fmt.Println("\t",ind+1," tx's sig is:", tx.Signature)
	//}


	//add some transactions
		//Note: How many should we add? Just going to default to 1/2 of pool
	var feeTotal float64
	numTXsToAdd := len(TRANSACTION_POOL)/2.0 + 1
	//fmt.Println("Starting pool len:", len(TRANSACTION_POOL))
	for x:=0; x < numTXsToAdd; x++ {

		//get front of pool (pop)
		tx, tmpPool := TRANSACTION_POOL[0], TRANSACTION_POOL[1:] //pops first transaction, Note: must double check that pop actually works
		TRANSACTION_POOL = tmpPool //updates pool

		//fmt.Println("Popped pool len:", len(TRANSACTION_POOL))

		//fmt.Println("Checking if pool was updated:", "\n\tActual Pool:", TRANSACTION_POOL, "\n\t   Tmp Pool:", tmpPool)

		//handle adding to mpt's
		resBalances, resTXs, err := tx.AddTransactionToMPTs(balances, transactions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error could not add transaction in GenerateNextMpt: %v\n", err)
			//break //is this the best response?
		} else {
			feeTotal += tx.Fee
			transactions = resTXs
			balances = resBalances
		}

		//what should we do if the tx can't go through? drop, or push to back of the pool?
			//currently just drop it
	}

	//fmt.Println("\nDone adding tx's:","\n\tBalances:", balances.Order_nodes(), "\n\tTransactions:", transactions.Order_nodes())


	//add x coins to self balance
		//Note: How can we do this well? Does it need a signature, or any verification?

	currBal, err2 := balances.Get(data.KeyToString(SELF_PUBLIC))
	if err2 != nil && err2.Error() == "reached_invalid_leaf" {
		//if we don't have a balance, need to create one! Duh
		balances.Insert(data.KeyToString(SELF_PUBLIC), fmt.Sprint(feeTotal+BLOCK_REWARD))

	} else if err2 != nil{
		fmt.Fprintf(os.Stderr, "Error could not get ballance of self in GenerateNextMpt: \"%v\"\n", err2)

		return balances, transactions, err2
	} else{
		balFloat, err3 := strconv.ParseFloat(currBal, 64)
		if err3 != nil{
			fmt.Fprintf(os.Stderr, "Error could not convert ballance of self in GenerateNextMpt: %v\n", err3)
			return balances, transactions, err3
		} else {
			//only update value if no previous errors
			//fmt.Println("New bal:", strconv.FormatFloat(balFloat+feeTotal+BLOCK_REWARD, 'f', -1, 64))
			//fmt.Println("bal:", balFloat,"Fee:",feeTotal,"Reward:",BLOCK_REWARD)
			balances.Insert(data.KeyToString(SELF_PUBLIC), strconv.FormatFloat(balFloat+feeTotal+BLOCK_REWARD, 'f', -1, 64)) //insert will update value, right? Must double check
		}
	}

	//fmt.Println("\nAt end of generate:","\n\tBalances:", balances.Order_nodes(), "\n\tTransactions:", transactions.Order_nodes())

	return balances, transactions, nil
}

func CreateTransaction(dest *rsa.PublicKey, amount float64){
	//note: should this also add the transaction to our own pool?

	//convert dst to string
	dstStr := data.KeyToString(dest)

	tx := data.NewTransaction(data.KeyToString(SELF_PUBLIC), dstStr, amount, amount*.05, time.Now().UnixNano())
	err := tx.SignTransaction(SELF_PRIVATE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error could not sign tx in CreateTransaction: %v\n", err)
		return
	}

	urlAddress := "/transaction/receive"
	httpType := "application/json"
	txJson, err2 := tx.TransactionToJson()

	if err2 != nil {
		fmt.Fprintf(os.Stderr, "Error could not Marshal TX in CreateTransaction: %v\n", err2)
		return
	}

	peerMap := Peers.Copy()

	for keyAddr := range peerMap {

		fmt.Println("Sent tx to:", keyAddr+urlAddress)
		resp, err3 := http.Post(keyAddr+urlAddress, httpType, bytes.NewBuffer([]byte(txJson)))
		//not really needed, since the response doesn't matter
		if err3 != nil {
			fmt.Fprintf(os.Stderr, "Error in response in startTransaction: %v\n", err3)
		}

		body, err4 := ioutil.ReadAll(resp.Body)
		if err4 != nil {
			fmt.Fprintf(os.Stderr, "Error in reading body startTransaction: %v\n", err4)
		} else {
			fmt.Println("\tPrinting the response in CreateTransaction:", string(body))
		}

	}

}
