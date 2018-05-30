package ctrls

import (
	"errors"
	"time"

	"github.com/tendermint/go-crypto"
)

const (
	CodeTypeOK            uint32 = 0
	CodeTypeEncodingError uint32 = 1
	CodeTypeBadNonce      uint32 = 2
	CodeTypeUnauthorized  uint32 = 3
)

type ActionStruct string

const (
	ADD_ACTION    = ActionStruct("add")
	REMOVE_ACTION = ActionStruct("remove")
	SEND_ACTION   = ActionStruct("send")
)

type DeliveryData struct {
	From   []byte  // public key
	To     *[]byte // public key
	Action ActionStruct
	Files  []string
}

type DeliveryResponse struct {
	Signature []byte //hex
	Data      DeliveryData
}

func (dr *DeliveryResponse) FromPubKeyAddress() (string, error) {
	pubkey, err := crypto.PubKeyFromBytes(dr.Data.From)
	if err != nil {
		return "", err
	}
	return pubkey.Address().String(), nil
}

func (dr *DeliveryResponse) ToPubKeyAddress() (string, error) {
	if dr.Data.To == nil {
		return "", errors.New("The public key of the receiver is empty.")
	}
	pubkey, err := crypto.PubKeyFromBytes(*dr.Data.To)
	if err != nil {
		return "", err
	}
	return pubkey.Address().String(), nil
}

type SpbQueryData struct {
	From     []byte
	Nonce    string
	Time     time.Time
	File     *string
	UserAddr *string
}

func (sq *SpbQuery) FromPubKeyAddress() (string, error) {
	pubkey, err := crypto.PubKeyFromBytes(sq.Data.From)
	if err != nil {
		return "", err
	}
	return pubkey.Address().String(), nil
}

type SpbQuery struct {
	Signature []byte
	Data      SpbQueryData
}

type OtopbQuery struct {
	From []byte
}

type QueryResponse struct {
	Files []string
}
