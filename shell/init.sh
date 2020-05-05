#!/bin/bash
for i in "$@"; do
    echo "begin init $i"
    if  test  -d ./data/${i} ;then
      rm -rf ./data/"$i"/platon/
    fi
    ./platon --identity "platon" --verbosity 5 --debug  --datadir ./data/$i   init ./genesis.json
    echo "init finsh $i"
done
