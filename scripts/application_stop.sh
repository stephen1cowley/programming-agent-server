#!/bin/bash
# Stop the server using the PID file.
if [ -f /home/ubuntu/server.pid ]; then
  kill $(cat /home/ubuntu/server.pid)
  rm /home/ubuntu/server.pid
fi
