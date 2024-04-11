#!/bin/bash

LND1_LNCLI_ARGS="--network=regtest --tlscertpath /data/lnd1/tls.cert --macaroonpath /data/lnd1/data/chain/bitcoin/regtest/admin.macaroon --rpcserver lnd1:10009"
lncli $LND1_LNCLI_ARGS "$@"