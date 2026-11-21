async (page) => {
  const checks = [];
  const check = (id, pass) => {
    checks.push({ id, pass: !!pass });
    if (!pass) throw Error(id);
  };
  await page.setViewportSize({ width: 390, height: 844 });
  const site = "http://127.0.0.1:5197";
  const home = await (await page.request.get(site + "/storefront/home")).json();
  const product = home.special_offers.find(
    (p) => p.name.includes("پاوربانک") && p.price < 100000000,
  );
  await page.goto(site + product.url);
  await page
    .getByRole("button", { name: "افزودن به سبد خرید", exact: true })
    .click();
  await page.getByRole("dialog").waitFor();
  check(
    "mobile-popup-no-overflow",
    await page
      .getByRole("dialog")
      .evaluate((e) => e.scrollWidth <= e.clientWidth),
  );
  await page
    .getByRole("link", { name: "مشاهده سبد خرید", exact: true })
    .click();
  await page
    .getByRole("heading", { name: product.name, exact: true })
    .waitFor();
  await page
    .getByRole("button", { name: "افزایش تعداد " + product.name, exact: true })
    .click();
  await page
    .getByRole("button", { name: "افزایش تعداد " + product.name, exact: true })
    .waitFor({ state: "visible" });
  await page.screenshot({
    path: "output/playwright/issue-30/cart-mobile.png",
    fullPage: true,
    animations: "disabled",
  });
  check(
    "mobile-cart-no-overflow",
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  );
  await page
    .getByRole("button", { name: "تغییر تنوع یا فروشنده", exact: true })
    .click();
  await page.getByRole("dialog").waitFor();
  check(
    "explicit-variant-seller-selector",
    (await page.getByRole("dialog").getByRole("combobox").count()) === 2,
  );
  await page.getByRole("button", { name: "بستن", exact: true }).click();
  await page.getByRole("link", { name: "ادامه و انتخاب آدرس" }).click();
  await page.getByRole("radio", { name: "انتخاب این آدرس" }).first().check();
  await page.locator(".buy-field select").first().selectOption("post-review");
  await page
    .getByRole("button", { name: "محاسبه مبلغ و روش‌های پرداخت" })
    .click();
  await page.getByRole("radio", { name: /پرداخت محلی/ }).check();
  await page.screenshot({
    path: "output/playwright/issue-30/checkout-mobile.png",
    fullPage: true,
    animations: "disabled",
  });
  check(
    "mobile-checkout-no-overflow",
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  );
  check(
    "non-gold-local-method-eligible",
    await page.getByRole("radio", { name: /پرداخت محلی/ }).isEnabled(),
  );
  await page.getByRole("button", { name: "بررسی نهایی سفارش" }).click();
  await page.getByRole("checkbox", { name: /مبلغ.*آدرس/ }).check();
  await page.getByRole("button", { name: "ثبت سفارش و ادامه پرداخت" }).click();
  await page.getByRole("textbox", { name: "دلیل لغو", exact: true }).waitFor();
  await page
    .getByRole("textbox", { name: "دلیل لغو", exact: true })
    .fill("پایان آزمون موبایل");
  await page.getByRole("button", { name: "لغو سفارش", exact: true }).click();
  await page.getByText("لغو شده", { exact: true }).first().waitFor();
  check("customer-unpaid-cancel-releases-order", true);
  return { checks, passed: checks.length };
}
