package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	crypto "github.com/tendermint/go-crypto"
	"github.com/urfave/cli"
)

type JsonKey struct {
	PrivateKey string
	PublicKey  string
	PublicAddr string
}

var GenerateKey = cli.Command{
	Name:    "generate",
	Aliases: []string{"g"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "filename",
			Usage: "the filename that the key will be saved",
		},
	},
	Usage: "generate the key in a file",
	Action: func(c *cli.Context) error {
		filename := c.String("filename")
		if len(filename) == 0 {
			return errors.New("Error: filename is missing")
		}
		edKey := crypto.GenPrivKeyEd25519()
		privHex := hex.EncodeToString(edKey.Bytes())
		jk := JsonKey{}
		jk.PrivateKey = privHex
		jk.PublicKey = hex.EncodeToString(edKey.PubKey().Bytes())
		jk.PublicAddr = edKey.PubKey().Address().String()
		b, _ := json.Marshal(jk)
		err := ioutil.WriteFile(filename, b, 0644)
		if err != nil {
			return errors.New("Error: " + err.Error())
		}
		fmt.Println("Created successfully the key.")
		return nil
	},
}

type TypeFile string

const (
	JSON_TYPE = TypeFile("json")
	FILE_TYPE = TypeFile("file")
	DIR_TYPE  = TypeFile("dir")
)

var Add = cli.Command{
	Name:    "add",
	Aliases: []string{"a"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "key",
			Usage: "the filename that contains the key in json file",
		},
		cli.StringFlag{
			Name:  "type",
			Usage: "the type of file",
		},
		cli.StringFlag{
			Name: "input",
			Usage: "the input based on the type.\n" +
				"When the type is json, it expects a string json.\n" +
				"When the type is directory, it expects a directory.\n" +
				"When the type is file, it expects a file",
		},
	},
	Usage: "add a hash to the blockchain",
	Action: func(c *cli.Context) error {
		key := c.String("key")
		if len(key) == 0 {
			return errors.New("Error: the key is missing")
		}
		tp := TypeFile(c.String("type"))
		if tp != FILE_TYPE && tp != JSON_TYPE && tp != DIR_TYPE {
			return errors.New("Error: the type needs to be 'json', 'dir' or 'file'")
		}

		input := c.String("input")
		if len(input) == 0 {
			return errors.New("Error: the input is empty")
		}

		edKey, err := fileKey(key)
		if err != nil {
			return err
		}
		hash := ""
		switch TypeFile(tp) {
		case JSON_TYPE:
			hash, err = ipfsAddJson([]byte(input))
		case FILE_TYPE:
			hash, err = ipfsAddFile(input)
		case DIR_TYPE:
			hash, err = ipfsAddFolder(input)
		}
		if err != nil {
			return errors.New("Error: IPFS problem " + err.Error())
		}

		_, err = AddRequest(*edKey, []string{hash})
		if err != nil {
			return errors.New("Error: the transaction failed: " + err.Error())
		}
		fmt.Println("Successfully added the hash " + hash)
		return nil
	},
}

var Remove = cli.Command{
	Name:    "remove",
	Aliases: []string{"r"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "key",
			Usage: "the filename that contains the key in json file",
		},
		cli.StringFlag{
			Name:  "hash",
			Usage: "the hash of the file",
		},
	},
	Usage: "remove a hash from the blockchain",
	Action: func(c *cli.Context) error {
		key := c.String("key")
		if len(key) == 0 {
			return errors.New("Error: the key is missing")
		}

		hash := c.String("hash")
		if len(hash) == 0 {
			return errors.New("Error: the hash is empty")
		}

		edKey, err := fileKey(key)
		if err != nil {
			return err
		}

		_, err = RemoveRequest(*edKey, []string{hash})
		if err != nil {
			return errors.New("Error: the transaction failed: " + err.Error())
		}
		fmt.Println("Successfully removed the hash " + hash)

		return nil
	},
}

var Send = cli.Command{
	Name:    "send",
	Aliases: []string{"s"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "key",
			Usage: "the filename that contains the key in json file",
		},
		cli.StringFlag{
			Name:  "receiver",
			Usage: "the public key of the receiver",
		},
		cli.StringFlag{
			Name:  "hash",
			Usage: "the hash of the file",
		},
	},
	Usage: "send a hash to another person",
	Action: func(c *cli.Context) error {
		key := c.String("key")
		if len(key) == 0 {
			return errors.New("Error: the key is missing")
		}

		hash := c.String("hash")
		if len(hash) == 0 {
			return errors.New("Error: the hash is empty")
		}

		receiver := c.String("receiver")
		if len(hash) == 0 {
			return errors.New("Error: the receiver is empty")
		}

		edKey, err := fileKey(key)
		if err != nil {
			return err
		}
		b, err := hex.DecodeString(receiver)
		if err != nil {
			return err
		}
		_, err = SendRequest(*edKey, b, []string{hash})
		if err != nil {
			return errors.New("Error: the transaction failed: " + err.Error())
		}
		fmt.Println("Successfully send the hash " + hash + " to " + receiver)
		return nil
	},
}

var Query = cli.Command{
	Name:    "query",
	Aliases: []string{"q"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "key",
			Usage: "the filename that contains the key in json file",
		},
		cli.StringFlag{
			Name:  "type",
			Usage: "the type of blockchain",
		},
		cli.StringFlag{
			Name:  "hash",
			Usage: "the file of the hash",
		},
		cli.StringFlag{
			Name:  "addr",
			Usage: "the user address",
		},
	},
	Usage: "query the files based on the key",
	Action: func(c *cli.Context) error {
		key := c.String("key")
		if len(key) == 0 {
			return errors.New("Error: the key is missing")
		}
		tp := BlockchainType(c.String("type"))
		if tp != SPB && tp != OtoOPB {
			return errors.New("Error: the type needs to be 'spb or 'otopb'")
		}

		hashStr := c.String("hash")
		var hash *string
		if len(hashStr) > 0 {
			hash = &hashStr
		}

		addrStr := c.String("addr")
		var addr *string
		if len(addrStr) > 0 {
			addr = &addrStr
		}

		edKey, err := fileKey(key)
		if err != nil {
			return err
		}
		qr := new(QueryResponse)
		if OtoOPB == tp {
			qr, _, err = OtopbQueryRequest(*edKey)
		} else if SPB == tp {
			qr, _, err = SpbQueryRequest(*edKey, hash, addr)
		}
		if err != nil {
			return err
		}
		fmt.Println("Files:")
		for _, v := range qr.Files {
			fmt.Println(v)
		}
		return nil
	},
}
