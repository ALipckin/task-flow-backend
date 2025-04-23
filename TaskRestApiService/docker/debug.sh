#!/bin/sh

go build -gcflags="all=-N -l" -o /tmp/app && \
dlv exec /tmp/app \
  --headless \
  --listen=:2345 \
  --api-version=2 \
  --accept-multiclient \
  --log \
  --continue