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

# Common Errors

go: github.com/Workiva/go-datastructures@v1.0.53: missing go.sum entry; to add it:
go mod download github.com/Workiva/go-datastructures
```
go mod tidy
```

## Video Tutorial
https://www.youtube.com/watch?v=Jo7BbL7Xdms