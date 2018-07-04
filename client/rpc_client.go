package main

import (
	"errors"

	client "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"
)

type AbciQuery struct {
	Data string `json:"data"`
	Path string `json:"path"`
}

func RpcBroadcastCommit(deliveryB []byte) (uint32, error) {
	cli := client.NewHTTP(Conf.AbciDaemon, "/websocket")
	btc, err := cli.BroadcastTxCommit(types.Tx(deliveryB))
	if err != nil {
		return CodeTypeClientError, err
	}

	if btc.CheckTx.Code > CodeTypeOK {
		return btc.CheckTx.Code, errors.New(btc.CheckTx.Log)
	}

	if btc.DeliverTx.Code > CodeTypeOK || btc.DeliverTx.Code < 0 {
		return btc.DeliverTx.Code, errors.New(btc.DeliverTx.Log)
	}
	return CodeTypeOK, nil
}

func RpcQuery(b []byte) ([]byte, uint32, error) {
	cli := client.NewHTTP(Conf.AbciDaemon, "/websocket")
	q, err := cli.ABCIQuery("", b)
	if err != nil {
		return nil, CodeTypeClientError, err
	}
	if q.Response.Code > CodeTypeOK {
		return nil, q.Response.Code, errors.New(q.Response.Log)
	}

	return q.Response.Value, q.Response.Code, nil
}
