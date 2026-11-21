"""آزمون مرورگر برای یک هویت فروشنده با دسترسی متفاوت در دو فروشگاه."""
from os import getenv
from playwright.sync_api import expect, sync_playwright

with sync_playwright() as playwright:
    browser = playwright.chromium.launch(headless=True, channel="chrome")
    page = browser.new_page()
    failures = []
    page.on("pageerror", lambda error: failures.append(str(error)))
    page.goto(getenv("SELLER_TEST_BASE", "http://127.0.0.1:8088"))
    page.wait_for_load_state("networkidle")
    page.get_by_label("شماره تلفن").fill("09000000002")
    page.get_by_role("button", name="دریافت کد", exact=True).click()
    page.get_by_role("button", name="ورود", exact=True).click()
    expect(page.get_by_role("heading", name="عرضه‌های فروشگاه")).to_be_visible()
    stock = page.get_by_role("spinbutton", name="موجودی p1")
    expect(stock).to_be_visible()
    target = "13" if stock.input_value() != "13" else "14"
    stock.fill(target)
    with page.expect_response(lambda res: res.request.method == "PUT") as update:
        page.get_by_role("button", name="ذخیره موجودی").click()
    assert update.value.status == 204, update.value.status
    expect(stock).to_have_value(target)
    page.get_by_role("combobox", name="فروشگاه").select_option("s2")
    expect(page.get_by_text("p2 · موجودی: 24", exact=True)).to_be_visible()
    expect(page.get_by_role("button", name="ذخیره موجودی")).to_have_count(0)
    page.get_by_role("combobox", name="فروشگاه").select_option("s1")
    expect(stock).to_have_value(target)
    page.get_by_role("button", name="خروج", exact=True).click()
    expect(page.get_by_role("heading", name="ورود فروشنده")).to_be_visible()
    assert not failures, failures
    browser.close()
print("Seller stock write/read-only scope and logout passed.")
