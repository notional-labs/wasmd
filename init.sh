#!/bin/sh
# set variables for the chain
VALIDATOR_NAME=validator1
CHAIN_ID=celeswasm
KEY_NAME=celeswasm-key
TOKEN_AMOUNT="10000000000000000000000000uwasm"
STAKING_AMOUNT=1000000000uwasm
CHAINFLAG="--chain-id ${CHAIN_ID}"
TXFLAG="--chain-id ${CHAIN_ID} --gas-prices 0uwasm --gas auto --gas-adjustment 1.3"
# create a random Namespace ID for your rollup to post blocks to
NAMESPACE=$(openssl rand -hex 8)
echo $NAMESPACE
# query the DA Layer start height, in this case we are querying
# an RPC endpoint provided by Celestia Labs. The RPC endpoint is
# to allow users to interact with Celestia's core network by querying
# the node's state and broadcasting transactions on the Celestia
# network. This is for Arabica, if using another network, change the RPC.
RPC_ADDRESS="127.0.0.1:26657"
DA_BLOCK_HEIGHT=$(curl "$RPC_ADDRESS/block" | jq -r '.result.block.header.height')
echo $DA_BLOCK_HEIGHT
# reset any existing genesis/chain data
wasmd tendermint unsafe-reset-all
wasmd init $VALIDATOR_NAME --chain-id $CHAIN_ID
# update wasmd configuration files to set chain details and enable necessary settings
# the sed commands here are editing various configuration settings for the wasmd instance
# such as setting minimum gas prices, enabling the api, setting the chain id, setting the rpc address,
# adjusting time constants, and setting the denomination for bonds and minting.
sed -i'' -e 's/^minimum-gas-prices *= .*/minimum-gas-prices = "0uwasm"/' "$HOME"/.wasmd/config/app.toml
sed -i'' -e '/\[api\]/,+3 s/enable *= .*/enable = true/' "$HOME"/.wasmd/config/app.toml
sed -i'' -e "s/^chain-id *= .*/chain-id = \"$CHAIN_ID\"/" "$HOME"/.wasmd/config/client.toml
sed -i'' -e '/\[rpc\]/,+3 s/laddr *= .*/laddr = "tcp:\/\/0.0.0.0:26657"/' "$HOME"/.wasmd/config/config.toml
sed -i'' -e 's/"time_iota_ms": "1000"/"time_iota_ms": "10"/' "$HOME"/.wasmd/config/genesis.json
sed -i'' -e 's/bond_denom": ".*"/bond_denom": "uwasm"/' "$HOME"/.wasmd/config/genesis.json
sed -i'' -e 's/mint_denom": ".*"/mint_denom": "uwasm"/' "$HOME"/.wasmd/config/genesis.json
# add a key to keyring-backend test
wasmd keys add $KEY_NAME --keyring-backend test
# add a genesis account
wasmd genesis add-genesis-account $KEY_NAME $TOKEN_AMOUNT --keyring-backend test
# set the staking amounts in the genesis transaction
wasmd genesis gentx $KEY_NAME $STAKING_AMOUNT --chain-id $CHAIN_ID --keyring-backend test
# collect gentxs
wasmd genesis collect-gentxs
# copy centralized sequencer address into genesis.json
# Note: validator and sequencer are used interchangeably here
ADDRESS=$(jq -r '.address' ~/.wasmd/config/priv_validator_key.json)
PUB_KEY=$(jq -r '.pub_key' ~/.wasmd/config/priv_validator_key.json)
jq --argjson pubKey "$PUB_KEY" '. + {"validators": [{"address": "'$ADDRESS'", "pub_key": $pubKey, "power": "1000", "name": "Rollkit Sequencer"}]}' ~/.wasmd/config/genesis.json > temp.json && mv temp.json ~/.wasmd/config/genesis.json
# generate an authorization token for the light client using the celestia binary
# this is for Arabica, if using another network, change the network name
export AUTH_TOKEN=$(docker exec $(docker ps -q)  celestia bridge --node.store /home/celestia/bridge/ auth admin)
# start the chain
wasmd start --rollkit.aggregator true --rollkit.da_layer celestia --rollkit.da_config='{"base_url":"http://localhost:26658","timeout":60000000000,"fee":600000,"gas_limit":6000000,"auth_token":"'$AUTH_TOKEN'"}' --rollkit.namespace_id $NAMESPACE --rollkit.da_start_height $DA_BLOCK_HEIGHT  --rpc.laddr tcp://127.0.0.1:36657 --p2p.laddr "0.0.0.0:36656"


# follow the cosmwasm tutorial to deploy the cw_nameservice contract
# TX_HASH=$(wasmd tx wasm store artifacts/cw_nameservice.wasm --from celeswasm-key --keyring-backend test --chain-id celeswasm --gas-prices 0uwasm --gas auto --gas-adjustment 1.3 --node http://127.0.0.1:26657 --output json -y | jq -r '.txhash') && echo $TX_HASH

# wasmd query tx --type=hash $TX_HASH --node http://127.0.0.1:36657 --output json | jq -r '.events[-1].attributes[1].value'
# INIT='{"purchase_price":{"amount":"100","denom":"uwasm"},"transfer_price":{"amount":"999","denom":"uwasm"}}'
# wasmd tx wasm instantiate $CODE_ID "$INIT" --from celeswasm-key --keyring-backend test --label "name service" --chain-id celeswasm --gas-prices 0uwasm --gas auto --gas-adjustment 1.3 -y --no-admin --node http://127.0.0.1:36657

#  wasmd query wasm list-contract-by-code $CODE_ID --output json --node http://127.0.0.1:36657  

## Get the instantiated contract address

# CONTRACT=$(wasmd query wasm list-contract-by-code $CODE_ID --output json --node http://127.0.0.1:36657 | jq -r '.contracts[-1]') 

#  wasmd tx wasm execute $CONTRACT "$REGISTER" --amount 100uwasm --from celeswasm-key --chain-id celeswasm --gas-prices 0uwasm --gas auto --gas-adjustment 1.3 --node http://127.0.0.1:36657 --keyring-backend test -y

# Query the registered name
# wasmd query wasm contract-state smart $CONTRACT "$NAME_QUERY" --node http://127.0.0.1:36657 --output json       1 ↵ 


