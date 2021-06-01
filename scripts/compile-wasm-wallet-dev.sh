GOOS=js GOARCH=wasm go build -o ./webassembly/dist/PandoraPay-wallet.wasm
cp ./webassembly/dist/PandoraPay-wallet.wasm ../PandoraPay-wallet/dist/dev