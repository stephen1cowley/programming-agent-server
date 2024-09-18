#!/bin/bash
# Start the server
nohup /home/ubuntu/server > /home/ubuntu/server.log 2>&1 &
echo $! > /home/ubuntu/server.pid
