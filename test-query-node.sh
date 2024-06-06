#!/bin/bash

# Define the node URL
NODE_URL="http://localhost:26657/"

CONTRACT_ADDRESS_KEY="021ab20506d0f843d11b6ab41f413a34df904b4b698b96e5da3aa903a06b01c1c40000000573746174651b19f123987e541388360466451c2c4d24d9d2fa"

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