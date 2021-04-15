#!/bin/bash

docker run --name flow-emulator -d -p 3569:3569 -p 8080:8080  gcr.io/flow-container-registry/emulator:v0.16.1
