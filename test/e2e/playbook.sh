#!/usr/bin/env bash
# E2E smoke playbook: assumes agent and proxy are already running (e.g. via make e2e-docker).
# Requires: curl. Agent default port 8080, proxy 8081.
# Optional: AGENT_URL=http://localhost:8080 PROXY_URL=http://localhost:8081

set -e
AGENT_URL="${AGENT_URL:-http://localhost:8080}"
PROXY_URL="${PROXY_URL:-http://localhost:8081}"

echo "E2E smoke: agent=$AGENT_URL proxy=$PROXY_URL"

# Health
curl -sf "$AGENT_URL/healthz" -o /dev/null || { echo "agent healthz failed"; exit 1; }
curl -sf "$PROXY_URL/healthz" -o /dev/null || { echo "proxy healthz failed"; exit 1; }

# Proxy without auth should return 401 when REQUIRE_AUTH is true
code=$(curl -s -o /dev/null -w "%{http_code}" "$PROXY_URL/graphql" 2>/dev/null || echo "000")
if [ "$code" = "401" ]; then
  echo "proxy unauthenticated -> 401 OK"
elif [ "$code" = "000" ]; then
  echo "warning: could not reach proxy (is it running?)"
  exit 1
else
  echo "proxy unauthenticated: expected 401, got $code"
  exit 1
fi

# Full-flow: login -> session -> refresh -> logout
COOKIE_JAR="$(mktemp -t authsentinel_cookiejar.XXXXXX)"
trap 'rm -f "$COOKIE_JAR"' EXIT

echo "E2E full flows: agent=$AGENT_URL proxy=$PROXY_URL"

# Login + callback: follow only the first two redirects (/login -> /authorize -> /callback)
# and stop before the /callback redirect to the app_base_url.
code=$(curl -s -L -o /dev/null --max-redirs 2 -w "%{http_code}" \
  -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
  "$AGENT_URL/login?redirect_to=/welcome" 2>/dev/null || echo "000")
if [ "$code" != "302" ]; then
  echo "agent login/callback expected 302, got $code"
  exit 1
fi

if [ ! -s "$COOKIE_JAR" ]; then
  echo "expected cookie jar to be populated after login"
  exit 1
fi

# Session verification
session_body="$(curl -sf -b "$COOKIE_JAR" "$AGENT_URL/session" 2>/dev/null || echo "")"
echo "$session_body" | grep -q '"is_authenticated":true' || { echo "session not authenticated"; exit 1; }
echo "$session_body" | grep -q '"sub":"test-user"' || { echo "session missing expected sub"; exit 1; }
echo "login/session OK"

# Refresh: should set a cookie.
refresh_headers="$(curl -s -D - -o /dev/null -b "$COOKIE_JAR" "$AGENT_URL/refresh" 2>/dev/null || true)"
echo "$refresh_headers" | grep -qi 'set-cookie' || { echo "expected Set-Cookie on refresh"; echo "$refresh_headers"; exit 1; }
echo "refresh OK"

# Logout: should clear the cookie and redirect.
logout_headers="$(curl -s -D - -o /dev/null -b "$COOKIE_JAR" "$AGENT_URL/logout?redirect_to=/loggedout" 2>/dev/null || true)"
logout_code="$(echo "$logout_headers" | head -n 1 | awk '{print $2}' || echo "")"
if [ "$logout_code" != "302" ]; then
  echo "logout expected 302, got $logout_code"
  exit 1
fi
echo "$logout_headers" | grep -qi 'max-age=0' || { echo "expected cookie clear on logout"; echo "$logout_headers"; exit 1; }
echo "logout OK"

# After logout, /session should report unauthenticated.
session_body_after="$(curl -s -b "$COOKIE_JAR" "$AGENT_URL/session" 2>/dev/null || echo "")"
echo "$session_body_after" | grep -q '"is_authenticated":false' || { echo "expected is_authenticated=false after logout"; echo "$session_body_after"; exit 1; }

echo "E2E full flows passed"
