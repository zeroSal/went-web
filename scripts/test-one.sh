#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BASE_DIR="$(dirname "$SCRIPT_DIR")"

echo -e "${YELLOW}=== Testing went-web examples ===${NC}"

FAILED=0

# Kill any lingering went processes
pkill -f "went-" 2>/dev/null || true
pkill -f "test-cookie" 2>/dev/null || true
sleep 1

test_example() {
    local name=$1
    local port=$2
    local dir="$BASE_DIR/examples/$name"
    local pid=
    local logfile="/tmp/went-$name.log"
    
    echo -e "\n${YELLOW}=== Testing $name on port $port ===${NC}"
    
    # Kill existing process on port
    if lsof -iTCP:$port -sTCP:LISTEN > /dev/null 2>&1; then
        echo "Killing existing process on port $port"
        kill $(lsof -t -iTCP:$port -sTCP:LISTEN) 2>/dev/null || true
        sleep 1
    fi
    
    # Build
    echo "Building..."
    if ! go build -o "/tmp/went-$name" "$dir/main.go" 2>/dev/null; then
        echo -e "${RED}✗ Build failed${NC}"
        FAILED=1
        return
    fi
    echo -e "${GREEN}✓ Built${NC}"
    
    # Start from example directory
    echo "Starting..."
    cd "$dir"
    "/tmp/went-$name" > "$logfile" 2>&1 &
    pid=$!
    cd "$BASE_DIR"
    
    # Wait for server
    echo "Waiting for server..."
    started=0
    for i in $(seq 1 30); do
        if ! kill -0 $pid 2>/dev/null; then
            echo -e "${RED}✗ Process died${NC}"
            cat "$logfile"
            break
        fi
        
        # Check if port is listening
        if lsof -iTCP:$port -sTCP:LISTEN > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Server started${NC}"
            started=1
            sleep 1
            break
        fi
        sleep 0.5
    done
    
    if [ $started -eq 0 ]; then
        echo -e "${RED}✗ Server failed to start${NC}"
        kill $pid 2>/dev/null || true
        wait $pid 2>/dev/null || true
        rm -f "/tmp/went-$name"
        FAILED=1
        return
    fi
    
    # Test cookie-auth (01)
    if [ "$name" = "01-cookie-auth" ]; then
        echo "Testing cookie auth..."
        
        # Test login
        resp=$(curl -s --max-time 5 "<http://localhost:$port/login>")
        echo "  Login response: $resp"
        if echo "$resp" | grep -q "Logged in"; then
            echo -e "${GREEN}✓ Login OK${NC}"
        else
            echo -e "${RED}✗ Login FAIL${NC}"
            FAILED=1
        fi
        
        # Test profile with cookie
        resp=$(curl -s --max-time 5 --cookie "SESSION_ID=user-session-123" "<http://localhost:$port/profile>")
        echo "  Profile response: $resp"
        if echo "$resp" | grep -q "Welcome"; then
            echo -e "${GREEN}✓ Profile OK${NC}"
        else
            echo -e "${RED}✗ Profile FAIL${NC}"
            FAILED=1
        fi
    fi
    
    # Stop
    echo "Stopping..."
    kill $pid 2>/dev/null || true
    wait $pid 2>/dev/null || true
    rm -f "/tmp/went-$name" "$logfile"
    echo -e "${GREEN}✓ Stopped${NC}"
}

# Test just one example first
test_example "01-cookie-auth" "8080"

echo -e "\n${YELLOW}=== Test completed ===${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ PASSED${NC}"
else
    echo -e "${RED}❌ FAILED${NC}"
fi

exit $FAILED
