curl -X POST http://localhost:9082/webhook \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: push" \
  -d '{
    "ref": "refs/heads/main",
    "repository": {
      "name": "git-hooks",
      "clone_url": "https://github.com/nsavinda/go-test-server.git"
    },
    "head_commit": {
      "id": "abc12354",
      "message": "Test commit"
    }
  }'

# !/bin/bash

# # Number of requests to send
# NUM_REQUESTS=25
# # Delay in seconds (0.01 = 10ms)
# DELAY=3

# # Loop to send multiple requests
# for ((i=1; i<=NUM_REQUESTS; i++)); do
#   echo "🚀 Sending request #$i"
  
#   curl -X POST http://localhost:9082/webhook \
#     -H "Content-Type: application/json" \
#     -H "X-GitHub-Event: push" \
#     -d '{
#       "ref": "refs/heads/main",
#       "repository": {
#         "name": "git-hooks",
#         "clone_url": "https://github.com/nsavinda/go-test-server.git"
#       },
#       "head_commit": {
#         "id": "abc12354",
#         "message": "Test commit"
#       }
#     }' &

#   # Delay before the next request
#   sleep $DELAY
# done

# # Wait for all background processes to finish
# wait
# echo "🎉 All requests sent."
