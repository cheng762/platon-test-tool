#!/bin/bash
# --pprof --pprofport 6062

posi=${1#data}

rpc_port=$(( 6771 + $posi )) 
port=$(( 16789 + $posi )) 

log="${1}.log"


if [ "$2" = "fast" ];then
    syncmode="fast"
else
    syncmode="full"
fi

echo "start ${1},rpc_port:${rpc_port},port:${port},log:${log},syncmode:${syncmode}"
./platon --identity "${1}"   --debug --db.nogc  --nodiscover  --syncmode "${syncmode}"   --nodekey ./$1/nodekey   --cbft.blskey ./$1/blskey --rpc  --rpcaddr "0.0.0.0"  --rpcport $rpc_port --port $port --datadir ./$1  --rpcapi "txpool,platon,net,web3,miner,admin,personal"    --verbosity 4  2> ./log/$log

