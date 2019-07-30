# Remove any existing compiled versions
rm -r bin

# Create the bin to place the binaries
mkdir bin

# Remove any vendor folder
rm -r vendor

# Ensure GO111MODULE=on
GO111MODULE=on

echo "Compiling for windows"
env GOOS=windows GOARCH=amd64 go build
mv pegnet.exe bin/pegnet-windows.exe

echo "Compiling for Linux"
env GOOS=linux GOARCH=amd64 go build
mv pegnet bin/pegnet-lin

echo "Compiling for mac"
env GOOS=darwin GOARCH=amd64 go build
mv pegnet bin/pegnet-mac
