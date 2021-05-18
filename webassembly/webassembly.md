# Configuring WebAssembly (WASM) for PandoraPay

```
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./webassembly/dist/
```

# Compiling to WebAssembly (WASM)

```
cd ./webassembly/
```

```
GOOS=js GOARCH=wasm go build -o ./dist/PandoraPay.wasm
```


## Video Tutorial
https://www.youtube.com/watch?v=Jo7BbL7Xdms