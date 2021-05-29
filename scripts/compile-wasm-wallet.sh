cd ../
GOOS=js GOARCH=wasm go build -o ./webassembly/dist/PandoraPay-wallet.wasm
mv ./webassembly/dist/PandoraPay-wallet.wasm ../wallet/dist/dev