---
id: FEAT-BO-REFINEMENT-004
title: حذف کدهای gsp از نمایش و قرارداد مصرفی خصوصیات پنل با نام‌گذاری chr
status: in-review
owner: eliassarband
updated: 2026-11-26
related: [PLAN-BO-REFINEMENT-001]
approved_by: []
implementation_status: done
execution_authorized: true
priority: P1
---

# CV-BOR-04 — حذف کدهای gsp از نمایش و قرارداد مصرفی خصوصیات پنل با نام‌گذاری chr

Issue: [#110](https://github.com/MohamadParsa/marketplace/issues/110)

<!-- cherriva-backoffice-refinement-20261125 CV-BOR-04 -->
مسئول: @eliassarband · والد برنامه #95 · اولویت پیشنهادی **P1**

**وضعیت: به انتخاب مالک تکمیل و تست شد.** [گزارش، تصاویر و محدودیت‌ها](../issue-110-verification.md).

## مسئله و نتیجه
کاربر پنل نیز هویت چریوا و عنوان خوانا ببیند، نه کد فنی مبدأ واردسازی.

## خواسته‌های تأییدشده
- نمونه gsp-axis-color هنوز در پنل دیده می‌شود؛ مالک نام‌گذاری chr و خوانایی را خواسته است.

## راه‌حل فنی و مرز کار
- #102 شناسه‌های عمومی storefront را اصلاح کرده و بسته است؛ این تسک شکاف پنل، code خصوصیات و مراجع مصرفی باقی‌مانده را ممیزی/تکمیل می‌کند.
- عنوان «رنگ» و نقش خصوصیت در حالت عادی نمایش داده شود؛ کد نمایشی canonical جدید با chr باشد و سرور ورودی سازگار قدیمی را resolve کند. قرارداد قالب کد خصوصیت در شروع کار ثبت شود؛ rename کور رشته‌ای انجام نشود.
- بین AttributeDefinition.ID، Code، alias عمومی و source reference تفاوت گذاشته شود؛ تصمیم قبلی حفظ PK/FK، سفارش و seed/manifest مبناست. اگر تغییر PK هم لازم دانسته شد، قبل از اجرا طرح migration و تصمیم جدا لازم است.
- همان خصوصیت مشترک، مقادیر محصول، binding دسته، محور، فیلتر و گروه گارانتی دو نسخه تکراری پیدا نکنند. واردسازی تکراری همچنان idempotent بماند.

## فرم‌ها و مدل‌های فعلیِ بررسی‌شده
- `apps/api/internal/application/backoffice/public_references.go`
- `apps/backoffice-web/src/features/management/shared-attributes.tsx`
- `apps/api/internal/application/backoffice/catalog_seed.go`

## ترتیب و عدم هم‌پوشانی
پیش‌نیاز پایه: #102؛ #108 برای UX کلی مرتبط است. alias/projection مستقل تکمیل شد و تبدیل محور یا UX کامل آن کار را اجرا نمی‌کند.

## معیار پذیرش
- [x] **FEAT-BO-REFINEMENT-004-AC01:** جست‌وجو، گرید و دیالوگ خصوصیات پنل به‌جای gsp-axis-color، عنوان خوانا و در صورت نیاز کد chr نشان دهند.
- [x] **FEAT-BO-REFINEMENT-004-AC02:** ورودی/لینک قدیمی به همان تعریف واقعی برسد؛ دو خصوصیت رنگ یا تنوع موازی ساخته نشوند.
- [x] **FEAT-BO-REFINEMENT-004-AC03:** مقادیر، قیمت، موجودی، گارانتی، سبد/سفارش و checksum snapshot ورودی محفوظ باشند؛ restart و retry واردسازی آزمون شوند.

## پرسش یا طراحی باقی‌مانده
- حل‌شده در [قرارداد](../../../architecture/contracts/catalog-attributes.md#تکمیل-مرز-پنل-و-کد-چریوا--110): alias نوع a از ID و Code مستقل، با حفظ ذخیره اصلی.

## قواعد تحویل و شرط شروع
طرح **A — متعادل و شیک** و نشان C مرجع‌اند. مجوز actor و scope کاربر/فروشگاه در سرور، version/idempotency متناسب، حفظ ورودی در خطا و حالت‌های loading/empty/error/forbidden/conflict لازم‌اند. تغییر معتبر عمومی باید cache و projection مرتبط را تازه کند؛ داده خصوصی/رمز/کد ملی وارد cache عمومی و log/export بی‌مجوز نشود. RTL، کیبورد، عرض محدود و اکشن/صفحه‌بندی قابل دسترس باشند؛ dropdown تازه جز استان/شهر استثنای تأییدشده ساخته نشود. نام و شماره و داده نمونه ساختگی باشند.

این کار با دستور صریح مالک اجرا شد؛ سند in-review و اجرا done هستند. دستور تازه مالک در۲۰۲۶-۱۱-۲۵ پوش پس از تکمیل را مجاز کرده است؛ انتشار لیارا جدا می‌ماند.
