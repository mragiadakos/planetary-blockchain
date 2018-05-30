package ctrls

import (
	"encoding/binary"

	"github.com/tendermint/abci/example/code"
	"github.com/tendermint/abci/types"
)

func (pba *PBApplication) CheckTx(tx []byte) types.ResponseCheckTx {
	return types.ResponseCheckTx{Code: code.CodeTypeOK}
}

func (pba *PBApplication) Commit() types.ResponseCommit {
	// Using a memdb - just return the big endian size of the db
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, pba.state.Size)
	pba.state.AppHash = appHash
	pba.state.Height += 1
	saveState(pba.state)
	return types.ResponseCommit{Data: appHash}
}
