package ctrls

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mragiadakos/planetary-blockchain/server/conf"
	"github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

func (pba *PBApplication) validateSpbQuery(sq SpbQuery) (uint32, error) {
	pubk, err := crypto.PubKeyFromBytes(sq.Data.From)
	if err != nil {
		return CodeTypeEncodingError, errors.New("Public key is not correct.")
	}
	b, _ := json.Marshal(sq.Data)
	sig, err := crypto.SignatureFromBytes(sq.Signature)
	if err != nil {
		return CodeTypeEncodingError, errors.New("Signature is not correct.")
	}
	isVerified := pubk.VerifyBytes(b, sig)
	if !isVerified {
		return CodeTypeUnauthorized, errors.New("The signature does not validate the data.")
	}
	now := time.Now().UTC()
	since := now.Sub(sq.Data.Time)
	if since > time.Duration(time.Duration(conf.Conf.WaitingSecondsQuery)*time.Second) {
		return CodeTypeUnauthorized, errors.New("The query passed its time.")
	}
	fromAddr, _ := sq.FromPubKeyAddress()
	if sq.Data.UserAddr != nil {
		_, ok := conf.Conf.GetAuthorizedAddresses()[fromAddr]
		if !ok {
			return CodeTypeUnauthorized, errors.New("You are not authorized to check other user's files.")
		}
	}
	return CodeTypeOK, nil
}

func (pba *PBApplication) validateOtopbQuery(oq OtopbQuery) (uint32, error) {
	_, err := crypto.PubKeyFromBytes(oq.From)
	if err != nil {
		return CodeTypeEncodingError, errors.New("Public key is not correct.")
	}
	return CodeTypeOK, nil
}

func (pba *PBApplication) spbQuery(sq SpbQuery) []byte {
	pubk, _ := crypto.PubKeyFromBytes(sq.Data.From)
	fromAddr := pubk.Address().String()

	qresp := QueryResponse{}

	if sq.Data.UserAddr == nil {
		filesBy := pba.state.db.Get(prefixUserKey(fromAddr))
		files := []string{}
		json.Unmarshal(filesBy, &files)
		if sq.Data.File == nil {
			qresp.Files = files
		} else {
			for _, v := range files {
				if *sq.Data.File == v {
					qresp.Files = []string{v}
					break
				}
			}
		}
	} else {
		filesBy := pba.state.db.Get(prefixUserKey(*sq.Data.UserAddr))
		files := []string{}
		json.Unmarshal(filesBy, &files)
		if sq.Data.File == nil {
			qresp.Files = files
		} else {
			for _, v := range files {
				if *sq.Data.File == v {
					qresp.Files = []string{v}
					break
				}
			}
		}

	}
	b, _ := json.Marshal(qresp)
	return b
}

func (pba *PBApplication) otopbQuery(oq OtopbQuery) []byte {
	pubk, _ := crypto.PubKeyFromBytes(oq.From)
	fromAddr := pubk.Address().String()

	qresp := QueryResponse{}
	filesBy := pba.state.db.Get(prefixUserKey(fromAddr))
	files := []string{}
	json.Unmarshal(filesBy, &files)

	qresp.Files = files
	b, _ := json.Marshal(qresp)
	return b

}

func (pba *PBApplication) Query(qreq types.RequestQuery) types.ResponseQuery {

	if conf.Conf.Blockchain == conf.SPB {
		sq := SpbQuery{}
		json.Unmarshal(qreq.Data, &sq)
		code, err := pba.validateSpbQuery(sq)
		if err != nil {
			return types.ResponseQuery{Code: code, Log: err.Error()}
		}
		bq := pba.spbQuery(sq)
		return types.ResponseQuery{Code: CodeTypeOK, Value: bq}
	} else if conf.Conf.Blockchain == conf.OtoOPB {
		oq := OtopbQuery{}
		json.Unmarshal(qreq.Data, &oq)
		fmt.Println(string(oq.From))
		code, err := pba.validateOtopbQuery(oq)
		if err != nil {
			return types.ResponseQuery{Code: code, Log: err.Error()}
		}
		bq := pba.otopbQuery(oq)
		return types.ResponseQuery{Code: CodeTypeOK, Value: bq}

	}
	qresp := types.ResponseQuery{Code: CodeTypeOK}
	return qresp
}
