if [ $# -eq 0 ]; then
  echo "argument missing"
fi

gitVersion=$(git log -n1 --format=format:"%H")
gitVersionShort=${gitVersion:0:12}


if [ "$1" == "dev" ]; then
  buildOutput="../webassembly/dist/PandoraPay.wasm"
fi
if [ "$1" == "wallet-dev" ]; then
  buildOutput="./webassembly/dist/PandoraPay-wallet.wasm"
fi
if [ "$1" == "wallet-build" ]; then
  buildOutput="./webassembly/dist/PandoraPay-wallet.wasm"
fi

GOOS=js GOARCH=wasm go build -ldflags "-X pandora-pay/config.BUILD_VERSION=${gitVersionShort}" -o ${buildOutput}

if [ "$1" == "dev" ]; then
  exit 1
fi

if [ "$1" == "wallet-dev" ]; then
  cp ${buildOutput} ../PandoraPay-wallet/dist/dev
  exit 1
fi

if [ "$1" == "wallet-build" ]; then

  stat --printf="%s \n" ${buildOutput}

  echo "Deleting..."
  if [ "$2" == "brotli" ]; then
    rm ${buildOutput}.br
  else
    rm ${buildOutput}.gz
  fi

  echo "Copy to wallet/build..."
  cp ${buildOutput} ../PandoraPay-wallet/dist/build

  if [ "$2" == "brotli" ]; then
    echo "Zipping using brotli..."
    if ! brotli -o ${buildOutput}.br ${buildOutput}; then
      echo "sudo apt-get install brotli"
      exit 1
    fi
    stat --printf="brotli size %s \n" ${buildOutput}.br
    echo "Copy to wallet/build..."
    cp ${buildOutput}.br ../PandoraPay-wallet/dist/build
  elif [ "$2" == "zopfli" ]; then
    echo "Zipping using zopfli..."
    if ! zopfli ${buildOutput}; then
      echo "sudo apt-get install zopfli"
      exit 1
    fi
    stat --printf="zopfli gzip size: %s \n" ${buildOutput}.gz
    echo "Copy to wallet/build..."
    cp ${buildOutput}.gz ../PandoraPay-wallet/dist/build
  else
    echo "Gzipping..."
    gzip --best ${buildOutput}
    stat --printf="gzip size %s \n" ${buildOutput}.gz
    echo "Copy to wallet/build..."
    cp ${buildOutput}.gz ../PandoraPay-wallet/dist/build
  fi

  exit 1
fi


echo "'dev', 'wallet-dev', 'wallet-build | brotli | zopfli'"
