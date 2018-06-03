package main

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	crypto "github.com/tendermint/go-crypto"
)

func TestDeliverSuccessfully(t *testing.T) {
	edKey := crypto.GenPrivKeyEd25519()
	u := uuid.NewV4()
	hash, err := ipfsAddJson(u.Bytes())
	assert.Nil(t, err)
	code, err := AddRequest(edKey, []string{hash})
	assert.Nil(t, err)
	assert.Equal(t, CodeTypeOK, code)
}

func TestSpbDeliverAndFindHashSuccessfully(t *testing.T) {
	edKey := crypto.GenPrivKeyEd25519()
	u := uuid.NewV4()
	hash, err := ipfsAddJson(u.Bytes())
	assert.Nil(t, err)
	code, err := AddRequest(edKey, []string{hash})
	assert.Nil(t, err)
	assert.Equal(t, CodeTypeOK, code)

	qr, code, err := SpbQueryRequest(edKey, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, CodeTypeOK, code)
	assert.Equal(t, []string{hash}, qr.Files)
}

func TestOtopbDeliverAndFindHashSuccessfully(t *testing.T) {
	edKey := crypto.GenPrivKeyEd25519()
	u := uuid.NewV4()
	hash, err := ipfsAddJson(u.Bytes())
	assert.Nil(t, err)
	code, err := AddRequest(edKey, []string{hash})
	assert.Nil(t, err)
	assert.Equal(t, CodeTypeOK, code)

	qr, code, err := OtopbQueryRequest(edKey)
	assert.Nil(t, err)
	assert.Equal(t, CodeTypeOK, code)
	assert.Equal(t, []string{hash}, qr.Files)
}
