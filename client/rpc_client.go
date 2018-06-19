package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/tendermint/abci/types"
)

type jsonRpcRequest struct {
	Method  string      `json:"method"`  //"method": "broadcast_tx_sync",
	Version string      `json:"jsonrpc"` //"jsonrpc": "2.0",
	Params  interface{} `json:"params"`  //"params": ,
	Id      string      `json:"id"`      //"id": "dontcare"
}

type jsonRpcResponseForDelivery struct {
	Method  string       `json:"method"`  //"method": "broadcast_tx_sync",
	Version string       `json:"jsonrpc"` //"jsonrpc": "2.0",
	Result  deliveryTx   `json:"result"`  //"result": ,
	Id      string       `json:"id"`      //"id": "dontcare"
	Error   *errorStatus `json:"error"`
}

type deliveryTx struct {
	CheckTx    types.ResponseCheckTx   `json:"check_tx"`
	DeliveryTx types.ResponseDeliverTx `json:"deliver_tx"`
	Height     int                     `json:"height"`
	Hash       string                  `json:"hash"`
}

type errorStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type jsonRpcResponseForQuery struct {
	Method  string        `json:"method"`  //"method": "broadcast_tx_sync",
	Version string        `json:"jsonrpc"` //"jsonrpc": "2.0",
	Result  responseQuery `json:"result"`  //"result": ,
	Id      string        `json:"id"`      //"id": "dontcare"
	Error   *errorStatus  `json:"error"`
}

type responseQuery struct {
	Response types.ResponseQuery `json:"response"`
}

func newJsonRpcRequest(method string, js interface{}) jsonRpcRequest {
	jr := jsonRpcRequest{}
	jr.Method = method
	jr.Id = "dontcare"
	jr.Params = js
	jr.Version = "2.0"
	return jr
}

type Tx struct {
	Tx string `json:"tx"`
}

type AbciQuery struct {
	Data string `json:"data"`
	Path string `json:"path"`
}

func RpcBroadcastCommit(deliveryB []byte) (uint32, error) {
	tx := Tx{}
	tx.Tx = base64.StdEncoding.EncodeToString(deliveryB)
	jr := newJsonRpcRequest("broadcast_tx_commit", tx)
	bout, _ := json.Marshal(jr)
	resp, err := http.Post(Conf.AbciDaemon, "text/plain", bytes.NewBuffer(bout))
	if err != nil {
		return CodeTypeClientError, err
	}
	bresp, _ := ioutil.ReadAll(resp.Body)
	jresp := jsonRpcResponseForDelivery{}
	json.Unmarshal(bresp, &jresp)
	if jresp.Error != nil {
		return CodeTypeClientError, errors.New(jresp.Error.Message + ": " + jresp.Error.Data)
	}
	if jresp.Result.CheckTx.Code > CodeTypeOK {
		return jresp.Result.CheckTx.Code, errors.New(jresp.Result.CheckTx.Log)
	}

	if jresp.Result.DeliveryTx.Code > CodeTypeOK || jresp.Result.DeliveryTx.Code < 0 {
		return jresp.Result.DeliveryTx.Code, errors.New(jresp.Result.DeliveryTx.Log)
	}
	return CodeTypeOK, nil
}

func RpcQuery(b []byte) (*types.ResponseQuery, error) {
	aq := AbciQuery{}
	aq.Data = hex.EncodeToString(b)
	jr := newJsonRpcRequest("abci_query", aq)
	bout, _ := json.Marshal(jr)
	resp, err := http.Post(Conf.AbciDaemon, "text/plain", bytes.NewBuffer(bout))
	if err != nil {
		return nil, err
	}
	bresp, _ := ioutil.ReadAll(resp.Body)
	jresp := jsonRpcResponseForQuery{}
	json.Unmarshal(bresp, &jresp)
	return &jresp.Result.Response, nil
}
