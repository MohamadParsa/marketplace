---
id: DEL-ORDER-REV-134-VERIFICATION
title: شواهد ثبت قبل از پرداخت و وضعیت مستقل سفارش
status: in-review
owner: eliassarband
updated: 2026-11-27
related: [FEAT-BO-ORDER-REV-004, FEAT-NEXT-009, API-ORDER-LIFECYCLE-001]
approved_by: []
implementation_status: done
---

# نتیجه #134

`AcceptQuote` موجود مرجع PlaceOrder باقی ماند؛ خرید شماره‌دار و سفارش/مرسوله‌های فروشگاه با snapshot پیش از تأمین وجه ساخته می‌شوند. موتور ثبت، پرداخت یا رزرو دیگری اضافه نشد. وضعیت مستقل placed/processing/completed/cancelled برای سفارش فروشگاه و والد و `fulfillment_status` جدا اضافه شد؛ پس از پرداخت همچنان placed است تا شروع دستی عملیات در #136.

نگاشت قدیمی فقط از وضعیت موجود است: awaiting-supply و پرداخت، شروع عملیات را ثابت نمی‌کنند؛ supplied/packing/shipped شاهد processing و delivered شاهد completed است. یک مرسوله تحویل‌شده، خرید چندمرسوله‌ای را تکمیل نمی‌کند؛ تحویل/لغو همه قسمت‌ها مطابق قرارداد تجمیع می‌شود. تاریخ/عامل/رویداد مالی جعلی backfill نمی‌شود. سفارش‌های کنارگذاشته‌شده پس از تأیید شرایط تازه، superseded و خارج فهرست جاری سلر می‌مانند؛ تاریخچه حذف نمی‌شود.

فیلدهای JSON اختیاری‌اند و شماره/شناسه و snapshot مالی حفظ شده؛ DDL/backfill پایگاه اصلی اجرا نشد. `Purchase.status` مالی/رزرو فعلی برای سازگاری حفظ و `order_status` نمای محصول مستقل است. label مشترک رابط‌ها نیز به عنوان فارسی درست متصل شد.

## بررسی اجراشده

- تست تازه `TestPlacedOrderPostgresBeforeFundingAndReplayAfterRestart`: دو فروشگاه، رد مبلغ دستکاری‌شده بدون اثر جزئی، ثبت دو مرسوله/رزرو بدون واریز یا debit، بازکردن مجدد repository/handler، replay دقیق همان پاسخ committed و جلوگیری از خرید/رزرو دوم و key با payload متفاوت؛ PostgreSQL با schema یک‌بارمصرف.
- تست تازه `TestPlacedOrderPaymentDoesNotInventProcessingAndPartialDelivery`: قبل و بعد از پرداخت، placed درست؛ تحویل بخشی/لغو بقیه و منع جعل history/ledger هنگام نگاشت.
- رگرسیون checkout/کوپن/اعتبار/محدودیت/ظرفیت/حمل با PostgreSQL موفق؛ سپس `go test ./...` با PostgreSQL و Redis آزمایشی DB15 و `go vet ./...` موفق بودند. race محلی اجرا نشد؛ CI لینوکس آن را پوشش می‌دهد.
- build/TypeScript بک‌آفیس و storefront موفق؛ هشدار موجود اندازه bundle بک‌آفیس باقی است. فرم تازه‌ای در #134 اضافه نشده؛ پذیرش رابط چرخه کامل در #140 و تست نهایی مالک جداست.

## ارتباط با #70

[ممیزی مرز و failure matrix](place-order-audit.md) مسیر محلی را به tx/lock و تست واقعی وصل می‌کند؛ این تست تازه مرز ثبت پیش از funding و بازیابی پاسخ پس از commit/restart را صریح تکمیل کرد. موتور اتمیک فعلی حفظ شده و شکاف‌های عملیاتی/استرداد در #136 تا #139 مالک مشخص دارند. provider/callback خارجی واقعی همچنان فعال نشده و پذیرش آن ادعا نمی‌شود؛ نتیجه نامشخص نمی‌تواند به معنی پرداخت قطعی فرض شود.

پوش، CI و کامنت بستن مستقل ثبت می‌شوند؛ merge/deploy عمومی انجام نشده است.
