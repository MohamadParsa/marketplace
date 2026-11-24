---
id: DEL-NEXT-ADJUSTMENTS-001
title: گزارش اجرای محلی کوپن، اعتبار انتخابی و گروه گارانتی
status: in-review
owner: eliassarband
updated: 2026-11-24
related: [FEAT-NEXT-006, FEAT-NEXT-007, FEAT-NEXT-008, FEAT-NEXT-017, API-CHECKOUT-ADJUSTMENTS-001]
approved_by: []
implementation_status: review
execution_authorized: true
---

# نتیجه #67، #68، #69 و #78

هر چهار قابلیت برای روش‌های نقدی فعلی در `D:\Cherriva` پیاده و بررسی شده‌اند؛ وضعیت review یعنی آماده بررسی مالک، نه انتشار. دستور تازه، اجرای هم‌زمان این چهار انتخاب را مجاز کرد؛ دستور قبلی «تغییرات بعدی را لوکال نگه دار» پابرجاست. **کامیت/پوش، PR، کامنت و بستن Issue در این نوبت انجام نشده‌اند.** مبنا: `52a8a09155fbcedae981eaf18a27ecea76120375`؛ کامیت جدیدی ادعا نمی‌شود. استقرار لیارا جداست.

[قرارداد و تصمیم‌های تأییدشده](../../architecture/contracts/checkout-adjustments.md) · [نتیجه ماشین‌خوان](issues-67-68-69-78-checks.json) · [صف کار](task-index.json).

## رفتار قابل مشاهده

- **#67:** بخش «کدهای تخفیف» در پنل؛ مبلغ ثابت تومان/درصد، هدف کالا/ارسال/جمع واجد شرایط، دسته با زیرشاخه، حداقل/حداکثر مبنا، سقف درصد، تاریخ، سهمیه کل/مشتری و انتخاب صریح نامحدود. سهم چریوا/سلر/ترکیبی در snapshot محفوظ است. مجوز خواندن، ساخت، ویرایش و فعال‌سازی مستقل و سمت سرور کنترل می‌شود؛ فهرست تعداد رزرو/مصرف و پیش‌نمایش بدون رزرو دارد.
- **#68:** باکس مستقل کوپن در مرحله دوم ویزارد؛ اعمال/حذف، پیام خطای روشن، حفظ ورودی و محاسبه سرور. تغییر ورودی/روش/شرایط، quote و رضایت قبلی را باطل می‌کند. درخواست در جریان با overlay و inert محافظت می‌شود؛ دوبار کلیک یک درخواست می‌سازد. پذیرش quote نسخه۲ به مبلغ و digest همان شرایط وابسته است.
- **#69:** کیف پول از روش‌های نسخه۲ حذف و به باکس اعتبار دلخواه تبدیل شده است؛ صفر/بخشی/کل با نمایش مانده و وجه لازم. اعتبار قبلی انتخاب‌نشده مصرف نمی‌شود؛ وجه تازه اول credit و سپس فقط تا نیاز همان خرید تخصیص می‌یابد. پذیرش خرید، hold موجودی/سهمیه/اعتبار را اتمیک می‌سازد؛ تأیید تیم یک debit دارد. انقضا/لغو و واریز دیرهنگام، اعتبار را مطابق قرارداد آزاد می‌کنند. محدودیت روش روی کل سفارش است؛ کسری زیر سقف، مجوز سفارش بالای سقف نیست. سفارش صفر گردش صفر ندارد و همچنان نیازمند تأیید تیم است.
- **#78:** در ویرایش خصوصیت، گروه نمایشی با نام و اعضای واقعی و وضعیت فعال؛ ثبت مستقل گروه‌ها با مجوز `attribute.groups.set` حتی برای تعریف واردشده پرگزینه، بدون تغییر schema واقعی. PDP یک چیپ برای گروه دارد و ارزان‌ترین عرضه موجود در ترکیب سایر محورها انتخاب می‌شود؛ نام واقعی کنار خرید، در عرضه‌ها، سبد و snapshot سفارش باقی است. URL قبلی raw variant حفظ می‌شود؛ انتخاب صریح سلر، همان عرضه واقعی است.

طرح A و رنگ‌های موجود حفظ شدند. روی storefront dropdown تازه اضافه نشده است. داده مالی، کوپن خصوصی و تخصیص اعتبار وارد کش عمومی نمی‌شوند؛ تغییر گروه dependency کاتالوگ را عوض می‌کند و schema کش v6 پاسخ قدیمی را جدا می‌کند.

## ردیابی پذیرش

| معیار | شاهد |
| --- | --- |
| FEAT-NEXT-006-AC01 | `TestCouponGoalsCategoriesRangesAndFunding`: هر سه هدف با مبلغ ثابت و درصد؛ محدودکردن تخفیف به ظرفیت هدف؛ سهم پلتفرم/سلر با جمع دقیق. |
| FEAT-NEXT-006-AC02 | همان تست: دسته پدر/فرزند، ارسال فروشگاه واجد، رنج مرزی و سقف؛ `TestCouponTimeRangeAndUserQuotaBoundaries` برای عدم اهلیت. |
| FEAT-NEXT-006-AC03 | `TestCouponValidationAccessVersionsAndSnapshot` و `TestCouponCreditHTTPAccessIdempotencyAndPostgresRestart`: ورودی، نقش، نسخه، idempotency و endpoint. |
| FEAT-NEXT-006-AC04 | `TestCouponQuotaConcurrentReservationExpiryAndCancellation` و تست free order: preview بدون مصرف، رزرو آخرین سهمیه هم‌زمان، آزادسازی و حفظ مصرف پس از لغو قطعی؛ تغییر قانون snapshot قبلی را عوض نمی‌کند. |
| FEAT-NEXT-007-AC01/02 | مرورگر اعمال/حذف و خطای کد، حفظ مرحله؛ تست دامنه برای هدف کالا/ارسال/order، تاریخ و بازه. |
| FEAT-NEXT-007-AC03 | مرورگر دوبار کلیک یک preview، overlay و ورودی غیرفعال؛ سرور digest/version، idempotency و کنترل مجدد شرایط زیر lock. |
| FEAT-NEXT-007-AC04 | مرورگر ترکیب کوپن/اعتبار و ثبت سفارش؛ تست snapshot و Money. اتصال روش فروش با مارکاپ غیرصفر در #73 باقی است. |
| FEAT-NEXT-008-AC01 | `TestSelectedWalletEarmarkNoDoubleDebitAndLegacyIsolation`، `TestWalletCoverageLimitsExpiryAndFreeOrder` و مرورگر برای صفر/جزئی/کل. |
| FEAT-NEXT-008-AC02 | `TestWalletConcurrentHoldAndRefundCannotStealCredit` و کنترل مبلغ منفی/بیش از کل/کسری؛ writer قدیمی هم hold جدید را خرج نمی‌کند. |
| FEAT-NEXT-008-AC03 | `TestSelectedWalletDoesNotBypassFundingOrderCap`: سفارش۲۰میلیونی با کسری زیر۱۵میلیون همچنان از کارت‌به‌کارت رد می‌شود؛ تست منع طلا برای local و policy کیف پول. |
| FEAT-NEXT-008-AC04 | انقضا، واریز دیرهنگام، لغو کامل از snapshot، یک debit و ledger قدیمی؛ reopen روی PostgreSQL. برگشت provider آنلاین خارج روش‌های فعلی است. |
| FEAT-NEXT-017-AC01/02 | تست مدل و مرورگر: چند نام واقعی زیر یک چیپ، ارزان‌ترین عرضه واقعی، ترکیب محور دیگر و انتخاب صریح سلر. |
| FEAT-NEXT-017-AC03 | `TestWarrantyMappingDoesNotAlterRealVariantOrHistoricalItem`: پس از تغییر نام گروه، تنوع و نام واقعی snapshot همان قبلی است. |
| FEAT-NEXT-017-AC04 | تست مجوز/overlap/عضو ناشناخته/نسخه/retry، نگاشت پرگزینه، URL قدیمی و invalidation/reopen؛ رفتار ناموجود از مسیر موجود PDP حفظ شده است. |

## بررسی انجام‌شده

- **۲۵۹ تست و زیرسناریوی Go موفق**، شامل تست‌های schema ایزوله PostgreSQL و کش Redis؛ `go vet ./...` موفق.
- دو skip صریح: `TestCatalogSnapshotPostgresRoundTrip` به fixture رسانه مستقل نیاز دارد و به این قابلیت مربوط نیست؛ `TestAdjustmentsBrowserHarness` فقط با opt-in دستی اجرا می‌شود و در suite عادی عمداً skip است.
- `npm run lint` و `npm run build` در هر دو app موفق؛ هشدار حجم chunk بک‌آفیس بیش از۵۰۰KB غیرمسدودکننده باقی است.
- **۳۱ بررسی مرورگر موفق، بدون runtime pageerror**: تعریف/پیش‌نمایش کوپن، گروه گارانتی/URL واقعی، سبد با نام واقعی، دو باکس پرداخت، اعتبار جزئی/کامل، رضایت تازه، یک برداشت، نمایش منابع و سهم هزینه در ادمین، رد دسترسی مشاهده‌گر، حذف کد و مبلغ منفی، overlay و دوبار کلیک؛ موبایل۳۹۰px بدون overflow صفحه.
- خطاهای۴۰۰/۴۰۳ در سناریوهای منفی عمداً ایجاد شده‌اند؛ «بدون خطای runtime» به معنای نبود پاسخ ناموفق موردانتظار شبکه نیست.
- تست مالی مرورگر فقط روی fixture حافظه‌ای جدا در پورت‌های۸۰۹۶/۵۱۹۶ اجرا شد؛ دیتابیس بررسی فعلی شما برای ساخت کوپن/سفارش/گردش ساختگی استفاده نشد. PostgreSQL integration هم schemaهای اختصاصی موقت خودش را داشت.
- race محلی اجرا نشد: CGO غیرفعال و ابزار C محلی موجود نیست؛ عبور تست هم‌زمانی معمولی ادعای اجرای race نیست.
- نسخه محلی واقعی با `Start-Cherriva.ps1 -NoBuild` به‌روز شد؛ هر دو سطح۲۰۰ و چهار ریل فعلی هرکدام۱۰محصول دارند. asset جدید storefront نیز تطبیق داده شد. حساب‌ها/رمزها/دفتر/بانک/کاتالوگ فعلی عمداً پاک یا جایگزین نشده‌اند.

## مشاهده محلی

- پنل: http://localhost:8098 — منوی «کدهای تخفیف»؛ نگاشت گارانتی در «کتابخانه خصوصیات» و ویرایش خصوصیت مربوط.
- سایت: http://localhost:5197 — مرحله دوم checkout برای کوپن و اعتبار. محصول نمونه واقعی واردشده: http://localhost:5197/product/chr-10082 و http://localhost:5197/product/chr-10092.
- نمونه‌های واقعی به‌صورت خودکار عضو گروه گارانتی آزمایشی نشده‌اند؛ گروه‌ها و کوپن آزمون فقط در harness جدا ساخته شدند. ادمین می‌تواند نگاشت واقعی موردنظر را در پنل ثبت کند.
- [حساب محلی موجود](../local-review-access.md) تغییر نکرده است. [راهنمای اجرا](../../local-development.md).

### تصاویر

[کوپن در موبایل](../../../output/playwright/issues-67-69-78/coupon-mobile.png) · [گارانتی گروهی و نام واقعی](../../../output/playwright/issues-67-69-78/warranty-mobile.png) · [باکس‌های خرید در موبایل](../../../output/playwright/issues-67-69-78/checkout-mobile.png) · [تفکیک مالی سفارش ادمین](../../../output/playwright/issues-67-69-78/admin-order.png) · [سفارش ادمین در موبایل](../../../output/playwright/issues-67-69-78/admin-order-mobile.png).

## فایل‌های اصلی و تکرار تست

- [Money/مدل usage](../../../apps/api/internal/domain/commerce/adjustments.go)، [کوپن](../../../apps/api/internal/application/backoffice/commerce_coupons.go)، [منابع اعتبار](../../../apps/api/internal/application/backoffice/commerce_credit.go)، [quote/پذیرش](../../../apps/api/internal/application/backoffice/commerce_checkout.go)، [تأیید/لغو](../../../apps/api/internal/application/backoffice/commerce_admin.go).
- [پنل کوپن](../../../apps/backoffice-web/src/features/commerce/coupons-page.tsx)، [مرحله خرید](../../../apps/storefront-web/src/commerce-pages.tsx)، [PDP](../../../apps/storefront-web/src/product-page.tsx)، [مدیریت گروه](../../../apps/backoffice-web/src/features/management/shared-attributes.tsx).
- [تست دامنه/سرویس](../../../apps/api/internal/application/backoffice/commerce_adjustments_test.go)، [HTTP/PostgreSQL](../../../apps/api/internal/bootstrap/commerce_adjustments_test.go)، [Money](../../../apps/api/internal/domain/commerce/adjustments_test.go).
- مرورگر: [سناریوی اصلی](../../../tests/e2e/checkout-adjustments-warranty.cjs)، [ادمین/دسترسی](../../../tests/e2e/checkout-admin-breakdown.cjs)، [خطا و هم‌زمانی UI](../../../tests/e2e/checkout-adjustment-edge-cases.cjs). harness حافظه‌ای باید تازه باشد؛ ابتدا سناریوی اصلی و سپس دو فایل دیگر اجرا شوند. harness opt-in با `CHERRIVA_ADJUSTMENTS_BROWSER=1` و `go test ./internal/bootstrap -run '^TestAdjustmentsBrowserHarness$' -count=1 -v -timeout 20m` از `apps/api` آغاز می‌شود؛ build هر دو app لازم است. برای اجرای مجدد سناریوی اصلی harness تازه لازم است.
- تست معمول: `go test ./...` و `go vet ./...` از `apps/api`. برای integration، متغیر خصوصی `MARKETPLACE_TEST_POSTGRES_URL` فقط به دیتابیس آزمون ایزوله و Redis آزمون همان محیط تنظیم شود؛ اطلاعات اتصال در این گزارش/گیت نیست.

## محدودیت‌ها و کار مستقل باقی‌مانده

موتور قیمت/کمیسیون/طلا و روش فروش/مارکاپ غیرصفر #72/#73 فعال نشده‌اند؛ روش‌های فعلی نقدی با مارکاپ صفرند. مثال۹۳۶ در تست Money و قرارداد نهایی ثبت شده، نه به‌عنوان درگاه اعتباری فعال. تسویه خودکار سهم کوپن، callback درگاه و لغو جزئی تازه خارج این تحویل‌اند؛ snapshot واحدها برای استفاده بعدی حفظ است. تحلیل جامع PlaceOrder #70 و گروه‌بندی عمومی فیلترهای PLP #79 نیز بسته نشده‌اند.

فلگ موجود `VITE_PDP_ADD_TO_CART` و توقف پیش‌فرض افزودن تازه در PDP تغییر نکرد؛ مرورگر برای آماده‌کردن سبد fixture از API خود برنامه استفاده کرد. این گزارش ادعای رفع آن توقف یا تست UI دکمه متوقف‌شده را ندارد. بنر/داده واقعی و مجوزها از این کار تغییر نمی‌کنند. برای انتشار بعدی، commit واقعی سورس/سند و بررسی CI و سپس تغییر Issue باید با دستور مالک انجام شود.
