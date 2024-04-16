# 2 EVM based blockchain network setup

## Prerequisite

- [Go v1.21+](https://go.dev/dl/)
- [Geth v1.13.x](https://github.com/ethereum/go-ethereum/tree/v1.13.14)

# Setup 2 Blockchain Networks

## Setup Environment

### Setup Environment variables

Copy .env.example file to .env and setup environment variables.

```
export GETH_DIR=$HOME/.geth
export KEYSTORE=$GETH_DIR/keystore
export CLEF_CONFIG_DIR=$GETH_DIR/clef

export CHAIN_DIR_A=$GETH_DIR/chain-a
export CHAIN_ID_A=90000001
export HTTP_PORT_A=8545
export WS_PORT_A=8546
export DISCOVERY_PORT_A=30303
export AUTHRPC_PORT_A=8551
export ETHERBASE_A="0xd1bf8388759c5adbaa2ca48bd161da7e163167c9"

export CHAIN_DIR_B=$GETH_DIR/chain-b
export CHAIN_ID_B=90000002
export HTTP_PORT_B=8645
export WS_PORT_B=8646
export DISCOVERY_PORT_B=31303
export AUTHRPC_PORT_B=8651
export ETHERBASE_B="0x1064b11d2a9691576e1e194557ed6a30a836b642"
```

### Setup Directory

```
mkdir $GETH_DIR
mkdir $KEYSTORE
mkdir $CLEF_CONFIG_DIR
```

## Setup Clef

### Init

```
clef --keystore $KEYSTORE --configdir $CLEF_CONFIG_DIR init
```

### Create Accounts

```
clef newaccount --keystore $KEYSTORE
clef newaccount --keystore $KEYSTORE
```

### Set Etherbase of Chain A and B

```
export ETHERBASE_A=
export ETHERBASE_B=
```

### Store passwords in clef

```
clef --keystore $KEYSTORE --configdir $CLEF_CONFIG_DIR --chainid $CHAIN_ID_A setpw $ETHERBASE_A
clef --keystore $KEYSTORE --configdir $CLEF_CONFIG_DIR --chainid $CHAIN_ID_B setpw $ETHERBASE_B
```

### Write Clef Approval Rules

```
function OnSignerStartup(info) {}

function ApproveListing() {
    return 'Approve';
}

function ApproveSignData(r) {
    if (r.content_type == 'application/x-clique-header') {
        for (var i = 0; i < r.messages.length; i++) {
            var msg = r.messages[i];
            if (msg.name == 'Clique header' && msg.type == 'clique') {
                return 'Approve';
            }
        }
    }
    return 'Reject';
}

// Approve transactions to a certain contract if value is below a certain limit
function ApproveTx(r) {
	return 'Approve';
}

function OnApprovedTx(resp) {
    var value = big(resp.tx.value);
    var txs = [];

    // Load stored transactions
    var stored = storage.get('txs');
    if (stored != '') {
      txs = JSON.parse(stored);
    }

    // Add this to the storage
    txs.push({ tstamp: new Date().getTime(), value: value });
    storage.put('txs', JSON.stringify(txs));
}
```

### Attest Rules

```
// Rules for Chain A
clef --keystore $KEYSTORE --configdir $CLEF_CONFIG_DIR --chainid $CHAIN_ID_A --suppress-bootwarn  attest  `shasum -a 256 $CLEF_CONFIG_DIR/rules.js | cut -f1`

// Rules for Chain B
clef --keystore $KEYSTORE --configdir $CLEF_CONFIG_DIR --chainid $CHAIN_ID_B --suppress-bootwarn  attest  `shasum -a 256 $CLEF_CONFIG_DIR/rules.js | cut -f1`
```

### Run Clef with Rules
```
// Terminal A
clef --keystore $KEYSTORE --configdir $CLEF_CONFIG_DIR --chainid $CHAIN_ID_A --rules $CLEF_CONFIG_DIR/rules.js

// Terminal A
clef --keystore $KEYSTORE --configdir $CLEF_CONFIG_DIR --chainid $CHAIN_ID_B --rules $CLEF_CONFIG_DIR/rules.js
```

## Setup Chain A

### Initialize Geth

```
geth init --datadir $CHAIN_DIR_A $CHAIN_DIR_A/genesis.json
geth init --datadir $CHAIN_DIR_B $CHAIN_DIR_B/genesis.json
```

### Start Chain

```
// Chain A
geth --datadir $CHAIN_DIR_A \
--port $DISCOVERY_PORT_A \
--authrpc.addr localhost \
--authrpc.port $AUTHRPC_PORT_A \
--authrpc.vhosts localhost \
--authrpc.jwtsecret $CHAIN_DIR_A/jwtsecret \
--http --http.port $HTTP_PORT_A --http.api eth,net,web3 \
--ws --ws.port $WS_PORT_A \
--signer=$CLEF_CONFIG_DIR/clef.ipc \
--mine --miner.etherbase=$ETHERBASE_A \
--nodiscover --networkid $CHAIN_DIR_B \
--verbosity 1

// Chain B
geth --datadir $CHAIN_DIR_B \
--port $DISCOVERY_PORT_B \
--authrpc.addr localhost \
--authrpc.port $AUTHRPC_PORT_B \
--authrpc.vhosts localhost \
--authrpc.jwtsecret $CHAIN_DIR_B/jwtsecret \
--http --http.port $HTTP_PORT_B --http.api eth,net,web3 \
--ws --ws.port $WS_PORT_B \
--signer=$CLEF_CONFIG_DIR/clef.ipc \
--mine --miner.etherbase=$ETHERBASE_B \
--nodiscover --networkid $CHAIN_DIR_B \
--verbosity 1
```

### Attach Chain A IPC

```
// Chain A
geth attach $CHAIN_DIR_A/geth.ipc

// Chain B
geth attach $CHAIN_DIR_B/geth.ipc
```

### Send Ether to contract deployer

```
eth.sendTransaction({ from: eth.accounts[0], to: "0x3605Ca39aC83b8F559B64C453feC6A22AEF99259", value: 1000000000000000000, gas: 21000 });
eth.sendTransaction({ from: eth.accounts[1], to: "0x3605Ca39aC83b8F559B64C453feC6A22AEF99259", value: 1000000000000000000, gas: 21000 });
```
