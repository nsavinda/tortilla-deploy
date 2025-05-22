#!/bin/bash

SECRET="abc"
BODY='{
  "ref": "refs/heads/main",
  "repository": {
    "clone_url": "https://github.com/nsavinda/go-test-server.git"
  },
  "head_commit": {
    "id": "abc12354"
  }
}'

SIGNATURE="sha256=$(echo -n "$BODY" | openssl dgst -sha256 -hmac "$SECRET" | cut -d ' ' -f2)"

# curl -X POST http://44.202.127.157/webhook \
curl -X POST http://localhost:9082/webhook \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: push" \
  -H "X-Hub-Signature-256: $SIGNATURE" \
  -d "$BODY"
