#linux
echo "build linux..."
GOOS=linux GOARCH=amd64 go build -o ./bin/pandora-amd64-linux
GOOS=linux GOARCH=386 go build -o ./bin/pandora-386-linux

# windows
echo "build windows..."
GOOS=windows GOARCH=amd64 go build -o ./bin/pandora-amd64-windows.exe
GOOS=windows GOARCH=386 go build -o ./bin/pandora-386-windows.exe

#macOS
echo "build darwin..."
GOOS=darwin GOARCH=amd64 go build -o ./bin/pandora-amd64-darwin
GOOS=darwin GOARCH=386 go build -o ./bin/pandora-386-darwin

echo "build success"