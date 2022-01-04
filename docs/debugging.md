### Debugging

Using profiling to debug memory leaks/CPU
0. Install graphviz by running `sudo apt install graphviz`
1. use `--debugging`
2. The following command will request for a 5s CPU
   profile and will launch a browser with an SVG file. `go tool pprof -web http://:6060/debug/pprof/profile?seconds=5`
4. To view the goroutines `go tool pprof -http :8080 http://:6060/debug/pprof/goroutine`

#### Debugging races
GORACE="log_path=/PandoraPay/pandora-pay-go/report" go run -race main.go 


# DISCLAIMER:
This source code is released for research purposes only, with the intent of researching and studying a decentralized p2p network protocol.

PANDORAPAY IS AN OPEN SOURCE COMMUNITY DRIVEN RESEARCH PROJECT. THIS IS RESEARCH CODE PROVIDED TO YOU "AS IS" WITH NO WARRANTIES OF CORRECTNESS. IN NO EVENT SHALL THE CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES. USE AT YOUR OWN RISK.

You may not use this source code for any illegal or unethical purpose; including activities which would give rise to criminal or civil liability.

Under no event shall the Licensor be responsible for the activities, or any misdeeds, conducted by the Licensee.