#!/bin/sh
read -r brief
printf '%s: %s' "$brief" "$(cat "$1")"
