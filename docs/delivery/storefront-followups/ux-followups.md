---
id: DEL-SF-UX-001
title: صف اصلاحات تجربه storefront پس از کش و URL
status: in-review
owner: eliassarband
updated: 2026-11-23
related: [FEAT-SF-001, API-PLP-SEARCH-001, API-PRODUCT-DETAIL-001, FEAT-SF-CACHE-001]
approved_by: []
implementation_status: todo
---

# اصلاحات جدید تجربه سایت — ۲۰۲۶-۱۱-۲۳

این بسته با ثبت۱۰تسک شروع شد. مالک ابتدا اجرای #45، #46 و #47 را خواست؛ هر سه با شواهد بسته شدند. اکنون **#48 و سپس #49** انتخاب شده‌اند؛ سایر تسک‌ها منتظر انتخاب‌اند. همه Issueها به `eliassarband` تخصیص می‌یابند و تا اجرای واقعی، ثبت شواهد و کامنت پایان باز می‌مانند. P1/P2/P3 پیشنهاد برنامه‌ریزی‌اند، نه دستور شروع. شماره ردیف، ترتیب متن مالک را دنبال می‌کند.

## قواعد مشترک

- طرح منتخب **A — متعادل و شیک**، رنگ primary فعلی و لوگوی C حفظ شوند؛ صفحه/فرانت تازه یا بازطراحی پالت در دامنه نیست.
- خواسته صریح مالک از پیشنهاد فنی جدا نوشته شده است؛ جزئیات ارتفاع، ردیف و روش پیاده‌سازی پیشنهادند. تصمیم سنجه پرفروش‌ترین از R20 موجود خوانده می‌شود، نه فرض تازه؛ SFUX-Q01 فقط جزئیات باز همان مسیر را مشخص می‌کند.
- تغییرات storefront هستند؛ بک‌آفیس فقط در تنظیم بازه سنجه فروشِ تسک01 دخیل است. درخواست به checkbox پنل‌ها، مدل مالی/قیمت طلا، موجودی، lead_days یا تعهد زمان تحویل تعمیم داده نشود.
- ریدایرکت و canonical جدید #42، کش نسخه‌دار و warmup #41، انتشار اجداد، قیمت/انتخاب معتبر، تومان و wizard دو مرحله‌ای خرید حفظ شوند. هیچ‌یک از #27تا30 و #41/#42 دوباره باز نمی‌شود.
- موبایل۳۹۰ و دسکتاپ۱۴۴۰، RTL، کیبورد/focus/اعلان دسترس‌پذیر، خطا/retry و شبکه کند متناسب با هر تسک بررسی شوند. داده آزمون ساختگی و ایزوله باشد.
- بعد از انتخاب و تکمیل هر تسک: سورس و گزارش/تصاویر کامیت و پوش، CI و بررسی متناسب، کامنت نتیجه و مسیر شواهد، سپس بستن همان Issue. انتشار لیارا با پارسا باقی می‌ماند.

## صف قابل انتخاب

| کد | کار | اولویت پیشنهادی | وضعیت |
| --- | --- | --- | --- |
| CV-SF-UX-01 · [#45](https://github.com/MohamadParsa/marketplace/issues/45) | [افزودن مرتب‌سازی پرفروش‌ترین با سنجه واقعی فروش](tasks/cv-sf-ux-01-best-selling-sort.md) | P2 | done؛ با کامنت بسته شد؛ [شواهد](issue-45-verification.md) |
| CV-SF-UX-02 · [#46](https://github.com/MohamadParsa/marketplace/issues/46) | [منوی بازشونده درختی دسته‌ها و خالی‌ماندن ادامه نوار اصلی](tasks/cv-sf-ux-02-hierarchical-category-menu.md) | P2 | done؛ با کامنت بسته شد؛ [شواهد](issue-46-verification.md) |
| CV-SF-UX-03 · [#47](https://github.com/MohamadParsa/marketplace/issues/47) | [بازطراحی مینیمال آیکن پروفایل و سبد با رنگ برند](tasks/cv-sf-ux-03-header-account-cart-icons.md) | P3 | done؛ با کامنت بسته شد؛ [شواهد](issue-47-verification.md) |
| CV-SF-UX-04 · [#48](https://github.com/MohamadParsa/marketplace/issues/48) | [اسلایدر قیمت و نمایش بیشتر فیلترهای دارای بیش از۵گزینه در PLP](tasks/cv-sf-ux-04-plp-price-slider-filter-disclosure.md) | P2 | done؛ با کامنت بسته شد؛ [شواهد](issue-48-verification.md) |
| CV-SF-UX-05 · [#49](https://github.com/MohamadParsa/marketplace/issues/49) | [جایگزینی نمایش checkboxهای storefront با Switch مشترک](tasks/cv-sf-ux-05-storefront-switch-controls.md) | P2 | review؛ [شواهد](issue-49-verification.md) |
| CV-SF-UX-06 · [#50](https://github.com/MohamadParsa/marketplace/issues/50) | [چیپ مرتب‌سازی بدون radio نمایان و با رنگ کامل primary](tasks/cv-sf-ux-06-sort-primary-chips.md) | P3 | todo؛ منتظر انتخاب مالک |
| CV-SF-UX-07 · [#51](https://github.com/MohamadParsa/marketplace/issues/51) | [Progress overlay یکپارچه و بدون جابه‌جایی محتوا در انتخاب، سبد و تمام مسیر خرید](tasks/cv-sf-ux-07-stable-overlay-progress.md) | P1 | todo؛ منتظر انتخاب مالک |
| CV-SF-UX-08 · [#52](https://github.com/MohamadParsa/marketplace/issues/52) | [ارتفاع محدود معرفی و مشخصات فنی PDP با مشاهده بیشتر/کمتر](tasks/cv-sf-ux-08-pdp-content-read-more.md) | P2 | todo؛ منتظر انتخاب مالک |
| CV-SF-UX-09 · [#53](https://github.com/MohamadParsa/marketplace/issues/53) | [یکسان‌سازی متن مشتری به «ارسال از … روز کاری»](tasks/cv-sf-ux-09-storefront-dispatch-wording.md) | P2 | todo؛ منتظر انتخاب مالک |
| CV-SF-UX-10 · [#54](https://github.com/MohamadParsa/marketplace/issues/54) | [چیدمان جمع‌وجور چیپ‌های محور تنوع، مخصوص وزن‌های متعدد طلا](tasks/cv-sf-ux-10-pdp-dense-variant-chips.md) | P2 | todo؛ منتظر انتخاب مالک |

## مرزهای جلوگیری از هم‌پوشانی

- 01 سنجه/API/تنظیم بازه پرفروش‌ترین؛ 06 فقط ظاهر گروه sort. هرکدام مستقل قابل انتخاب‌اند.
- 02 ساختار منوی درختی و خالی‌بودن ادامه ردیف؛ 03 فقط آیکن پروفایل/سبد در هدر مشترک.
- 04 slider و جمع‌شدن گزینه‌های فیلتر؛ 05 Switch boolean مشترک در کل storefront. انتخاب‌های رادیویی، کارت روش پرداخت و ارسال Switch نمی‌شوند.
- 07 مالک واحد progress overlay مشترک و پوشش تمام مراحل خرید است؛ دیگر کارها اگر آماده باشد آن را مصرف می‌کنند، جایگزین مستقل دوم نمی‌سازند.
- 08 معرفی/جدول مشخصات PDP؛ 10 چیپ‌های محور تنوع؛ هر دو نمایش بیشتر دارند ولی داده و رفتار جداست.
- 09 فقط اصطلاح زمان ارسال در سایت مشتری؛ عملیات تأمین/تقویم #10 یا اسلات #16 را دوباره پیاده نمی‌کند.

[فهرست ماشین‌خوان](ux-task-index.json) · [سناریوهای رفتاری مستند](../../bdd/features/experiences/storefront-ux-followups.feature) · [پرسش محدود سنجه SFUX-Q01](../../product/open-questions.md#sfux-q01) · [صف و تحویل‌های قبلی](README.md).
