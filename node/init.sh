#!/bin/bash
for i in "$@"; do
    echo "begin init $i"
    if  test  -d ${i} ;then
      cd $i
      rm -rf ./platon/
      cd ..
    fi
    ./platon --identity "platon" --verbosity 5 --debug  --datadir ./$i   init ./genesis.json
    echo "init finsh $i"
done
