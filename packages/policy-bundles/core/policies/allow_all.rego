package authsentinel

# Minimal example policy: always allow.
#
# The proxy evaluates `data.authsentinel.decision` and expects an object result
# compatible with pkg/policy.Decision.
decision := {
  "allow": true,
  "status_code": 200,
  "reason": "",
  "headers": {},
  "obligations": {},
}

