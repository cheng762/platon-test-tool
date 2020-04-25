#!/bin/bash


PublicKey=""

funWithParam(){
  while read line
  do
    prk=$(echo $line  | grep "PrivateKey"|awk '{print $2}') 
    if [ "$prk" != ""  ]
    then
    PrivateKey=$prk
    fi
    puk=$(echo $line  | grep "PublicKey"|awk '{print $3}')
    if [ "$puk" != ""  ]
    then
    PublicKey=$puk
    fi
  done < <(./ethkey genkeypair)

  echo "$1 set nodekey $PrivateKey "
  if [ ! -d "$1" ]; then
      mkdir -p $1/platon
  fi
  echo $PrivateKey > $1/nodekey
}

funWithParam2(){
  while read line
  do
    prk=$(echo $line  | grep "PrivateKey"|awk '{print $2}') 
    if [ "$prk" != ""  ]
    then
    PrivateKey=$prk
    fi
    puk=$(echo $line  | grep "PublicKey"|awk '{print $3}')
    if [ "$puk" != ""  ]
    then
    PublicKey=$puk
    fi
  done < <(./ethkey genblskeypair)

  echo "$1 set blskey $PrivateKey "
  echo $PrivateKey > $1/blskey
}


echo "begin new accout"
#ADD1=$( ./platon --datadir data0 account new -password passwd | grep "Address"|awk '{print $2}'| sed  's/{\(.*\)}/\1/') 
#ADD2=$( ./platon --datadir data1 account new -password passwd | grep "Address"|awk '{print $2}'| sed  's/{\(.*\)}/\1/') 
#ADD3=$( ./platon --datadir data2 account new -password passwd | grep "Address"|awk '{print $2}'| sed  's/{\(.*\)}/\1/') 
#ADD4=$( ./platon --datadir data3 account new -password passwd | grep "Address"|awk '{print $2}'| sed  's/{\(.*\)}/\1/') 


echo "begin generate node private and public key"
funWithParam data0
PublicKey1=$PublicKey
echo "generate node0 publickey $PublicKey1"

funWithParam data1
PublicKey2=$PublicKey
echo "generate node1 publickey $PublicKey2"


funWithParam data2
PublicKey3=$PublicKey
echo "generate node2 publickey $PublicKey3"


funWithParam data3
PublicKey4=$PublicKey
echo "generate node3 publickey $PublicKey4"




funWithParam2 data0
BlsKey1=$PublicKey
echo "generate node0 blskey $BlsKey1"


funWithParam2 data1
BlsKey2=$PublicKey
echo "generate node1 blskey $BlsKey2"


funWithParam2 data2
BlsKey3=$PublicKey
echo "generate node2 blskey $BlsKey3"


funWithParam2 data3
BlsKey4=$PublicKey
echo "generate node3 blskey $BlsKey4"



echo "finish node"

echo "begin set genesis.json"
sed -e "s/your-node-pubkey1/$PublicKey1/g" -e "s/your-node-pubkey2/$PublicKey2/g" -e "s/your-node-pubkey3/$PublicKey3/g" -e "s/your-node-pubkey4/$PublicKey4/g" -e "s/blsPubKey1/$BlsKey1/g" -e "s/blsPubKey2/$BlsKey2/g" -e "s/blsPubKey3/$BlsKey3/g" -e "s/blsPubKey4/$BlsKey4/g"  ./genesis.bak > ./genesis.json
echo "done"


mkdir -p data5/platon

funWithParam data5
PublicKey5=$PublicKey
echo "generate node5 publickey $PublicKey5"

funWithParam2 data5
BlsKey5=$PublicKey
echo "generate node5 blskey $BlsKey5"


