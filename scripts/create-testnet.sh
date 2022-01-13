if [ $# -eq 0 ]; then
  echo "argument missing"
  echo "include 'normal' for starting in normal mode"
  echo "include 'race' to enable the race detection"
  echo "include 'pprof' to enable debugging using profiling"
  echo "include 'debug' to enable debug info"
  exit 1
fi

SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

nodes=4
race=false
extraArgs=""

for arg in $@
do
  if [ $arg = "pprof" ]; then
    extraArgs+=" --pprof "
  fi
  if [ $arg = "race" ]; then
      race=true
  fi
  if [ $arg = "debug" ]; then
      extraArgs=" --debug "
    fi
done

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
    xterm -e go run main.go --instance="devnet" --instance-id="$i" --network="devnet" --wallet-derive-delegated-stake="0,0,delegated.stake" --exit
    mv ./_build/devnet_$i/DEV/delegated.stake ./_build/devnet_0/DEV/$i.stake
    echo "runned"

    str+="$i.stake"

    if [ $i != $(( nodes-1 )) ]; then
      str+=","
    fi

done

echo "creating genesis $str"
xterm -e go run main.go --instance="devnet" --instance-id="0" --network="devnet" --create-new-genesis="$str" --exit

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
    qterminal GORACE="log_path=/$SCRIPTPATH/report"  -e go run -race main.go --instance="devnet" --instance-id="$i" --tcp-server-port="5230" --new-devnet --network="devnet" --set-genesis="file" --staking --hcaptcha-site-key="10000000-ffff-ffff-ffff-000000000001" --hcaptcha-secret="0x0000000000000000000000000000000000000000" --faucet-testnet-enabled="true" --delegates-allowed-enabled="true" $extraArgs &
  else
    xterm -e go run main.go --instance="devnet" --instance-id="$i" --tcp-server-port="5230" --new-devnet --network="devnet" --set-genesis="file" --staking --hcaptcha-site-key="10000000-ffff-ffff-ffff-000000000001" --hcaptcha-secret="0x0000000000000000000000000000000000000000" --faucet-testnet-enabled="true" --delegates-allowed-enabled="true" $extraArgs &
  fi
done

wait

echo "finished"
exit 1
