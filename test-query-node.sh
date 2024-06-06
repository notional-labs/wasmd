#!/bin/bash

# Define the node URL
NODE_URL="http://localhost:26657/"

CONTRACT_ADDRESS_KEY="${1}"

# Create the payload
cat > payload.json <<EOF
{
  "jsonrpc": "2.0",
  "id": 864499175752,
  "method": "abci_query",
  "params": {
    "path": "/store/wasm/key",
    "data": "$CONTRACT_ADDRESS_KEY",
    "prove": false
  }
}
EOF

# Perform the curl command
curl -d @payload.json $NODE_URL -H 'Content-Type: application/json'