#!/bin/bash

# Build Go shared library for Python bindings

echo "Building Go shared library for Python bindings..."

# Build the shared library
go build -buildmode=c-shared -o lib_olane.so main.go

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "✅ Successfully built lib_olane.so"
    echo "📁 Files created:"
    ls -la lib_olane.*
    echo ""
    echo "🐍 You can now use this library with Python:"
    echo "   python3 example.py"
else
    echo "❌ Build failed"
    exit 1
fi
