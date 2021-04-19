#!/bin/bash

FLOW_CLI_IMG="flow-cli"
FLOW_EMULATOR="flow-emulator"

# Build flow-cli image if needed
if [[ "$(docker images -q ${FLOW_CLI_IMG} 2> /dev/null)" == "" ]]; then
  docker build \
    --build-arg UID=$(id -u) \
    --build-arg GID=$(id -g) \
    -t ${FLOW_CLI_IMG} \
    ./scripts/flow-cli
fi

# Init if needed
if [[ ! -f "flow.json" ]]; then
  docker run --rm \
    -v $(pwd)/:/app \
    ${FLOW_CLI_IMG} init
fi

if [ ! "$(docker ps -q -f name=$FLOW_EMULATOR)" ]; then
    if [ "$(docker ps -aq -f status=exited -f name=$FLOW_EMULATOR)" ]; then
      # Start old
      docker start $FLOW_EMULATOR
    else
      # Create new
      docker run -d \
        --name $FLOW_EMULATOR \
        -p 3569:3569 \
        -p 8080:8080  \
        -v $(pwd)/:/app \
        ${FLOW_CLI_IMG} emulator -v
    fi
fi

docker logs -f $FLOW_EMULATOR
