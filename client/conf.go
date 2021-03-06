package main

type BlockchainType string

const (
	SPB    = BlockchainType("spb")
	OtoOPB = BlockchainType("otopb")
)

type configuration struct {
	IpfsConnection string
	Blockchain     BlockchainType
	AbciDaemon     string
}

var Conf = configuration{}

func init() {
	Conf.IpfsConnection = "127.0.0.1:5001"
	Conf.AbciDaemon = "http://0.0.0.0:26657"
	Conf.Blockchain = SPB
}
