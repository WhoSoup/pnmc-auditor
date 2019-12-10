package main

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/cenkalti/backoff/v3"
	"gopkg.in/ini.v1"
)

var cli *factom.Client
var paying factom.EsAddress
var signing factom.FsAddress
var chainID factom.Bytes32

func BackOff() *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     500 * time.Millisecond,
		RandomizationFactor: .5,
		Multiplier:          1.5,
		MaxInterval:         time.Second,
		MaxElapsedTime:      10 * time.Second,
		Clock:               backoff.SystemClock,
	}
	b.Reset()
	return b
}

func WriteEntry(content []byte) error {
	e := new(factom.Entry)
	e.ChainID = &chainID

	now := fmt.Sprintf("%d", time.Now().Unix())
	e.Content = content
	e.ExtIDs = append(e.ExtIDs, []byte(now))
	e.ExtIDs = append(e.ExtIDs, []byte(signing.PublicKey()))
	e.ExtIDs = append(e.ExtIDs, Signature(e))

	operation := func() error {

		b32, err := e.ComposeCreate(nil, cli, paying)
		if err != nil {
			return err
		}
		log.Println("Wrote Entry:", b32)
		return nil
	}

	err := backoff.Retry(operation, BackOff())
	return err
}

func Signature(e *factom.Entry) []byte {
	data := make([]byte, 0)
	data = append(data, chainID[:]...)
	data = append(data, e.ExtIDs[0]...)
	data = append(data, e.Content...)

	sig := ed25519.Sign(signing.PrivateKey(), data)

	return sig
}

func GetResponses() []byte {
	responses := make(map[string]string)

	if apiResponse, err := FetchAPIResponse(pnmcURL); err != nil {
		log.Println("unable to fetch PNMC:", err)
		responses["PNMC"] = "unable to fetch rates"
	} else {
		responses["PNMC"] = apiResponse
	}

	if apiResponse, err := FetchAPIResponse(factoshiURL); err != nil {
		log.Println("unable to fetch Factoshi:", err)
		responses["Factoshi"] = "unable to fetch rates"
	} else {
		responses["Factoshi"] = apiResponse
	}

	content, err := json.Marshal(responses)
	if err != nil {
		log.Println("error marshalling responses:", err)
	}

	return content
}

func Audit() {
	responses := GetResponses()
	if err := WriteEntry(responses); err != nil {
		log.Println("failed to write entry:", err)
	}
}

func main() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatal(err)
	}

	audit := cfg.Section("audit")

	cfgInterval, err := audit.Key("interval").Int()
	if err != nil {
		log.Fatalf("unable to convert config[audit.interval] to int: %v", err)
	}

	cfgPaying := audit.Key("paying").String()
	if paying, err = factom.NewEsAddress(cfgPaying); err != nil {
		log.Fatalf("unable to convert config[audit.paying] to address: %v", err)
	}

	cfgSigning := audit.Key("signing").String()
	if signing, err = factom.NewFsAddress(cfgSigning); err != nil {
		log.Fatalf("unable to convert config[audit.signing] to address: %v", err)
	}

	cfgChain := audit.Key("chain").String()
	if chainID = factom.NewBytes32(cfgChain); chainID.IsZero() {
		log.Fatalf("unable to convert config[audit.chain] to chain: %v", err)
	}

	cli = factom.NewClient()
	cli.FactomdServer = audit.Key("factomd").String()

	// audit loop
	Audit()
	ticker := time.NewTicker(time.Second * time.Duration(cfgInterval))
	for range ticker.C {
		Audit()
	}
}
