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
  buildOutput="./dist/PandoraPay-wallet-dev"
elif [[ "$*" == *wallet-build* ]]; then
  buildOutput="./dist/PandoraPay-wallet-build"
else
  echo "argument dev|wallet-dev|wallet-build missing"
  exit 1
fi

if [[ "$*" == *main* ]]; then
  buildOutput+="-main"
  src="./"
elif [[ "$*" == *helper* ]]; then
  buildOutput+="-helper"
  src="./webassembly_helper/"
else
  echo "argument main|helper missing"
  exit 1
fi


buildOutput+=".wasm"

echo ${buildOutput}

go version
(cd ${src} && GOOS=js GOARCH=wasm go build -ldflags "-X pandora-pay/config.BUILD_VERSION=${gitVersionShort}" -o ${buildOutput} )

buildOutput=${src}${buildOutput}

finalOutput="../PandoraPay-wallet/"

cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" "${finalOutput}src/webworkers/dist/wasm_exec.js"
sriOutput="${finalOutput}src/webworkers/dist/sri/"

finalOutput+="dist/"

mkdir -p "${finalOutput}"
mkdir -p "${sriOutput}"

if [[ "$*" == *wallet-dev* ]]; then
  finalOutput+="dev/"
elif [[ "$*" == *wallet-build* ]]; then
  finalOutput+="build/"
  sriOutput+="build"
fi

mkdir -p "${finalOutput}"

stat --printf="%s \n" ${buildOutput}

echo "Deleting..."

rm ${buildOutput}.br 2>/dev/null
rm ${buildOutput}.gz 2>/dev/null

sha256_wasm=$(sha256sum  "${buildOutput}" | awk '{print $1}')
mkdir -p "${finalOutput}wasm"

if [[ "$*" == *main* ]]; then
  finalOutput+="wasm/PandoraPay-wallet-main.wasm"
  sriOutput+="-main.js"
elif [[ "$*" == *helper* ]]; then
  finalOutput+="wasm/PandoraPay-wallet-helper.wasm"
  sriOutput+="-helper.js"
fi

if [[ "$*" == *wallet-build* ]]; then
  echo "export default {
    'wasm': '${sha256_wasm}',
  }" > "${sriOutput}"
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