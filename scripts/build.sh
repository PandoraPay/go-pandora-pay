output="./bin/pandora"

#linux
echo "build linux..."
GOOS=linux GOARCH=amd64 go build -o ${output}-linux-amd64
GOOS=linux GOARCH=386 go build -o ${output}-linux-386

# windows
echo "build windows..."
GOOS=windows GOARCH=amd64 go build -o ${output}-windows-amd64.exe
GOOS=windows GOARCH=386 go build -o ${output}-windows-386.exe

#macOS
echo "build darwin..."
GOOS=darwin GOARCH=amd64 go build -o ${output}-darwin-amd64
GOOS=darwin GOARCH=386 go build -o ${output}-darwin-386

echo "build success"
