package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Factom-Asset-Tokens/factom"
)

type Entry struct {
	Timestamp time.Time
	Factoshi  float64
	PNMC      float64
}

var PNMCKey []byte
var Chain []byte

func init() {
	PNMCKey, _ = hex.DecodeString("90a5ad85e62dbc535f98c424429a3ea6e285538231ab1324136403cbdc459ae1")
	Chain, _ = hex.DecodeString("843dbee7a49a9b9510d399759fbce24b1f700268c94508085abce352d70ed1f6")
}

func ParseEntry(rawtime, raw []byte) (Entry, error) {
	i, err := strconv.ParseInt(string(rawtime), 10, 64)
	if err != nil {
		return Entry{}, err
	}
	timestamp := time.Unix(i, 0)

	var wrapper map[string]string
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return Entry{}, err
	}

	var pricePNMC, priceFactoshi float64
	if b, ok := wrapper["PNMC"]; ok && string(b) != "unable to fetch rates" {
		var PNMC map[string]interface{}
		if err := json.Unmarshal([]byte(wrapper["PNMC"]), &PNMC); err != nil {
			return Entry{}, err
		}
		// entry is signed by pnmc, so data is expected
		if price, err := strconv.ParseFloat(PNMC["exchange_price"].(string), 64); err != nil {
			fmt.Println(PNMC["exchange_price"].(string))
			return Entry{}, err
		} else {
			pricePNMC = price
		}
	}

	if b, ok := wrapper["Factoshi"]; ok && string(b) != "unable to fetch rates" {
		var Factoshi map[string]interface{}
		if err := json.Unmarshal([]byte(wrapper["Factoshi"]), &Factoshi); err != nil {
			return Entry{}, err
		}
		if price, ok := Factoshi["price"].(float64); ok {
			priceFactoshi = price
		} else {
			return Entry{}, errors.New("unable to convert factoshi price to float")
		}
	}

	return Entry{
		Timestamp: timestamp,
		Factoshi:  priceFactoshi,
		PNMC:      pricePNMC,
	}, nil
}

func Verify(e factom.Entry) error {
	if !bytes.Equal(PNMCKey, e.ExtIDs[1]) {
		return errors.New("unrecognized public key")
	}

	data := Chain
	data = append(data, e.ExtIDs[0]...)
	data = append(data, e.Content...)

	if !ed25519.Verify(PNMCKey, data, e.ExtIDs[2]) {
		return errors.New("signature did not match")
	}
	return nil
}

func main() {
	chain := factom.NewBytes32("843dbee7a49a9b9510d399759fbce24b1f700268c94508085abce352d70ed1f6")
	cli := factom.NewClient()
	cli.FactomdServer = "https://api.factomd.net/v2"

	head := new(factom.EBlock)
	head.ChainID = &chain
	if _, err := head.GetChainHead(nil, cli); err != nil {
		log.Fatal(err)
	}

	eblocks, err := head.GetPrevAll(nil, cli)
	if err != nil {
		log.Fatal(err)
	}

	var data []Entry

	for i := len(eblocks) - 1; i > 0; i-- {
		eb := eblocks[i]
		if err := eb.GetEntries(nil, cli); err != nil {
			log.Println("unable to get entries:", err)
			continue
		}

		for _, e := range eb.Entries {
			if err := Verify(e); err != nil {
				// the first entry created the chain and does not contain data
				if e.Hash.String() != "d89c60e7fb6efb91bb722d459a2a2392f5b6849ba9079b253e969220d796c66a" {
					fmt.Println("unable to verify", e.Hash, "=", err)
				}
				continue
			}

			d, err := ParseEntry(e.ExtIDs[0], e.Content)
			if err != nil {
				fmt.Println("unable to parse", e.Hash, "=", err)
				continue
			}
			data = append(data, d)
		}
	}

	fmt.Println("Time,PNMC,Factoshi")
	for _, d := range data {
		fmt.Printf("%s,%1.8f,%1.8f\n", d.Timestamp.UTC(), d.PNMC, d.Factoshi)
	}
}
