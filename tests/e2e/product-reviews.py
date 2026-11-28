"""Exercise #89 against a fresh loopback APP_ENV=test API and storefront."""
import argparse
from pathlib import Path
from urllib.parse import urlparse
from uuid import uuid4

from playwright.sync_api import expect, sync_playwright


def exercise(browser, backoffice, storefront):
    panel = browser.new_page(viewport={"width": 1440, "height": 960})
    site = browser.new_page(viewport={"width": 1440, "height": 960})
    page_errors = []
    for page in (panel, site):
        page.set_default_timeout(10000)
        page.on("pageerror", lambda error: page_errors.append(str(error)))

    panel.goto(backoffice)
    panel.wait_for_load_state("networkidle")
    panel.get_by_label("شماره موبایل یا ایمیل").fill("09000000001")
    panel.get_by_role("button", name="ادامه", exact=True).click()
    with panel.expect_response(lambda r: r.url.endswith("/backoffice/auth/request-otp")) as otp:
        panel.get_by_role("button", name="دریافت کد یک‌بارمصرف", exact=True).click()
    panel.get_by_label("کد یک‌بارمصرف").fill(otp.value.json()["test_code"])
    with panel.expect_response(lambda r: r.url.endswith("/backoffice/auth/verify-otp")) as verification:
        panel.get_by_role("button", name="ورود به پنل", exact=True).click()
    headers = {"Authorization": "Bearer " + verification.value.json()["token"]}

    def workspace():
        response = panel.request.get(backoffice + "/backoffice/workspace", headers=headers)
        assert response.ok, response.text()
        return response.json()

    def command(**values):
        response = panel.request.post(
            backoffice + "/backoffice/commands",
            data={**values, "idempotency_key": str(uuid4())},
            headers=headers,
        )
        assert response.ok, response.text()

    unique = uuid4().hex[:8]
    category_name = "دسته آزمون نظر " + unique
    command(action="catalog.category.create", catalog={
        "name": category_name, "english_name": "Review category " + unique, "published": True,
    })
    category = next(c for c in workspace()["categories"] if c["name"] == category_name)
    product_name = "محصول آزمون نظر " + unique
    catalog = {
        "name": product_name,
        "category_id": category["id"],
        "published": False,
        "content": {"reviews_enabled": True},
    }
    command(action="catalog.product.create", catalog=catalog)
    product = next(p for p in workspace()["products"] if p["category_id"] == category["id"])
    product_id = product["id"]
    photos = Path(__file__).resolve().parents[2] / "docs/review-packages/cherriva-2026-11-17-v1-1/01-storefront/assets"
    photo = photos / "iphone.jpg"
    upload = panel.request.post(
        backoffice + f"/backoffice/catalog/product/{product_id}/icon",
        headers=headers,
        multipart={
            "version": str(product["version"]),
            "idempotency_key": str(uuid4()),
            "file": {"name": photo.name, "mimeType": "image/jpeg", "buffer": photo.read_bytes()},
        },
    )
    assert upload.ok, upload.text()
    product = next(p for p in workspace()["products"] if p["id"] == product_id)
    variant = next(v for v in workspace()["variants"] if v["product_id"] == product_id)
    command(
        action="offer.create",
        target=product_id,
        version=product["variant_version"],
        offer={"product_id": product_id, "variant_id": variant["id"], "store_id": "s1", "price": 1000000, "stock": 5},
    )
    catalog["published"] = True
    command(action="catalog.product.update", target=product_id, version=product["version"], catalog=catalog)
    product_path = f"/storefront/products/{product_id}"
    detail = site.request.get(storefront + product_path)
    assert detail.ok, detail.text()
    url = storefront + detail.json()["url"] + "#reviews"
    site.goto(url, wait_until="domcontentloaded")
    reviews = site.locator("#reviews")
    expect(reviews.get_by_text("هنوز نظری منتشر نشده است.")).to_be_visible()
    reviews.get_by_label("نام نمایشی").fill("مهمان نمونه")
    reviews.get_by_label("انتشار بدون نام").check()
    reviews.get_by_role("radio", name="۵", exact=True).check()
    anonymous_review = "نظر. متن نمونه <b>به صورت متن</b>"
    reviews.get_by_label("متن نظر").fill(anonymous_review)
    review_url = storefront + product_path + "/reviews"
    site.route(review_url, lambda route: route.fulfill(status=503))
    reviews.get_by_role("button", name="ارسال برای بررسی").click()
    expect(reviews.get_by_role("alert")).to_contain_text("ثبت نظر ممکن نشد")
    expect(reviews.get_by_label("متن نظر")).to_have_value(anonymous_review)
    site.unroute(review_url)
    reviews.get_by_role("button", name="ارسال برای بررسی").click()
    expect(reviews.get_by_role("status")).to_contain_text("نظر شما ثبت شد")
    expect(reviews.get_by_text(anonymous_review, exact=True)).to_have_count(0)

    for i in range(5):
        response = site.request.post(
            storefront + product_path + "/reviews",
            data={"name": "کاربر آزمون", "rating": 4, "text": f"نظر نمونه {i}", "idempotency_key": str(uuid4())},
        )
        assert response.status == 204, response.text()

    panel.get_by_role("link", name="نظرات محصولات", exact=True).click()
    expect(panel.get_by_role("heading", name="نظرات محصولات", exact=True).first).to_be_visible()
    product_cards = panel.get_by_role("article").filter(has_text=product_name)
    for remaining in range(6, 0, -1):
        approvals = product_cards.get_by_role("button", name="تأیید و انتشار", exact=True)
        with panel.expect_response(lambda r: r.url.endswith("/decisions") and r.status == 204):
            approvals.first.click()
        expect(approvals).to_have_count(remaining - 1)
    expect(panel.get_by_text("نظری در این بخش وجود ندارد.")).to_be_visible()
    panel.get_by_role("tab", name="تأییدشده", exact=True).click()
    anonymous_card = product_cards.filter(has_text=anonymous_review)
    expect(anonymous_card.get_by_text("مهمان نمونه", exact=False)).to_be_visible()
    anonymous_card.get_by_role("button", name="نوشتن پاسخ").click()
    reply = "پاسخ آزمایشی چریوا"
    anonymous_card.get_by_label("متن پاسخ تیم").fill(reply)
    anonymous_card.get_by_role("button", name="انتشار پاسخ", exact=True).click()
    expect(anonymous_card.get_by_text(reply, exact=False)).to_be_visible()

    site.reload(wait_until="domcontentloaded")
    groups = reviews.locator(".pdp-review-group")
    expect(groups).to_have_count(1)
    expect(groups.first.get_by_role("article")).to_have_count(5)
    site.set_viewport_size({"width": 390, "height": 844})
    reviews.get_by_role("button", name="رفتن به نظرات بیشتر").click()
    expect(groups).to_have_count(2)
    expect(groups.last.get_by_role("article")).to_have_count(1)
    expect(groups.last.get_by_text("بدون نام", exact=True)).to_be_in_viewport()
    expect(groups.last.get_by_text(anonymous_review, exact=True)).to_be_visible()
    expect(groups.last.get_by_text(reply, exact=True)).to_be_visible()
    expect(groups.last.locator("p b")).to_have_count(0)
    expect(reviews.get_by_text("مهمان نمونه", exact=True)).to_have_count(0)

    product = next(p for p in workspace()["products"] if p["id"] == product_id)
    catalog["content"]["reviews_enabled"] = False
    command(action="catalog.product.update", target=product_id, version=product["version"], catalog=catalog)
    site.reload(wait_until="domcontentloaded")
    expect(reviews.get_by_text("ثبت نظر برای این محصول غیرفعال است.")).to_be_visible()
    expect(reviews.get_by_role("button", name="ارسال برای بررسی")).to_have_count(0)
    expect(groups.first.get_by_role("article")).to_have_count(5)
    assert not page_errors, page_errors


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--backoffice", required=True)
    parser.add_argument("--storefront", required=True)
    parser.add_argument("--chromium-path")
    args = parser.parse_args()
    for origin in (args.backoffice, args.storefront):
        if urlparse(origin).hostname not in ("localhost", "127.0.0.1"):
            parser.error("Use a local synthetic test instance.")
    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(
            executable_path=args.chromium_path, headless=True
        )
        try:
            exercise(browser, args.backoffice, args.storefront)
            print("Product reviews browser acceptance passed.", flush=True)
        finally:
            browser.close()
