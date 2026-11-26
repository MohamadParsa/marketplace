---
id: VERIFY-STORE-ELIGIBILITY-20261126
title: تحویل مستقل مجوز دسته و روش فروش فروشگاه — #114
status: in-review
owner: eliassarband
updated: 2026-11-26
related: [FEAT-BO-REFINEMENT-008]
approved_by: []
implementation_status: done
execution_authorized: true
---

# خروجی

در `/stores` ← ویرایش فروشگاه، بخش «دسته‌ها و روش‌های فروش مجاز» انتخاب درختی، همه/فقط انتخابی، چیپ‌های قابل حذف ممنوع/مجاز، روش‌ها و نام اختیاری هر روش را دارد. فیلتر روش فروش به گرید اضافه شده است. سیاست و نام اصلی/شناسه/پروفایل/رسانه/وقفه و سیاست محتوا مستقل می‌مانند.

[قرارداد و API](../../architecture/contracts/store-sales-eligibility.md) · [نتیجه ماشین‌خوان](issue-114-checks.json) · [BDD](../../bdd/features/experiences/store-sales-policy.feature).

## تصمیم نهایی مالک

PDP جاری فقط قیمت نقدی و نام اصلی دارد؛ با انتخاب روش پرداخت در خرید، نام متناظر روش فروش عوض می‌شود. نمایش فروشندگان و قیمت محاسبه‌شده بر اساس روش در PDP به دستور مالک در [#118](https://github.com/MohamadParsa/marketplace/issues/118) ثبت و به eliassarband اساین شد و **در این تحویل اجرا نشده است**. قبل از انتخاب روش، نام اصلی و برای alias خالی نیز نام اصلی استفاده می‌شود.

## ردیابی پذیرش

| معیار | شاهد و نتیجه |
| --- | --- |
| AC01 | همه/فقط انتخابی، تقدم deny، فرزند تازه و چرخه نامعتبر؛ [تست واحد](../../../apps/api/internal/application/backoffice/store_eligibility_test.go)، TestStoreEligibilityTreeDenyPriorityLegacyAndNewDescendant |
| AC02 | guard مشترک عرضه و خرید، رد درخواست مستقیم سلرِ عضو و کاربر بی‌مجوز؛ DTO پنل قابلیت اکشن را از سیاست جاری می‌گیرد. تست MixedCartMethodsAndDirectOfferGuards و [HTTP](../../../apps/api/internal/bootstrap/store_eligibility_test.go) |
| AC03 | نام اصلی پیش از روش؛ نام نقدی/اعتباری پس از روش با StoreID ثابت؛ snapshot در quote و سفارش و حفظ سفارش قطعی/دفتر پس از تغییر alias؛ تست AliasOnlyAfterSelectionAndHistoricalSnapshot و [مرورگر](../../../tests/e2e/store-sales-policy.cjs) |
| AC04 | نسخه سیاست در line snapshot؛ رد accept قدیمی و نیاز به تأیید دوباره رزرو، شامل wallet-only؛ وجه خرج نمی‌شود. نسخه/idempotency/رقابت و سلب مجوز، restart PostgreSQL و تازه‌شدن PDP/PLP آزموده شدند. |

## کنترل‌ها و داده

۳۶۸تست/زیرتست Go موفق،۴skip اختیاری و صفر شکست؛ PostgreSQL با schemaهای ایزوله و Redis فعال. go vet و build/TypeScript بک‌آفیس و storefront موفق؛ dependency تازه اضافه نشده و هشدار قبلی chunk بزرگ بک‌آفیس باقی است. race محلی اجرا نشده؛ نتیجه CI جدا ثبت می‌شود.

۱۳کنترل مرورگر موفق با صفر runtime pageerror: ساخت سیاست صریح روی legacy، درخت/deny/پیش‌نمایش، نام قبل/بعد روش، حفظ draft در۴۰۹، موبایل، PDP ممنوع بدون خرید، ویزارد تا سفارش با alias و نقش فقط‌خواندنی واقعی. fixture فقط حافظه ایزوله بود؛ BDD سند است و runner مستقل ندارد. روش «اعتباری آزمون» در fixture به پرداخت محلی ساختگی نگاشت شد و درگاه واقعی فعال نشد.

نسخه اصلی `D:\Cherriva` با startup و warmup بازراه‌اندازی شد. **۲۸بخش** هویت/رمز/نقش/عضویت، سفارش/دفتر/بانک و کاتالوگ/قیمت/موجودی/برند/فروشگاه/قواعد فروش/آرشیو و audit پیش/پس restart یکسان بودند. هیچ سیاستی روی داده فعلی خودکار ذخیره یا واریزی خرج نشد. مدل در JSONB موجود افزایشی است و migration/backfill جدید لازم نبود. snapshot کاتالوگ و بسته تاریخی در این کار تغییر نکردند؛ چهار مغایرت قبلی manifest در گزارش #109 همان سابقه قبلی‌اند.

## تصاویر

[پنل سیاست](../../../output/playwright/issue-114/store-policy.png) · [موبایل](../../../output/playwright/issue-114/store-policy-mobile.png) · [PDP ممنوع](../../../output/playwright/issue-114/restricted-pdp.png) · [نام با انتخاب روش](../../../output/playwright/issue-114/checkout-method-alias.png) · [سفارش](../../../output/playwright/issue-114/order-alias.png). تصاویر ساختگی و بازبینی شدند.

## نسخه و مرز

سورس مستقل [47b798a](https://github.com/MohamadParsa/marketplace/commit/47b798a0b22a40f5f0309dd18978b25b9f6829f8). موتور مارکاپ/کوپن/اعتبار، دفتر و هویت مالی از #73 و مدل موجود استفاده می‌کنند. مجوز global مستقل stores.eligibility.write قابل انتساب است؛ [هماهنگی #71](coordination-71.md) به‌روز شد و ممیزی کامل آن انجام‌شده محسوب نمی‌شود.

رکورد بدون سیاست اختصاصی مجوز موجود را حفظ می‌کند؛ پس از ثبت، روش‌ها صریح‌اند. حذف مجوز نقدی در تنظیمات، خرید نقدی فعلی PDP را غیرفعال می‌کند و UI آن را توضیح می‌دهد؛ توسعه روش‌محور متعلق به #118 است. فیلتر/صفحه‌بندی سرور هنوز از snapshot کامل مخزن استفاده می‌کند و benchmark تولید ادعا نمی‌شود. main/Liara ادغام یا مستقر نشده‌اند؛ رسید پوش و CI پس از ثبت افزوده می‌شود. #115 تحویل بعدی مستقل است.

## رسید نهایی

[سورس](https://github.com/MohamadParsa/marketplace/commit/47b798a0b22a40f5f0309dd18978b25b9f6829f8) و [مدارک اولیه](https://github.com/MohamadParsa/marketplace/commit/eb4105889fcbded38a70206e2ea3697eaa0a3520) روی شاخه PR #106 پوش شدند؛ [کامنت پایان](https://github.com/MohamadParsa/marketplace/issues/114#issuecomment-7190989002) ثبت و #114 completed بسته شد. هر۵job [CI](https://github.com/MohamadParsa/marketplace/actions/runs/45811990945) روی مدارک با source tree یکسان موفق‌اند؛ race/vet شامل آن است. [رسید ماشین‌خوان](issue-114-delivery.json). #114/#115 جدا و main/Liara بدون استقرار باقی‌اند.
