package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
	crypto "github.com/tendermint/go-crypto"
)

func query(b []byte) (*QueryResponse, uint32, error) {

	resp, err := RpcQuery(b)
	if err != nil {
		return nil, CodeTypeClientError, err
	}

	if resp.Code > CodeTypeOK {
		return nil, resp.Code, errors.New(resp.Log)
	}
	qresp := QueryResponse{}
	json.Unmarshal(resp.Value, &qresp)
	return &qresp, CodeTypeOK, nil
}

func ipfsAddJson(block []byte) (string, error) {
	sh := shell.NewShell(Conf.IpfsConnection)
	return sh.BlockPut(block)
}

func ipfsAddFile(filename string) (string, error) {
	sh := shell.NewShell(Conf.IpfsConnection)
	file, err := os.Open(filename)
	if err != nil {
		return "", errors.New("Could not read file: " + err.Error())
	}
	return sh.Add(file)
}

func ipfsAddFolder(folder string) (string, error) {
	sh := shell.NewShell(Conf.IpfsConnection)
	return sh.AddDir(folder)
}

func AddRequest(from crypto.PrivKeyEd25519, fileHashes []string) (uint32, error) {
	dd := DeliveryData{}
	dd.From = from.PubKey().Bytes()
	dd.Action = ADD_ACTION
	dd.Files = fileHashes
	b, _ := json.Marshal(dd)
	dr := DeliveryRequest{}
	dr.Signature = from.Sign(b).Bytes()
	dr.Data = dd
	b, _ = json.Marshal(dr)
	return RpcBroadcastCommit(b)
}

func RemoveRequest(from crypto.PrivKeyEd25519, fileHashes []string) (uint32, error) {
	dd := DeliveryData{}
	dd.From = from.PubKey().Bytes()
	dd.Action = REMOVE_ACTION
	dd.Files = fileHashes
	b, _ := json.Marshal(dd)
	dr := DeliveryRequest{}
	dr.Signature = from.Sign(b).Bytes()
	dr.Data = dd
	b, _ = json.Marshal(dr)
	return RpcBroadcastCommit(b)
}

func SendRequest(from crypto.PrivKeyEd25519, toPublicKey []byte, fileHashes []string) (uint32, error) {
	dd := DeliveryData{}
	dd.From = from.PubKey().Bytes()
	dd.Action = SEND_ACTION
	dd.To = &toPublicKey
	dd.Files = fileHashes
	b, _ := json.Marshal(dd)
	dr := DeliveryRequest{}
	dr.Signature = from.Sign(b).Bytes()
	dr.Data = dd
	b, _ = json.Marshal(dr)
	return RpcBroadcastCommit(b)
}

func SpbQueryRequest(from crypto.PrivKeyEd25519, file, userAddr *string) (*QueryResponse, uint32, error) {
	q := SpbQuery{}
	data := SpbQueryData{}
	data.From = from.PubKey().Bytes()
	if file != nil {
		data.File = file
	}
	if userAddr != nil {
		data.UserAddr = userAddr
	}
	data.Time = time.Now().UTC()
	b, _ := json.Marshal(data)

	q.Data = data
	q.Signature = from.Sign(b).Bytes()
	b, _ = json.Marshal(q)
	return query(b)
}

func OtopbQueryRequest(from crypto.PrivKeyEd25519) (*QueryResponse, uint32, error) {
	q := OtopbQuery{
		From: from.PubKey().Bytes(),
	}
	b, _ := json.Marshal(q)
	return query(b)
}

func fileKey(filename string) (*crypto.PrivKeyEd25519, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.New("Error: " + err.Error())
	}

	jk := JsonKey{}
	err = json.Unmarshal(b, &jk)
	if err != nil {
		return nil, errors.New("Error: json problem with the key " + err.Error())
	}

	bKey, err := hex.DecodeString(jk.PrivateKey)
	if err != nil {
		return nil, errors.New("Error: hex decoding problem with the key " + err.Error())
	}

	edKey, err := crypto.PrivKeyFromBytes(bKey)
	if err != nil {
		return nil, errors.New("Error: private key decoding problem with the key " + err.Error())
	}
	key := edKey.(crypto.PrivKeyEd25519)
	return &key, nil
}
