#!/bin/bash

set -eu -o pipefail

jq -rc '.response.docs[] | {"title": .title, "author": .author, "year": .publishDateSort, "url": .url}' < "${1:-/dev/stdin}"
