cd ../
echo "Deleting..."
rm ./webassembly/dist/PandoraPay-wallet.wasm.gz
echo "Compiling..."
GOOS=js GOARCH=wasm go build -o ./webassembly/dist/PandoraPay-wallet.wasm

echo "Copy to wallet/dev..."
cp ./webassembly/dist/PandoraPay-wallet.wasm ../PandoraPay-wallet/dist/dev
echo "Copy to wallet/build..."
cp ./webassembly/dist/PandoraPay-wallet.wasm ../PandoraPay-wallet/dist/build

echo "Gzipping..."
gzip --best ./webassembly/dist/PandoraPay-wallet.wasm
echo "Copy to wallet/dev..."
cp ./webassembly/dist/PandoraPay-wallet.wasm.gz ../PandoraPay-wallet/dist/dev
echo "Copy to wallet/build..."
cp ./webassembly/dist/PandoraPay-wallet.wasm.gz ../PandoraPay-wallet/dist/build
