#!/bin/bash
file_path=$(jq -r '.tool_input.file_path')
[[ -z "$file_path" ]] && exit 0

if [[ "$file_path" == *.go ]]; then
  gofmt -w "$file_path" 2>/dev/null || true
fi

exit 0
