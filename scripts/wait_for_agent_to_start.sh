#!/usr/bin/env bash

wait_for_it () {
    connection=$(curl --silent --show-error localhost:8088/health)
    if [[ "$connection" == '{"status":"ok"}' ]]; then
        echo "Agent server is up and running:" "$connection"
        return 0
    else
        return 1
    fi
}

echo "Connecting to agent server..."
timeout=$((SECONDS + 30))      # 30 s timeout limit to prevent infinite loop

while ! wait_for_it; do
    if (( SECONDS > timeout )); then
        echo "Error! Timeout exceeded. Agent server not started?"
        exit 1
    else
        sleep 1
    fi
done

echo "Connection established."
