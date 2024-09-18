#!/bin/bash
# Validate if the server is running
if pgrep -f server > /dev/null; then
  echo "Server is running."
  exit 0
else
  echo "Server is not running."
  exit 1
fi
