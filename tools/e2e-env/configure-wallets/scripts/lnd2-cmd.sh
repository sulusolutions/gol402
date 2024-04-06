#!/bin/bash

LND2_LNCLI_ARGS="--network=regtest --tlscertpath /data/lnd2/tls.cert --macaroonpath /data/lnd2/data/chain/bitcoin/regtest/admin.macaroon --rpcserver lnd2:10010"
lncli $LND2_LNCLI_ARGS "$@"