extraArgs=""

for arg in $@; do
  extraArgs+=" $arg "
done

./scripts/create-testnet.sh mode=normal --pprof --light-computations --tcp-server-port="5230" $extraArgs