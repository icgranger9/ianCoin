package main

import (
	"./p3"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func main() {
	//Note, requires as args the port number, and the id.
	//for simplicity, I usually use port:99XX, and id: XX, just to keep them straight
	//starts with node1, with Id 01. All other nodes will connect to this on start to get the blockchain

	//creates a seed for our random numbers
	//not positive this is where it should go
	rand.Seed(time.Now().UnixNano())

	fmt.Println("running main")

	router := p3.NewRouter()
	if len(os.Args) > 1 {
		log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	} else {
		log.Fatal(http.ListenAndServe(":6686", router))
	}
}
