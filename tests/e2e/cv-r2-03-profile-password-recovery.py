"""پوشش مرورگر و API برای CV-R2-03 در محیط داده آزمون."""
import json
import re
from pathlib import Path
from playwright.sync_api import sync_playwright, expect

BASE = "http://127.0.0.1:8087"
ARTIFACT = Path("docs/delivery/review-v1-1/evidence/cv-r2-03-profile-password-recovery.png")

with sync_playwright() as p:
    browser = p.chromium.launch(headless=True, channel="chrome")
    page = browser.new_page(viewport={"width": 1440, "height": 1000})
    page.goto(BASE)
    challenge = page.request.post(f"{BASE}/backoffice/auth/request-otp", data={"phone": "09000000001"})
    expect(challenge).to_be_ok()
    challenge_json = challenge.json()
    verified = page.request.post(f"{BASE}/backoffice/auth/verify-otp", data={"request_id": challenge_json["request_id"], "code": challenge_json["test_code"]})
    expect(verified).to_be_ok()
    token = verified.json()["token"]
    expect(page.get_by_role("heading", name="ورود به بک‌آفیس")).to_be_visible()
    headers = {"Authorization": f"Bearer {token}"}

    profile = page.request.get(f"{BASE}/backoffice/account/profile", headers=headers)
    expect(profile).to_be_ok()
    before = profile.json()
    updated = page.request.patch(f"{BASE}/backoffice/account/profile", headers=headers, data={
        "name": before["name"], "email": "admin@example.test", "password": "test-password-123"
    })
    expect(updated).to_be_ok()
    assert updated.json()["user_id"] == before["user_id"]
    password_login = page.request.post(f"{BASE}/backoffice/auth/password-login", data={
        "phone": "09000000001", "password": "test-password-123"
    })
    expect(password_login).to_be_ok()
    email_login = page.request.post(f"{BASE}/backoffice/auth/password-login", data={
        "phone": "admin@example.test", "password": "test-password-123"
    })
    expect(email_login).to_be_ok()
    reset = page.request.post(f"{BASE}/backoffice/auth/admin-password-reset", headers=headers, data={
        "surface": "backoffice", "phone": "09000000002", "password": "reset-password-123"
    })
    expect(reset).to_be_ok()
    target_login = page.request.post(f"{BASE}/backoffice/auth/password-login", data={
        "phone": "09000000002", "password": "reset-password-123"
    })
    expect(target_login).to_be_ok()
    page.screenshot(path=str(ARTIFACT), full_page=True)
    print(json.dumps({"profile_user_id": before["user_id"], "same_user_email_login": True, "admin_reset": True}, ensure_ascii=False))
    browser.close()
