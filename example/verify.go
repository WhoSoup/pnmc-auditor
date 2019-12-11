package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"time"
)

// example verification of entry 3c715dba3a70a0cbcc6b10dee73dbcb3f09141f776a63457679db85d821e6780
func main() {
	// "download" the entry data
	chain, _ := hex.DecodeString("843dbee7a49a9b9510d399759fbce24b1f700268c94508085abce352d70ed1f6")
	timestamp := []byte("1576061423")
	content := []byte(`{"Factoshi":"{\"price\":0.0034144444,\"updated_at\":1576061416,\"quote\":\"USD\",\"base\":\"PEG\"}","PNMC":"{\"ticker_symbol\":\"PEG\",\"exchange_price\":\"0.00341280\",\"exchange_price_dateline\":1576061415}"}`)
	pubkey, _ := hex.DecodeString("90a5ad85e62dbc535f98c424429a3ea6e285538231ab1324136403cbdc459ae1")
	signature, _ := hex.DecodeString("7c7b9273681ecac228c6548e341163d972d3e13215b97f9a3ca8f4f4b1c4cd272a00464afec32b2de0f34c95620857f521485594d530720b6afa46b942b05804")

	// build the data for the signature
	data := chain
	data = append(data, timestamp...)
	data = append(data, content...)

	// verify the signature
	fmt.Println(ed25519.Verify(pubkey, data, signature))

	// verify the time
	fmt.Println("Local timezone:", time.Unix(1576061423, 0))
	fmt.Println("UTC", time.Unix(1576061423, 0).UTC())
}
