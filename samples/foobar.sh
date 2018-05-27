#!/bin/bash

echo "Hello"
echo

trap cleanup  INT

function cleanup() {
  exit 0
}

while true; do
  echo "Hello World!" >&2
  sleep 1
done
