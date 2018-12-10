# gethalyzer

Demo tool meant to analyze custom geth logs for an interview with Kaliedo.

For the sake of time, I have am only supporting a few basic tests. The code base is pretty messy and given more time I would like to organize things into libraries and create a more extensible test suit.

---

## Testing

This section details how to setup and run the tests.

### Test Accounts

#### Main

account: 81edc9fc800e1b9c76be2f83e5c1dcc73f62980d
private: 6dfafde8135f35253f6482d80df34bd3ec52ad5733ab1edda4ba46110663d7d4

#### Hash Rebel

account: f61bb995b5fb19aa0a38c7ecc9b52ff6199a69ca
private: 7676e3fd889435ef3349ee57775f1cc2c151446f2d1660bce48d3ed00a7bafbd

#### Secondary

Account: c567982f00db259c2af4a6c7ed7b7e8ba393d695

### Tests Cases

* [x] Standard eth transaction
* [x] Standard contract creation transaction
* [ ] Update a contract update transaction
* [x] A contract which runs out of gas
* [ ] A contract which runs out of gas due to too much data
* [x] Nonce is too high (I think this can only happen when there is a race condition against the transaction pool and the miner)
* [x] Nonce is too low  (This doesn't make it past the transaction pool)
* [ ] Execution depth test (core/vm/evm.go:378)
* [ ] Contract collision (core/vm/evm.go:395, not sure if this is possible without costume test hacks)
* [ ] Incompatible EVM code (core/vn/evm.go:69)
* [ ] Non refundable contract failure

---

## Demo

### Setup for my private server

My custom geth can be found [here](https://github.com/HashRebel/go-ethereum/tree/mining-trx-logs)

chainid = "hashrebel" = 0x48 0x61 0x73 0x68 0x20 0x52 0x65 0x62 0x65 0x6C = 0x4861736820526562656C = 4.8617×10¹⁹

Download branch to your local go directory (e.g. ```$GOPATH/src/github.com/ethereum```) and build geth

```bash
cd $GOPATH
mkdir -p $GOPATH/src/github.com/ethereum
cd $GOPATH/src/github.com/ethereum
git clone git@github.com:HashRebel/go-ethereum.git
git checkout mining-trx-logs

make geth
```

I also made a quick alias to help run the right development version of geth

```bash
# Ubuntu 18.04
echo alias gethdev='$GOPATH/src/github.com/ethereum/go-ethereum/build/bin/geth' >> $HOME/.bash_aliases
```

I use this [guide](https://hackernoon.com/setup-your-own-private-proof-of-authority-ethereum-network-with-geth-9a0a3750cda8) to setup a private server. I skipped all bootnode steps and only created one node.

Next step is to start up the dev geth miner and pipe the logs into ```node1-miner.log```

```bash
# where gethdev is an alias pointing to the dev build of geth since I already have geth installed
# Completely opened the rpc API for testing purposes.
# tragetgaslimit keeps the gas static
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
```

### Setup ethalyzer tool in order to kick off tests and see the log output

Download and build ethalyzer

```bash
cd $GOPATH
mkdir -p $GOPATH/src/github.com/HashRebel
cd $GOPATH/src/github.com/HashRebel
git clone git@github.com:bhenze/gethalyzer.git

go build
./gethalyzer
./gethalyzer -l {path-to-log-file}
```

---

## Retrospective

Things I learned, things I would have done different and open questions.

For the sake of time and because I am still learning golang, I made took lots of shortcuts and created one monolithic program. I would have made this much more modular if I had more time. (don't judge the copy/paste too much!)

### Open Questions

* ~~Should I make a new log level for the logs? (answer: ended up using a stamp in the logs~~
* When an error occurs in the worker, why do we shift() vs pop() some tx requests vs others? - e.g. miner/miner/worker.go:commitTransactions:777 (too low of nonce would mean it is taken, but too high and you can just set the nonce higher?)
* I would expect the balances on my accounts to change once the transaction has been added to block. However, after the block is mined the balances for both accounts remain this same. What am I missing here?
* When I run the out of gas test it looks like the miner attempts to put the transaction in the block a whole bunch of time. How does it determine retries?

### Gethalyzer 2.0

* Learn more about the transaction pool.
* Cleanup and beautify the log returns from geth.
* Would be kinda fund to have some sort of stack visualization tool. Maybe another add a ```stack``` command to gethalyzer.
* Setup docker containers in order to easily orchestrate and replicate my private test data.
* Would cleanup and restructure gethalyzer (It is just one big monolithic mess at the moment).
  * Package for handling the geth communication
  * Package for all managing and running the actual tests
  * More extensible test suite
  * Test function should take in new interface in order to pass around data