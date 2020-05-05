#!/bin/bash
# --pprof --pprofport 6062

posi=${1#data}

rpc_port=$(( 6771 + $posi )) 
port=$(( 16788 + $posi )) 

log="${1}.log"


if [ "$2" = "fast" ];then
    syncmode="fast"
else
    syncmode="full"
fi

net=""

if [ "$3" = "testnet" ];then
    net="--testnet"
fi


echo "start ${1},rpc_port:${rpc_port},port:${port},log:${log},syncmode:${syncmode},net:${3}"
./platon --identity "${1}"   --debug --db.nogc  --nodiscover  $net  --syncmode "${syncmode}"   --nodekey ./data/$1/nodekey   --cbft.blskey ./data/$1/blskey --rpc  --rpcaddr "0.0.0.0"  --rpcport $rpc_port --port $port --datadir ./data/$1  --rpcapi "txpool,platon,net,web3,miner,admin,personal"    --verbosity 4  2> ./log/$log

