package ctrls

import (
	"encoding/json"

	"github.com/tendermint/abci/types"
)

func (pba *PBApplication) CheckTx(tx []byte) types.ResponseCheckTx {
	dr := DeliveryRequest{}
	err := json.Unmarshal(tx, &dr)
	if err != nil {
		return types.ResponseCheckTx{Code: CodeTypeEncodingError, Log: "The response is not correct."}
	}
	code, err := pba.deliverTxValidator(dr)
	if err != nil {
		return types.ResponseCheckTx{Code: code, Log: err.Error()}
	}

	return types.ResponseCheckTx{Code: CodeTypeOK}
}
