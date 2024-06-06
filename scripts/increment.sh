#!/bin/bash

BINARY=${1:-wasmd}
CONTINUE=${CONTINUE:-"false"}
HOME_DIR=mytestnet
ENV=${ENV:-""}
CHAIN_ID="test"
KEYRING="test"
KEY="test0"
DENOM=${2:-stake}

echo "Waiting for the node to start..."
sleep 3

WALLET=$($BINARY keys show $KEY -a --keyring-backend $KEYRING --home $HOME_DIR)
echo "Wallet: $WALLET"

# Deploy the smart contract on chain to test the callbacks. (find the source code under the following url: `~/scripts/tests/ibc-hooks/counter/src/contract.rs`)
echo "Deploying counter contract"
TX_HASH=$($BINARY tx wasm store $(pwd)/scripts/counter/artifacts/counter.wasm --from $WALLET  --home $HOME_DIR --gas 10000000  --keyring-backend test -y  -o  json| jq -r '.txhash')

echo "Tx hash: $TX_HASH"

sleep 1
CODE_ID=$($BINARY query tx $TX_HASH -o json --home $HOME_DIR  | jq -r '.logs[0].events[1].attributes[1].value')
echo "Code id: $CODE_ID"

RANDOM_HASH=$(hexdump -vn16 -e'4/4 "%08X" 1 "\n"' /dev/urandom)
TX_HASH=$($BINARY tx wasm instantiate2 $CODE_ID '{"count": 0}' $RANDOM_HASH --no-admin --label="Label with $RANDOM_HASH" --from $WALLET --home $HOME_DIR --keyring-backend test -y --gas 10000000 --fees 6000000$DENOM -o json | jq -r '.txhash')

echo "TX hash: $TX_HASH"
sleep 1
CONTRACT_ADDRESS=$($BINARY query tx $TX_HASH -o json --home $HOME_DIR | jq -r '.logs[0].events[1].attributes[0].value')
echo "Contract address: $CONTRACT_ADDRESS"

# increment the counter 
wasmd tx wasm execute $CONTRACT_ADDRESS  '{"increment":{}}' --from $WALLET --home $HOME_DIR --keyring-backend test -y

sleep 1

$BINARY query wasm contract-state smart "$CONTRACT_ADDRESS" '{"get_count": {"addr": "'"$WALLET"'"}}'


