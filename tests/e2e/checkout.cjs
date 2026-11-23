async (page) => {
  const customerPhone = "090" + String(Date.now()).slice(-8);
  const results = [];
  const check = (name, ok) => {
    results.push({ name, passed: !!ok });
    if (!ok) throw Error(name);
  };
  const runtimeErrors = [];
  page.on("pageerror", (e) => runtimeErrors.push(e.message));
  const wait = async () => page.waitForTimeout(250);
  const shot = async (name) => {
    await page.screenshot({
      path: "output/playwright/purchase-refinements/" + name + ".png",
      fullPage: true,
      animations: "disabled",
    });
  };
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page.goto("http://127.0.0.1:5197/");
  await page.evaluate(() => {
    sessionStorage.removeItem("cherriva:customer-token");
    localStorage.removeItem("cherriva:guest-cart");
  });
  await page.reload();
  const product = await page.evaluate(async () => {
    const h = await (await fetch("/storefront/home")).json();
    return h.special_offers.find(
      (p) => p.name.includes("هندزفری") && p.price < 100000000,
    );
  });
  check("published purchasable product", !!product);
  await page.goto("http://127.0.0.1:5197" + product.url);
  await page
    .getByRole("button", { name: "افزودن به سبد خرید", exact: true })
    .waitFor();
  check(
    "single-designed-PDP-purchase-control",
    (await page.locator(".purchase-control").count()) === 1 &&
      (await page.locator(".pdp-quantity").count()) === 0,
  );
  await page
    .getByRole("button", { name: "افزودن به سبد خرید", exact: true })
    .click();
  await page.getByRole("dialog").waitFor();
  check(
    "PDP add opens actual cart popup",
    (await page
      .getByRole("dialog")
      .getByText(product.name, { exact: true })
      .count()) === 1,
  );
  await shot("pdp-added");
  await page
    .getByRole("link", { name: "مشاهده سبد خرید", exact: true })
    .click();
  await page.getByRole("heading", { name: "سبد خرید", exact: true }).waitFor();
  await shot("cart-desktop");
  check(
    "cart-uses-same-diagonal-purchase-control",
    await page
      .locator(".purchase-control-less")
      .evaluate((e) => getComputedStyle(e).clipPath.startsWith("polygon")),
  );
  check(
    "cart has price and seller",
    (await page.getByText("جمع کالاها", { exact: true }).count()) === 1,
  );
  await page.getByRole("link", { name: "ادامه و انتخاب آدرس" }).click();
  await page.getByRole("link", { name: "ورود و ادامه" }).click();
  await page
    .getByRole("textbox", { name: "موبایل یا ایمیل" })
    .fill(customerPhone);
  await page.getByRole("button", { name: "ورود با OTP / ثبت‌نام" }).click();
  const otpResponse = page.waitForResponse(
    (r) =>
      r.url().endsWith("/storefront/auth/request-otp") && r.status() === 202,
  );
  await page
    .getByRole("button", { name: "دریافت کد تأیید", exact: true })
    .click();
  const otp = await (await otpResponse).json();
  await page
    .getByRole("textbox", { name: "کد تأیید", exact: true })
    .fill(otp.test_code);
  await page.getByRole("button", { name: "تأیید و ورود" }).click();
  await page
    .getByRole("heading", { name: "تکمیل خرید", exact: true })
    .waitFor();
  await page.getByRole("button", { name: "افزودن آدرس" }).waitFor();
  check(
    "OTP login returns to checkout with merged cart",
    (await page.getByRole("button", { name: "افزودن آدرس" }).count()) === 1,
  );
  await page.getByRole("button", { name: "افزودن آدرس" }).click();
  const address = page.getByRole("dialog");
  await address
    .getByRole("textbox", { name: "نام گیرنده", exact: true })
    .fill("خریدار ساختگی آزمون");
  await address
    .getByRole("textbox", { name: "موبایل گیرنده", exact: true })
    .fill(customerPhone);
  await address
    .getByRole("combobox", { name: "استان", exact: true })
    .selectOption("تهران");
  await address
    .getByRole("combobox", { name: "شهر", exact: true })
    .selectOption("تهران");
  await address
    .getByRole("textbox", { name: "نشانی کامل", exact: true })
    .fill("نشانی ساختگی برای آزمون مسیر خرید");
  await address
    .getByRole("textbox", { name: "کد پستی", exact: true })
    .fill("1234567890");
  await address
    .getByRole("button", { name: "ذخیره آدرس", exact: true })
    .click();
  await address.waitFor({ state: "hidden" });
  check(
    "saved owner address selected",
    await page.getByRole("radio", { name: "انتخاب این آدرس" }).isChecked(),
  );
  check(
    "wizard-step-one-hides-payment",
    (await page
      .getByRole("heading", { name: "روش پرداخت", exact: true })
      .count()) === 0 && (await page.locator("select").count()) === 0,
  );
  await page
    .locator(".buy-shipping-group")
    .first()
    .getByRole("radio")
    .first()
    .check();
  await page
    .getByRole("button", { name: "تأیید آدرس و ارسال؛ ادامه به پرداخت" })
    .click();
  await page
    .getByRole("heading", { name: "روش پرداخت", exact: true })
    .waitFor();
  check(
    "wizard-step-two-hides-editable-address",
    (await page
      .getByRole("radio", { name: "انتخاب این آدرس", exact: true })
      .count()) === 0,
  );
  await page
    .getByRole("button", {
      name: "← بازگشت به مرحله آدرس و ارسال",
      exact: true,
    })
    .click();
  check(
    "wizard-back-preserves-address-and-shipping",
    (await page
      .getByRole("radio", { name: "انتخاب این آدرس", exact: true })
      .isChecked()) &&
      (await page.locator(".buy-shipping-group input:checked").count()) === 1,
  );
  await page
    .getByRole("button", {
      name: "تأیید آدرس و ارسال؛ ادامه به پرداخت",
      exact: true,
    })
    .click();
  await page.getByRole("radio", { name: /کارت‌به‌کارت/ }).check();
  await shot("checkout-desktop");
  await page.getByRole("button", { name: "بررسی نهایی سفارش" }).click();
  await page.getByRole("switch", { name: /مبلغ.*آدرس/ }).check();
  await page.getByRole("button", { name: "ثبت سفارش و ادامه پرداخت" }).click();
  await page
    .getByRole("heading", { name: "جزئیات سفارش", exact: true })
    .waitFor();
  const purchaseID = page.url().split("/").pop();
  check(
    "friendly-numeric-order-number",
    /^\d{4,}$/.test(purchaseID) && Number(purchaseID) >= 1000,
  );

  await page.getByText("پرداخت نشده", { exact: true }).first().waitFor();
  check(
    "registered order is unpaid",
    await page.getByText("پرداخت نشده", { exact: true }).first().isVisible(),
  );
  await shot("order-unpaid");
  await page.getByRole("button", { name: "اعلام پرداخت / واریز" }).click();
  const deposit = page.getByRole("dialog");
  await deposit
    .getByRole("textbox", { name: "شماره پیگیری / مرجع", exact: true })
    .fill("BROWSER-" + Date.now());
  await deposit
    .getByRole("button", { name: "ثبت اعلام پرداخت", exact: true })
    .click();
  await deposit.waitFor({ state: "hidden" });
  await page.getByText("در انتظار بررسی", { exact: false }).first().waitFor();
  check(
    "declaration does not pay order",
    await page.getByText("پرداخت نشده", { exact: true }).first().isVisible(),
  );
  await page
    .locator("input[type=file]")
    .setInputFiles("tests/e2e/fixtures/catalog-icon.png");
  await page.getByRole("img", { name: "تصویر خصوصی رسید" }).waitFor();
  await shot("order-with-private-receipt");
  // Use a second independent browser tab for the authorized finance/operator UI.
  const admin = await page.context().newPage();
  admin.on("pageerror", (e) => runtimeErrors.push(e.message));
  await admin.goto("http://127.0.0.1:8098/commerce");
  await admin
    .getByRole("textbox", { name: "شماره موبایل یا ایمیل", exact: true })
    .fill("review-admin@example.invalid");
  await admin.getByRole("button", { name: /ادامه/ }).click();
  await admin.getByLabel(/^رمز عبور/).fill("Review-only-27-pass");
  await admin
    .getByRole("button", { name: "ورود با رمز عبور", exact: true })
    .click();
  await admin
    .getByRole("heading", { name: "سفارش‌ها و پرداخت‌های سایت", exact: true })
    .waitFor();
  await admin.getByRole("tab", { name: /اعلام واریز/ }).click();
  await admin
    .getByRole("button", { name: "مشاهده و بررسی واریز" })
    .first()
    .click();
  const review = admin.getByRole("dialog");
  await review.getByRole("img", { name: "تصویر رسید خصوصی" }).waitFor();
  const token = await page.evaluate(() =>
    sessionStorage.getItem("cherriva:customer-token"),
  );
  const current = await page.evaluate(
    async ({ id, token }) => {
      const r = await fetch("/storefront/account/orders/" + id, {
        headers: { Authorization: "Bearer " + token },
      });
      return r.json();
    },
    { id: purchaseID, token },
  );
  const d = current.deposits[0];
  await review
    .getByRole("textbox", { name: "مبلغ واقعی تطبیق‌شده (تومان)", exact: true })
    .fill(d.amount_toman);
  await review
    .getByRole("textbox", { name: "مرجع واقعی تراکنش", exact: true })
    .fill("VERIFIED-" + Date.now());
  await review.getByRole("checkbox").check();
  await review
    .getByRole("button", { name: "تأیید و شارژ کیف پول", exact: true })
    .click();
  await review.waitFor({ state: "hidden" });
  await page.getByRole("button", { name: "تازه‌سازی وضعیت" }).click();
  await page.getByText("تأیید و شارژ شد", { exact: false }).first().waitFor();
  check(
    "finance approval credits only; order remains unpaid",
    await page.getByText("پرداخت نشده", { exact: true }).first().isVisible(),
  );
  await admin.getByRole("tab", { name: /سفارش‌ها/ }).click();
  await admin
    .getByRole("button", { name: "جزئیات سفارش", exact: true })
    .first()
    .click();
  const order = admin.getByRole("dialog");
  check(
    "admin-order-details-in-readable-boxes",
    (await order.locator(".commerce-order-box").count()) === 5 &&
      (await order
        .getByRole("heading", { name: "#" + purchaseID, exact: true })
        .isVisible()),
  );
  check(
    "admin order separates payment and order status",
    (await order
      .locator(".commerce-status>div")
      .filter({ hasText: "وضعیت سفارش" })
      .count()) === 1 &&
      (await order
        .locator(".commerce-status>div")
        .filter({ hasText: "وضعیت پرداخت" })
        .count()) === 1,
  );
  await admin.screenshot({
    path: "output/playwright/purchase-refinements/backoffice-order.png",
    fullPage: true,
    animations: "disabled",
  });
  await order.getByRole("checkbox").check();
  await order
    .getByRole("button", { name: "تأیید نهایی و پرداخت سفارش" })
    .click();
  await order.waitFor({ state: "hidden" });
  await page.getByRole("button", { name: "تازه‌سازی وضعیت" }).click();
  await page.getByText("پرداخت شده", { exact: true }).first().waitFor();
  check(
    "separate admin confirmation pays order",
    await page
      .getByText("در انتظار تأمین", { exact: true })
      .first()
      .isVisible(),
  );
  await page.getByText("تاریخچه سفارش و پرداخت", { exact: true }).click();
  await shot("order-paid-desktop");
  await page.reload();
  await page.getByText("پرداخت شده", { exact: true }).first().waitFor();
  check("refresh preserves account/order state", true);
  await page.setViewportSize({ width: 390, height: 844 });
  await shot("order-mobile");
  check(
    "390px order no overflow",
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  );
  await page.goto("http://127.0.0.1:5197/wallet");
  await page.getByRole("heading", { name: "گردش کیف پول" }).waitFor();
  await shot("wallet-mobile");
  check(
    "wallet shows credit and debit separately",
    (await page.getByText("شارژ کیف پول", { exact: true }).count()) > 0 &&
      (await page.getByText("پرداخت سفارش", { exact: false }).count()) > 0,
  );
  check(
    "390px wallet no overflow",
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  );
  await admin.getByRole("link", { name: "تنظیمات خرید و حمل" }).click();
  await admin
    .getByRole("heading", { name: "تنظیمات خرید، حمل و کسب‌وکار" })
    .waitFor();
  await admin
    .getByRole("button", { name: "ذخیره روش‌های پرداخت و مقصد" })
    .waitFor();
  await admin.setViewportSize({ width: 390, height: 844 });
  await admin.screenshot({
    path: "output/playwright/purchase-refinements/backoffice-settings-mobile.png",
    fullPage: true,
    animations: "disabled",
  });
  check(
    "390px admin settings no overflow",
    await admin.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  );
  check("no page runtime errors", runtimeErrors.length === 0);
  return { checks: results, runtimeErrors, purchaseID };
}
