$ErrorActionPreference = "Stop"

$AGENT_URL = $env:AGENT_URL
if ([string]::IsNullOrWhiteSpace($AGENT_URL)) { $AGENT_URL = "http://localhost:8080" }

$PROXY_URL = $env:PROXY_URL
if ([string]::IsNullOrWhiteSpace($PROXY_URL)) { $PROXY_URL = "http://localhost:8081" }

Write-Host "E2E full flows (PowerShell): agent=$AGENT_URL proxy=$PROXY_URL"

$cookieJar = [IO.Path]::GetTempFileName()
Remove-Item -Force $cookieJar -ErrorAction SilentlyContinue
New-Item -ItemType File -Path $cookieJar -Force | Out-Null

try {
  # Health
  & curl.exe -sf "$AGENT_URL/healthz" -o NUL
  & curl.exe -sf "$PROXY_URL/healthz" -o NUL

  # Proxy unauthenticated should be 401 when REQUIRE_AUTH is true
  $code = & curl.exe -s -o NUL -w "%{http_code}" "$PROXY_URL/graphql"
  if ($code -ne "401") { throw "proxy unauthenticated expected 401, got $code" }

  # Login -> callback: follow redirects so cookie jar gets populated
  $codeLogin = & curl.exe -s -o NUL -L --max-redirs 2 -w "%{http_code}" -c $cookieJar -b $cookieJar "$AGENT_URL/login?redirect_to=/welcome"
  if ($codeLogin -ne "302") { throw "agent login/callback expected 302, got $codeLogin" }
  if ((Get-Item $cookieJar).Length -lt 1) { throw "expected cookie jar to be populated after login" }

  # Session verification
  $sessionBody = & curl.exe -sf -b $cookieJar "$AGENT_URL/session"
  if ($sessionBody -notmatch '"is_authenticated":true') { throw "session not authenticated" }
  if ($sessionBody -notmatch '"sub":"test-user"') { throw "session missing expected sub" }
  Write-Host "login/session OK"

  # Refresh: should set a new cookie
  $refreshHeaders = & curl.exe -s -D - -o - -b $cookieJar "$AGENT_URL/refresh" | Out-String
  if ($refreshHeaders -notmatch '(?i)set-cookie:') {
    throw "expected Set-Cookie on refresh. refresh response:`n$refreshHeaders"
  }
  Write-Host "refresh OK"

  # Logout: should redirect (302) and clear cookie (max-age=0)
  $logoutHeaders = & curl.exe -s -D - -o - -b $cookieJar "$AGENT_URL/logout?redirect_to=/loggedout" | Out-String
  $logoutCodeLine = ($logoutHeaders -split "`r?`n" | Select-Object -First 1).Trim()
  if ($logoutCodeLine -notmatch 'HTTP/1.1 302') { throw "logout expected 302, got $logoutCodeLine" }
  if ($logoutHeaders -notmatch '(?i)max-age\s*=\s*0') { throw "expected cookie clear on logout. logout response:`n$logoutHeaders" }
  Write-Host "logout OK"

  # After logout, /session should report unauthenticated
  $sessionAfter = & curl.exe -s -b $cookieJar "$AGENT_URL/session"
  if ($sessionAfter -notmatch '"is_authenticated":false') { throw "expected is_authenticated=false after logout" }

  Write-Host "E2E full flows passed"
}
finally {
  if (Test-Path $cookieJar) { Remove-Item $cookieJar -Force -ErrorAction SilentlyContinue }
}

