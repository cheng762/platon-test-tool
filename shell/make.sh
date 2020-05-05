#!/bin/bash
cd $GOPATH/src/github.com/PlatONnetwork/PlatON-Go || exit
make platon
cp ./build/bin/platon  /Users/chenglin/develop/platon-node/platon
echo "cp to dir finish"