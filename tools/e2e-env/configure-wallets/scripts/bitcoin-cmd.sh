#!/bin/bash

BITCOIN_CLI_ARGS="-regtest -rpcpassword=pass -rpcuser=user -rpcconnect=bitcoind"
bitcoin-cli $BITCOIN_CLI_ARGS "$@"