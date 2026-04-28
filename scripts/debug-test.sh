#!/bin/bash

BASE_DIR="/Users/sal/Documents/Developing/went-web"
name="01-cookie-auth"
port=8080
dir="$BASE_DIR/examples/$name"
logfile="/tmp/went-test.log"

echo "=== Killing old processes ==="
pkill -f "went-01-cookie-auth" 2>/dev/null || true
pkill -f "test-cookie" 2>/dev/null || true
sleep 2

echo "=== Building ==="
cd "$dir"
go build -o "/tmp/went-$name" main.go
echo "Build OK"

echo "=== Starting server ==="
"/tmp/went-$name" > "$logfile" 2>&1 &
pid=$!
echo "Server PID: $pid"

echo "=== Waiting 3 seconds ==="
sleep 3

echo "=== Checking if process is alive ==="
if ! kill -0 $pid 2>/dev/null; then
    echo "PROCESS DIED!"
    cat "$logfile"
    exit 1
fi
echo "Process is alive"

echo "=== Checking port ==="
if ! lsof -iTCP:$port -sTCP:LISTEN > /dev/null 2>&1; then
    echo "PORT NOT LISTENING!"
    cat "$logfile"
    exit 1
fi
echo "Port $port is listening"

echo "=== Testing curl ==="
echo "Testing /login..."
curl --max-time 5 "http://127.0.0.1:$port/login"
echo ""
echo "Exit code: $?"

echo ""
echo "Testing /profile with cookie..."
curl --max-time 5 --cookie "SESSION_ID=user-session-123" "http://127.0.0.1:$port/profile"
echo ""
echo "Exit code: $?"

echo ""
echo "=== Stopping server ==="
kill $pid 2>/dev/null || true
wait $pid 2>/dev/null || true
rm -f "/tmp/went-$name"
echo "Done"
