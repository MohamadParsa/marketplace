// Isolated local review data only. Refuses an API without exposed test OTP.
const fs = require("fs"),
  path = require("path");
(async () => {
  const base = "http://127.0.0.1:8098";
  let token = "";
  async function call(url, body, method) {
    const r = await fetch(base + url, {
      method: method ?? (body ? "POST" : "GET"),
      headers: {
        "Content-Type": "application/json",
        ...(body ? { "Idempotency-Key": crypto.randomUUID() } : {}),
        ...(token ? { Authorization: "Bearer " + token } : {}),
      },
      ...(body ? { body: JSON.stringify(body) } : {}),
    });
    if (!r.ok) throw Error(url + " HTTP " + r.status);
    return r.status === 204 ? null : r.json();
  }
  const otp = await call("/backoffice/auth/request-otp", {
    phone: "09000000001",
  });
  if (!otp.test_code) throw Error("Requires isolated APP_ENV=test fixture");
  token = (
    await call("/backoffice/auth/verify-otp", {
      request_id: otp.request_id,
      code: otp.test_code,
    })
  ).token;
  let w;
  const refresh = async () => {
    w = await call("/backoffice/workspace");
  };
  await refresh();
  const command = async (c) => {
    await call("/backoffice/commands", {
      ...c,
      idempotency_key: crypto.randomUUID(),
    });
    await refresh();
  };
  for (const [code, name, options] of [
    ["pdp-color", "رنگ بدنه نمونه", ["مشکی", "سفید"]],
    ["pdp-warranty", "گارانتی نمونه", ["۱۸ ماهه نمونه", "۲۴ ماهه نمونه"]],
  ])
    if (!w.attributes.some((a) => a.code === code))
      await command({
        action: "attribute.create",
        attribute: { code, name, kind: "enum", options },
      });
  const color = w.attributes.find((a) => a.code === "pdp-color"),
    warranty = w.attributes.find((a) => a.code === "pdp-warranty");
  const assets =
    "docs/review-packages/cherriva-2026-11-17-v1-1/01-storefront/assets";
  const content = {
    short_description:
      "محصول ساختگی برای مشاهده رنگ، گارانتی و قیمت هر انتخاب؛ قابل استفاده در بررسی محلی.",
    full_description:
      "<p>نمونه بررسی رابط چریوا؛ این داده برای فروش واقعی نیست.</p>",
    advantages: ["انتخاب رنگ و گارانتی", "مقایسه فروشنده‌ها"],
    disadvantages: ["داده ساختگی"],
    reviews_enabled: false,
    qa_enabled: false,
    shipping: { weight_unit: "g", dimension_unit: "mm" },
    seo: { meta_title: "", meta_description: "", meta_keywords: "" },
    noindex: true,
    nofollow: true,
  };
  const examples = [
    [
      "گوشی نمونه چریوا — انتخاب رنگ و گارانتی",
      "موبایل نمونه تنوع",
      "iphone.jpg",
      120000000,
    ],
    [
      "هندزفری نمونه چریوا — انتخاب رنگ و گارانتی",
      "هندزفری نمونه تنوع",
      "airpods.jpg",
      25000000,
    ],
    [
      "ساعت هوشمند نمونه چریوا — انتخاب رنگ و گارانتی",
      "ساعت هوشمند نمونه تنوع",
      "apps/storefront-web/public/images/watch-silver.webp",
      45000000,
    ],
  ];
  const stores = w.stores.filter((s) => s.active).slice(0, 2);
  if (stores.length < 2) throw Error("Two synthetic stores required");
  const products = [];
  for (const [name, catName, photo, price] of examples) {
    let cat = w.categories.find((c) => c.name === catName);
    if (!cat) {
      await command({
        action: "catalog.category.create",
        catalog: {
          name: catName,
          published: true,
          description: "محصولات ساختگی بررسی تنوع",
        },
      });
      cat = w.categories.find((c) => c.name === catName);
    }
    if (!cat.bindings?.some((b) => b.attribute_id === color.id && b.axis))
      await command({
        action: "category.bindings.set",
        target: cat.id,
        version: cat.version,
        bindings: [
          { attribute_id: color.id, axis: true, position: 1 },
          { attribute_id: warranty.id, axis: true, position: 2 },
        ],
      });
    let p = w.products.find((p) => p.name === name);
    if (!p) {
      await command({
        action: "catalog.product.create",
        catalog: {
          name,
          english_name: "Cherriva variant review sample",
          category_id: cat.id,
          brand_id: w.brands.find((b) => b.active)?.id ?? "",
          published: false,
          content,
          attributes: [],
        },
      });
      p = w.products.find((p) => p.name === name);
    }
    if (!p.image_id) {
      const file = photo.includes("/") ? photo : path.join(assets, photo);
      const form = new FormData();
      form.append("version", String(p.version));
      form.append("idempotency_key", crypto.randomUUID());
      form.append(
        "file",
        new Blob([fs.readFileSync(file)], {
          type: photo.endsWith(".webp") ? "image/webp" : "image/jpeg",
        }),
        path.basename(photo),
      );
      const r = await fetch(
        base + "/backoffice/catalog/product/" + p.id + "/icon",
        {
          method: "POST",
          headers: { Authorization: "Bearer " + token },
          body: form,
        },
      );
      if (!r.ok) throw Error("image HTTP " + r.status);
      await refresh();
      p = w.products.find((x) => x.id === p.id);
    }
    for (const c of color.options)
      for (const g of warranty.options) {
        let variant = w.variants.find(
          (v) =>
            v.product_id === p.id &&
            v.values.some(
              (x) => x.attribute_id === color.id && x.value === c,
            ) &&
            v.values.some(
              (x) => x.attribute_id === warranty.id && x.value === g,
            ),
        );
        if (!variant) {
          p = w.products.find((x) => x.id === p.id);
          await command({
            action: "variant.create",
            target: p.id,
            version: p.variant_version,
            variant: {
              product_id: p.id,
              sku:
                "SHOW-" +
                examples.findIndex((x) => x[0] === name) +
                "-" +
                color.options.indexOf(c) +
                "-" +
                warranty.options.indexOf(g),
              values: [
                { attribute_id: color.id, value: c },
                { attribute_id: warranty.id, value: g },
              ],
            },
          });
          variant = w.variants.find(
            (v) =>
              v.product_id === p.id &&
              v.values.some((x) => x.value === c) &&
              v.values.some((x) => x.value === g),
          );
        }
        for (let i = 0; i < stores.length; i++)
          if (
            !w.offers.some(
              (o) => o.variant_id === variant.id && o.store_id === stores[i].id,
            )
          ) {
            const unavailable = c === "سفید" && g === "۲۴ ماهه نمونه";
            p = w.products.find((x) => x.id === p.id);
            await command({
              action: "offer.create",
              target: p.id,
              version: p.variant_version,
              offer: {
                product_id: p.id,
                variant_id: variant.id,
                store_id: stores[i].id,
                price:
                  price +
                  color.options.indexOf(c) * 100000 +
                  warranty.options.indexOf(g) * 500000 +
                  i * 200000,
                regular_price: price + 1500000,
                stock: unavailable ? 0 : 20,
              },
            });
          }
      }
    p = w.products.find((x) => x.id === p.id);
    if (p.status !== "published")
      await command({
        action: "catalog.product.update",
        target: p.id,
        version: p.version,
        catalog: {
          name: p.name,
          english_name: p.english_name,
          category_id: p.category_id,
          brand_id: p.brand_id,
          published: true,
          content: p.content,
          attributes: p.attributes,
        },
      });
    products.push({
      name,
      url: "/product/" + p.id,
      category: "/category/" + cat.id,
    });
  }
  let settings = await call("/backoffice/commerce/settings");
  if (!settings.geography?.length) {
    await call(
      "/backoffice/commerce/settings",
      {
        version: settings.version,
        geography: [
          {
            id: "review-tehran",
            name: "تهران",
            active: true,
            cities: [
              { id: "review-tehran-city", name: "تهران", active: true },
              { id: "review-rey-city", name: "ری", active: true },
            ],
          },
          {
            id: "review-qom",
            name: "قم",
            active: true,
            cities: [{ id: "review-qom-city", name: "قم", active: true }],
          },
        ],
      },
      "PATCH",
    );
    settings = await call("/backoffice/commerce/settings");
  }
  if (!settings.shipping.some((m) => m.id === "review-courier-tehran")) {
    await call(
      "/backoffice/commerce/settings",
      {
        version: settings.version,
        shipping: {
          methods: [
            ...settings.shipping,
            {
              id: "review-courier-tehran",
              name: "پیک تهران — داده آزمون",
              active: true,
              all_cities: false,
              cities: ["تهران"],
              store_ids: [],
              fee_toman: "40000",
              min_days: 0,
              max_days: 1,
            },
          ],
          closed_weekdays: settings.closed_weekdays,
          holidays: settings.holidays,
        },
      },
      "PATCH",
    );
  }
  console.log(JSON.stringify({ products }, null, 2));
})().catch((e) => {
  console.error(e.message);
  process.exitCode = 1;
});
