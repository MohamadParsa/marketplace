// Run seed-content-review.cjs first; this uses synthetic review identities only.
async (page) => {
  const checks = [],
    errors = [];
  const check = (id, pass) => {
    checks.push({ id, pass: !!pass });
    if (!pass) throw Error(id);
  };
  page.on("pageerror", (e) => errors.push(e.message));
  const api = "http://127.0.0.1:8098",
    site = "http://127.0.0.1:5197";
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page.goto(api);
  await page
    .getByLabel(/شماره موبایل یا ایمیل/)
    .fill("review-admin@example.invalid");
  await page.getByRole("button", { name: "ادامه", exact: true }).click();
  await page.getByLabel(/^رمز عبور/).fill("Review-only-27-pass");
  const auth = page.waitForResponse(
    (r) => r.url().endsWith("/backoffice/auth/password-login") && r.ok(),
  );
  await page
    .getByRole("button", { name: "ورود با رمز عبور", exact: true })
    .click();
  const adminToken = (await (await auth).json()).token;
  await page
    .getByRole("heading", { name: "نمای کلی بازارگاه", exact: true })
    .waitFor();
  const get = async (path, token) => {
    const r = await page.request.get(api + path, {
      headers: { Authorization: "Bearer " + token },
    });
    if (!r.ok()) throw Error(path + " " + r.status());
    return r.json();
  };
  const original = await get("/backoffice/workspace", adminToken);
  const p = original.products.find(
    (p) => p.name === "محصول مستقل آزمون بررسی محتوا",
  );
  if (!p || p.status !== "published")
    throw Error("Run seed-content-review.cjs to create a published product");
  const offersBefore = JSON.stringify(
    original.offers.filter((o) => o.product_id === p.id),
  );
  const publicProduct = async () => {
    const r = await page.request.get(site + "/storefront/products/" + p.id);
    return r.json();
  };
  const openProduct = async (tab) => {
    const menu = tab.getByRole("button", { name: "باز کردن منو", exact: true });
    if (await menu.isVisible()) await menu.click();
    await tab.locator('nav a[href="/products"]').click();
    await tab
      .getByRole("row")
      .filter({ hasText: p.id })
      .getByRole("link", { name: "جزئیات محصول", exact: true })
      .click();
  };
  const seller = await page.context().newPage();
  seller.on("pageerror", (e) => errors.push(e.message));
  await seller.goto(api + "/seller-login");
  await seller.getByLabel(/شماره موبایل یا ایمیل/).fill("09120000002");
  await seller.getByRole("button", { name: "ادامه", exact: true }).click();
  const challenge = seller.waitForResponse(
    (r) => r.url().endsWith("/seller/auth/request-otp") && r.ok(),
  );
  await seller
    .getByRole("button", { name: "دریافت کد یک‌بارمصرف", exact: true })
    .click();
  const otp = await (await challenge).json();
  await seller.getByLabel(/^کد یک‌بارمصرف/).fill(otp.test_code);
  const verified = seller.waitForResponse(
    (r) => r.url().endsWith("/seller/auth/verify-otp") && r.ok(),
  );
  await seller
    .getByRole("button", { name: "ورود به پنل", exact: true })
    .click();
  const sellerToken = (await (await verified).json()).token;
  await seller
    .getByRole("heading", { name: "محصولات و بررسی محتوا", exact: true })
    .waitFor();
  check("shared-panel-seller-login", seller.url().endsWith("/products"));
  await openProduct(seller);
  await seller
    .getByRole("button", { name: "ویرایش محتوای فروشگاه", exact: true })
    .click();
  let dialog = seller.getByRole("dialog").last();
  await dialog
    .getByRole("textbox", { name: "نام فارسی", exact: true })
    .fill("عنوان جدید فروشنده — بررسی");
  await dialog
    .getByRole("textbox", { name: "توضیح کوتاه", exact: true })
    .fill("شرح جدید منتظر بررسی");
  await dialog
    .getByRole("button", { name: "ذخیره / ارسال محتوا", exact: true })
    .click();
  await dialog.waitFor({ state: "hidden" });
  await seller
    .getByRole("heading", {
      name: "ویرایش فروشنده: در انتظار بررسی",
      exact: true,
    })
    .waitFor();
  check(
    "pending-preserves-public-approved-name",
    (await publicProduct()).name === p.name,
  );
  check(
    "seller-cannot-self-approve-ui",
    (await seller
      .getByRole("button", {
        name: "تأیید و اعمال ویرایش فروشنده",
        exact: true,
      })
      .count()) === 0,
  );
  await seller.setViewportSize({ width: 390, height: 844 });
  await seller.screenshot({
    path: "output/playwright/issue-25/seller-pending-mobile.png",
    fullPage: true,
    animations: "disabled",
  });
  check(
    "seller-mobile-no-overflow",
    await seller.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  );
  await openProduct(page);
  await page
    .getByRole("heading", {
      name: "ویرایش فروشنده: در انتظار بررسی",
      exact: true,
    })
    .waitFor();
  await page.screenshot({
    path: "output/playwright/issue-25/admin-content-comparison.png",
    fullPage: true,
    animations: "disabled",
  });
  await page
    .getByRole("textbox", { name: "یادداشت بررسی / دلیل رد", exact: true })
    .fill("متن نمونه نیازمند بازبینی است.");
  await page
    .getByRole("button", { name: "رد ویرایش با دلیل", exact: true })
    .click();
  await page
    .getByRole("heading", {
      name: "ویرایش فروشنده: نیازمند اصلاح",
      exact: true,
    })
    .waitFor();
  check(
    "rejection-preserves-public-approved-name",
    (await publicProduct()).name === p.name,
  );
  await openProduct(seller);
  await seller
    .getByRole("heading", {
      name: "ویرایش فروشنده: نیازمند اصلاح",
      exact: true,
    })
    .waitFor();
  await seller
    .getByRole("button", { name: "ویرایش محتوای فروشگاه", exact: true })
    .click();
  dialog = seller.getByRole("dialog").last();
  check(
    "rejected-draft-preserved",
    (await dialog
      .getByRole("textbox", { name: "نام فارسی", exact: true })
      .inputValue()) === "عنوان جدید فروشنده — بررسی",
  );
  await dialog
    .getByRole("button", { name: "ذخیره / ارسال محتوا", exact: true })
    .click();
  await dialog.waitFor({ state: "hidden" });
  await openProduct(page);
  await page
    .getByRole("heading", {
      name: "ویرایش فروشنده: در انتظار بررسی",
      exact: true,
    })
    .waitFor();
  await page
    .getByRole("button", { name: "تأیید و اعمال ویرایش فروشنده", exact: true })
    .click();
  await page
    .getByRole("heading", { name: "ویرایش فروشنده: تأیید شده", exact: true })
    .waitFor();
  check(
    "approved-content-visible-in-pdp",
    (await publicProduct()).name === "عنوان جدید فروشنده — بررسی",
  );
  const after = await get("/backoffice/workspace", adminToken);
  check(
    "content-review-preserves-offer-price-stock-reservation",
    JSON.stringify(after.offers.filter((o) => o.product_id === p.id)) ===
      offersBefore,
  );
  await page.screenshot({
    path: "output/playwright/issue-25/admin-approved-content.png",
    fullPage: true,
    animations: "disabled",
  });
  await page.locator('nav a[href="/stores"]').click();
  const policy = page.getByRole("combobox", {
    name: "سیاست محتوا خانه سبز",
    exact: true,
  });
  const changed = page.waitForResponse(
    (r) =>
      r.url().endsWith("/backoffice/commands") &&
      r.request().postDataJSON()?.action === "store.content-policy.set",
  );
  await policy.selectOption("direct");
  check("store-policy-save-success", (await changed).ok());
  check("store-policy-editor-present", (await policy.count()) === 1);
  check("no-runtime-page-errors", errors.length === 0);
  return { checks, passed: checks.length, errors, productId: p.id };
}
