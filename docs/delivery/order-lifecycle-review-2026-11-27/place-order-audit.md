---
id: DEL-ORDER-PLACE-AUDIT-001
title: مرز اتمیک فعلی و ماتریس شکست ثبت سفارش
status: in-review
owner: eliassarband
updated: 2026-11-27
related: [FEAT-NEXT-009, API-ORDER-LIFECYCLE-001, FEAT-BO-ORDER-REV-004]
approved_by: []
---

# ممیزی منبع فعلی برای #70/#133/#134

مرجع بررسی سورس `21ebcd5`؛ تغییر #132 فقط چیدمان کوپن است. این گزارش، مرز تراکنش و شاهدهای موجود را ثبت می‌کند؛ provider واقعی فعال نیست و موفقیت خارجی ادعا نمی‌شود. [قرارداد پذیرفته‌شده برش](../../architecture/contracts/order-lifecycle.md).

## نقطه اتمیک

`mutateCommerce` در [commerce_base.go](../../../apps/api/internal/application/backoffice/commerce_base.go) مجوز، expiry، fingerprint/idempotency، apply و ذخیره نتیجه را در `repository.Update` انجام می‌دهد. [PostgreSQL Update](../../../apps/api/internal/adapters/postgresrepo/repository.go) با `BeginFunc` و قفل `review_workspace FOR UPDATE` بارگذاری و save را در همان tx دارد؛ `saveCommerce` دفتر مالی، مرجع واریز و receipt را همان‌جا نگه می‌دارد. خطای apply یا save کل tx را rollback می‌کند. adapter حافظه نیز copy/lock قبل از commit دارد؛ جای آزمون PG را نمی‌گیرد.

AcceptQuote موجود، خرید/شماره، سفارش‌های فروشگاه، stock reserved، اعتبار، ظرفیت و نتیجه idempotency را در همین مرز ثبت می‌کند. ConfirmPurchase پس از شرایط معتبر، debit، مصرف موجودی، commit ظرفیت و وضعیت paid را همان‌جا قطعی می‌کند. ReviewDeposit یک tx جدا برای credit است؛ جدا بودن آن خواسته مالی است و تراکنش شکسته سفارش تلقی نمی‌شود. مرورگر/درگاه/اعلان هرگز داخل این ACID محلی نیستند.

## ماتریس failure/timeout/retry

| نقطه | شکست/تکرار | نتیجه مجاز | شاهد موجود / تکمیل |
| --- | --- | --- | --- |
| preview | اعتبار/موجودی عوض شد | رزرو/سفارش/پرداخت نسازد | TestCommercePreviewHasNoReservationAndPurchaseHasTwoGroups |
| پذیرش quote | مبلغ/digest/نشانی/روش تغییر کرد |۴۰۹؛ رضایت تازه، بدون منابع نیمه‌ثبت | TestCommerceValidSessionFreezesPriceAndExpiredNeedsCustomerConsent؛ TestShippingTermsPostgresTwoShipmentsRestartAndSellerScope |
| PlaceOrder | رقابت آخرین stock/credit/quota/capacity | فقط منابع موجود؛ بازنده بدون اثر جزئی | TestCommerceConcurrentLastStockAndRefundHolds؛ TestCouponQuotaConcurrentReservationExpiryAndCancellation؛ TestShippingCapacityPostgresConcurrentLastUnitRestartAndAPIChoices |
| ثبت/commit | خطای ذخیره یا قطع پاسخ پس از commit | rollback کامل یا نتیجه قابل بازیابی همان عملیات | idempotency محلی موجود؛ #134 سناریوی HTTP تکرار و snapshot پس از restart را صریح تکمیل می‌کند |
| idempotency | همان key/payload یا key با payload دیگر | همان نتیجه؛ یا۴۰۹بدون ساخت دوم | TestCommercePublicNumbersConcurrentPrivateAndRetry؛ TestCouponCreditHTTPAccessIdempotencyAndPostgresRestart |
| انقضا |۳۰دقیقه گذشت، آزادسازی تکراری | یک‌بار آزادسازی hold و نیاز تأیید مجدد | TestWalletCoverageLimitsExpiryAndFreeOrder؛ TestShippingCapacityPostgresConcurrentLastUnitRestartAndAPIChoices |
| واریز | مرجع تأیید تکراری/تغییر مبلغ | credit دوم یا تطبیق جعلی ممنوع | TestCommerceReviewPermissionsAndVerifiedReferenceUniqueness؛ TestCommerceHTTPRegistrationCreditAndPayment |
| وجه دیررس/اضافی | سفارش منقضی/لغو‌شده | فقط credit کیف پول؛ پرداخت/اسلات خودکار نه | TestSelectedWalletEarmarkNoDoubleDebitAndLegacyIsolation؛ #135 projection تلاش و تست نتیجه دیررس را تکمیل می‌کند |
| ConfirmPurchase | موجودی/روش ممنوع/کسری | debit/مصرف ناقص نشود؛ شرایط تازه طبق قرارداد | TestCommerceInsufficientWalletDoesNotDebitOrConsumeStock؛ TestCommerceMethodRestrictionRecheckedWithoutRepricingFrozenOrder |
| لغو پرداخت‌شده | retry | یک credit، حفظ محدودیت منبع | TestCommercePaidCancellationReturnsWalletOnce و تست‌های wallet policy؛ #138 مسیر جدید را به مانده تاریخی متصل می‌کند |
| ارسال/لغو هم‌زمان | version قدیمی | فقط یک نتیجه؛ ظرفیت پس از ارسال آزاد نشود | قرارداد ظرفیت موجود؛ آزمون عملیاتی تازه #136/#137 لازم است |
| برگشت/استرداد جزئی هم‌زمان | یک تعداد/مبلغ دو درخواست | سقف واحد/وجه/hold رعایت شود | #137/#138 لازم است؛ فعلاً اجرای محصول ادعا نشده |
| اعلان | پیام یا مصرف‌کننده قطع است | سفارش/مالی rollback نشود؛ رویداد تکراری پیام دوم نسازد | قرارداد رویداد #139؛ provider واقعی خارج برش |
| callback درگاه واقعی | تکراری/نامرتب/unknown/capture | بدون verify، paid نشود؛ retry کور نه | خارج adapter فعال؛ #135 قرارداد ایزوله و ممنوعیت فعال‌سازی را حفظ می‌کند |

## جمع‌بندی ممیزی

موتور اتمیک دوم لازم نیست. شکاف اثبات‌شده در این برش، وضعیت مستقل/نمایش و چرخه عملیاتی واقعی و پرونده/استرداد مرتبط است؛ به #134 تا #140 وصل شده است. شواهد موجود نام تست منبع است، نه ادعای اجرای تازه آن در #133. در #134 رگرسیون محلی/PG و شاهد مرز ثبت پیش از پرداخت ثبت می‌شود؛ #70 با صرف نوشتن این فایل بسته اعلام نمی‌شود. دامنه خارجی تا انتخاب provider معوق و با کیف پول نقدی فعلی یکی نیست.
