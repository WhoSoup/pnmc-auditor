package main

import (
	"io/ioutil"
	"net/http"

	"github.com/cenkalti/backoff/v3"
)

const pnmcURL = "https://pegnetmarketcap.com/api/asset/PEG?columns=ticker_symbol,exchange_price,exchange_price_dateline"
const factoshiURL = "https://pegapi.factoshi.io/"

func FetchAPIResponse(url string) (string, error) {
	var body string

	call := func() error {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		bodyT, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		body = string(bodyT)
		return nil
	}

	if err := backoff.Retry(call, BackOff()); err != nil {
		return "", err
	}
	return body, nil
}
