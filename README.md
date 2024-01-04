# How to Run

This guide will walk you through the process of setting up a light node with test tokens on the Mocha testnet and setting up a Wasmd chain.

## Step 1: Set Up a Celestia Node

1. Follow this [tutorial](https://docs.celestia.org/developers/node-tutorial) to set up a light node running with test tokens on the Mocha testnet.
2. Once you've completed the tutorial, start up your node.

## Step 2: Set Up a Wasmd Chain

We run a demo chain based on **Wasmd** [cosmos-sdk v0.50.x](https://github.com/notional-labs/cosmos-sdk-rollkit) and [Rollkit](https://github.com/notional-labs/rollkit). Rollkit is a modular framework for rollups, with an ABCI-compatible client interface.

Follow these steps to set up the Wasmd chain:

1. Clone the Wasmd repository: `git clone https://github.com/notional-labs/wasmd.git`
2. Check out the specific version: `git checkout v0.50.0-notional`
3. Tidy up the Go modules: `go mod tidy`
4. Install the necessary components: `make install`
5. Initialize the setup: `bash init.sh`
6. Start the setup: `bash start.sh`

```code=
git clone https://github.com/notional-labs/wasmd.git
git checkout v0.50.0-notional
go mod tidy
make install
bash init.sh
bash start.sh
```
