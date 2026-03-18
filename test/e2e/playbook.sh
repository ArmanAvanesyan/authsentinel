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

# Agent /login should redirect (302) to IdP or return 200
code=$(curl -s -o /dev/null -w "%{http_code}" "$AGENT_URL/login" 2>/dev/null || echo "000")
if [ "$code" = "302" ] || [ "$code" = "200" ]; then
  echo "agent /login responded OK"
else
  echo "agent /login: expected 302 or 200, got $code"
fi

echo "E2E smoke passed"
