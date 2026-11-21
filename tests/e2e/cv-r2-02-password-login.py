"""پوشش مرورگر CV-R2-02 برای انتخاب رمز پس از واردکردن موبایل."""
from pathlib import Path
from playwright.sync_api import sync_playwright, expect

BASE = "http://127.0.0.1:5173"
ARTIFACT = Path("docs/delivery/review-v1-1/evidence/cv-r2-02-password-login.png")

with sync_playwright() as p:
    browser = p.chromium.launch(headless=True, channel="chrome")
    page = browser.new_page(viewport={"width": 1440, "height": 1000})
    page.goto(BASE)
    page.wait_for_load_state("networkidle")
    page.get_by_role("button", name="دریافت کد ورود").click()
    expect(page.get_by_role("button", name="ورود با رمز")).to_be_visible()
    page.get_by_role("button", name="ورود با رمز").click()
    page.get_by_label("رمز ورود").fill("test-password-123")
    page.get_by_role("button", name="ورود با رمز").click()
    expect(page.get_by_role("heading", name="نمای کلی بازارگاه")).to_be_visible()
    page.screenshot(path=str(ARTIFACT), full_page=True)
    browser.close()
