---
id: DEL-NEXT-001
title: صف توسعه بعدی چریوا پس از اصلاحات تجربه فروشگاه
status: in-review
owner: eliassarband
updated: 2026-11-24
related: [DEL-SF-UX-001, ADR-009]
approved_by: []
implementation_status: in-progress
execution_authorized: true
---

# برنامه درخواست‌های جدید — ۲۰۲۶-۱۱-۲۳

۳۰Issue ثبت و به الیاس (`eliassarband`) اساین شده‌اند. وضعیت جاری: هشت تحویل قبلی در main، دو تحویل پارسا #89/#90 روی شاخه‌های او و چهار تحویل #67/#68/#69/#78 در [PR #94](https://github.com/MohamadParsa/marketplace/pull/94) انجام و Issueهایشان بسته شده‌اند؛۱۶کار دیگر همچنان باز است. چهار Issue اخیر طبق دستور صریح مالک با کامنت شواهد بسته شدند؛ PR #94 هنوز ادغام نشده است. [گزارش چهار قابلیت](issues-67-68-69-78-verification.md). نسخه یکپارچه پارسا و اصلاح دکمه تا دستور انتشار تازه لوکال می‌مانند؛ پیگیری #105. انتشار لیارا و انتخاب بقیه کارها جداست. صف #45تا#54 done/closed باقی است.

هر دو سرویس نظر #89 و پرسش‌وپاسخ #90 پارسا در ادغام محلی حاضرند؛ Issueها توسط پارسا بسته‌اند و انتشار main از این وضعیت نتیجه نمی‌شود.

[تصمیم تازه کیف پول/کوپن/گارانتی](../../decisions/adr-009-checkout-credit-and-warranty-direction.md) · [فهرست ماشین‌خوان](task-index.json) · [پرسش‌ها](../../product/open-questions.md#next-phase-questions).

## پوشش و اولویت پیشنهادی

| کد | کار | اولویت / موج | وضعیت اجرا |
| --- | --- | --- | --- |
| CV-NEXT-01 · [#62](https://github.com/MohamadParsa/marketplace/issues/62) | [تخفیف ویژه: لیست پنل و اولویت انتخابی با Drag & Drop](tasks/cv-next-01-special-offer-priority.md) | P2 / W1 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-02 · [#63](https://github.com/MohamadParsa/marketplace/issues/63) | [علاقه‌مندی مشتری و آیکن کنار گالری PDP](tasks/cv-next-02-customer-wishlist.md) | P2 / W2 | done؛ [شواهد](issue-63-verification.md) |
| CV-NEXT-03 · [#64](https://github.com/MohamadParsa/marketplace/issues/64) | [ثبت و نمایش محصولات اخیراً دیده‌شده هر مشتری](tasks/cv-next-03-customer-recently-viewed.md) | P2 / W2 | done؛ [شواهد](issue-64-verification.md) |
| CV-NEXT-04 · [#65](https://github.com/MohamadParsa/marketplace/issues/65) | [مقایسه محصولات از یک دسته‌بندی](tasks/cv-next-04-same-category-comparison.md) | P2 / W2 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-05 · [#66](https://github.com/MohamadParsa/marketplace/issues/66) | [قرارداد مبلغ سفارش، CreditUsage و DiscountUsage](tasks/cv-next-05-checkout-amounts-usage-contract.md) | P1 / W0 | done؛ [تحلیل](issue-66-analysis.md) |
| CV-NEXT-06 · [#67](https://github.com/MohamadParsa/marketplace/issues/67) | [کد تخفیف: تعریف و محدودیت‌ها در پنل و ارزیابی سرور](tasks/cv-next-06-coupon-management.md) | P1 / W3 | done/closed؛ سورس در PR #94، ادغام‌نشده؛ [شواهد چهار قابلیت](issues-67-68-69-78-verification.md) |
| CV-NEXT-07 · [#68](https://github.com/MohamadParsa/marketplace/issues/68) | [باکس مستقل کد تخفیف در مرحله پرداخت خرید](tasks/cv-next-07-checkout-coupon-box.md) | P1 / W3 | done/closed؛ سورس در PR #94، ادغام‌نشده؛ [شواهد چهار قابلیت](issues-67-68-69-78-verification.md) |
| CV-NEXT-08 · [#69](https://github.com/MohamadParsa/marketplace/issues/69) | [باکس اعتبار کیف پول و مبلغ دلخواه کنار روش پرداخت](tasks/cv-next-08-checkout-wallet-credit.md) | P1 / W3 | done/closed؛ سورس در PR #94، ادغام‌نشده؛ [شواهد چهار قابلیت](issues-67-68-69-78-verification.md) |
| CV-NEXT-09 · [#70](https://github.com/MohamadParsa/marketplace/issues/70) | [PlaceOrder: شفاف‌سازی قرارداد اتمیک و سناریوهای شکست](tasks/cv-next-09-atomic-place-order.md) | P1 / W0 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-10 · [#71](https://github.com/MohamadParsa/marketplace/issues/71) | [بازبینی دقیق دسترسی و محدوده داده فروشندگان](tasks/cv-next-10-seller-access-audit.md) | P1 / W0 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-11 · [#72](https://github.com/MohamadParsa/marketplace/issues/72) | [بازبینی قیمت‌گذاری دیجیتال/طلا و تخفیف مستقیم محصول](tasks/cv-next-11-pricing-direct-discounts.md) | P1 / W0 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-12 · [#73](https://github.com/MohamadParsa/marketplace/issues/73) | [روش فروش، اتصال به روش پرداخت و مارکاپ در خرید](tasks/cv-next-12-sales-method-markup.md) | P1 / W0 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-13 · [#74](https://github.com/MohamadParsa/marketplace/issues/74) | [SEO: ممیزی، فهرست تغییرات مصوب و اجرای اصلاحات](tasks/cv-next-13-seo-audit-remediation.md) | P2 / W2 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-14 · [#75](https://github.com/MohamadParsa/marketplace/issues/75) | [ActivityLog: سابقه تغییرات و عملیات کسب‌وکار در پنل](tasks/cv-next-14-business-activity-log.md) | P1 / W0 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-15 · [#76](https://github.com/MohamadParsa/marketplace/issues/76) | [CustomerActivity و Analytics رفتار مشتری](tasks/cv-next-15-customer-activity-analytics.md) | P2 / W2 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-16 · [#77](https://github.com/MohamadParsa/marketplace/issues/77) | [لاگ فنی عملیات حساس، خطا و شکست با correlation](tasks/cv-next-16-sensitive-operation-logs.md) | P1 / W0 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-17 · [#78](https://github.com/MohamadParsa/marketplace/issues/78) | [گروه نمایشی گارانتی و حفظ گارانتی واقعی هر عرضه](tasks/cv-next-17-warranty-display-groups.md) | P2 / W2 | done/closed؛ سورس در PR #94، ادغام‌نشده؛ [شواهد چهار قابلیت](issues-67-68-69-78-verification.md) |
| CV-NEXT-18 · [#79](https://github.com/MohamadParsa/marketplace/issues/79) | [گروه‌بندی مقادیر خصوصیات پرتعداد در گروه‌های کوچک‌تر](tasks/cv-next-18-attribute-value-groups.md) | P2 / W2 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-19 · [#80](https://github.com/MohamadParsa/marketplace/issues/80) | [بازبینی و اصلاح نمایش فیلترهای PLP](tasks/cv-next-19-plp-filter-experience.md) | P2 / W1 | done؛ [شواهد](issue-80-verification.md) |
| CV-NEXT-20 · [#81](https://github.com/MohamadParsa/marketplace/issues/81) | [ممیزی تمام گریدهای پنل، فیلترها و خروجی Excel](tasks/cv-next-20-backoffice-grid-capabilities.md) | P2 / W2 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-21 · [#82](https://github.com/MohamadParsa/marketplace/issues/82) | [سرویس ایمیل تراکنشی، قالب و پیگیری ارسال](tasks/cv-next-21-transactional-email-service.md) | P2 / W2 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-22 · [#83](https://github.com/MohamadParsa/marketplace/issues/83) | [تب مستقل مزایا و معایب کنار مشخصات فنی PDP](tasks/cv-next-22-pdp-pros-cons-tab.md) | P2 / W1 | done؛ [شواهد](issue-83-verification.md) |
| CV-NEXT-23 · [#84](https://github.com/MohamadParsa/marketplace/issues/84) | [بازشدن زیرمنو با hover متن یا فلش و حفظ کلیک](tasks/cv-next-23-category-menu-hover.md) | P2 / W1 | done؛ [شواهد](issue-84-verification.md) |
| CV-NEXT-24 · [#85](https://github.com/MohamadParsa/marketplace/issues/85) | [رفع اسکرول عمودی و دسترسی به تمام گره‌های منو](tasks/cv-next-24-category-menu-scroll.md) | P1 / W1 | done؛ [شواهد](issue-85-verification.md) |
| CV-NEXT-25 · [#86](https://github.com/MohamadParsa/marketplace/issues/86) | [اولویت نمایش دسته و ترتیب پایدار گره‌های منو](tasks/cv-next-25-category-display-priority.md) | P2 / W1 | done؛ [شواهد](issue-86-verification.md) |
| CV-NEXT-26 · [#87](https://github.com/MohamadParsa/marketplace/issues/87) | [تکمیل محتوای فوتر و صفحات درباره چریوا و راهنما](tasks/cv-next-26-footer-business-content.md) | P2 / W1 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-27 · [#88](https://github.com/MohamadParsa/marketplace/issues/88) | [تکمیل مدیریت سکشن‌ها با فرم نوع‌دار و انتشار](tasks/cv-next-27-typed-section-management.md) | P2 / W2 | todo؛ اجرا انتخاب نشده |
| CV-NEXT-28 · [#89](https://github.com/MohamadParsa/marketplace/issues/89) | [سرویس نظرات کاربران؛ ادامه CV-WEB-15](tasks/cv-next-28-product-reviews-service.md) | P2 / W2 | done/closed؛ کد شاخه پارسا، ادغام محلی |
| CV-NEXT-29 · [#90](https://github.com/MohamadParsa/marketplace/issues/90) | [سرویس پرسش و پاسخ محصول؛ ادامه CV-WEB-16](tasks/cv-next-29-product-qa-service.md) | P2 / W2 | done/closed؛ کد شاخه پارسا، ادغام محلی · [شواهد](issue-90-verification.md) |
| CV-NEXT-30 · [#91](https://github.com/MohamadParsa/marketplace/issues/91) | [لندینگ اقساطی متصل به روش فروش و شرایط واقعی](tasks/cv-next-30-installment-landing.md) | P2 / W3 | todo؛ اجرا انتخاب نشده |

## موج‌ها؛ ترتیب قابل انتخاب

- **W0 — قرارداد و ممیزی:** مبالغ/usage، PlaceOrder، قیمت و مارکاپ، دسترسی سلر و ردیابی حساس. PlaceOrder طبق خواست مالک نیازمند شفاف‌سازی بعدی است؛ تحلیل، اجازه اجرای کد مالی نیست.
- **W1 — تجربه و رفع ایراد:** اسکرول منو اولویت بالاتر دارد؛ hover، ترتیب دسته، فیلترها، مزایا/معایب، اولویت تخفیف ویژه و محتوای فوتر. W1های بدون وابستگی می‌توانند مستقل از تصمیم مالی شروع شوند، فقط با انتخاب مالک.
- **W2 — خدمات و مدیریت:** علاقه‌مندی/سابقه/مقایسه، گروه‌بندی، پنل، ایمیل، SEO، audit و تعاملات محصول. برش انتخابی فیلتر19 طبق دستور checkbox/چیپ مالک وابستگی سخت به گروه‌بندی18 ندارد؛ گروه‌بندی معنایی همچنان کار انتخاب‌نشده مستقل است.
- **W3 — خرید مالی/اقساطی:** قرارداد05 و قیمت11 → روش فروش12 و کوپن06 → طراحی اتمیک09 → کوپن checkout07 → اعتبار checkout08. لندینگ30 از روش فروش12 مصرف می‌کند؛ provider واقعی شرط فعال‌شدن پرداخت است، نه نوشتن متن پیشنهادی.

## مالکیت یکتا و جلوگیری از هم‌پوشانی

| موضوع | مالک این مرحله | مرز محفوظ |
| --- | --- | --- |
| تخفیف ویژه با اولویت منتخب | CV-NEXT-01 · [#62](https://github.com/MohamadParsa/marketplace/issues/62) | eligibility تخفیف پویاست؛ ریل دستی عمومی #8 محصول دلخواه دارد و تسک دیگری است |
| موتور/قیمت/کوپن | 05 قرارداد مبلغ؛11 قیمت مستقیم؛12 مارکاپ؛06 تعریف کوپن؛07 مصرف در خرید | محاسبه سرور مشترک؛ هیچ resolver کلاینت/دفتر مالی موازی ساخته نشود |
| اعتبار کیف پول | 08 با قرارداد05 و PlaceOrder09 | اعتبار و تخفیف و روش پرداخت سه مفهوم جدا؛ R02/R11 قدیمی طبق ADR-009 بازنگری می‌خواهند |
| ActivityLog / Analytics / لاگ خطا | به‌ترتیب14 / 15 / 16 | حسابرسی کسب‌وکار، رفتار مشتری و خطای فنی یکی نیستند؛ correlation می‌تواند مشترک باشد |
| گروه‌بندی و گارانتی | 18 زیرساخت گروه مقادیر؛17 رفتار دامنه گارانتی | مقدار/گارانتی واقعی تنوع و سفارش حفظ شود؛ ادغام مخرب یا دسته محصول تازه فرض نشود |
| منو |23 تعامل hover؛24 scroll؛25 اولویت داده | مالک مشترک جزء منو، تست رگرسیون ترکیبی؛ #46 باز نمی‌شود |
| مدیریت سکشن |27 هسته نوع‌دار | #7 بنر responsive تحویل شده؛ #8 ریل دستی و #9 sticky مالکان اجرایی موجودند؛ Issue کپی ساخته نشود |
| نظر و پرسش/پاسخ |28 و29 | ادامه کارهای تاریخی CV-WEB-15/16 و گیت‌های #25؛ گیت مشترک از #25 است. نظر در #89 و پرسش‌وپاسخ در #90 پیگیری می‌شوند |
| هویت و نقش پایه | #1/#2/#3 | CV-NEXT-10 ممیزی scoped فروشنده است؛ نقص بنیانی به مالک موجود لینک شود، نه ورود/هویت موازی |
| تنظیمات عمومی | #15 | کلیدهای مشترک از این مسیر؛ بازنویسی کل تنظیمات شرط کار مستقل نیست |

ثبت۳۰Issue به معنی فعال‌سازی #8/#9 یا بستن #11/#12/#15 نیست. وضعیت آن‌ها همان وضعیت واقعی می‌ماند و فقط رابطه برنامه به آن‌ها اضافه می‌شود. بسته تاریخی design-review تغییر نمی‌کند.

## آمادگی و شواهد این مرحله

قواعد صریح از پیشنهادها جدا هستند؛ پرسش‌های مالی/گارانتی/ردیابی قبل از مسیر وابسته پاسخ می‌خواهند. [معیار آمادگی](../readiness.md) برای هر قابلیت جداست. در مرحله ثبت فقط پوشش۳۰خواسته، گراف وابستگی، لینک‌ها/شناسه‌ها و همگام‌سازی GitHub بررسی می‌شوند؛ هیچ تست محصول یا AC اجراشده ادعا نمی‌شود. [گزارش ثبت](planning-verification.md) پس از sync تکمیل می‌شود.

## تحویل اجرایی قبلی

#80، #83، #84، #85 و #86 پیاده و محلی آزموده شده‌اند؛۸۴کنترل مرورگر و۲۱۷کنترل Go موفق‌اند. هر۵ job CI موفق و PR #92 ادغام شد؛ هر پنج ایشو با کامنت شواهد بسته شدند. سورس و مدارک روی main و نسخه محلی D:\Cherriva موجودند؛ لیارا منتشر نشده و انتخاب‌های بعدی و تحلیل محلی #66 در بخش بعد آمده‌اند. [گزارش #80](issue-80-verification.md) · [#83](issue-83-verification.md) · [#84](issue-84-verification.md) · [#85](issue-85-verification.md) · [#86](issue-86-verification.md).

## تحویل‌های جاری #66/#63/#64

- [تحلیل #66](issue-66-analysis.md): قرارداد مبلغ/مصرف اعتبار و تخفیف، مثال‌ها و طرح مهاجرت؛ موتور مالی و migration تازه اجرا نشده‌اند.
- [علاقه‌مندی #63](issue-63-verification.md): قلب PDP، آیکن کنار سبد برای واردشده و فهرست شخصی؛ نمایش خالی/پُر بدون متن/کادر طبق اصلاح مالک.
- [بازدیدهای اخیر #64](issue-64-verification.md):۵۰کالا/۳۰روز، حذف/clear، privacy و cleanup startup/ساعتی؛ breadcrumb با `>` در RTL اصلاح شد.

تغییرات در [PR #93](https://github.com/MohamadParsa/marketplace/pull/93) پس از موفقیت هر۵ job CI، شامل Go race/vet، روی main با کامیت `a253d5edbe0adb467e1e7fb2ab5903608fb64bdc` ادغام شدند. این Issue با کامنت شواهد بسته شد. استقرار لیارا انجام نشد؛ نسخه D:\Cherriva نیز با main همگام است. تغییرهای بعد از این دسته تا دستور صریح تازه مالک لوکال می‌مانند.

نسخه ترکیبی، کنترل خرید و نمایش دسته‌های جدید main را حفظ می‌کند؛ build/lint و۶۵کنترل مرورگر روی آن موفق‌اند. نتایج Go محلی۲۲۷تست/زیرتست و یک skip اختیاری بود؛ CI همین PR نیز race و هر۵ job را موفق بررسی کرد. وضعیت سند همچنان in-review و وضعیت اجرای این هشت کار done است. این آمار مربوط به تحویل قبلی است؛ وضعیت جاری چهار انتخاب مالی/گارانتی در بخش بعد آمده است.

## اجرای چهارتایی محلی۲۰۲۶-۱۱-۲۴

با دستور مالک، #67/#68/#69/#78 یکجا اجرا و سورس در commit `1dd7ccd616ba0d6380ce75adcdf9aaf38310a64d` و PR #94 پوش شد. پس از پنج job موفق CI و دستور تازه مالک در۲۰۲۶-۱۱-۲۴، هر چهار Issue با کامنت شواهد completed/closed شدند. PR هنوز باز است؛ ادغام main و استقرار انجام نشده است. [قرارداد](../../architecture/contracts/checkout-adjustments.md) · [گزارش و تصاویر](issues-67-68-69-78-verification.md).

انتشار مجاز این چهار انتخاب: [سورس 1dd7ccd](https://github.com/MohamadParsa/marketplace/commit/1dd7ccd616ba0d6380ce75adcdf9aaf38310a64d) · [PR #94](https://github.com/MohamadParsa/marketplace/pull/94). وضعیت‌های «تحویل لوکال» در شرح آزمون بالا سابقه تحویل اول‌اند؛ هم‌اکنون همان بسته روی شاخه قابل بررسی است.

[صف یکتا و ممیزی۲۰۲۶-۱۱-۲۴؛ وضعیت واقعی، تسک‌های تازه و جلوگیری از اجرای تکراری](../backlog-review-2026-11-24/README.md) · [والد گیت #95](https://github.com/MohamadParsa/marketplace/issues/95).
