#!/bin/sh
# query the DA Layer start height, in this case we are querying
# an RPC endpoint provided by Celestia Labs. The RPC endpoint is
# to allow users to interact with Celestia’s core network by querying
# the node’s state and broadcasting transactions on the Celestia
# network. This is for Arabica, if using another network, change the RPC.
DA_BLOCK_HEIGHT=$(curl https://rpc-mocha.pops.one/block | jq -r '.result.block.header.height')
echo $DA_BLOCK_HEIGHT
# generate an authorization token for the light client using the celestia binary
# this is for Arabica, if using another network, change the network name
export AUTH_TOKEN=$(celestia light auth write --p2p.network mocha)
# start the chain
wasmd start --rollkit.aggregator true --rollkit.da_address 127.0.0.1:26650 --rollkit.da_start_height $DA_BLOCK_HEIGHT --rollkit.block_time 1s