#!/bin/bash
# Stop any running instance of the server
if pgrep -f server > /dev/null; then
  echo "Stopping old server instance..."
  pkill -f server
fi


if [ -d "/home/ubuntu/server" ]; then
  rm -rf /home/ubuntu/server
fi
