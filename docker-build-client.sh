#!/bin/bash

docker build --tag wat-client -f wat-client/Dockerfile .

echo "The wat-client can now be run with:"
echo "    docker run -it wat-client -nick <name> -host <server hostname or ip>"