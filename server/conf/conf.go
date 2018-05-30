package conf

import (
	"encoding/json"
	"log"

	"github.com/ipfs/go-ipfs-api"
)

type BlockchainType string

const (
	SPB    = BlockchainType("spb")
	OtoOPB = BlockchainType("otopb")
)

type configuration struct {
	IpfsConnection              string
	Blockchain                  BlockchainType
	WaitingSecondsQuery         int
	AuthorizedAddressesIpfsHash string
	authorizedAddresses         map[string]string
}

func (c *configuration) GetAuthorizedAddresses() map[string]string {
	return c.authorizedAddresses
}

func (c *configuration) SetAuthorizedAddresses() {
	sh := shell.NewShell(c.IpfsConnection)
	if len(c.AuthorizedAddressesIpfsHash) == 0 {
		return
	}

	listBy, err := sh.BlockGet(c.AuthorizedAddressesIpfsHash)
	if err != nil {
		log.Fatal("The ipfs hash for authorized addresses has a problem, " + err.Error())
	}
	list := []string{}
	err = json.Unmarshal(listBy, &list)
	if err != nil {
		log.Fatal("The ipfs hash is not a json")
	}
	for _, v := range list {
		c.authorizedAddresses[v] = v
	}
}

var Conf = configuration{}

func init() {
	Conf.IpfsConnection = "127.0.0.1:5001"
	Conf.Blockchain = SPB
	Conf.WaitingSecondsQuery = 5
	Conf.authorizedAddresses = map[string]string{}
	Conf.SetAuthorizedAddresses()
}
