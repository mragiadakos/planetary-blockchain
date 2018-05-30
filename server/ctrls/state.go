package ctrls

import (
	"encoding/json"

	dbm "github.com/tendermint/tmlibs/db"
)

var (
	stateKey = []byte("stateKey")
	fileKey  = []byte("fileKey:")
	userKey  = []byte("userKey:")
)

type State struct {
	db      dbm.DB
	Size    int64  `json:"size"`
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

func loadState(db dbm.DB) State {
	stateBytes := db.Get(stateKey)
	var state State
	if len(stateBytes) != 0 {
		err := json.Unmarshal(stateBytes, &state)
		if err != nil {
			panic(err)
		}
	}
	state.db = db
	return state
}

func saveState(state State) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	state.db.Set(stateKey, stateBytes)
}

func prefixUserKey(key string) []byte {
	b := []byte(key)
	return append(userKey, b...)
}

func prefixFileKey(key string) []byte {
	b := []byte(key)
	return append(fileKey, b...)
}
