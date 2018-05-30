package ctrls

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ipfs/go-ipfs-api"
	"github.com/mragiadakos/planetary-blockchain/server/conf"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
)

type forTestUtils struct{}

func (f forTestUtils) createAddOrRemoveDelivery(t *testing.T, from crypto.PrivKeyEd25519, action ActionStruct, input [][]byte) DeliveryResponse {
	dd := DeliveryData{}
	if action != ADD_ACTION && action != REMOVE_ACTION {
		assert.Error(t, errors.New("use the method for add or remove"))
	}
	dd.Action = action
	sh := shell.NewShell(conf.Conf.IpfsConnection)
	dd.Files = []string{}
	for _, v := range input {
		red := bytes.NewReader(v)
		hash, err := sh.Add(red)
		assert.Nil(t, err)
		dd.Files = append(dd.Files, hash)
	}
	dd.From = from.PubKey().Bytes()
	b, _ := json.Marshal(dd)
	dr := DeliveryResponse{}
	dr.Signature = from.Sign(b).Bytes()
	dr.Data = dd
	return dr
}

func (f forTestUtils) createSendDelivery(t *testing.T, from crypto.PrivKeyEd25519, to *crypto.PrivKeyEd25519,
	action ActionStruct, files []string) DeliveryResponse {
	dd := DeliveryData{}
	if action != SEND_ACTION {
		assert.Error(t, errors.New("use the method for send only"))
	}
	dd.Action = action
	dd.Files = files
	dd.From = from.PubKey().Bytes()
	if to != nil {
		toB := to.PubKey().Bytes()
		dd.To = &toB
	}
	b, _ := json.Marshal(dd)
	dr := DeliveryResponse{}
	dr.Signature = from.Sign(b).Bytes()
	dr.Data = dd
	return dr
}

func TestDeliverySuccesfulAdd(t *testing.T) {
	pba := NewPBApplication()
	edKey := crypto.GenPrivKeyEd25519()
	utils := forTestUtils{}
	dr := utils.createAddOrRemoveDelivery(t, edKey, ADD_ACTION, [][]byte{[]byte("random")})
	b, _ := json.Marshal(dr)

	assert.Equal(t, types.ResponseDeliverTx{Code: CodeTypeOK}, pba.DeliverTx(b))
}

func TestDeliveryFailOnSignature(t *testing.T) {
	pba := NewPBApplication()
	edKey := crypto.GenPrivKeyEd25519()
	utils := forTestUtils{}
	sneakyDr := utils.createAddOrRemoveDelivery(t, edKey, ADD_ACTION, [][]byte{[]byte("random2")})

	dr := utils.createAddOrRemoveDelivery(t, edKey, ADD_ACTION, [][]byte{[]byte("random")})
	dr.Data.Files = append(dr.Data.Files, sneakyDr.Data.Files[0])
	b, _ := json.Marshal(dr)

	assert.Equal(t, CodeTypeUnauthorized, pba.DeliverTx(b).Code)
}

func TestDeliveryFailOnSameAdd(t *testing.T) {
	pba := NewPBApplication()
	edKey := crypto.GenPrivKeyEd25519()
	utils := forTestUtils{}
	dr := utils.createAddOrRemoveDelivery(t, edKey, ADD_ACTION, [][]byte{[]byte("random")})
	b, _ := json.Marshal(dr)

	assert.Equal(t, types.ResponseDeliverTx{Code: CodeTypeOK}, pba.DeliverTx(b))

	utils = forTestUtils{}
	dr = utils.createAddOrRemoveDelivery(t, edKey, ADD_ACTION, [][]byte{[]byte("random")})
	b, _ = json.Marshal(dr)

	assert.Equal(t, CodeTypeUnauthorized, pba.DeliverTx(b).Code)
}

func TestSpbDeliveryFailToRemoveTwice(t *testing.T) {
	conf.Conf.Blockchain = conf.SPB
	pba := NewPBApplication()
	edKey := crypto.GenPrivKeyEd25519()
	utils := forTestUtils{}
	input := [][]byte{[]byte("random")}
	dr := utils.createAddOrRemoveDelivery(t, edKey, ADD_ACTION, input)
	b, _ := json.Marshal(dr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	dr = utils.createAddOrRemoveDelivery(t, edKey, REMOVE_ACTION, input)
	b, _ = json.Marshal(dr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	dr = utils.createAddOrRemoveDelivery(t, edKey, REMOVE_ACTION, input)
	b, _ = json.Marshal(dr)
	assert.Equal(t, CodeTypeUnauthorized, pba.DeliverTx(b).Code)
}

func TestSpbDeliveryFailToRemoveOthersHash(t *testing.T) {
	conf.Conf.Blockchain = conf.SPB
	pba := NewPBApplication()
	otherEdKey := crypto.GenPrivKeyEd25519()
	utils := forTestUtils{}
	input := [][]byte{[]byte("random")}
	dr := utils.createAddOrRemoveDelivery(t, otherEdKey, ADD_ACTION, input)
	b, _ := json.Marshal(dr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	myEdKey := crypto.GenPrivKeyEd25519()
	dr = utils.createAddOrRemoveDelivery(t, myEdKey, REMOVE_ACTION, input)
	b, _ = json.Marshal(dr)
	assert.Equal(t, CodeTypeUnauthorized, pba.DeliverTx(b).Code)
}

func TestOtopbDeliveryFailToAddTwiceWithTheSameKey(t *testing.T) {
	conf.Conf.Blockchain = conf.OtoOPB
	pba := NewPBApplication()
	edKey := crypto.GenPrivKeyEd25519()
	utils := forTestUtils{}
	input := [][]byte{[]byte("random1"), []byte("random2")}
	dr := utils.createAddOrRemoveDelivery(t, edKey, ADD_ACTION, input)
	b, _ := json.Marshal(dr)
	assert.Equal(t, CodeTypeUnauthorized, pba.DeliverTx(b).Code)
}

func TestOtopbDeliveryFailToAddTwoHashesWithTheSameKey(t *testing.T) {
	conf.Conf.Blockchain = conf.OtoOPB
	pba := NewPBApplication()
	edKey := crypto.GenPrivKeyEd25519()
	utils := forTestUtils{}

	input := [][]byte{[]byte("random1")}
	dr := utils.createAddOrRemoveDelivery(t, edKey, ADD_ACTION, input)
	b, _ := json.Marshal(dr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	input = [][]byte{[]byte("random2")}
	dr = utils.createAddOrRemoveDelivery(t, edKey, ADD_ACTION, input)
	b, _ = json.Marshal(dr)
	assert.Equal(t, CodeTypeUnauthorized, pba.DeliverTx(b).Code)
}

func TestOtopbDeliveryToRemoveTwiceTheSameKey(t *testing.T) {
	conf.Conf.Blockchain = conf.OtoOPB
	pba := NewPBApplication()
	edKey := crypto.GenPrivKeyEd25519()
	utils := forTestUtils{}

	input := [][]byte{[]byte("random1")}
	dr := utils.createAddOrRemoveDelivery(t, edKey, ADD_ACTION, input)
	b, _ := json.Marshal(dr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	dr = utils.createAddOrRemoveDelivery(t, edKey, REMOVE_ACTION, input)
	b, _ = json.Marshal(dr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)
}

func TestSpbDeliverySendFailOnEmptyTo(t *testing.T) {
	conf.Conf.Blockchain = conf.SPB
	pba := NewPBApplication()
	fromEdKey := crypto.GenPrivKeyEd25519()
	utils := forTestUtils{}

	input := [][]byte{[]byte("random1")}
	addDr := utils.createAddOrRemoveDelivery(t, fromEdKey, ADD_ACTION, input)
	b, _ := json.Marshal(addDr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	sendDr := utils.createSendDelivery(t, fromEdKey, nil, SEND_ACTION, addDr.Data.Files)
	b, _ = json.Marshal(sendDr)
	assert.Equal(t, CodeTypeUnauthorized, pba.DeliverTx(b).Code)
}

func TestSpbDeliverySendSuccessfully(t *testing.T) {
	conf.Conf.Blockchain = conf.SPB
	pba := NewPBApplication()
	fromEdKey := crypto.GenPrivKeyEd25519()
	toEdKey := crypto.GenPrivKeyEd25519()
	utils := forTestUtils{}

	input := [][]byte{[]byte("random1")}
	addDr := utils.createAddOrRemoveDelivery(t, fromEdKey, ADD_ACTION, input)
	b, _ := json.Marshal(addDr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	sendDr := utils.createSendDelivery(t, fromEdKey, &toEdKey, SEND_ACTION, addDr.Data.Files)
	b, _ = json.Marshal(sendDr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	remDrForTo := utils.createAddOrRemoveDelivery(t, toEdKey, REMOVE_ACTION, input)
	b, _ = json.Marshal(remDrForTo)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)
}

func TestSpbDeliverySendFailOnSendingSomethingThatIsNotOwned(t *testing.T) {
	conf.Conf.Blockchain = conf.SPB
	pba := NewPBApplication()

	thirdEdKey := crypto.GenPrivKeyEd25519()
	fromEdKey := crypto.GenPrivKeyEd25519()
	toEdKey := crypto.GenPrivKeyEd25519()
	utils := forTestUtils{}

	input := [][]byte{[]byte("random1")}
	addDr := utils.createAddOrRemoveDelivery(t, thirdEdKey, ADD_ACTION, input)
	b, _ := json.Marshal(addDr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	sendDr := utils.createSendDelivery(t, fromEdKey, &toEdKey, SEND_ACTION, addDr.Data.Files)
	b, _ = json.Marshal(sendDr)
	assert.Equal(t, CodeTypeUnauthorized, pba.DeliverTx(b).Code)

}

func TestSpbDeliverySendFailOnSelf(t *testing.T) {
	conf.Conf.Blockchain = conf.SPB
	pba := NewPBApplication()
	utils := forTestUtils{}

	fromEdKey := crypto.GenPrivKeyEd25519()
	input := [][]byte{[]byte("random1")}
	addDr := utils.createAddOrRemoveDelivery(t, fromEdKey, ADD_ACTION, input)
	b, _ := json.Marshal(addDr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	sendDr := utils.createSendDelivery(t, fromEdKey, &fromEdKey, SEND_ACTION, addDr.Data.Files)
	b, _ = json.Marshal(sendDr)
	assert.Equal(t, CodeTypeUnauthorized, pba.DeliverTx(b).Code)

}

func TestOtopbDeliverySendFailOnKeyThatExists(t *testing.T) {
	conf.Conf.Blockchain = conf.OtoOPB
	pba := NewPBApplication()
	fromEdKey := crypto.GenPrivKeyEd25519()
	toEdKey := crypto.GenPrivKeyEd25519()
	utils := forTestUtils{}

	input := [][]byte{[]byte("random1")}
	addDrFrom := utils.createAddOrRemoveDelivery(t, fromEdKey, ADD_ACTION, input)
	b, _ := json.Marshal(addDrFrom)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	input = [][]byte{[]byte("random2")}
	addDrTo := utils.createAddOrRemoveDelivery(t, toEdKey, ADD_ACTION, input)
	b, _ = json.Marshal(addDrTo)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	sendDr := utils.createSendDelivery(t, fromEdKey, &toEdKey, SEND_ACTION, addDrFrom.Data.Files)
	b, _ = json.Marshal(sendDr)
	assert.Equal(t, CodeTypeUnauthorized, pba.DeliverTx(b).Code)

}

func TestOtopbDeliverySendSuccessWhenReusesTheSameKeyAfterSendForDifferentFile(t *testing.T) {
	conf.Conf.Blockchain = conf.OtoOPB
	pba := NewPBApplication()
	fromEdKey := crypto.GenPrivKeyEd25519()
	toEdKey := crypto.GenPrivKeyEd25519()
	utils := forTestUtils{}

	input := [][]byte{[]byte("random1")}
	addDrFrom := utils.createAddOrRemoveDelivery(t, fromEdKey, ADD_ACTION, input)
	b, _ := json.Marshal(addDrFrom)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	sendDr := utils.createSendDelivery(t, fromEdKey, &toEdKey, SEND_ACTION, addDrFrom.Data.Files)
	b, _ = json.Marshal(sendDr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	input = [][]byte{[]byte("random2")}
	addDrFrom = utils.createAddOrRemoveDelivery(t, fromEdKey, ADD_ACTION, input)
	b, _ = json.Marshal(addDrFrom)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

}
