cd ../
echo "Compiling..."
GOOS=js GOARCH=wasm go build -o ./webassembly/dist/PandoraPay.wasm