if [ $# -eq 0 ]; then
  echo "arguments missing"
fi

if [[ "$*" == "help" ]]; then
    echo "main|helper, dev|wallet-dev|wallet-build, brotli|zopfli|gzip"
    exit 1
fi

gitVersion=$(git log -n1 --format=format:"%H")
gitVersionShort=${gitVersion:0:12}

src=""
buildOutput=""

if [[ "$*" == *dev* ]]; then
  buildOutput="./dist/PandoraPay"
elif [[ "$*" == *wallet-dev* ]]; then
  buildOutput="./dist/PandoraPay-wallet"
elif [[ "$*" == *wallet-build* ]]; then
  buildOutput="./dist/PandoraPay-wallet"
else
  echo "argument dev|wallet-dev|wallet-build missing"
  exit 1
fi

if [[ "$*" == *main* ]]; then
  buildOutput="./webassembly/"${buildOutput}"-main"
  src="./"
elif [[ "$*" == *helper* ]]; then
  buildOutput+="-helper"
  src="webassembly_helper/"
else
  echo "argument main|helper missing"
  exit 1
fi

buildOutput+=".wasm"

echo ${buildOutput}

go version
(cd ${src} && GOOS=js GOARCH=wasm go build -ldflags "-X pandora-pay/config.BUILD_VERSION=${gitVersionShort}" -o ${buildOutput} )

buildOutput=${src}${buildOutput}

finalOutput="../PandoraPay-wallet/dist/"

if [[ "$*" == *wallet-dev* ]]; then
  finalOutput+="dev/wasm/"
elif [[ "$*" == *wallet-build* ]]; then
  finalOutput+="build/wasm/"
fi

stat --printf="%s \n" ${buildOutput}

echo "Deleting..."

rm ${buildOutput}.br 2>/dev/null
rm ${buildOutput}.gz 2>/dev/null

mkdir -p ${finalOutput}

if [[ "$*" == *main* ]]; then
  finalOutput+="PandoraPay-wallet-main.wasm"
elif [[ "$*" == *helper* ]]; then
  finalOutput+="PandoraPay-wallet-helper.wasm"
fi

echo "Copy to wallet/build..."
cp ${buildOutput} ${finalOutput}

if [[ "$*" == *wallet-build* ]]; then

  if [[ "$*" == *brotli* ]]; then
    echo "Zipping using brotli..."
    if ! brotli -o ${buildOutput}.br ${buildOutput}; then
      echo "sudo apt-get install brotli"
      exit 1
    fi
    stat --printf="brotli size %s \n" ${buildOutput}.br
    echo "Copy to wallet/build..."
    cp ${buildOutput}.br ${finalOutput}.br
  fi

  if [[ "$*" == *zopfli* ]]; then
    echo "Zipping using zopfli..."
    if ! zopfli ${buildOutput}; then
      echo "sudo apt-get install zopfli"
      exit 1
    fi
    stat --printf="zopfli gzip size: %s \n" ${buildOutput}.gz
    echo "Copy to wallet/build..."
    cp ${buildOutput}.gz ${finalOutput}.gz
  elif [[ "$*" == *gzip* ]]; then
    echo "Gzipping..."
    gzip --best ${buildOutput}
    stat --printf="gzip size %s \n" ${buildOutput}.gz
    echo "Copy to wallet/build..."
    cp ${buildOutput}.gz ${finalOutput}.gz
  fi

fi