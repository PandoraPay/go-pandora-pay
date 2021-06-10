if [ $# -eq 0 ]; then
  echo "argument missing"
fi


if [ "$1" == "dev" ]; then
  GOOS=js GOARCH=wasm go build -o ../webassembly/dist/PandoraPay.wasm
  exit 1
fi

if [ "$1" == "wallet-dev" ]; then
  GOOS=js GOARCH=wasm go build -o ./webassembly/dist/PandoraPay-wallet.wasm
  cp ./webassembly/dist/PandoraPay-wallet.wasm ../PandoraPay-wallet/dist/dev
  exit 1
fi

if [ "$1" == "wallet-build" ]; then

  echo "Deleting..."
  if [ "$2" == "brotli" ]; then
    rm ./webassembly/dist/PandoraPay-wallet.wasm.br
  else
    rm ./webassembly/dist/PandoraPay-wallet.wasm.gz
  fi

  echo "Compiling..."
  GOOS=js GOARCH=wasm go build -o ./webassembly/dist/PandoraPay-wallet.wasm

  echo "Copy to wallet/build..."
  cp ./webassembly/dist/PandoraPay-wallet.wasm ../PandoraPay-wallet/dist/build

  if [ "$2" == "brotli" ]; then
    echo "Zipping using brotli..."
    if ! brotli -o ./webassembly/dist/PandoraPay-wallet.wasm.br ./webassembly/dist/PandoraPay-wallet.wasm; then
      echo "sudo apt-get install brotli"
      exit 1
    fi
    echo "Copy to wallet/build..."
    cp ./webassembly/dist/PandoraPay-wallet.wasm.br ../PandoraPay-wallet/dist/build
  elif [ "$2" == "zopfli" ]; then
    echo "Zipping using zopfli..."
    if ! zopfli ./webassembly/dist/PandoraPay-wallet.wasm; then
      echo "sudo apt-get install zopfli"
      exit 1
    fi
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


echo "'dev', 'wallet-dev', 'wallet-build | brotli | zopfli'"
