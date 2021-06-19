if [ $# -eq 0 ]; then
  echo "argument missing"
  echo "'race' or 'empty'"
  exit 1
fi

SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

nodes=4
race=false
if [ $1 = "race" ]; then
  race=true
fi

str="genesis.data,"

echo "test2 $race"

go build main.go

for (( i=0; i < $nodes; ++i ))
do
  echo "deleting $i"
  rm -r ./_build/devnet_$i
done

sleep 0.5

for (( i=0; i < $nodes; ++i ))
do
    echo "running $i"
    xterm -e go run main.go --instance="devnet_$i" --network="devnet" --wallet-derive-delegated-stake="0,0,delegated" --exit
    mv ./_build/devnet_$i/DEV/delegated.delegatedStake ./_build/devnet_0/DEV/$i.delegatedStake
    echo "runned"

    str+="$i.delegatedStake"

    if [ $i != $(( nodes-1 )) ]; then
      str+=","
    fi

done

echo "creating genesis $str"
xterm -e go run main.go --instance="devnet_0" --network="devnet" --create-new-genesis="$str" --exit

sleep 0.1

for (( i=1; i < $nodes; ++i ))
do
  echo "copying genesis $i"
  cp ./_build/devnet_0/DEV/genesis.data ./_build/devnet_$i/DEV/genesis.data
done

for (( i=0; i < $nodes; ++i ))
do
  rm ./_build/devnet_$i/DEV/store/blockchain_store.bolt
done

sleep 0.1

for (( i=0; i < $nodes; ++i ))
do
  echo "opening $i"
  if $race ; then
    qterminal GORACE="log_path=/$SCRIPTPATH/../report"  -e go run -race main.go --instance="devnet_$i" --new-devnet --network="devnet" --tcp-server-port="523$i" --set-genesis="file" --staking &
  else
    xterm -e go run main.go --instance="devnet_$i" --new-devnet --network="devnet" --tcp-server-port="523$i" --set-genesis="file" --staking &
  fi
done

wait

echo "finished"
exit 1
