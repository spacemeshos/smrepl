package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spacemeshos/CLIWallet/accounts"
	"github.com/spacemeshos/CLIWallet/log"
	"io"
	"net/http"
	"strconv"
)

const ServerAddress = "http://localhost:9090/v1"
const Nonce = "nonce"
const Balance = "balance"
const SendTransmission = "submittransaction"

type Requester interface {
	Get(api, data string) []byte
}

type HTTPRequester struct {
	*http.Client
	url string
}

func NewHTTPRequester(url string) *HTTPRequester {
	return &HTTPRequester{&http.Client{}, url}
}

func (hr *HTTPRequester) Get(api, data string) (map[string]interface{}, error) {
	var jsonStr = []byte(data)
	url := hr.url + "/" + api
	log.Info("Sending to url: %v request : %v ", url, string(jsonStr))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
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
	log.Info("resp: %v len %v", buf.String(), buf.Len())
	resp.Body.Close()
	var f interface{}
	err = json.NewDecoder(buf).Decode(&f)
	if err != nil {
		return nil, err
	}

	return f.(map[string]interface{}), nil

}

type HttpClient struct {
	Requester
}

func (hr *HTTPRequester) NodeURL() string {
	return hr.url
}

func printBuffer(b []byte) string {
	st := "["
	for _, byt := range b {
		st += strconv.Itoa(int(byt)) + ","
	}
	st = st[:len(st)-1] + "]"
	return st
}

func (m HTTPRequester) AccountInfo(id string) (*accounts.AccountInfo, error) {
	str := fmt.Sprintf(`{ "address": "0x%s"}`, id)
	output, err := m.Get(Nonce, str)
	if err != nil {
		return nil, fmt.Errorf("cant get nonce %v", err)
	}

	acc := accounts.AccountInfo{}
	if val, ok := output["value"]; ok {
		acc.Nonce = val.(string)
	} else {
		return nil, fmt.Errorf("cant get nonce %v", output)
	}

	output, err = m.Get(Balance, str)
	if err != nil {
		return nil, fmt.Errorf("cant get balance: %v", err)
	}

	if val, ok := output["value"]; ok {
		acc.Balance = val.(string)
	} else {
		return nil, fmt.Errorf("cant get nonce %v", output)
	}

	return &acc, nil
}

func (m HTTPRequester) Send(b []byte) error {
	str := fmt.Sprintf(`{ "tx": %s}`, printBuffer(b))
	_, err := m.Get(SendTransmission, str)
	if err != nil {
		return err
	}
	return nil
}
