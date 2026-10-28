#!/usr/bin/env bash
set -euo pipefail
base=${PLUTO_EXAMPLE_DEMO_BASE_URL:-http://127.0.0.1:8080}
origin=${PLUTO_EXAMPLE_DEMO_ORIGIN:-$base}
umask 077
headers=$(mktemp)
trap 'rm -f "$headers"' EXIT
login() {
  curl --silent --show-error --fail-with-body --dump-header "$headers" \
    --output /dev/null -X POST "$base/auth/session" \
    -H "Origin: $origin" -H "Authorization: Bearer $1"
  python3 - "$headers" <<'PYTHON'
import sys
from http.cookies import SimpleCookie
from pathlib import Path
for line in Path(sys.argv[1]).read_text().splitlines():
    if line.lower().startswith('set-cookie:'):
        cookie=SimpleCookie(line.split(':',1)[1].strip())
        if '__Host-pluto_session' in cookie:
            print(cookie['__Host-pluto_session'].value)
            break
else:
    raise SystemExit('Login did not issue a session cookie')
PYTHON
}
alice=$(login demo_alice_mJ7Kq9vL2xR8)
bob=$(login demo_bob_tW4nC6yP1sF3)
carol=$(login demo_carol_hG8dA5uZ9eQ2)
request() {
  local token=$1 organization=$2 method=$3 path=$4
  shift 4
  local action_key
  action_key=$(python3 -c 'import uuid; print(uuid.uuid4())')
  curl --silent --show-error --fail-with-body \
    -H "Cookie: __Host-pluto_session=$token" -H "Origin: $origin" -H "X-Organization-ID: $organization" \
    -H "Idempotency-Key: $action_key" -H 'Content-Type: application/json' -X "$method" "$base$path" "$@"
}
echo '1. Create a draft as Alice in Northstar'
created=$(request "$alice" northstar POST /api/purchases \
  --data '{"title":"Field workstation","amount_cents":185000}')
python3 -m json.tool <<< "$created"
id=$(python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["id"])' <<< "$created")
echo "Request ID: $id"
echo '2. Submit the draft as Alice'
request "$alice" northstar POST "/api/purchases/$id/submit"
echo '3. Approve it as Bob'
request "$bob" northstar POST "/api/purchases/$id/decision" --data '{"decision":"approved"}'
echo '4. Harbor cannot read the Northstar ID (expected HTTP 404)'
curl --silent --show-error --write-out '\nHTTP %{http_code}\n' \
  -H "Cookie: __Host-pluto_session=$alice" -H "Origin: $origin" -H 'X-Organization-ID: harbor' "$base/api/purchases/$id"
echo '5. Read Alice personal and Bob organization audit histories'
request "$alice" northstar GET /api/audit
request "$bob" northstar GET /api/audit
echo '6. Read visible flags, then export (requires startup export flag=true)'
request "$bob" northstar GET /api/features
request "$bob" northstar GET /api/purchases/export
echo '7. Carol promotes Alice to approver; Alice still cannot decide her own request (HTTP 403)'
request "$carol" northstar PATCH /api/memberships/alice --data '{"role":"approver"}'
curl --silent --show-error --write-out '\nHTTP %{http_code}\n' \
  -H "Cookie: __Host-pluto_session=$alice" -H "Origin: $origin" -H 'X-Organization-ID: northstar' \
  -H 'Idempotency-Key: expected-self-decision' -H 'Content-Type: application/json' -X POST "$base/api/purchases/$id/decision" --data '{"decision":"approved"}'
echo '8. Restore Alice to member'
request "$carol" northstar PATCH /api/memberships/alice --data '{"role":"member"}'

echo '9. Revoke Alice session (logout does not require organization membership)'
request "$alice" '' DELETE /auth/session
