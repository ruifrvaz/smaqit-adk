#!/bin/sh
read -r task
printf '%s: %s' "$task" "$(cat "$1")"
