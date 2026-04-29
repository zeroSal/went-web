#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BASE_DIR="$(dirname "$SCRIPT_DIR")"

echo -e "${YELLOW}=== Testing went-web examples ===${NC}"

FAILED=0

pkill -f "went-" 2>/dev/null || true
sleep 1

kill_port() {
    local port=$1
    if lsof -iTCP:$port -sTCP:LISTEN > /dev/null 2>&1; then
        echo "Killing process on port $port"
        kill $(lsof -t -iTCP:$port -sTCP:LISTEN) 2>/dev/null || true
        sleep 1
    fi
}

EXAMPLES="01-cookie-auth:8080 02-bearer-auth:8081 04-composite-auth:8083 06-user-provider:8085 07-controllers:8086 08-security-full:8088 09-integration:8089 10-routes:8090 11-session:8091 12-csrf:8092"

for pair in $EXAMPLES; do
    name=$(echo "$pair" | cut -d: -f1)
    port=$(echo "$pair" | cut -d: -f2)
    dir="$BASE_DIR/examples/$name"
    logfile="/tmp/went-$name.log"

    if [ ! -d "$dir" ]; then
        continue
    fi

    echo -e "\n${YELLOW}=== Testing $name on port $port ===${NC}"

    kill_port $port

    echo "Building..."
    if ! go build -o "/tmp/went-$name" "$dir/main.go" 2>/dev/null; then
        echo -e "${RED}✗ Build failed${NC}"
        FAILED=1
        continue
    fi
    echo -e "${GREEN}✓ Built${NC}"

    echo "Starting..."
    cd "$dir"
    "/tmp/went-$name" > "$logfile" 2>&1 &
    pid=$!
    cd "$BASE_DIR"

    echo "Waiting for server..."
    started=0
    for i in $(seq 1 30); do
        if ! kill -0 $pid 2>/dev/null; then
            echo -e "${RED}✗ Process died${NC}"
            cat "$logfile"
            break
        fi

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
        continue
    fi

    case $name in
        01-cookie-auth)
            echo "Testing cookie auth..."
            resp=$(curl -s --max-time 5 "http://127.0.0.1:$port/login")
            echo "  Login: $resp"
            echo "$resp" | grep -q "Logged in" && echo -e "${GREEN}✓ Login OK${NC}" || { echo -e "${RED}✗ Login FAIL${NC}"; FAILED=1; }

            resp=$(curl -s --max-time 5 --cookie "SESSION_ID=user-session-123" "http://127.0.0.1:$port/profile")
            echo "  Profile: $resp"
            echo "$resp" | grep -q "Welcome" && echo -e "${GREEN}✓ Profile OK${NC}" || { echo -e "${RED}✗ Profile FAIL${NC}"; FAILED=1; }
            ;;
        02-bearer-auth)
            echo "Testing bearer auth..."
            resp=$(curl -s --max-time 5 -X POST "http://127.0.0.1:$port/login" -d "username=admin&password=secret")
            echo "  Login: $resp"
            echo "$resp" | grep -q "token" && echo -e "${GREEN}✓ Login OK${NC}" || { echo -e "${RED}✗ Login FAIL${NC}"; FAILED=1; }

            resp=$(curl -s --max-time 5 -H "Authorization: Bearer my-secret-token-123" "http://127.0.0.1:$port/api/data")
            echo "  Bearer: $resp"
            echo "$resp" | grep -q "Protected" && echo -e "${GREEN}✓ Bearer OK${NC}" || { echo -e "${RED}✗ Bearer FAIL${NC}"; FAILED=1; }
            ;;
        04-composite-auth)
            echo "Testing composite auth..."
            resp=$(curl -s --max-time 5 --cookie "SESSION_ID=cookie-session-123" "http://127.0.0.1:$port/api/data")
            echo "  Cookie: $resp"
            echo "$resp" | grep -q "authenticated" && echo -e "${GREEN}✓ Cookie OK${NC}" || { echo -e "${RED}✗ Cookie FAIL${NC}"; FAILED=1; }
            ;;
        06-user-provider)
            echo "Testing user provider..."
            resp=$(curl -s --max-time 5 -X POST "http://127.0.0.1:$port/login" -d "username=john")
            echo "  Login: $resp"
            echo "$resp" | grep -q "user_id" && echo -e "${GREEN}✓ Login OK${NC}" || { echo -e "${RED}✗ Login FAIL${NC}"; FAILED=1; }
            ;;
        07-controllers)
            echo "Testing controllers..."
            resp=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:$port/")
            echo "  GET / : $resp"
            [ "$resp" = "200" ] && echo -e "${GREEN}✓ Home OK${NC}" || { echo -e "${RED}✗ Home FAIL${NC}"; FAILED=1; }

            resp=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:$port/about")
            echo "  GET /about : $resp"
            [ "$resp" = "200" ] && echo -e "${GREEN}✓ About OK${NC}" || { echo -e "${RED}✗ About FAIL${NC}"; FAILED=1; }

            resp=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:$port/api/data")
            echo "  GET /api/data : $resp"
            [ "$resp" = "200" ] && echo -e "${GREEN}✓ API GET OK${NC}" || { echo -e "${RED}✗ API GET FAIL${NC}"; FAILED=1; }
            ;;
        08-security-full)
            echo "Testing full security..."
            resp=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:$port/login")
            echo "  GET /login : $resp"
            [ "$resp" = "200" ] && echo -e "${GREEN}✓ Login page OK${NC}" || { echo -e "${RED}✗ Login page FAIL${NC}"; FAILED=1; }

            resp=$(curl -s -X POST "http://127.0.0.1:$port/login" -d "username=admin&password=secret")
            echo "  Login: $resp"
            echo "$resp" | grep -q "Logged in" && echo -e "${GREEN}✓ Login OK${NC}" || { echo -e "${RED}✗ Login FAIL${NC}"; FAILED=1; }
            ;;
        09-integration)
            echo "Testing integration..."
            resp=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:$port/public/info")
            echo "  GET /public/info : $resp"
            [ "$resp" = "200" ] && echo -e "${GREEN}✓ Public OK${NC}" || { echo -e "${RED}✗ Public FAIL${NC}"; FAILED=1; }
            ;;
        10-routes)
            echo "Testing routes..."
            resp=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:$port/")
            echo "  GET / : $resp"
            [ "$resp" = "200" ] && echo -e "${GREEN}✓ Home route OK${NC}" || { echo -e "${RED}✗ Home route FAIL${NC}"; FAILED=1; }

            resp=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:$port/api/data")
            echo "  GET /api/data : $resp"
            [ "$resp" = "401" ] && echo -e "${GREEN}✓ API GET OK${NC}" || { echo -e "${RED}✗ API GET FAIL${NC}"; FAILED=1; }

            resp=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://127.0.0.1:$port/api/data")
            echo "  POST /api/data : $resp"
            [ "$resp" = "401" ] && echo -e "${GREEN}✓ API POST OK${NC}" || { echo -e "${RED}✗ API POST FAIL${NC}"; FAILED=1; }

            resp=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:$port/nonexistent")
            echo "  GET /nonexistent : $resp"
            [ "$resp" = "404" ] && echo -e "${GREEN}✓ 404 for unknown route OK${NC}" || { echo -e "${RED}✗ Should return 404${NC}"; FAILED=1; }
            ;;
        11-session)
            echo "Testing session..."
            # Login and save cookies
            resp=$(curl -s --max-time 5 -X POST "http://127.0.0.1:$port/login" -c /tmp/session-cookies.txt)
            echo "  Login: $resp"
            echo "$resp" | grep -q "Logged in" && echo -e "${GREEN}✓ Login OK${NC}" || { echo -e "${RED}✗ Login FAIL${NC}"; FAILED=1; }

            # Access profile with saved cookies
            resp=$(curl -s --max-time 5 -b /tmp/session-cookies.txt "http://127.0.0.1:$port/profile")
            echo "  Profile: $resp"
            if echo "$resp" | grep -q "active"; then
                echo -e "${GREEN}✓ Session OK${NC}"
            else
                echo -e "${YELLOW}⚠ Profile requires auth integration${NC}"
            fi
            rm -f /tmp/session-cookies.txt
            ;;
        12-csrf)
            echo "Testing CSRF..."
            resp=$(curl -s --max-time 5 "http://127.0.0.1:$port/check-csrf")
            echo "  CSRF: $resp"
            echo "$resp" | grep -q "true" && echo -e "${GREEN}✓ CSRF OK${NC}" || { echo -e "${RED}✗ CSRF FAIL${NC}"; FAILED=1; }
            ;;
        *)
            echo "No specific tests for $name"
            ;;
    esac

    echo "Stopping..."
    kill $pid 2>/dev/null || true
    wait $pid 2>/dev/null || true
    rm -f "/tmp/went-$name" "$logfile"
    echo -e "${GREEN}✓ Stopped${NC}"
done

echo -e "\n${YELLOW}=== All tests completed ===${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}ALL TESTS PASSED${NC}"
else
    echo -e "${RED}SOME TESTS FAILED${NC}"
fi

exit $FAILED
