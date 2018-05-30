There will be created two different validators.
The first validator will put the files in one single key called SPB (single key blockchain)
The second validator will put each file to be owned by one key only called OtoOPB (one key to one file blockchain)
However, both will work the same, except that multi-key validator will check that key and file are one to one

- Blockhain API
It will exchange file hashes based on a public key.
The transactions will have 3 actions 'send', 'add' and 'remove'

POST /Delivery 
RESPONSE 
Signature: signature
Data: {
    From : public key
    To: *public key
    action: string
    Files :[]string
}
REQUEST:
  Error scenarios:
    - For send and remove, the same file exist in other public key that does not equal to From
    - For OtoOPB, one-key for one file
    - The file does not exists in the IPFS
    - For OtoOPB, it has more than one file on the same delivery

POST /query
RESPONSE
Two options to return files for each blockchain:
1)For SPB when we want only the user or the admin to see what he have 
Signature: signature
Data: {
   From: public key
   Nonce: string
   Time: time // with this the query will validate if it created on the validated time, so no other person can use the same query again
   File: *string  // if it is empty then it will return all the files from owner of the public key or it will return yes
   User: *public key // if it is empty, it will check the files based on the "From" or else it will check the user as long as the "From" is authorized 
}

2) For OtoOPB
From: public key // it will return the file that the public key represent



