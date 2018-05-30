package ctrls

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ipfs/go-ipfs-api"
	"github.com/mragiadakos/planetary-blockchain/server/conf"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

func (f forTestUtils) querySpb(t *testing.T, from crypto.PrivKeyEd25519, file *string, userAddr *string) SpbQuery {
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
	return q
}

func TestSpbQuerySuccessfully(t *testing.T) {
	conf.Conf.Blockchain = conf.SPB
	pba := NewPBApplication()
	utils := forTestUtils{}

	fromEdKey := crypto.GenPrivKeyEd25519()
	input := [][]byte{[]byte("random1")}
	addDr := utils.createAddOrRemoveDelivery(t, fromEdKey, ADD_ACTION, input)
	b, _ := json.Marshal(addDr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	q := utils.querySpb(t, fromEdKey, nil, nil)
	b, _ = json.Marshal(q)
	req := types.RequestQuery{}
	req.Data = b

	// the value that we expect
	qr := QueryResponse{}
	qr.Files = addDr.Data.Files
	b, _ = json.Marshal(qr)
	res := types.ResponseQuery{Code: CodeTypeOK}
	res.Value = b
	assert.Equal(t, res, pba.Query(req))
}

func TestSpbQueryFailSignature(t *testing.T) {
	conf.Conf.Blockchain = conf.SPB
	pba := NewPBApplication()
	utils := forTestUtils{}

	fromEdKey := crypto.GenPrivKeyEd25519()
	q := utils.querySpb(t, fromEdKey, nil, nil)
	q.Data.Nonce = "a"
	b, _ := json.Marshal(q)
	req := types.RequestQuery{}
	req.Data = b

	assert.Equal(t, CodeTypeUnauthorized, pba.Query(req).Code)
}

func TestSpbQueryReturnEmptyFromKeyThatDoesNotHaveFiles(t *testing.T) {
	conf.Conf.Blockchain = conf.SPB
	pba := NewPBApplication()
	utils := forTestUtils{}

	fromEdKey := crypto.GenPrivKeyEd25519()
	q := utils.querySpb(t, fromEdKey, nil, nil)
	b, _ := json.Marshal(q)
	req := types.RequestQuery{}
	req.Data = b

	// the value that we expect
	qr := QueryResponse{}
	qr.Files = []string{}
	b, _ = json.Marshal(qr)
	res := types.ResponseQuery{Code: CodeTypeOK}
	res.Value = b
	assert.Equal(t, res, pba.Query(req))
}

func TestSpbQuerySuccessfullyGiveOneFile(t *testing.T) {
	conf.Conf.Blockchain = conf.SPB
	pba := NewPBApplication()
	utils := forTestUtils{}

	fromEdKey := crypto.GenPrivKeyEd25519()
	input := [][]byte{[]byte("random1"), []byte("random2")}
	addDr := utils.createAddOrRemoveDelivery(t, fromEdKey, ADD_ACTION, input)
	b, _ := json.Marshal(addDr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	q := utils.querySpb(t, fromEdKey, &addDr.Data.Files[0], nil)
	b, _ = json.Marshal(q)
	req := types.RequestQuery{}
	req.Data = b

	// the value that we expect
	qr := QueryResponse{}
	qr.Files = []string{addDr.Data.Files[0]}
	b, _ = json.Marshal(qr)
	res := types.ResponseQuery{Code: CodeTypeOK}
	res.Value = b
	assert.Equal(t, res, pba.Query(req))
}

func TestSpbQueryFailOnTimeSigningAfter1Second(t *testing.T) {
	conf.Conf.Blockchain = conf.SPB
	conf.Conf.WaitingSecondsQuery = 1
	pba := NewPBApplication()
	utils := forTestUtils{}

	fromEdKey := crypto.GenPrivKeyEd25519()
	input := [][]byte{[]byte("random1")}
	addDr := utils.createAddOrRemoveDelivery(t, fromEdKey, ADD_ACTION, input)
	b, _ := json.Marshal(addDr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	q := utils.querySpb(t, fromEdKey, nil, nil)
	b, _ = json.Marshal(q)
	req := types.RequestQuery{}
	req.Data = b

	time.Sleep(2 * time.Second)
	assert.Equal(t, CodeTypeUnauthorized, pba.Query(req).Code)
}

func TestSpbQuerySuccessOnQueryOtherUsersFiles(t *testing.T) {
	conf.Conf.Blockchain = conf.SPB
	conf.Conf.WaitingSecondsQuery = 1
	pba := NewPBApplication()
	utils := forTestUtils{}
	// we authorized user by submitting its address to the configuration file
	authorizedEdKey := crypto.GenPrivKeyEd25519()
	sh := shell.NewShell(conf.Conf.IpfsConnection)
	b, _ := json.Marshal([]string{authorizedEdKey.PubKey().Address().String()})
	hash, err := sh.BlockPut(b)
	assert.Nil(t, err)
	conf.Conf.AuthorizedAddressesIpfsHash = hash
	conf.Conf.SetAuthorizedAddresses()

	fromEdKey := crypto.GenPrivKeyEd25519()
	input := [][]byte{[]byte("random1")}
	addDr := utils.createAddOrRemoveDelivery(t, fromEdKey, ADD_ACTION, input)
	b, _ = json.Marshal(addDr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	fromAddr := fromEdKey.PubKey().Address().String()
	q := utils.querySpb(t, authorizedEdKey, nil, &fromAddr)
	b, _ = json.Marshal(q)
	req := types.RequestQuery{}
	req.Data = b

	// the value that we expect
	qr := QueryResponse{}
	qr.Files = addDr.Data.Files
	b, _ = json.Marshal(qr)
	res := types.ResponseQuery{Code: CodeTypeOK}
	res.Value = b
	assert.Equal(t, res, pba.Query(req))

}

func TestSpbQueryFailOnQueryOtherUsersFiles(t *testing.T) {
	conf.Conf.Blockchain = conf.SPB
	conf.Conf.WaitingSecondsQuery = 1
	pba := NewPBApplication()
	utils := forTestUtils{}

	authorizedEdKey := crypto.GenPrivKeyEd25519()

	fromEdKey := crypto.GenPrivKeyEd25519()
	input := [][]byte{[]byte("random1")}
	addDr := utils.createAddOrRemoveDelivery(t, fromEdKey, ADD_ACTION, input)
	b, _ := json.Marshal(addDr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	fromAddr := fromEdKey.PubKey().Address().String()
	q := utils.querySpb(t, authorizedEdKey, nil, &fromAddr)
	b, _ = json.Marshal(q)
	req := types.RequestQuery{}
	req.Data = b

	assert.Equal(t, CodeTypeUnauthorized, pba.Query(req).Code)
}

func TestOtopbQuerySuccesfully(t *testing.T) {
	conf.Conf.Blockchain = conf.OtoOPB
	pba := NewPBApplication()
	utils := forTestUtils{}

	fromEdKey := crypto.GenPrivKeyEd25519()
	input := [][]byte{[]byte("random1")}
	addDr := utils.createAddOrRemoveDelivery(t, fromEdKey, ADD_ACTION, input)
	b, _ := json.Marshal(addDr)
	assert.Equal(t, CodeTypeOK, pba.DeliverTx(b).Code)

	oq := OtopbQuery{}
	oq.From = fromEdKey.PubKey().Bytes()
	b, _ = json.Marshal(oq)
	req := types.RequestQuery{}
	req.Data = b

	qr := QueryResponse{}
	qr.Files = addDr.Data.Files
	b, _ = json.Marshal(qr)
	res := types.ResponseQuery{Code: CodeTypeOK}
	res.Value = b
	assert.Equal(t, res, pba.Query(req))

}
