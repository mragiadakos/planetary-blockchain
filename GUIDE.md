Guide

Make sure that IPFS daemon and tendermint node is working

Start the server

$ server --type=spb

Generate a key in a file
$ ./client g --filename key.json
Created successfully the key.

Add a json to the blockchain and receive hash
$ ./client a --key=key.json  --type=json --input='{"coin":1}'
Successfully added the hash QmfJudNdQPrGLcaxaxvH1eCMLU7buTAFCcFGrt4etWv7rq

Remove a hash
$ ./client r --key=key.json --hash=QmfJudNdQPrGLcaxaxvH1eCMLU7buTAFCcFGrt4etWv7rq
Successfully removed the hash QmfJudNdQPrGLcaxaxvH1eCMLU7buTAFCcFGrt4etWv7rq

Generate a key in a file for the other person to send
$ ./client g --filename other.json
Created successfully the key.

Create a new hash that we will send
$ ./client a --key=key.json  --type=json --input='{"coin":3}'
Successfully added the hash Qmco6ZUSMApGbyVh2sWajtx2JTJiRnVDDbvPHeAEvsUgHT

Send the hash to the other persin based on his public key (in hex)
$ ./client s --key=key.json  --receiver=1624de6220926bf0d423eac065b0807df5e3014b224222daf9d3345693130b663c8a06f449 --hash=Qmco6ZUSMApGbyVh2sWajtx2JTJiRnVDDbvPHeAEvsUgHT
Successfully send the hash Qmco6ZUSMApGbyVh2sWajtx2JTJiRnVDDbvPHeAEvsUgHT to 1624de6220926bf0d423eac065b0807df5e3014b224222daf9d3345693130b663c8a06f449

Query the hashes based on the receiver's key and the blockchain type
$ ./client q --key=other.json --type=spb
Files:
Qmco6ZUSMApGbyVh2sWajtx2JTJiRnVDDbvPHeAEvsUgHT
