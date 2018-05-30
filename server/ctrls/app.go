package ctrls

import (
	"github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
)

var _ types.Application = (*PBApplication)(nil)

type PBApplication struct {
	types.BaseApplication

	state State
}

func NewPBApplication() *PBApplication {
	state := loadState(dbm.NewMemDB())
	return &PBApplication{state: state}
}
