### Debugging

Using profiling to debug memory leaks/CPU
0. Install graphviz by running `sudo apt install graphviz`
1. use `--debugging`
2. The following command will request for a 5s CPU
   profile and will launch a browser with an SVG file. `go tool pprof -web http://:6060/debug/pprof/profile?seconds=5`
4. To view the goroutines `go tool pprof -http :8080 http://:6060/debug/pprof/goroutine`

#### Debugging races
GORACE="log_path=/PandoraPay/pandora-pay-go/report" go run -race main.go 
