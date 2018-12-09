# gethalyzer
Demo tool meant to analyze custom geth logs for an interview with Kaliedo 

## Open Questions

* Should I make a new log level for the logs?
* When an error occurs in the worker, why do we shift() vs pop() some tx requests vs others? - e.g. miner/miner/worker.go:commitTransactions:777

---

## Testing

chainid = "hashrebel" = 0x48 0x61 0x73 0x68 0x20 0x52 0x65 0x62 0x65 0x6C = 0x4861736820526562656C = 4.8617×10¹⁹

### Tests Cases

A contract which runs out of gas
A contract which runs out of gas due to too much data
Nonce is too high
Nonce is too low
Non replay protected transaction (pre eip155)

Execution depth test (core/vm/evm.go:378)
Contract collision (core/vm/evm.go:395, not sure if this is possible without costume test hacks)
Incompatible EVM code (core/vn/evm.go:69)
Non refundable contract failure

---

## Demo

### Setup Costume Private node

Download branch (TODO: put branch here) to your local GOPATH directory

```bash
cd $GOPATH
git clone git@github.com:HashRebel/go-ethereum.git
git checkout mining-trx-logs

make geth
'''

'''bash
# where gethdev is an alias pointing to the dev build of geth since I already have geth installed
gethdev \
--datadir node1/ \
--syncmode 'full' \
--targetgaslimit 94000000 \
--port 30311 \
--rpc \
--rpcaddr 'localhost' \
--rpcport 8501 \
--rpcapi 'admin,personal,db,debug,eth,net,web3,txpool,miner' \
--networkid 1515 \
--gasprice '1' \
--unlock '0x81edc9fc800e1b9c76be2f83e5c1dcc73f62980d' \
--password node1/password.txt \
--debug \
--mine \
2>> node1-miner.log

#--vmodule miner/*=5 \
```