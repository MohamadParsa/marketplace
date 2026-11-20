/* Run with playwright-cli run-code --filename tests/e2e/catalog-icons.cjs.
   Start API with APP_ENV=test PORT=8094 WEB_ROOT=../backoffice-web/build. */
async (page) => {
  const checks = [],
    errors = [];
  let uploads = 0;
  page.on("pageerror", (e) => errors.push(e.message));
  page.on("request", (r) => {
    if (r.method() === "POST" && r.url().includes("/backoffice/catalog/"))
      uploads++;
  });
  const check = (id, pass) => {
    checks.push({ id, pass: !!pass });
    if (!pass) throw new Error(id);
  };
  async function login(viewer = false) {
    if (viewer)
      await page
        .getByRole("button", { name: "مشاهده‌گر", exact: true })
        .click();
    await page
      .getByRole("button", { name: "دریافت کد ورود", exact: true })
      .click();
    const alert = await page
      .getByRole("alert")
      .filter({ hasText: "کد آزمون" })
      .innerText();
    await page.getByLabel("کد یک‌بارمصرف").fill(alert.match(/\d{6}/)[0]);
    await page
      .getByRole("button", { name: "ورود به پنل", exact: true })
      .click();
    await page.locator(".content h1").waitFor();
  }
  const file = "tests/e2e/fixtures/catalog-icon.png";
  const firstRow = () => page.locator("tbody tr").first();
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page.goto("http://127.0.0.1:8094/");
  await login();
  await page.locator('nav a[href="/categories"]').click();
  check(
    "category-default",
    (await firstRow()
      .getByRole("img", { name: /آیکن پیش‌فرض/ })
      .count()) === 1,
  );
  await firstRow().getByRole("button", { name: "آیکن", exact: true }).click();
  await page.getByLabel("انتخاب فایل آیکن").setInputFiles(file);
  await page.getByRole("img", { name: /پیش‌نمایش آیکن/ }).waitFor();
  check("preview-without-upload", uploads === 0);
  await page.screenshot({
    path: "output/playwright/issue-4/category-preview.png",
    fullPage: true,
    animations: "disabled",
  });
  await page.getByRole("button", { name: "انصراف", exact: true }).click();
  check(
    "cancel-does-not-publish",
    uploads === 0 &&
      (await firstRow()
        .getByRole("img", { name: /آیکن پیش‌فرض/ })
        .count()) === 1,
  );
  await firstRow().getByRole("button", { name: "آیکن", exact: true }).click();
  await page.getByLabel("انتخاب فایل آیکن").setInputFiles(file);
  await page.getByRole("img", { name: /پیش‌نمایش آیکن/ }).waitFor();
  await page.getByRole("button", { name: "ثبت آیکن", exact: true }).click();
  await page.getByRole("dialog").waitFor({ state: "hidden" });
  await firstRow().locator("img.catalog-icon").waitFor();
  check(
    "category-upload",
    uploads === 1 &&
      (await firstRow()
        .locator("img.catalog-icon")
        .evaluate((x) => x.complete && x.naturalWidth > 0)),
  );
  await page.screenshot({
    path: "output/playwright/issue-4/category-saved.png",
    fullPage: true,
    animations: "disabled",
  });
  await firstRow().getByRole("button", { name: "آیکن", exact: true }).click();
  await page
    .getByLabel("انتخاب فایل آیکن")
    .setInputFiles("tests/e2e/fixtures/invalid-icon.png");
  await page
    .getByRole("alert")
    .filter({ hasText: "تصویر قابل خواندن نیست" })
    .waitFor();
  check(
    "invalid-keeps-current",
    (await page.getByRole("dialog").locator("img.catalog-icon").count()) ===
      1 &&
      (await page
        .getByRole("button", { name: "ثبت آیکن", exact: true })
        .isDisabled()) &&
      uploads === 1,
  );
  await page.screenshot({
    path: "output/playwright/issue-4/invalid-file.png",
    fullPage: true,
    animations: "disabled",
  });
  await page.getByRole("button", { name: "انصراف", exact: true }).click();
  await page.route("**/catalog/icons/*", (r) => r.abort());
  await page.locator('nav a[href="/brands"]').click();
  await page.locator('nav a[href="/categories"]').click();
  await firstRow()
    .getByRole("img", { name: /آیکن پیش‌فرض/ })
    .waitFor();
  check(
    "broken-image-fallback-and-name",
    await firstRow().getByText("خانه و زندگی", { exact: true }).isVisible(),
  );
  await page.unroute("**/catalog/icons/*");
  await firstRow().getByRole("button", { name: "آیکن", exact: true }).click();
  await page
    .getByRole("button", { name: "بازگشت به پیش‌فرض", exact: true })
    .click();
  await page.getByRole("dialog").waitFor({ state: "hidden" });
  check(
    "category-remove",
    (await firstRow()
      .getByRole("img", { name: /آیکن پیش‌فرض/ })
      .count()) === 1,
  );
  await page.locator('nav a[href="/brands"]').click();
  check(
    "brand-default",
    (await firstRow()
      .getByRole("img", { name: /آیکن پیش‌فرض برند/ })
      .count()) === 1,
  );
  await firstRow().getByRole("button", { name: "آیکن", exact: true }).click();
  await page.getByLabel("انتخاب فایل آیکن").setInputFiles(file);
  await page.getByRole("img", { name: /پیش‌نمایش آیکن/ }).waitFor();
  await page.setViewportSize({ width: 390, height: 844 });
  check(
    "mobile-dialog-no-overflow",
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth + 1,
    ),
  );
  await page.screenshot({
    path: "output/playwright/issue-4/brand-mobile-preview.png",
    fullPage: true,
    animations: "disabled",
  });
  await page.getByRole("button", { name: "ثبت آیکن", exact: true }).click();
  await page.getByRole("dialog").waitFor({ state: "hidden" });
  await firstRow().locator("img.catalog-icon").waitFor();
  check(
    "brand-upload-mobile",
    uploads === 2 &&
      (await firstRow()
        .locator("img.catalog-icon")
        .evaluate((x) => x.complete && x.naturalWidth > 0)),
  );
  check(
    "mobile-list-no-overflow",
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth + 1,
    ),
  );
  await page.screenshot({
    path: "output/playwright/issue-4/brands-mobile.png",
    fullPage: true,
    animations: "disabled",
  });
  await firstRow().getByRole("button", { name: "آیکن", exact: true }).click();
  await page
    .getByRole("button", { name: "بازگشت به پیش‌فرض", exact: true })
    .click();
  await page.getByRole("dialog").waitFor({ state: "hidden" });
  check(
    "brand-remove-mobile",
    (await firstRow()
      .getByRole("img", { name: /آیکن پیش‌فرض برند/ })
      .count()) === 1,
  );
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page.getByRole("button", { name: "خروج از حساب", exact: true }).click();
  await login(true);
  await page.locator('nav a[href="/brands"]').click();
  check(
    "viewer-write-disabled",
    await firstRow()
      .getByRole("button", { name: "آیکن", exact: true })
      .isDisabled(),
  );
  check("runtime-errors", errors.length === 0);
  return {
    checks,
    passed: checks.filter((x) => x.pass).length,
    total: checks.length,
    runtimeErrors: errors,
    uploads,
  };
}
