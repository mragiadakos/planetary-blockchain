package ctrls

import (
	"encoding/json"
	"errors"

	"github.com/ipfs/go-ipfs-api"
	"github.com/mragiadakos/planetary-blockchain/server/conf"
	"github.com/tendermint/go-crypto"

	"github.com/tendermint/abci/types"
)

func (pba *PBApplication) addActionValidation(dr DeliveryRequest) (uint32, error) {
	if conf.Conf.Blockchain == conf.OtoOPB {
		if len(dr.Data.Files) > 1 {
			return CodeTypeUnauthorized,
				errors.New("For one to one blockchain, you can not add more than one file in one delivery.")
		}
		addr, _ := dr.FromPubKeyAddress()
		if pba.state.db.Has(prefixUserKey(addr)) {
			return CodeTypeUnauthorized,
				errors.New("For one to one blockchain, you can not use the same key.")
		}

	}
	// check if the files already exists
	for _, v := range dr.Data.Files {
		has := pba.state.db.Has(prefixFileKey(v))
		if has {
			return CodeTypeUnauthorized, errors.New("The hash " + v + " already exists.")
		}
	}
	return CodeTypeOK, nil
}

func (pba *PBApplication) sendActionValidation(dr DeliveryRequest) (uint32, error) {
	if dr.Data.To == nil {
		return CodeTypeUnauthorized, errors.New("The public key of the receiver does not exists.")
	}

	to, err := crypto.PubKeyFromBytes(*dr.Data.To)
	if err != nil {
		return CodeTypeEncodingError, errors.New("The public key of the receiver is not correct.")
	}

	from, err := crypto.PubKeyFromBytes(dr.Data.From)
	if err != nil {
		return CodeTypeEncodingError, errors.New("The public key of the sender is not correct.")
	}

	if from == to {
		return CodeTypeUnauthorized, errors.New("The public key of the receiver is the same as the senders.")
	}
	toAddr, _ := dr.ToPubKeyAddress()
	if conf.Conf.Blockchain == conf.OtoOPB {
		if pba.state.db.Has(prefixUserKey(toAddr)) {
			return CodeTypeUnauthorized, errors.New("The public key of the receiver exists in the DB.")
		}
	}

	fromAddr, _ := dr.FromPubKeyAddress()
	for _, v := range dr.Data.Files {
		user := pba.state.db.Get(prefixFileKey(v))
		if string(user) != fromAddr {
			return CodeTypeUnauthorized, errors.New("You dont own The hash " + v + ".")
		}
	}

	return CodeTypeOK, nil
}

func (pba *PBApplication) removeActionValidation(dr DeliveryRequest) (uint32, error) {
	fromAddr, _ := dr.FromPubKeyAddress()
	for _, v := range dr.Data.Files {
		addr := pba.state.db.Get(prefixFileKey(v))
		if len(addr) == 0 {
			return CodeTypeUnauthorized, errors.New("The hash " + v + " doesn't exists.")
		}
		if string(addr) != fromAddr {
			return CodeTypeUnauthorized, errors.New("The hash " + v + " is not owned by you.")
		}
	}
	return CodeTypeOK, nil
}

func (pba *PBApplication) deliverTxValidator(dr DeliveryRequest) (uint32, error) {
	pubk, err := crypto.PubKeyFromBytes(dr.Data.From)
	if err != nil {
		return CodeTypeEncodingError, errors.New("Public key is not correct.")
	}
	b, _ := json.Marshal(dr.Data)
	sig, err := crypto.SignatureFromBytes(dr.Signature)
	if err != nil {
		return CodeTypeEncodingError, errors.New("Signature is not correct.")
	}
	isVerified := pubk.VerifyBytes(b, sig)
	if !isVerified {
		return CodeTypeUnauthorized, errors.New("The signature does not validate the data.")
	}

	// check if the hashes exist in the IPFS
	sh := shell.NewShell(conf.Conf.IpfsConnection)
	for _, v := range dr.Data.Files {
		_, err := sh.BlockGet(v)
		if err != nil {
			return CodeTypeEncodingError,
				errors.New("The file " + v + " does not exists.")
		}
	}

	switch action := dr.Data.Action; action {
	case ADD_ACTION:
		code, err := pba.addActionValidation(dr)
		if err != nil {
			return code, err
		}
	case REMOVE_ACTION:
		code, err := pba.removeActionValidation(dr)
		if err != nil {
			return code, err
		}
	case SEND_ACTION:
		code, err := pba.sendActionValidation(dr)
		if err != nil {
			return code, err
		}
	}

	return CodeTypeOK, nil
}

func (pba *PBApplication) addActionState(dr DeliveryRequest) {
	fromAddr, _ := dr.FromPubKeyAddress()
	pba.addFilesToUserKey(fromAddr, dr.Data.Files)
	for _, v := range dr.Data.Files {
		pba.state.db.Set(prefixFileKey(v), []byte(fromAddr))
	}
}

func (pba *PBApplication) addFilesToUserKey(fromAddr string, addFiles []string) {
	filesBy := pba.state.db.Get(prefixUserKey(fromAddr))
	files := []string{}
	json.Unmarshal(filesBy, &files)
	files = append(files, addFiles...)
	b, _ := json.Marshal(files)
	pba.state.db.Set(prefixUserKey(fromAddr), b)
}

func (pba *PBApplication) removeFilesFromUserKey(fromAddr string, delFiles []string) {
	filesBy := pba.state.db.Get(prefixUserKey(fromAddr))
	files := []string{}
	json.Unmarshal(filesBy, &files)
	for i := 0; i < len(files); i++ {
		for _, v := range delFiles {
			if v == files[i] {
				files = append(files[:i], files[i+1:]...)
				i -= 1
				break
			}
		}
	}
	if len(files) == 0 {
		pba.state.db.Delete(prefixUserKey(fromAddr))
	} else {
		b, _ := json.Marshal(files)
		pba.state.db.Set(prefixUserKey(fromAddr), b)
	}
}

func (pba *PBApplication) removeActionState(dr DeliveryRequest) {
	fromAddr, _ := dr.FromPubKeyAddress()
	pba.removeFilesFromUserKey(fromAddr, dr.Data.Files)
	for _, v := range dr.Data.Files {
		pba.state.db.Delete(prefixFileKey(v))
	}
}

func (pba *PBApplication) sendActionState(dr DeliveryRequest) {
	toAddr, _ := dr.ToPubKeyAddress()
	fromAddr, _ := dr.FromPubKeyAddress()
	pba.removeFilesFromUserKey(fromAddr, dr.Data.Files)
	pba.addFilesToUserKey(toAddr, dr.Data.Files)

	if conf.Conf.Blockchain == conf.OtoOPB {
		pba.state.db.Set(prefixUserKey(toAddr), nil)
		pba.state.db.Delete(prefixUserKey(fromAddr))
	}
	for _, v := range dr.Data.Files {
		pba.state.db.Set(prefixFileKey(v), []byte(toAddr))
	}
}

func (pba *PBApplication) DeliverTx(tx []byte) types.ResponseDeliverTx {
	dr := DeliveryRequest{}
	err := json.Unmarshal(tx, &dr)
	if err != nil {
		return types.ResponseDeliverTx{Code: CodeTypeEncodingError, Log: "The response is not correct."}
	}
	code, err := pba.deliverTxValidator(dr)
	if err != nil {
		return types.ResponseDeliverTx{Code: code, Log: err.Error()}
	}

	switch action := dr.Data.Action; action {
	case ADD_ACTION:
		pba.addActionState(dr)
	case REMOVE_ACTION:
		pba.removeActionState(dr)
	case SEND_ACTION:
		pba.sendActionState(dr)
	}

	return types.ResponseDeliverTx{Code: code}
}
