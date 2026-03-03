#!/usr/bin/env bash
set -eou pipefail

# Usage:
# ./script/visitor_jwt.sh <key>
# Require jq installed
#
KEY=$1

# Header
header=$(echo -n '{"alg":"HS256"}' | openssl base64 -e -A | tr '+/' '-_' | tr -d '=')

# Payload
current_time=$(date +%s)
expiration_time=$(($current_time + 864000)) # Ten days from now
payload=$(echo -n '{"exp":'$expiration_time'}' | openssl base64 -e -A | tr '+/' '-_' | tr -d '=')

# Signature
signature=$(echo -n "$header.$payload" | openssl dgst -sha256 -hmac $KEY -binary | openssl base64 -e -A | tr '+/' '-_' | tr -d '=')

# JWT
jwt="$header.$payload.$signature"

echo $jwt
