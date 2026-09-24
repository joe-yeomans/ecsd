#!/bin/bash
set -e

cd "$(dirname "${BASH_SOURCE[0]}")/.."

echo "Building ecsd..."
mkdir -p "$HOME/bin"
go build -o "$HOME/bin/ecsd" .

echo "Done!"
