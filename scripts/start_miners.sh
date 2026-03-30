#!/bin/bash

# devgen.car should be copied from the first node. also, the address of the client and miner of the first node still need to be known.

#step1: start the daemon
./lotus daemon --genesis=devgen.car

#step2: connect to the first node (to get sync, otherwise, the miner can't be initilized)
./lotus net connect </ip4/172.29.244.239/tcp/36995/p2p/12D3KooWJ7Ec6h4yWeqYgqCaKf8PdZ6kPtQq2qBLepeLTE46h5vq> #address of the client 
./lotus net connect </ip4/172.29.244.239/tcp/46591/p2p/12D3KooWGKYdW2FBWyLdAdQ7Yh9Dki6swXHwMRFPTwxoRDpJrYiu> #address of the miner

#step3: create a wallet 
./lotus wallet new bls
#then, return to the first node and exec: ./lotus send <wallet address> <balance> to send some balance to this wallet, otherwise, the miner can't be initilized

#step4: init the miner 
./lotus-miner init --owner=<wallet address> --worker=<wallet address> --no-local-storage --sector-size=8MiB

#step5: start the miner and connect to the first node
./lotus-miner run
./lotus-miner net connect </ip4/172.29.244.239/tcp/36995/p2p/12D3KooWJ7Ec6h4yWeqYgqCaKf8PdZ6kPtQq2qBLepeLTE46h5vq> #address of the client 
./lotus-miner net connect </ip4/172.29.244.239/tcp/46591/p2p/12D3KooWGKYdW2FBWyLdAdQ7Yh9Dki6swXHwMRFPTwxoRDpJrYiu> #address of the miner

#the following command can get the miner ID
./lotus-miner info

#step6: attach storage 

mkdir unseal
mkdir sealed
./lotus-miner storage attach --init --store /home/jiahao/go/src/lotus/unseal
./lotus-miner storage attach --init --seal /home/jiahao/go/src/lotus/sealed
