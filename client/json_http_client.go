package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/CLIWallet/accounts"
	"github.com/CLIWallet/log"
	"io"
	"net/http"
)

const ServerAddress = "http://localhost:3030"
const Nonce = "nonce"
const Balance = "balance"
const SendTransmission = "trans"

type Requester interface {
	Get(api, data string) []byte
}

type HTTPRequester struct {
	*http.Client
	url string
}




func NewHTTPRequester(url string) *HTTPRequester {
	return &HTTPRequester{ &http.Client{}, url}
}

func (hr *HTTPRequester) Get(api, data string) (map[string]interface{}, error) {
	var jsonStr = []byte(data)
	log.Info("Sending oracle request : %v ", jsonStr)
	req, err := http.NewRequest("POST", hr.url + "/" + api, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := hr.Do(req)

	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buf, resp.Body)

	if err != nil {
		return nil, err
	}

	resp.Body.Close()
	var f interface{}
	json.NewDecoder(buf).Decode(&f)

	return f.(map[string]interface{}), nil

}

// OracleClient is a temporary replacement fot the real oracle. its gets accurate results from a server.
type HttpClient struct {
	Requester
}

func (m HTTPRequester) AccountInfo(id string) (*accounts.AccountInfo, error) {
	str := fmt.Sprintf(`{ "address": %s}`, id)
	output, err := m.Get(Nonce, str)
	if err != nil {
		return nil, fmt.Errorf("cant get nonce")
	}

	acc := accounts.AccountInfo{}
	if val , ok := output["value"]; ok {
		acc.Nonce = val.(string)
	}

	output, err = m.Get(Balance, str)
	if err != nil {
		return nil, fmt.Errorf("cant get nonce")
	}

	if val , ok := output["value"]; ok {
		acc.Balance = val.(string)
	}

	return &acc, nil
}

func (m HTTPRequester) Transfer(from, to, amount, nonce, passphrase string) error {
	str := fmt.Sprintf(`{ "sender": %s, receiver: %s, nonce: %s, amount:%s}`,from, to,nonce,amount )
	_, err := m.Get(SendTransmission, str)
	if err != nil {
		return err
	}
	return nil
}

