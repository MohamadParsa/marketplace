// Run only against the isolated loopback review API (requires exposed test OTP).
const fs = require("fs");
(async () => {
  const api = "http://127.0.0.1:8098";
  async function call(path, body, token = "") {
    const r = await fetch(api + path, {
      method: body ? "POST" : "GET",
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: "Bearer " + token } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
    });
    if (!r.ok) throw Error(path + " " + r.status + " " + (await r.text()));
    return r.status === 204 ? null : r.json();
  }
  const challenge = await call("/backoffice/auth/request-otp", {
    phone: "09000000001",
  });
  if (!challenge.test_code)
    throw Error("Only isolated review fixture with test OTP is supported");
  const session = await call("/backoffice/auth/verify-otp", {
    request_id: challenge.request_id,
    code: challenge.test_code,
  });
  const token = session.token;
  const command = (c) =>
    call(
      "/backoffice/commands",
      { ...c, idempotency_key: crypto.randomUUID() },
      token,
    ).catch((e) => {
      throw Error(c.action + ": " + e.message);
    });
  let w = await call("/backoffice/workspace", null, token);
  let role = w.access.roles.find(
    (r) => r.name === "ویرایش محتوای مستقل — آزمون",
  );
  if (!role) {
    await command({
      action: "role.create",
      name: "ویرایش محتوای مستقل — آزمون",
      scope: "store",
      permissions: ["offers.read", "products.content.write"],
    });
    w = await call("/backoffice/workspace", null, token);
    role = w.access.roles.find((r) => r.name === "ویرایش محتوای مستقل — آزمون");
  }
  if (
    !w.access.grants.some(
      (g) =>
        g.user_id === "seller-test" &&
        g.role_id === role.id &&
        g.store_id === "s1",
    )
  )
    await command({
      action: "user.role-grant",
      target: "seller-test",
      value: role.id,
      store_id: "s1",
    });
  const store = w.stores.find((s) => s.id === "s1");
  await command({
    action: "store.content-policy.set",
    target: "s1",
    value: "review",
    version: store.content_policy_version ?? 0,
  });
  let category = w.categories.find(
    (c) => c.name === "دسته ساختگی بازبینی محتوا",
  );
  if (!category) {
    await command({
      action: "catalog.category.create",
      catalog: { name: "دسته ساختگی بازبینی محتوا", published: true },
    });
    w = await call("/backoffice/workspace", null, token);
    category = w.categories.find((c) => c.name === "دسته ساختگی بازبینی محتوا");
  }
  const content = {
    short_description: "نسخه عمومی تأییدشده قبلی",
    full_description: "<p>محتوای پذیرفته‌شده فروشگاه</p>",
    advantages: [],
    disadvantages: [],
    reviews_enabled: false,
    qa_enabled: false,
    shipping: { weight_unit: "g", dimension_unit: "mm" },
    seo: { meta_title: "", meta_keywords: "", meta_description: "" },
    noindex: false,
    nofollow: false,
  };
  const input = {
    name: "محصول مستقل آزمون بررسی محتوا",
    english_name: "Independent review sample",
    description: "",
    category_id: category.id,
    brand_id: "",
    model: "independent",
    owner_store_id: "s1",
    published: false,
    content,
    attributes: [],
  };
  let p = w.products.find(p=>p.name===input.name || (p.english_name===input.english_name&&p.model==='independent'&&p.owner_store_id==='s1'));
  if (!p) {
    await command({ action: "catalog.product.create", catalog: input });
    w = await call("/backoffice/workspace", null, token);
    p = w.products.find((p) => p.name === input.name);
  }
  if (!p.image_id) {
    const file = new Blob(
      [fs.readFileSync("tests/e2e/fixtures/catalog-icon.png")],
      { type: "image/png" },
    );
    const form = new FormData();
    form.set("file", file, "sample.png");
    form.set("version", String(p.version));
    form.set("idempotency_key", crypto.randomUUID());
    const r = await fetch(
      api + "/backoffice/catalog/product/" + p.id + "/icon",
      {
        method: "POST",
        headers: { Authorization: "Bearer " + token },
        body: form,
      },
    );
    if (!r.ok) throw Error("image " + r.status);
  }
  w = await call("/backoffice/workspace", null, token);
  p = w.products.find((x) => x.id === p.id);
  const variant = w.variants.find((v) => v.product_id === p.id);
  if (!variant) throw Error("Seed requires category without variation axes");
  if (!w.offers.some((o) => o.product_id === p.id))
    await command({
      action: "offer.create",
      target: p.id,
      version: p.variant_version,
      offer: {
        product_id: p.id,
        variant_id: variant.id,
        store_id: "s1",
        price: 1000000,
        stock: 10,
      },
    });
  w = await call("/backoffice/workspace", null, token);
  p = w.products.find((x) => x.id === p.id);
  await command({
    action: "catalog.product.update",
    target: p.id,
    version: p.version,
    catalog: { ...input, published: true },
  });
  const out = {
    product_id: p.id,
    category_id: category.id,
    store_id: "s1",
    seller_phone: "09120000002",
  };
  console.log(JSON.stringify(out));
})().catch((e) => {
  console.error(e.message);
  process.exit(1);
});
