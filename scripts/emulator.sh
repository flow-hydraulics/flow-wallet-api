#!/bin/bash

FLOW_CLI="flow-cli"
FLOW_EMULATOR="flow-emulator"

# Build flow-cli image if needed
if [[ "$(docker images -q ${FLOW_CLI} 2> /dev/null)" == "" ]]; then
  docker build \
    --build-arg UID=$(id -u) \
    --build-arg GID=$(id -g) \
    -t ${FLOW_CLI} \
    ./scripts/flow-cli
fi

# Create config folder
if [[ ! -d ".flow" ]]
then
  mkdir -p .flow
fi

# Init if needed
if [[ ! -f ".flow/flow.json" ]]; then
  docker run --rm \
    -v $(pwd)/.flow/:/config \
    ${FLOW_CLI} init
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
        -v $(pwd)/.flow/:/config \
        ${FLOW_CLI} emulator -v
    fi
fi

docker logs -f $FLOW_EMULATOR
