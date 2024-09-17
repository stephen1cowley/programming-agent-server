#!/bin/bash

cd home/ubuntu
./server > /dev/null 2>&1 &
#Don't block the command line while running the server!! --reduce latency!
