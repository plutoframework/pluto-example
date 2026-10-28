#!/usr/bin/env python3
"""Synthetic loopback walkthrough. Secrets stay in memory and Mailpit only."""
import http.client
import json
import os
from pathlib import Path
import re
import secrets
import subprocess
import time
from urllib.parse import urlsplit
import uuid

BASE = os.environ.get("PLUTO_EXAMPLE_DEMO_BASE_URL", "http://127.0.0.1:8080")
MAIL = os.environ.get("PLUTO_EXAMPLE_DEMO_MAILPIT_URL", "http://127.0.0.1:8025")


def loopback(value):
    url = urlsplit(value)
    if url.scheme != "http" or url.hostname not in ("127.0.0.1", "::1") or url.path not in ("", "/"):
        raise SystemExit("Walkthrough requires a numeric loopback HTTP base")
    return url


loopback(BASE)
loopback(MAIL)


def call(method, path, data=None, cookie=None, tenant=None, want=200):
    url = loopback(BASE)
    connection = http.client.HTTPConnection(url.hostname, url.port, timeout=15)
    headers = {"Origin": BASE, "Content-Type": "application/json", "Idempotency-Key": str(uuid.uuid4())}
    if cookie:
        headers["Cookie"] = "__Host-pluto_session=" + cookie
    if tenant:
        headers["X-Organization-ID"] = tenant
    body = json.dumps(data) if data is not None else None
    try:
        connection.request(method, path, body, headers)
        response = connection.getresponse()
        payload = response.read()
        if response.status != want:
            raise SystemExit(f"{method} {path}: HTTP {response.status}, expected {want}; use a fresh disposable database")
        token = None
        for key, value in response.getheaders():
            if key.lower() == "set-cookie":
                token = value.split(";", 1)[0].split("=", 1)[1]
        return json.loads(payload), token
    finally:
        connection.close()


def mail_token(email, purpose):
    url = loopback(MAIL)
    for _ in range(30):
        connection = http.client.HTTPConnection(url.hostname, url.port, timeout=5)
        try:
            connection.request("GET", "/api/v1/messages?limit=100")
            response = connection.getresponse()
            if response.status != 200:
                raise SystemExit("Mailpit unavailable")
            messages = json.loads(response.read()).get("messages", [])
            for message in messages:
                if not any(item["Address"] == email for item in message.get("To", [])) or purpose not in message.get("Subject", ""):
                    continue
                connection.request("GET", "/api/v1/message/" + message["ID"])
                detail = connection.getresponse()
                if detail.status != 200:
                    raise SystemExit("Mailpit message unavailable")
                text = json.loads(detail.read()).get("Text", "")
                match = re.search(r"^Token: ([A-Za-z0-9_-]{43})\s*$", text, re.MULTILINE)
                if match:
                    return match[1]
        finally:
            connection.close()
        time.sleep(0.1)
    raise SystemExit("No local notification; inspect Mailpit and request an explicit resend")


cookies = {}
subjects = {}
passwords = {}
for name in ("alice", "bob", "carol", "dana"):
    email = name + "@example.test"
    password = secrets.token_urlsafe(24)
    passwords[name] = password
    call("POST", "/auth/enroll", {"email": email, "password": password}, want=202)
    call("POST", "/auth/signin", {"email": email, "password": password}, want=401)
    token = mail_token(email, "verify_email")
    call("POST", "/auth/verify", {"token": token})
    call("POST", "/auth/verify", {"token": token}, want=400)
    _, cookies[name] = call("POST", "/auth/signin", {"email": email, "password": password})
    metadata, _ = call("GET", "/auth/session", cookie=cookies[name])
    subjects[name] = metadata["data"]["actor_id"]
    call("GET", "/api/purchases", cookie=cookies[name], tenant="northstar", want=401)
print("1. Enrolled, verified and signed in four synthetic accounts; none received a membership")

# Explicit trusted local administration. No password/token goes to the command.
command = ["go", "run", "./cmd/dev-fixtures"]
for name, subject in subjects.items():
    command += ["--" + name + "-id", subject]
subprocess.run(command, cwd=Path(__file__).resolve().parent.parent, check=True)
print("2. Bound explicit account IDs to DEVELOPMENT fixture memberships")

alice, bob, carol = (cookies[name] for name in ("alice", "bob", "carol"))
created, _ = call("POST", "/api/purchases", {"title": "Field workstation", "amount_cents": 185000}, alice, "northstar", 201)
purchase_id = created["data"]["id"]
call("POST", f"/api/purchases/{purchase_id}/submit", cookie=alice, tenant="northstar")
call("POST", f"/api/purchases/{purchase_id}/decision", {"decision": "approved"}, bob, "northstar")
call("GET", f"/api/purchases/{purchase_id}", cookie=alice, tenant="harbor", want=404)
call("GET", "/api/purchases", cookie=bob, tenant="harbor", want=401)
for cookie in (alice, bob):
    call("GET", "/api/audit", cookie=cookie, tenant="northstar")
page, _ = call("GET", "/api/purchases?limit=1", cookie=bob, tenant="northstar")
call("GET", "/api/features", cookie=bob, tenant="northstar")
call("GET", "/api/purchases/export", cookie=bob, tenant="northstar")
call("PATCH", "/api/memberships/" + subjects["alice"], {"role": "approver"}, carol, "northstar")
call("POST", f"/api/purchases/{purchase_id}/decision", {"decision": "approved"}, alice, "northstar", 403)
call("PATCH", "/api/memberships/" + subjects["alice"], {"role": "member"}, carol, "northstar")
print("3. Purchase submitted/approved; tenant isolation, audit, cursor page, flags and role policy checked")

call("POST", "/auth/recovery", {"email": "alice@example.test"}, want=202)
recovery = mail_token("alice@example.test", "recover_password")
replacement = secrets.token_urlsafe(24)
call("POST", "/auth/reset", {"token": recovery, "password": replacement})
call("GET", "/api/purchases", cookie=alice, tenant="northstar", want=401)
call("POST", "/auth/signin", {"email": "alice@example.test", "password": passwords["alice"]}, want=401)
_, fresh = call("POST", "/auth/signin", {"email": "alice@example.test", "password": replacement})
call("POST", "/auth/logout", cookie=fresh)
call("GET", "/auth/session", cookie=fresh, want=401)
for name in ("bob", "carol", "dana"):
    call("POST", "/auth/logout", cookie=cookies[name])
print("4. Recovery/reset revoked the old session/password; new sign-in and logout checked")
print("Walkthrough passed. Passwords/tokens were not printed or written to disk. Discard the database and Mailpit now.")
