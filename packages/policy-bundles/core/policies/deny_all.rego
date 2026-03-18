package authsentinel

# Minimal example policy: always deny with 403.
decision := {
  "allow": false,
  "status_code": 403,
  "reason": "denied by policy",
  "headers": {},
  "obligations": {},
}

