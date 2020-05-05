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
  done < <(./keytool genkeypair)

  echo "$1 node prikey $PrivateKey "
  if [ ! -d "$1" ]; then
      mkdir -p ./data/$1
  fi
  echo $PrivateKey > ./data/$1/nodekey
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
  done < <(./keytool genblskeypair)

  echo "$1 bls prikey $PrivateKey "
  echo $PrivateKey > ./data/$1/blskey
}

if [ x"$1" = x ]; then 
    echo "no cmd param!"
    exit 1
fi

if [[ "$1" =~ ^[1-9]+$ ]] ; then 
  true
else
  echo "input is wrong,must num"
  exit
fi



rm -rf ./data*


json_string=""



for((i=0;i<$1;i++));  
do   
funWithParam "data$i"
PublicKey1=$PublicKey
echo "data$i node pubkey $PublicKey1"

funWithParam2 "data$i"
BlsKey1=$PublicKey
echo "data$i bls pubkey $BlsKey1"

port=$(( 16788 + $i )) 

JSON_FMT='{\"node\":\"enode:\/\/Pubkey@127.0.0.1:Port\",\"blsPubKey\":\"Blskey\"}'


tmp1=${JSON_FMT/Pubkey/$PublicKey1}	
tmp2=${tmp1/Port/$port}	
tmp3=${tmp2/Blskey/$BlsKey1}	

if [ $i -ne 0 ] ;then
  json_string="$json_string,$tmp3"
else
  json_string="$tmp3"
fi

done  


printf "\nfinish node\n"

echo "begin set genesis.json"
sed  "s/nodereplace/$json_string/g"  ./genesis.bak | python3 -m json.tool > ./genesis.json

echo $2

if [ x"$2" = x ]; then 
    echo "done"
    exit 1
fi

if [[ "$2" =~ ^[1-9]+$ ]] ; then 
  true
else
  echo "input is wrong,must num"
  exit
fi

for((i=0;i<$2;i++));  
  do   
  tmp=$(( $1 + $i )) 

  funWithParam "data$tmp"
  PublicKey1=$PublicKey
  echo "data$tmp node pubkey $PublicKey1"

  funWithParam2 "data$tmp"
  BlsKey1=$PublicKey
  echo "data$tmp bls pubkey $BlsKey1"
done  

echo "done"
