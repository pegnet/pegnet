#!/bin/bash

if [ -n "$(gofmt -l .)" ]; then
  echo "Go code is not properly formatted:"
  gofmt -d .
  echo "Run 'gofmt -s -w .' to fix"
  exit 1
fi