package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	kitlog "github.com/go-kit/kit/log"
	"github.com/mragiadakos/planetary-blockchain/server/conf"
	"github.com/mragiadakos/planetary-blockchain/server/ctrls"
	absrv "github.com/tendermint/abci/server"
	cmn "github.com/tendermint/tmlibs/common"
	tmlog "github.com/tendermint/tmlibs/log"
)

func main() {
	logger := tmlog.NewTMLogger(kitlog.NewSyncWriter(os.Stdout))
	flagAbci := "socket"
	ipfsDaemon := flag.String("ipfs", "127.0.0.1:5001", "the URL for the IPFS's daemon")
	node := flag.String("node", "tcp://0.0.0.0:46658", "the TCP URL for the ABCI daemon")
	ipfsAuthorizedUserHash := flag.String("auth", "", "the IPFS hash with the JSON list of public key addresses")
	waitSec := flag.Int("wait", 5, "the seconds for an acceptable query")
	blockchainType := flag.String("type", "spb", "the blockchain types are allowed SPB as 'spb' and OtoOPB as 'otoopb'")
	flag.Parse()

	if len(*ipfsAuthorizedUserHash) > 0 {
		conf.Conf.AuthorizedAddressesIpfsHash = *ipfsAuthorizedUserHash
	}
	conf.Conf.AbciDaemon = *node
	conf.Conf.IpfsConnection = *ipfsDaemon
	conf.Conf.WaitingSecondsQuery = *waitSec
	if *blockchainType != "spb" && *blockchainType != "otopb" {
		log.Fatal("There is not such a type, try 'spb' or 'otopb'")
	}
	conf.Conf.Blockchain = conf.BlockchainType(*blockchainType)

	app := ctrls.NewPBApplication()
	srv, err := absrv.NewServer(conf.Conf.AbciDaemon, flagAbci, app)
	if err != nil {
		fmt.Println("Error ", err)
		return
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if err := srv.Start(); err != nil {
		fmt.Println("Error ", err)
		return
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})
}
