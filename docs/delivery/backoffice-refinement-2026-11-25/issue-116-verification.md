---
id: VERIFY-PRODUCT-MANAGEMENT-20261126
title: شواهد تکمیل مستقل فرم و فعال‌بودن محصول — #116
status: in-review
owner: eliassarband
updated: 2026-11-26
related: [FEAT-BO-REFINEMENT-010, API-PRODUCT-MANAGEMENT-001]
approved_by: []
implementation_status: done
execution_authorized: true
---

# تحویل #116

`/products` با فیلتر مستقل نام فارسی/انگلیسی، کد، دسته، برند، انتشار، فعال‌بودن و موجودی آزاد تکمیل شد. فرم مرجع، انتخاب نوع/برند/مالک/enum قابل جست‌وجو، منشأ خصوصیات دسته و لینک آن، کد خودکار فقط‌خواندنی و فعال‌بودن مستقل دارد. اکشن‌های گالری، تنوع، عرضه، قیمت، موجودی و گردش محتوای سلر حفظ شده‌اند. انتشار/رد به مجوز واقعی همان عملیات وابسته‌اند.

طبق پاسخ صریح مالک، محصول منتشرشده ولی غیرفعال از خانه حذف و در PLP/PDP با «غیرقابل خرید»، بدون قیمت و افزودن می‌ماند. سبد/checkout و تأیید نهایی همین قاعده را در سرور کنترل می‌کنند. نسخه کش عمومی به v8 ارتقا یافت تا پاسخ قدیمی فاقد Active مصرف نشود؛ TTL و warmup قبلی برقرارند.

[سورس مستقل](https://github.com/MohamadParsa/marketplace/commit/99ee075a58b45ff93c8116500ea3a62c8540b425) · [قرارداد](../../architecture/contracts/product-management.md) · [تسک](tasks/product-form-active-state-and-filters.md) · [نتیجه ماشین‌خوان](issue-116-checks.json).

| معیار | شاهد |
| --- | --- |
| AC01 فرم کامل و کد خودکار | مرورگر ایجاد محتوا/RTE/وزن/SEO، تخصیص chr واقعی در schema جدا PostgreSQL؛ رد کد تحمیلی، ثبات کد در restart در `apps/api/internal/bootstrap/product_management_test.go` |
| AC02 خصوصیات/تنوع | UI فقط رمِ binding غیرمحور را نشان می‌دهد؛ رنگ در تنوع است؛ ویژگی نامرتبط/محور تازه روی Product رد و تغییر دسته ناسازگار409 می‌شود؛ duplicate قدیمی بدون تغییر محفوظ |
| AC03 چهار حالت و حفاظت خرید | چهار ترکیب Active/Published و انتشار پیش‌نویسِ غیرفعال؛ PDP/PLP/خانه و درخواست مستقیم سبد، سبد قدیمی بدون قیمت؛ preview منقضی‌شده پس از غیرفعال‌سازی رزرو نمی‌شود؛ پس از شارژ، تأیید خرید اعتبار مصرف نمی‌کند و hold را آزاد می‌کند؛ سفارش قطعی و snapshot قبلی ثابت |
| AC04 گرید/مجوز/عرض محدود | ۴۶محصول ساختگی در PostgreSQL جدا، صفحه بعد، فیلتر ترکیبی/خالی و تغییر category URL روی همان صفحه؛ فرم/جدول موبایل، اکشن سریع تنوع و قیمت، منع ویرایش viewer و درخواست seller از commands؛ نسخه/idempotency و تعارض |

## نتیجه اجرا

- `go test -json ./...` با متغیرهای خصوصی `MARKETPLACE_TEST_POSTGRES_URL` و `MARKETPLACE_TEST_REDIS_URL`: **۳۸۱ تست/زیرتست موفق،۴ skip اختیاری و صفر شکست**. تست مستقل حفاظت خرید در `apps/api/internal/application/backoffice/product_active_test.go` و HTTP/restart/cache در فایل bootstrap بالا است.
- `go vet ./...` و build هر دو storefront/backoffice موفق؛ هشدار شناخته‌شده chunk بزرگ پنل باقی است. dependency جدید نداریم. race محلی به‌دلیل نبود CGO/toolchain اجرا نشد؛ نتیجه CI در رسید نهایی جدا می‌آید.
- `tests/e2e/product-management.cjs`: **۲۱کنترل موفق**، صفر runtime pageerror. fixtures روی schema مجزای PostgreSQL و پورت8097/5199، بدون استفاده از کاربر/بانک/سفارش اصلی؛ خطای409 عمداً شبیه‌سازی شد.
- نسخه اصلی از `D:\Cherriva` با `Start-Cherriva.ps1 -NoBuild` به‌روز شد؛ Redis و startup warmup فعال‌اند. مقایسه **۲۸ بخش داده** پس از restart بدون تغییر است؛ کاربر/رمز/نقش، سفارش، کیف پول، دفتر، بانک، پرداخت، ارسال، کاتالوگ و رسانه اصلی محفوظ‌اند. هیچ اتصال به SQL Server مبدأ و هیچ SQL مستقیم برای جایگزینی داده انجام نشد.
- snapshot کاتالوگ و بسته طراحی تاریخی تغییر نکردند. اصل تصویرهای زیر ساختگی و برای تست همین فلو هستند.

## تصاویر

[فرم و اثر غیرفعال‌سازی](../../../output/playwright/issue-116/product-inactive-edit.png) · [PDP غیرفعال](../../../output/playwright/issue-116/inactive-pdp.png) · [PLP غیرفعال](../../../output/playwright/issue-116/inactive-plp.png) · [سبد قبلی](../../../output/playwright/issue-116/inactive-cart.png) · [فرم موبایل و کد واقعی](../../../output/playwright/issue-116/product-mobile-edit.png) · [گرید موبایل](../../../output/playwright/issue-116/products-mobile.png).

## محدودیت‌ها و انتشار

فیلتر/صفحه‌بندی روی workspace مجاز فعلی در کلاینت است، نه query حجیم سرور؛ موجودی آزاد فیزیکی با قابلیت خرید یکی نیست. create پیش‌نویس است و تصویر/تنوع/عرضه از اکشن همان ردیف تکمیل می‌شوند. تغییر دسته/مدل/مالک در update فعلی409 است؛ تبدیل محور #109 مسیر مستقل دارد. داده قدیمیِ فاقد Active فعال تفسیر می‌شود و bulk rewrite ندارد. واحدها هنوز #117، PDP روش‌محور #118 و ستاره/میانگین نظرات #119 هستند؛ این سه اجرا نشده‌اند. سند in-review؛ main و Liara ادغام/مستقر نشده‌اند.

## بررسی نسخه اصلی محلی

سایت5197 و پنل8098 پاسخ200 دادند؛ نمونه [محصول واقعیِ واردشده](http://localhost:5197/product/chr-10082) پس از به‌روزرسانی active=true و دارای تصویر/عرضه بود. این خواندن صرفاً smoke است؛ تست غیرفعال‌سازی با داده اصلی انجام نشد.
