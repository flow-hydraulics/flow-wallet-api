FROM alpine:latest
ARG VERSION=0.20.2
ARG UNAME=appuser
ARG GNAME=appgroup
ARG UID=1000
ARG GID=1000
RUN apk add --no-cache curl
RUN curl -fsSL https://storage.googleapis.com/flow-cli/install.sh | sh -s -- v$VERSION
RUN mv /root/.local/bin/flow /usr/local/bin/flow
RUN addgroup -g $GID -S $GNAME && adduser --uid $UID -S $UNAME -G $GNAME
RUN mkdir /app && chown appuser /app && chgrp $GNAME /app
WORKDIR /app
VOLUME [ "/app" ]
USER appuser
ENTRYPOINT [ "/usr/local/bin/flow" ]
