Guide

Generate a key in a file
$ client g --filename key.json

Add a json to the blockchain and receive hash
$ client a --key=key.json  --type=json --input='{"coin":1}'
successfully added the QmfJudNdQPrGLcaxaxvH1eCMLU7buTAFCcFGrt4etWv7rq

Remove a hash
$ client r --key=key.json --hash=QmfJudNdQPrGLcaxaxvH1eCMLU7buTAFCcFGrt4etWv7rq
successfully removed the QmfJudNdQPrGLcaxaxvH1eCMLU7buTAFCcFGrt4etWv7rq

Query the hashes based on the key and the blockchain type
$ client q --key=key.json --type=spb
Files:
QmfJudNdQPrGLcaxaxvH1eCMLU7buTAFCcFGrt4etWv7rq

$ client q --key=key.json --type=spb --hash=QmfJudNdQPrGLcaxaxvH1eCMLU7buTAFCcFGrt4etWv7rq
Files:
QmfJudNdQPrGLcaxaxvH1eCMLU7buTAFCcFGrt4etWv7rq
