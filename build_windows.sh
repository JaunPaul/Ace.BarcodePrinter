#!/bin/bash
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o printer.exe .
if [ $? -eq 0 ]; then
    echo "Success! printer.exe created."
else
    echo "Build failed."
fi
