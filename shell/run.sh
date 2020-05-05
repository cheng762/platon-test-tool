#!/bin/bash
./make.sh

cnode=4
node=1

./generate.sh $cnode $node



for((i=0;i<$(( cnode + node ));i++));  
do   
  ./init.sh "data$i"
done  



