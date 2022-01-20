FROM ubuntu:20.04

ARG FLOW_CLI_VERSION

RUN apt-get update && apt-get install -y \
  ca-certificates \
  curl \
  && rm -rf /var/lib/apt/lists/*

RUN sh -c "$(curl -fsSL https://storage.googleapis.com/flow-cli/install.sh)" 0 $FLOW_CLI_VERSION \
  && mv /root/.local/bin/flow /usr/local/bin

ENTRYPOINT ["/usr/local/bin/flow"]
