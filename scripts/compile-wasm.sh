if [ $# -eq 0 ]
then
  echo "argument missing"
  exit 1
fi


if [ $1 == "dev" ]
then
  GOOS=js GOARCH=wasm go build -o ../webassembly/dist/PandoraPay.wasm
  exit 1
fi

if [ $1 == "wallet-dev" ]
then
  GOOS=js GOARCH=wasm go build -o ./webassembly/dist/PandoraPay-wallet.wasm
  cp ./webassembly/dist/PandoraPay-wallet.wasm ../PandoraPay-wallet/dist/dev
  exit 1
fi

if [ $1 == "wallet-build" ]
then
  echo "Deleting..."
  rm ./webassembly/dist/PandoraPay-wallet.wasm.gz
  rm ./webassembly/dist/PandoraPay-wallet.wasm.br

  echo "Compiling..."
  GOOS=js GOARCH=wasm go build -o ./webassembly/dist/PandoraPay-wallet.wasm

  echo "Copy to wallet/build..."
  cp ./webassembly/dist/PandoraPay-wallet.wasm ../PandoraPay-wallet/dist/build

  if [ $2 == "brotli" ]
  then
    echo "Gzipping using brotli..."
    brotli -o ./webassembly/dist/PandoraPay-wallet.wasm.br ./webassembly/dist/PandoraPay-wallet.wasm
    echo "Copy to wallet/build..."
    cp ./webassembly/dist/PandoraPay-wallet.wasm.gz ../PandoraPay-wallet/dist/build
  else
    echo "Gzipping..."
    gzip --best ./webassembly/dist/PandoraPay-wallet.wasm
    echo "Copy to wallet/build..."
    cp ./webassembly/dist/PandoraPay-wallet.wasm.gz ../PandoraPay-wallet/dist/build
  fi

  exit 1
fi


echo "'dev', 'wallet-dev', 'wallet-build | brotli'"
