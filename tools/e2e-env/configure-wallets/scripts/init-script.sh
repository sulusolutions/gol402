#!/bin/bash

wait_for_bitcoin() {
    until bitcoin-cmd.sh getblockchaininfo; do
        echo "Waiting for Bitcoin node..."
        sleep 5
    done
}

wait_for_lnd() {
    until lnd1-cmd.sh getinfo; do
        echo "Waiting for lnd1..."
        sleep 5
    done

    until lnd2-cmd.sh getinfo; do
        echo "Waiting for lnd2..."
        sleep 5
    done
}

# Function to check if the default wallet already exists and load it
bitcoin_wallet_exists_and_load() {
    # Try to load the wallet
    if bitcoin-cmd.sh loadwallet default 2>/dev/null; then
        echo "Default wallet loaded successfully."
        return 0
    else
        # Check if the wallet exists and loaded
        if bitcoin-cmd.sh listwallets | grep -q "default"; then
            echo "Default wallet exists but could not be loaded."
            return 0
        else
            echo "Default wallet does not exist."
            return 1
        fi
    fi
}

wait_for_bitcoin

# Create a new wallet and generate funds if the wallet doesn't exist
if ! bitcoin_wallet_exists_and_load; then
    echo "Creating new wallet..."
    bitcoin-cmd.sh createwallet default

    echo "Generating new address and funding..."
    bitcoin_wallet_address=$(bitcoin-cmd.sh getnewaddress default)
    bitcoin-cmd.sh generatetoaddress 101 $bitcoin_wallet_address
else 
    bitcoin_wallet_address=$(bitcoin-cmd.sh getaddressesbylabel default | jq -r "keys | .[0]")
    bitcoin-cmd.sh generatetoaddress 101 $bitcoin_wallet_address
fi

echo "Default wallet already exists, wallet balance:"
echo bitcoin-cmd.sh getbalance

wait_for_lnd

lnd1_address=$(lnd1-cmd.sh newaddress np2wkh | jq -r '.address')
bitcoin-cmd.sh sendtoaddress $lnd1_address 10
bitcoin-cmd.sh generatetoaddress 6 $lnd1_address
echo lnd1-cmd.sh walletbalance

# add 10 btc to lighting node 2
lnd2_address=$(lnd2-cmd.sh newaddress np2wkh | jq -r '.address')
bitcoin-cmd.sh sendtoaddress $lnd2_address 10
bitcoin-cmd.sh generatetoaddress 6 $lnd2_address
echo lnd2-cmd.sh walletbalance

# Setup Connection between nodes and creates channels
AMOUNT=50000 # Example amount in satoshis
# Connect lnd1 to lnd2
LND2_PEER=$(lnd2-cmd.sh getinfo | jq -r '.uris[0]')
echo "Creating connection from lnd1 to lnd2 at address ${LND2_PEER}"
NODE2_KEY=$(echo $LND2_PEER | cut -d'@' -f1)
HOST2_PORT=$(echo $LND2_PEER | cut -d'@' -f2)
lnd1-cmd.sh connect $LND2_PEER

# Connect lnd2 to lnd1
LND1_PEER=$(lnd1-cmd.sh getinfo | jq -r '.uris[0]')
echo "Creating connection from lnd2 to lnd1 at address ${LND1_PEER}"
NODE1_KEY=$(echo $LND1_PEER | cut -d'@' -f1)
HOST1_PORT=$(echo $LND1_PEER | cut -d'@' -f2)
lnd2-cmd.sh connect $LND1_PEER

# Create channels
lnd1-cmd.sh openchannel --node_key=$NODE2_KEY --connect=$HOST2_PORT --local_amt=$AMOUNT
bitcoin-cmd.sh generatetoaddress 6 $bitcoin_wallet_address
bitcoin-cmd.sh generatetoaddress 6 $lnd1_address
bitcoin-cmd.sh generatetoaddress 6 $lnd2_address
sleep 5 # Wait for the nodes to be in sync
lnd2-cmd.sh openchannel --node_key=$NODE1_KEY --connect=$HOST1_PORT --local_amt=$AMOUNT


# Generate bitcoins in the background to speed up anything related to the bitcoin
while true
do
    bitcoin-cmd.sh generatetoaddress 6 $bitcoin_wallet_address
    bitcoin-cmd.sh generatetoaddress 6 $lnd1_address
    bitcoin-cmd.sh generatetoaddress 6 $lnd2_address
    sleep 30
done
