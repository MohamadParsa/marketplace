"""Exercise product Q&A against a fresh loopback API in test mode."""
import os
import re
from pathlib import Path
from urllib.parse import urlparse
from uuid import uuid4

from playwright.sync_api import expect, sync_playwright


API = os.environ.get("QA_TEST_API", "http://127.0.0.1:39390")
STOREFRONT = os.environ.get("QA_TEST_STOREFRONT", "http://127.0.0.1:39391")
BACKOFFICE = os.environ.get("QA_TEST_BACKOFFICE", "http://127.0.0.1:39392")
OUTPUT = Path(os.environ.get("QA_TEST_OUTPUT", "/tmp/marketplace-issue-90-browser"))
for origin in (API, STOREFRONT, BACKOFFICE):
    assert urlparse(origin).hostname in ("127.0.0.1", "localhost"), origin


def checked(response, status):
    assert response.status == status, (response.url, response.status, response.text())
    return response.json() if status != 204 else None


with sync_playwright() as playwright:
    request = playwright.request.new_context(base_url=API)
    browser = playwright.chromium.launch(headless=True, channel="chrome")
    panel = browser.new_page(viewport={"width": 1440, "height": 1000})
    panel.set_default_timeout(10000)
    errors = []
    panel.on("pageerror", lambda error: errors.append(str(error)))
    panel.goto(BACKOFFICE)
    panel.wait_for_load_state("networkidle")
    panel.get_by_label("شماره موبایل یا ایمیل").fill("09000000001")
    panel.get_by_role("button", name="ادامه", exact=True).click()
    panel.get_by_role("button", name="دریافت کد یک‌بارمصرف").click()
    code = re.search(r"\d{6}", panel.get_by_role("alert").filter(has_text="کد آزمون").inner_text()).group()
    panel.get_by_label("کد یک‌بارمصرف").fill(code)
    with panel.expect_response("**/backoffice/auth/verify-otp") as verified:
        panel.get_by_role("button", name="ورود به پنل", exact=True).click()
    session = checked(verified.value, 200)
    admin = {"Authorization": "Bearer " + session["token"]}

    def command(**body):
        checked(request.post("/backoffice/commands", headers=admin, data={
            **body, "idempotency_key": str(uuid4()),
        }), 204)

    def workspace():
        return checked(request.get("/backoffice/workspace", headers=admin), 200)

    category_name = "دسته آزمون پرسش محصول " + uuid4().hex[:6]
    command(action="catalog.category.create", catalog={"name": category_name, "english_name": "QA test " + uuid4().hex[:6], "published": True})
    category = next(c for c in workspace()["categories"] if c["name"] == category_name)
    content = {"qa_enabled": True, "shipping": {"weight_unit": "g", "dimension_unit": "mm"}}
    catalog = {
        "category_id": category["id"], "name": "محصول " + category_name,
        "model": "shared", "content": content,
    }
    command(action="catalog.product.create", catalog=catalog)
    product = next(p for p in workspace()["products"] if p["name"] == catalog["name"])
    product_id = product["id"]
    icon = Path(__file__).with_name("fixtures") / "catalog-icon.png"
    checked(request.post("/backoffice/catalog/product/" + product_id + "/icon", headers=admin, multipart={
        "file": {"name": "sample.png", "mimeType": "image/png", "buffer": icon.read_bytes()},
        "version": str(product["version"]), "idempotency_key": str(uuid4()),
    }), 204)
    state = workspace()
    product = next(p for p in state["products"] if p["id"] == product_id)
    variant = next(v for v in state["variants"] if v["product_id"] == product_id)
    command(action="offer.create", target=product_id, version=product["variant_version"], offer={
        "product_id": product_id, "variant_id": variant["id"], "store_id": "s1", "price": 1000000, "stock": 10,
    })

    def toggle(enabled):
        product = next(p for p in workspace()["products"] if p["id"] == product_id)
        command(action="catalog.product.update", target=product_id, version=product["version"],
                catalog={**catalog, "published": True, "content": {**content, "qa_enabled": enabled}})

    toggle(True)
    detail = checked(request.get("/storefront/products/" + product_id), 200)
    url = STOREFRONT + detail["url"]
    mobile = browser.new_page(viewport={"width": 390, "height": 844}, is_mobile=True, has_touch=True)
    mobile.set_default_timeout(10000)
    mobile.on("pageerror", lambda error: errors.append(str(error)))
    mobile.goto(url, wait_until="domcontentloaded")
    qa = mobile.locator("#questions")
    qa.get_by_role("link", name="برای ثبت پرسش وارد شوید").click()
    mobile.get_by_label("موبایل یا ایمیل").fill("09000000005")
    mobile.get_by_role("button", name="ورود با OTP / ثبت‌نام").click()
    mobile.get_by_role("button", name="دریافت کد تأیید", exact=True).click()
    code = re.search(r"\d{6}", mobile.get_by_text("کد محیط آزمون:", exact=False).inner_text()).group()
    mobile.get_by_label("کد تأیید").fill(code)
    mobile.get_by_role("button", name="تأیید و ورود").click()
    field = qa.get_by_label("پرسش شما")
    expect(field).to_be_visible()
    question = "آیا این محصول برای استفاده روزانه مناسب است؟"
    answer = "بله، برای استفاده روزانه مناسب است."
    keys = []

    def fail_once(route):
        if route.request.method != "POST":
            route.continue_()
            return
        keys.append(route.request.headers["idempotency-key"])
        if len(keys) == 1:
            route.fulfill(status=503, json={"code": "qa_unavailable"})
        else:
            route.continue_()

    mobile.route("**/storefront/products/*/questions", fail_once)
    field.fill(question)
    field.press("Tab")
    submit = qa.get_by_role("button", name="ثبت پرسش")
    expect(submit).to_be_focused()
    submit.press("Enter")
    expect(qa.get_by_role("alert")).to_be_visible()
    expect(field).to_have_value(question)
    submit.click()
    expect(qa.get_by_role("status")).to_contain_text("پرسش ثبت شد")
    assert len(keys) == 2 and keys[0] == keys[1], keys
    expect(qa.get_by_text(question, exact=True)).to_have_count(0)

    panel.locator('nav a[href="/products"]').click()
    panel.locator(f'a[href="/products/{product_id}"]').click()
    panel.get_by_role("tab", name="پرسش و پاسخ").click()
    editor = panel.get_by_role("article").filter(has_text=question)
    expect(editor.get_by_text("در انتظار تأیید")).to_be_visible()
    editor.get_by_role("button", name="انتشار پرسش").click()
    expect(editor.get_by_text("تأییدشده")).to_be_visible()
    toggle(False)
    editor.get_by_label("پاسخ رسمی").fill(answer)
    editor.get_by_role("button", name="ثبت پاسخ برای بررسی").click()
    expect(editor.get_by_text(answer, exact=True)).to_be_visible()
    mobile.reload()
    expect(qa.get_by_text(question, exact=True)).to_be_visible()
    expect(qa.get_by_text(answer, exact=True)).to_have_count(0)
    expect(qa.get_by_text("ثبت پرسش جدید برای این محصول غیرفعال است.")).to_be_visible()
    editor.get_by_role("button", name="انتشار پاسخ 1").click()
    expect(editor.get_by_text("تأییدشده")).to_have_count(2)
    mobile.reload()
    expect(qa.get_by_text(answer, exact=True)).to_be_visible()
    assert mobile.evaluate("document.documentElement.scrollWidth === innerWidth")
    OUTPUT.mkdir(parents=True, exist_ok=True)
    qa.screenshot(path=OUTPUT / "storefront-mobile.png")
    editor.screenshot(path=OUTPUT / "backoffice-moderation.png")
    editor.get_by_role("button", name="رد پاسخ 1").click()
    expect(editor.get_by_text("ردشده")).to_be_visible()
    mobile.reload()
    expect(qa.get_by_text(question, exact=True)).to_be_visible()
    expect(qa.get_by_text(answer, exact=True)).to_have_count(0)
    assert not errors, errors
    print("Product Q&A browser checks passed")
    browser.close()
