#!/bin/bash

FLOW_EMULATOR="flow-emulator"
FLOW_CLI="docker exec -it $FLOW_EMULATOR flow "

if [ ! "$(docker ps -q -f name=$FLOW_EMULATOR)" ]; then
    if [ "$(docker ps -aq -f status=exited -f name=$FLOW_EMULATOR)" ]; then
      # Start old
      docker start $FLOW_EMULATOR
    else
      echo "Run ./scripts/emulator.sh first"
    fi
fi

$FLOW_CLI accounts add-contract NonFungibleToken ./cadence/contracts/NonFungibleToken.cdc
$FLOW_CLI accounts add-contract KittyItems ./cadence/contracts/KittyItems.cdc
