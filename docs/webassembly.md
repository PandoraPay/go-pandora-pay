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


## How it works.

Because the GOWASM is compatible and can work as a WebWorker, the code works with bytes instead of strings for blobs because they can be transferable from one worker to another one (main) 


# DISCLAIMER:
This source code is released for research purposes only, with the intent of researching and studying a decentralized p2p network protocol.

PANDORAPAY IS AN OPEN SOURCE COMMUNITY DRIVEN RESEARCH PROJECT. THIS IS RESEARCH CODE PROVIDED TO YOU "AS IS" WITH NO WARRANTIES OF CORRECTNESS. IN NO EVENT SHALL THE CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES. USE AT YOUR OWN RISK.

You may not use this source code for any illegal or unethical purpose; including activities which would give rise to criminal or civil liability.

Under no event shall the Licensor be responsible for the activities, or any misdeeds, conducted by the Licensee.