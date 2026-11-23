// Run seed-variant-showcase.cjs first, only against the owned local review servers.
async (page) => {
  const site = "http://127.0.0.1:5197",
    checks = [],
    errors = [];
  const check = (id, pass) => {
    checks.push({ id, pass: !!pass });
    if (!pass) throw Error(id);
  };
  page.on("pageerror", (e) => errors.push(e.message));
  const shot = async (name) =>
    page.screenshot({
      path: "output/playwright/purchase-refinements/" + name + ".png",
      fullPage: true,
      animations: "disabled",
    });
  await page.goto(site);
  await page.evaluate(() => {
    sessionStorage.removeItem("cherriva:customer-token");
    localStorage.removeItem("cherriva:guest-cart");
  });
  await page.reload();
  await page.setViewportSize({ width: 1440, height: 1000 });
  const names = [
    "گوشی نمونه چریوا — انتخاب رنگ و گارانتی",
    "هندزفری نمونه چریوا — انتخاب رنگ و گارانتی",
    "ساعت هوشمند نمونه چریوا — انتخاب رنگ و گارانتی",
  ];
  const examples = [];
  for (const name of names) {
    const listing = await (
      await page.request.get(
        site + "/storefront/catalog?q=" + encodeURIComponent(name),
      )
    ).json();
    const item = listing.products.find((p) => p.name === name);
    if (!item) throw Error("Missing seeded " + name);
    const d = await (
      await page.request.get(site + "/storefront/products/" + item.id)
    ).json();
    examples.push(d);
    await page.goto(site + item.url);
    await page.locator(".pdp-overview").waitFor();
    check(
      "two-visible-axes-" + examples.length,
      (await page.locator(".pdp-axes fieldset").count()) === 2 &&
        d.variants.length === 4,
    );
    const before = await page.locator(".pdp-price strong").innerText();
    const expensive = d.variants.find(
      (v) =>
        v.values.some((x) => x.value === "مشکی") &&
        v.values.some((x) => x.value === "۲۴ ماهه نمونه"),
    );
    await page
      .locator(".pdp-axes fieldset")
      .nth(1)
      .getByRole("button", { name: "۲۴ ماهه نمونه", exact: true })
      .click();
    await page.waitForFunction(
      (id) =>
        document
          .querySelector(".pdp-purchase")
          ?.getAttribute("data-variant-id") === id,
      expensive.id,
    );
    check(
      "axis-changes-selected-price-" + examples.length,
      (await page.locator(".pdp-price strong").innerText()) !== before,
    );
    const missing = d.variants.find(
      (v) =>
        v.values.some((x) => x.value === "سفید") &&
        v.values.some((x) => x.value === "۲۴ ماهه نمونه"),
    );
    await page
      .locator(".pdp-axes fieldset")
      .first()
      .getByRole("button", { name: "سفید", exact: true })
      .click();
    await page.waitForFunction(
      (id) =>
        document
          .querySelector(".pdp-purchase")
          ?.getAttribute("data-variant-id") === id,
      missing.id,
    );
    check(
      "unavailable-combination-keeps-other-choices-" + examples.length,
      (await page
        .getByRole("button", { name: "ناموجود", exact: true })
        .isDisabled()) &&
        (await page.locator(".pdp-axes fieldset").count()) === 2,
    );
  }
  for (let i = 0; i < 2; i++) {
    const d = examples[i];
    await page.goto(site + "/product/" + d.id);
    await page.locator(".pdp-overview").waitFor();
    if (i === 1) {
      const seller = page.locator(".pdp-sellers article").last();
      await seller
        .getByRole("button", { name: "انتخاب فروشنده", exact: true })
        .click();
      await page.waitForFunction(
        (id) =>
          document
            .querySelector(".pdp-purchase")
            ?.getAttribute("data-offer-id") === id,
        d.offers.at(-1).id,
      );
    }
    if (i === 0) {
      await page.locator(".pdp-large-photo img").evaluate((i) => i.decode());
      await shot("variant-pdp");
    }
    await page
      .getByRole("button", { name: "افزودن به سبد خرید", exact: true })
      .click();
    await page.getByRole("dialog").waitFor();
    await page.getByRole("button", { name: "بستن", exact: true }).click();
  }
  await page.goto(site + "/cart");
  await page.locator(".buy-cart-row").first().waitFor();
  check("two-seller-cart", (await page.locator(".buy-cart-row").count()) === 2);
  const row = page.locator(".buy-cart-row").first();
  await row.getByRole("button", { name: /افزایش تعداد/ }).click();
  await page.waitForFunction(
    () =>
      document.querySelector(".buy-cart-row .purchase-control output")
        ?.textContent === "۲",
  );
  await row.getByRole("button", { name: /کاهش تعداد/ }).click();
  await page.waitForFunction(
    () =>
      document.querySelector(".buy-cart-row .purchase-control output")
        ?.textContent === "۱",
  );
  check("cart-quantity-up-and-down-confirmed", true);
  await page.setViewportSize({ width: 390, height: 844 });
  await shot("two-seller-cart-mobile");
  check(
    "mobile-cart-no-overflow",
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  );
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page
    .getByRole("link", { name: "ادامه و انتخاب آدرس", exact: true })
    .click();
  await page.getByRole("link", { name: "ورود و ادامه", exact: true }).click();
  const phone = "090" + String(Date.now()).slice(-8);
  await page
    .getByRole("textbox", { name: "موبایل یا ایمیل", exact: true })
    .fill(phone);
  await page
    .getByRole("button", { name: "ورود با OTP / ثبت‌نام", exact: true })
    .click();
  const response = page.waitForResponse(
    (r) =>
      r.url().endsWith("/storefront/auth/request-otp") && r.status() === 202,
  );
  await page
    .getByRole("button", { name: "دریافت کد تأیید", exact: true })
    .click();
  const otp = await (await response).json();
  if (!otp.test_code) throw Error("Isolated review OTP required");
  await page
    .getByRole("textbox", { name: "کد تأیید", exact: true })
    .fill(otp.test_code);
  await page.getByRole("button", { name: "تأیید و ورود", exact: true }).click();
  await page.getByRole("button", { name: "افزودن آدرس", exact: true }).click();
  const modal = page.getByRole("dialog");
  for (const [label, value] of [
    ["نام گیرنده", "گیرنده ساختگی ویزارد"],
    ["موبایل گیرنده", phone],
    ["نشانی کامل", "نشانی نمونه برای بررسی شهر و ارسال"],
    ["کد پستی", "1234567890"],
  ])
    await modal.getByRole("textbox", { name: label, exact: true }).fill(value);
  check("address-cascade-starts-with-disabled-city",await modal.getByRole("combobox",{name:"شهر",exact:true}).isDisabled());
  await modal
    .getByRole("combobox", { name: "استان", exact: true })
    .selectOption("تهران");
  await modal
    .getByRole("combobox", { name: "شهر", exact: true })
    .selectOption("تهران");
  await modal.getByRole("button", { name: "ذخیره آدرس", exact: true }).click();
  await modal.waitFor({ state: "hidden" });
  check(
    "shipping-cards-per-seller-and-city",
    (await page.locator(".buy-shipping-group").count()) === 2 &&
      (await page.getByRole("radio", { name: /پیک تهران/ }).count()) === 2 &&
      (await page.locator("select").count()) === 0,
  );
  for (const group of await page.locator(".buy-shipping-group").all())
    await group.getByRole("radio", { name: /پیک تهران/ }).check();
  await shot("wizard-step-1");
  await page.setViewportSize({ width: 390, height: 844 });
  await shot("wizard-step-1-mobile");
  check(
    "mobile-wizard-step-one-no-overflow",
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  );
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page
    .getByRole("button", {
      name: "تأیید آدرس و ارسال؛ ادامه به پرداخت",
      exact: true,
    })
    .click();
  await page
    .getByRole("heading", { name: "روش پرداخت", exact: true })
    .waitFor();
  await page.getByRole("radio", { name: /کارت‌به‌کارت/ }).check();
  await page
    .getByRole("button", { name: "بررسی نهایی سفارش", exact: true })
    .click();
  await page.getByRole("switch", { name: /مبلغ.*آدرس/ }).waitFor();
  check(
    "two-confirmed-shipments-on-payment-step",
    (await page.locator(".buy-delivery-confirmed .buy-shipment").count()) === 2,
  );
  await shot("wizard-step-2");
  await page.setViewportSize({ width: 390, height: 844 });
  await shot("wizard-step-2-mobile");
  check(
    "mobile-wizard-step-two-no-overflow",
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  );
  await page
    .getByRole("button", {
      name: "← بازگشت به مرحله آدرس و ارسال",
      exact: true,
    })
    .click();
  await page
    .locator(".buy-address.selected")
    .getByRole("button", { name: "ویرایش", exact: true })
    .click();
  await modal
    .getByRole("combobox", { name: "استان", exact: true })
    .selectOption("قم");
  check("province-switch-clears-city",await modal.getByRole("combobox",{name:"شهر",exact:true}).inputValue()===""&&await modal.getByRole("combobox",{name:"شهر",exact:true}).locator('option').allTextContents().then(v=>!v.includes("تهران")));
  await modal
    .getByRole("combobox", { name: "شهر", exact: true })
    .selectOption("قم");
  await modal.getByRole("button", { name: "ذخیره آدرس", exact: true }).click();
  await modal.waitFor({ state: "hidden" });
  check(
    "changed-city-removes-ineligible-shipping",
    (await page.getByRole("radio", { name: /پیک تهران/ }).count()) === 0 &&
      (await page.locator(".buy-shipping-group input:checked").count()) === 0 &&
      (await page
        .getByRole("button", {
          name: "تأیید آدرس و ارسال؛ ادامه به پرداخت",
          exact: true,
        })
        .isDisabled()),
  );
  for (const group of await page.locator(".buy-shipping-group").all())
    await group.getByRole("radio").first().check();
  await page
    .getByRole("button", {
      name: "تأیید آدرس و ارسال؛ ادامه به پرداخت",
      exact: true,
    })
    .click();
  await page
    .getByRole("heading", { name: "روش پرداخت", exact: true })
    .waitFor();
  check(
    "address-change-invalidates-old-payment-consent",
    (await page.locator("input[name=payment_method]:checked").count()) === 0 &&
      (await page.getByRole("switch", { name: /مبلغ.*آدرس/ }).count()) === 0,
  );
  await page.goto(site + "/");
  await page.locator(".lh-categories a").first().waitFor();
  const home = await (await page.request.get(site + "/storefront/home")).json();
  check(
    "homepage-only-available-category-shortcuts",
    home.categories.every((c) => c.count > 0) &&
      (await page.locator(".lh-categories a").count()) ===
        home.categories.length,
  );
  const category = examples[0].category;
  await page.goto(site + category.url);
  await page
    .getByRole("group", { name: "مرتب‌سازی کالاها", exact: true })
    .waitFor();
  await page
    .getByRole("group", { name: "مرتب‌سازی کالاها", exact: true })
    .getByRole("radio", { name: "ارزان‌ترین", exact: true })
    .locator("..")
    .click();
  await page.waitForURL(/sort=price-asc/);
  check(
    "PLP-sort-without-dropdown",
    (await page.locator("select").count()) === 0,
  );
  check(
    "mobile-PLP-no-overflow",
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  );
  check("no-runtime-pageerror", errors.length === 0);
  return {
    checks,
    passed: checks.length,
    errors,
    products: examples.map((p) => ({ name: p.name, url: "/product/" + p.id })),
  };
}
