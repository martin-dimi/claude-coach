#!/bin/bash
# Thin wrapper for standalone (non-plugin) installations
OUTPUT=$(coach check 2>/dev/null)
if [ -n "$OUTPUT" ]; then
  echo "$OUTPUT"
fi
exit 0
