if [ $# -eq 0 ]; then
  echo "argument missing"
fi

gitVersion=$(git log -n1 --format=format:"%H")
gitVersionShort=${gitVersion:0:12}

src=""
buildOutput=""

if [ "$2" == "dev" ]; then
  buildOutput="./dist/PandoraPay"
fi
if [ "$2" == "wallet-dev" ]; then
  buildOutput="./dist/PandoraPay-wallet"
fi
if [ "$2" == "wallet-build" ]; then
  buildOutput="./dist/PandoraPay-wallet"
fi

if [ "$1" == "main" ]; then
  buildOutput="./webassembly/"${buildOutput}"-main"
  src="./"
fi
if [ "$1" == "helper" ]; then
  buildOutput=${buildOutput}"-helper"
  src="webassembly_helper/"
fi

buildOutput=${buildOutput}".wasm"

go version
(cd ${src} && GOOS=js GOARCH=wasm go build -ldflags "-X pandora-pay/config.BUILD_VERSION=${gitVersionShort}" -o ${buildOutput} )

buildOutput=${src}${buildOutput}

if [ "$2" == "wallet-dev" ]; then
  cp ${buildOutput} ../PandoraPay-wallet/dist/dev/wasm
  exit 1
fi

if [ "$2" == "wallet-build" ]; then

  stat --printf="%s \n" ${buildOutput}

  echo "Deleting..."
  if [ "$3" == "brotli" ]; then
    rm ${buildOutput}.br
  else
    rm ${buildOutput}.gz
  fi

  finalOutput="../PandoraPay-wallet/dist/build/wasm"

  echo "Copy to wallet/build..."
  cp ${buildOutput} ${finalOutput}

  if [ "$3" == "brotli" ]; then
    echo "Zipping using brotli..."
    if ! brotli -o ${buildOutput}.br ${buildOutput}; then
      echo "sudo apt-get install brotli"
      exit 1
    fi
    stat --printf="brotli size %s \n" ${buildOutput}.br
    echo "Copy to wallet/build..."
    cp ${buildOutput}.br ${finalOutput}
  elif [ "$3" == "zopfli" ]; then
    echo "Zipping using zopfli..."
    if ! zopfli ${buildOutput}; then
      echo "sudo apt-get install zopfli"
      exit 1
    fi
    stat --printf="zopfli gzip size: %s \n" ${buildOutput}.gz
    echo "Copy to wallet/build..."
    cp ${buildOutput}.gz ${finalOutput}
  else
    echo "Gzipping..."
    gzip --best ${buildOutput}
    stat --printf="gzip size %s \n" ${buildOutput}.gz
    echo "Copy to wallet/build..."
    cp ${buildOutput}.gz ${finalOutput}
  fi

  exit 1
fi


echo "'main' | 'helper', 'dev', 'wallet-dev', 'wallet-build | brotli | zopfli'"
