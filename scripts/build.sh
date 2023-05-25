output="./bin/pandora"

#linux
echo "build linux..."
GOOS=linux GOARCH=amd64 go build -o ${output}-linux-amd64
GOOS=linux GOARCH=386 go build -o ${output}-linux-386
GOOS=linux GOARCH=arm64 go build -o ${output}-linux-arm64
GOOS=linux GOARCH=arm GOARM=7 go build -o ${output}-linux-armv7l

# windows
echo "build windows..."
GOOS=windows GOARCH=amd64 go build -o ${output}-windows-amd64.exe
GOOS=windows GOARCH=386 go build -o ${output}-windows-386.exe
GOOS=windows GOARCH=arm64 go build -o ${output}-windows-arm64.exe

#macOS
echo "build darwin..."
GOOS=darwin GOARCH=amd64 go build -o ${output}-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o ${output}-darwin-arm64

echo "build success"
