#!/bin/bash

BINARY=${1:-wasmd}
CONTINUE=${CONTINUE:-"false"}
HOME_DIR=mytestnet
ENV=${ENV:-""}
CHAIN_ID="test"
KEYRING="test"
KEY="test0"
DENOM=${2:-stake}


if [ "$CONTINUE" == "true" ]; then
    $BINARY start --p2p.pex=false --home $HOME_DIR --log_level debug
    exit 0
fi

rm -rf mytestnet
pkill wasmd


SED_BINARY=sed
# check if this is OS X
if [[ "$OSTYPE" == "darwin"* ]]; then
    # check if gsed is installed
    if ! command -v gsed &> /dev/null
    then
        echo "gsed could not be found. Please install it with 'brew install gnu-sed'"
        exit
    else
        SED_BINARY=gsed
    fi
fi


# Function updates the config based on a jq argument as a string
update_test_genesis () {
    # update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
    cat $HOME_DIR/config/genesis.json | jq "$1" > $HOME_DIR/config/tmp_genesis.json && mv $HOME_DIR/config/tmp_genesis.json $HOME_DIR/config/genesis.json
}

$BINARY init --chain-id $CHAIN_ID moniker --home $HOME_DIR
echo "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry" | $BINARY keys add $KEY --keyring-backend $KEYRING  --recover --home $HOME_DIR

# Allocate genesis accounts (cosmos formatted addresses)
$BINARY genesis add-genesis-account $KEY "1000000000000${DENOM}" --keyring-backend $KEYRING --home $HOME_DIR


# enable rest server and swagger
$SED_BINARY -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0stake"/' $HOME_DIR/config/app.toml
$SED_BINARY -i 's/timeout_commit = "5s"/timeout_commit = "500ms"/' $HOME_DIR/config/config.toml


# Sign genesis transaction
$BINARY genesis gentx $KEY "1000000${DENOM}" --keyring-backend $KEYRING --chain-id $CHAIN_ID --home $HOME_DIR

# Collect genesis tx
$BINARY genesis collect-gentxs --home $HOME_DIR
# Run this to ensure everything worked and that the genesis file is setup correctly

$BINARY genesis validate-genesis --home $HOME_DIR
screen -L -dms wasmd $BINARY start --p2p.pex=false --home $HOME_DIR 

