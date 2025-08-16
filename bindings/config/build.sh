#!/bin/bash

# Build Go shared library for Python config bindings

echo "Building Go config shared library for Python bindings..."

# Build the shared library
go build -buildmode=c-shared -o lib_olane_config.so main.go

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "✅ Successfully built lib_olane_config.so"
    echo "📁 Files created:"
    ls -la lib_olane_config.*
    echo ""
    echo "🐍 You can now use this library with Python:"
    echo "   python3 example.py"
    echo ""
    echo "🧪 Run tests:"
    echo "   python3 test_config.py"
else
    echo "❌ Build failed"
    exit 1
fi
