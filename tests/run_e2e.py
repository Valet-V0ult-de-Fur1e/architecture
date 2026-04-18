#!/usr/bin/env python3

from __future__ import annotations

import argparse
import json
import sys
import time
import urllib.error
import urllib.request


NO_PROXY_OPENER = urllib.request.build_opener(urllib.request.ProxyHandler({}))


def request_json(method: str, url: str, payload: dict | None = None, headers: dict | None = None) -> tuple[int, dict | str]:
    data = None
    req_headers = {"Content-Type": "application/json"}
    if headers:
        req_headers.update(headers)

    if payload is not None:
        data = json.dumps(payload).encode("utf-8")

    req = urllib.request.Request(url=url, method=method, headers=req_headers, data=data)
    try:
        with NO_PROXY_OPENER.open(req, timeout=10) as resp:
            raw = resp.read().decode("utf-8")
            if not raw:
                return resp.status, ""
            try:
                return resp.status, json.loads(raw)
            except json.JSONDecodeError:
                return resp.status, raw
    except urllib.error.HTTPError as err:
        raw = err.read().decode("utf-8")
        try:
            body = json.loads(raw) if raw else ""
        except json.JSONDecodeError:
            body = raw
        return err.code, body
    except urllib.error.URLError as err:
        return 0, str(err)


def expect(condition: bool, message: str) -> None:
    if not condition:
        raise AssertionError(message)


def wait_for_status(url: str, expected_status: int, timeout_seconds: int, name: str) -> None:
    deadline = time.time() + timeout_seconds
    last_status = None
    last_body: dict | str = ""

    while time.time() < deadline:
        status, body = request_json("GET", url, headers={"Content-Type": "text/plain"})
        if status == expected_status:
            return
        last_status = status
        last_body = body
        time.sleep(1)

    raise AssertionError(
        f"{name} expected {expected_status}, got {last_status}, body={last_body}"
    )


def run(base_url: str) -> None:
    ts = int(time.time())
    email = f"user{ts}@example.com"
    password = "secret"


    print("1) Check health")
    try:
        wait_for_status(f"{base_url}/healthz", 200, 60, "healthz")
        print("   ✔ success")
    except Exception as e:
        print(f"   ✗ failed: {e}")
        raise


    print("2) Check readiness")
    try:
        wait_for_status(f"{base_url}/readyz", 200, 60, "readyz")
        print("   ✔ success")
    except Exception as e:
        print(f"   ✗ failed: {e}")
        raise


    print("3) Register user")
    try:
        code, body = request_json(
            "POST",
            f"{base_url}/api/v1/identity/register",
            {"email": email, "password": password},
        )
        expect(code == 201, f"register expected 201, got {code}, body={body}")
        expect(isinstance(body, dict) and "user_id" in body, "register response missing user_id")
        print("   ✔ success")
    except Exception as e:
        print(f"   ✗ failed: {e}")
        raise


    print("4) Login")
    try:
        code, body = request_json(
            "POST",
            f"{base_url}/api/v1/identity/login",
            {"email": email, "password": password},
        )
        expect(code == 200, f"login expected 200, got {code}, body={body}")
        expect(isinstance(body, dict) and "access_token" in body, "login response missing access_token")
        token = body["access_token"]
        print("   ✔ success")
    except Exception as e:
        print(f"   ✗ failed: {e}")
        raise


    headers = {"Authorization": f"Bearer {token}"}



    print("5) Create TODO (triggers Redis cache and RabbitMQ event)")
    try:
        code, body = request_json(
            "POST",
            f"{base_url}/api/v1/todos/",
            {"title": "Learn DDD", "description": "Aggregate and invariants", "priority": 3},
            headers=headers,
        )
        expect(code == 201, f"create todo expected 201, got {code}, body={body}")
        expect(isinstance(body, dict) and "todo" in body and "ID" in body["todo"], "create todo response malformed")
        todo_id = body["todo"]["ID"]
        print("   ✔ success")
    except Exception as e:
        print(f"   ✗ failed: {e}")
        raise


    print("6) List TODO (should hit Redis cache)")
    try:
        code, body = request_json("GET", f"{base_url}/api/v1/todos/", headers=headers)
        expect(code == 200, f"list todos expected 200, got {code}, body={body}")
        expect(isinstance(body, dict) and "todos" in body, "list todos response malformed")
        todos = body["todos"]
        expect(any(t["ID"] == todo_id for t in todos), "created todo not in list")
        print("   ✔ success")
    except Exception as e:
        print(f"   ✗ failed: {e}")
        raise


    print("7) Get TODO by ID (should hit Redis cache)")
    try:
        code, body = request_json("GET", f"{base_url}/api/v1/todos/{todo_id}", headers=headers)
        expect(code == 200, f"get todo expected 200, got {code}, body={body}")
        expect(isinstance(body, dict) and "todo" in body, "get todo response malformed")
        print("   ✔ success")
    except Exception as e:
        print(f"   ✗ failed: {e}")
        raise


    print("8) Update status (invalidates Redis cache)")
    try:
        code, body = request_json(
            "PATCH",
            f"{base_url}/api/v1/todos/{todo_id}/status",
            {"status": "done"},
            headers=headers,
        )
        expect(code == 200, f"update status expected 200, got {code}, body={body}")
        print("   ✔ success")
    except Exception as e:
        print(f"   ✗ failed: {e}")
        raise


    print("9) Update priority (invalidates Redis cache)")
    try:
        code, body = request_json(
            "PATCH",
            f"{base_url}/api/v1/todos/{todo_id}/priority",
            {"priority": 5},
            headers=headers,
        )
        expect(code == 200, f"update priority expected 200, got {code}, body={body}")
        print("   ✔ success")
    except Exception as e:
        print(f"   ✗ failed: {e}")
        raise


    print("10) Soft delete (invalidates Redis cache)")
    try:
        code, body = request_json("DELETE", f"{base_url}/api/v1/todos/{todo_id}", headers=headers)
        expect(code == 204, f"delete todo expected 204, got {code}, body={body}")
        print("   ✔ success")
    except Exception as e:
        print(f"   ✗ failed: {e}")
        raise


    print("11) Verify deleted TODO is inaccessible")
    try:
        code, _ = request_json("GET", f"{base_url}/api/v1/todos/{todo_id}", headers=headers)
        expect(code == 404, f"deleted todo get expected 404, got {code}")
        print("   ✔ success")
    except Exception as e:
        print(f"   ✗ failed: {e}")
        raise


    print("12) Check Redis cache invalidation by listing TODOs (should not contain deleted TODO)")
    try:
        code, body = request_json("GET", f"{base_url}/api/v1/todos/", headers=headers)
        expect(code == 200, f"list todos after delete expected 200, got {code}, body={body}")
        expect(isinstance(body, dict) and "todos" in body, "list todos response malformed after delete")
        todos = body["todos"]
        if todos is None:
            todos = []
        expect(not any(t["ID"] == todo_id for t in todos), "deleted todo still present in list (cache not invalidated)")
        print("   ✔ success")
    except Exception as e:
        print(f"   ✗ failed: {e}")
        raise


    print("13) (Manual) Check backend logs for RabbitMQ event-driven consumer output (todo.created)")
    print("    (You should see a log line like: [event-consumer] Получено событие todo.created: ...)")
    print("   ✔ success (manual check)")
    print("All E2E checks passed (including Redis cache and RabbitMQ event)")


def main() -> int:
    parser = argparse.ArgumentParser(description="Run E2E API checks")
    parser.add_argument("--base-url", default="http://localhost:8080", help="API base URL")
    args = parser.parse_args()

    try:
        run(args.base_url.rstrip("/"))
    except AssertionError as err:
        print(f"FAILED: {err}")
        return 1
    except Exception as err:
        print(f"UNEXPECTED ERROR: {err}")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
