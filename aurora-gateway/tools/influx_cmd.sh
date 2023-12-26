#!/bin/bash

execute_influx_command() {
    local command=$1
    local full_command="docker exec influxdb influx $command"
    echo $full_command
    eval "$full_command"
}

exchange=""
parse_json_command() {
    local json_file=$(cat "$1")
    name=$(echo "$json_file" | jq -r '.[0].spec.name')
    every=$(echo "$json_file" | jq -r '.[0].spec.every')
    query=$(echo "$json_file" | jq -r '.[0].spec.query')

    exchange="option task = {name: \"$name\", every: $every} $query"
}

if [ "$1" = "task" ]; then
    parse_json_command "$2"
    tmp=$(printf "task create '%s'\n" "$exchange")
    execute_influx_command "$tmp"
else
    execute_influx_command "$1 $2"
fi
