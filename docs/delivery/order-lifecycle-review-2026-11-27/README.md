---
id: DEL-ORDER-LIFECYCLE-REVIEW-20261127
title: بررسی پرسش‌وپاسخ، تخفیف و چرخه سفارش و استرداد
status: in-review
owner: eliassarband
updated: 2026-11-28
related: [DOM-ORDER-LIFECYCLE-001, FEAT-NEXT-009]
approved_by: []
implementation_status: done
execution_authorized: true
---

# بررسی و اجرای ترتیبی چرخه سفارش

اجرای فنی #131 تا #141 با گزارش مستقل هر کار در بخش «پیشرفت اجرا» تکمیل شده است؛ آزمون دستی یک‌جای مالک و انتشار سرور، مرحله جدا هستند. بخش‌های بررسی سورس و برنامه‌ریزی زیر، سابقهٔ پیش از اجرا را نگه می‌دارند.

درخواست اولیه مالک: **کامل بررسی و تسک‌بندی شود تا بعداً اجرای آن انتخاب شود**. سپس مالک اجرای ترتیبی #131 تا #140 با تست، commit/push جدا و بستن تک‌به‌تک پس از شواهد را انتخاب کرد. آزمون یک‌جای دستی مالک پس از تحویل همه انجام می‌شود. کامیت برنامه‌ریزی `5fdbe8d` فقط مستندات بود. تحویل‌های اجرایی بعدی جدا گزارش می‌شوند؛ تغییر درگاه واقعی/محیط عمومی از این دستور نتیجه نمی‌شود.

[قواعد مشترک و تفکیک تصمیم/پیشنهاد](../../domains/orders/lifecycle-review-2026-11-27.md) · [پرسش‌های باز](../../product/open-questions.md#order-lifecycle-review-20261127) · [فهرست ماشین‌خوان](task-index.json) · [سناریوهای پذیرش مستند](../../bdd/features/experiences/order-lifecycle-review.feature)

## بازبینی اولیه بر اساس سورس — پیش از اجرا

بازبینی از HEAD مدارک `98a89f6f7fdf027d605001f1a87e4efc0b9cb867` و [آخرین سورس تحویل‌شده `2ac2da5`](https://github.com/MohamadParsa/marketplace/commit/2ac2da55a47d5cd444d251995734412c85f83ed7) است؛ ادعای تست یا پذیرش مرورگر تازه در این نوبت نداریم.

| بخش | یافته قابل مشاهده در کد | شکاف و کار لازم |
| --- | --- | --- |
| Q&A | [صفحه](../../../apps/backoffice-web/src/features/management/product-questions-page.tsx) بدون محصول فقط پیام انتخاب می‌دهد. [سرویس](../../../apps/api/internal/application/backoffice/product_qa.go) و [query PostgreSQL](../../../apps/api/internal/adapters/postgresrepo/product_qa.go) محصول‌محور و cursor بیست‌تایی‌اند. [نظرات](../../../apps/backoffice-web/src/features/management/product-reviews-page.tsx) فهرست سراسری با فیلتر pending/approved/rejected دارد. | حذف الزام محصول در UI کافی نیست؛ query/API سراسری با همان مجوز بررسی، فیلتر و صفحه‌بندی لازم است. endpoint عمومی PDP باید محصول‌محور بماند. |
| تخفیف | [کوپن](../../../apps/backoffice-web/src/features/commerce/coupons-page.tsx) از Dialog با `maxWidth="xl"` استفاده می‌کند؛ fieldsetها flex و `.coupon-choices` در [CSS](../../../apps/backoffice-web/src/styles.css) تمام‌عرض‌اند. دو انتخاب متفاوت هم label یکسان «نوع محدودیت» دارند. | اندازه/چیدمان تطبیقی، دو کنترل در عرض مناسب و label دقیق؛ تغییر موتور کوپن لازم نیست. مراد بررسی‌شده از «مودال تخفیف»، فرم کد تخفیف است. |
| دو منوی سفارش | [ناوبری](../../../apps/backoffice-web/src/app/app.tsx): `/commerce` با عنوان «سفارش و پرداخت سایت» و `/orders` با عنوان «سفارش‌ها». [CommercePage](../../../apps/backoffice-web/src/features/commerce/commerce-page.tsx) خرید والد، واریز، بازپرداخت و audit را نشان می‌دهد؛ [OperationsPage](../../../apps/backoffice-web/src/features/operations/operations-page.tsx) سفارش‌های فروشگاهی و عملیات را. | دو نمای یک رابطه تجاری‌اند؛ تجربه یکپارچه با خرید والد/سفارش فروشگاه/مرسوله و تب‌های تفکیک‌شده پیشنهاد می‌شود، بدون حذف داده یا از بین بردن لینک قدیمی. |
| ثبت سفارش و مرسوله | `AcceptQuote` در [checkout](../../../apps/api/internal/application/backoffice/commerce_checkout.go) خرید شماره‌دار، رزرو موجودی/اعتبار/حمل و `syncPurchaseOrders` را ثبت می‌کند. snapshot مرسوله‌ها پیش از تأیید مالی موجود است؛ وضعیت خرید `awaiting-confirmation` و سفارش فروشگاهی `awaiting-payment` است. | پایه ایجاد قبل از پرداخت وجود دارد؛ نباید موتور ثبت دوم ساخت. مرز PlaceOrder نسبت به درگاه و زمان commit منابع نیازمند قرارداد و ممیزی #70 است. |
| پرداخت | `ReviewDeposit` در [commerce_admin](../../../apps/api/internal/application/backoffice/commerce_admin.go) واریز را به کیف پول شارژ می‌کند. `ConfirmPurchase` سپس کل مبلغ را برداشت، موجودی/ظرفیت را قطعی و سفارش‌های فروشگاهی را `paid/awaiting-supply` می‌کند. درگاه آنلاین در این مسیر فعلی فعال نیست. | مدل مستقل تلاش پرداخت، معنای «مجاز» و نسبت درگاه با کیف پول لازم است؛ شارژ و مصرف یک مرحله فرض نشوند. |
| وضعیت مستقل | `purchaseStatuses` در [commerce_base](../../../apps/api/internal/application/backoffice/commerce_base.go) با `ConfirmedAt` وضعیت نمایشی سفارش را همیشه `awaiting-supply` می‌سازد. [Purchase](../../../apps/api/internal/domain/commerce/model.go) رویداد دارد، ولی [Order](../../../apps/api/internal/domain/orders/order.go) `Status` عملیاتی و `Payment` دارد. | مدل سفارش ثبت/پردازش/تکمیل از عملیات جدا نشده و تجمیع تحویل/تکمیل خرید والد وجود ندارد. نگاشت داده قدیمی بدون جعل تاریخچه لازم است. |
| عملیات واقعی | [خواندن سفارش سلر](../../../apps/api/internal/application/backoffice/commerce_seller.go) scoped است. جست‌وجوی commandهای backend مسیر کامل تأمین/آماده‌سازی/ارسال سفارش checkout را نشان نداد؛ `order.deliver` در [support_commands](../../../apps/api/internal/application/backoffice/support_commands.go) سفارش دارای `CheckoutID` را صریحاً رد می‌کند. | وجود وضعیت یا دکمه «ثبت دریافت» شاهد قابل‌اجرا بودن چرخه سایت نیست؛ API و فرم عملیاتی واقعی با مجوز scoped، transition و تاریخچه لازم است. |
| کد ارسال | snapshot روش/وعده موجود؛ `Packages` در Order آرایه رشته‌ای است. نوع tracking/delivery-code و ثبت/اعتبارسنجی آن به عنوان سیاست روش در مدل فعلی نیست. | تنظیم نوع شناسه، snapshot در مرسوله، دریافت هنگام ارسال و نمایش مجاز؛ قواعد چندبسته/چندکد قبلی حفظ شوند. |
| کنسلی | `CancelPurchase` برای خرید confirmed فقط boolean درخواست و دلیل ثبت می‌کند. `AdminCancelPurchase` کل خرید را تنها پیش از ارسال لغو، موجودی/ظرفیت را آزاد و وجه را با حفظ منابع به کیف پول برمی‌گرداند. | درخواست دارای چرخه رسیدگی، قلم/تعداد/مرسوله، پذیرش/رد و مرجوعی پس از تحویل وجود ندارد؛ مسیر موجود حفظ و تکمیل شود. |
| بازپرداخت فعلی | `RequestRefund` در [commerce_wallet](../../../apps/api/internal/application/backoffice/commerce_wallet.go) مبلغی از اعتبار قابل‌برداشت را برای برداشت دستی رزرو می‌کند. `ResolveRefund` مرجع/تاریخ انتقال را ثبت و کیف پول را debit می‌کند. RefundRequest مرجع سفارش/اقلام ندارد. | این قابلیت برداشت کیف پول است؛ معادل استرداد مرتبط با سفارش/مرجوعی نیست. اتصال مالی با جلوگیری از دو بار پرداخت و حفظ محدودیت منابع لازم است. مسیر قدیمی تستی `case.refund` نیز checkout را رد می‌کند. |
| تاریخچه و مشتری | رویدادهای پذیرش/پرداخت/درخواست لغو/لغو و audit مالی موجودند؛ مدل رویداد و projection کل چرخه عملیاتی کامل نیست. | تاریخچه واحد سفارش و رویدادهای قابل مصرف برای حساب مشتری/اعلان لازم است؛ لاگ فنی #77 و ActivityLog جامع #75 مالک‌های جدا هستند. |

## پیشنهاد تجربه بک‌آفیس

یک ورودی «سفارش‌ها» با فیلترهای مستقل وضعیت سفارش، پرداخت و عملیات؛ در جزئیات، سربرگ خرید و بخش فروشگاه‌های آن، با تب‌های «خلاصه و اقلام»، «پرداخت‌ها»، «عملیات و مرسوله‌ها»، «درخواست‌ها و استرداد» و «تاریخچه». عملیات از خود سفارش قابل انجام باشد. واریز مستقل کیف پول، سفارش نیست و صف رسیدگی مالی مستقلش باید باقی بماند؛ لینک از تب پرداخت به مرجع واریز مرتبط بدهد.

این طراحی پیشنهادی است؛ حفظ دسترسی مالی و محدوده فروشگاه شرط آن است. درخت منو صرفاً یک‌بار سفارش را نمایش دهد، URLهای قبلی به نمای مناسب هدایت شوند و داده نمونه قدیمی با سفارش واقعی مخلوط و حذف نشود. فیلترها در لود اول بسته و جست‌وجو primary طبق [قرارداد گریدهای پنل](../../architecture/contracts/backoffice-grids.md) باشد.

## صف کار و ترتیب پیشنهادی

اجرای #131 تا #140 به همین ترتیب با دستور صریح مالک انتخاب شد؛ همه مسئول @eliassarband و `execution_authorized: true` هستند. وضعیت واقعی هر ردیف در فهرست ماشین‌خوان است. ده Issue تازه #131 تا #140 باز و منتسب شدند؛ #70 موجود برای ممیزی حفظ شد. شماره‌ها در فهرست ماشین‌خوان و خود شرح‌ها ثبت‌اند.

| کار | شرح و معیار اصلی | وابستگی/مانع |
| --- | --- | --- |
| [#131](https://github.com/MohamadParsa/marketplace/issues/131) — CV-ORDER-REV-01 | [پرسش‌های همه محصولات، فیلتر و moderation](tasks/global-product-questions.md) | مستقل، بدون نیاز به تصمیم مالی |
| [#132](https://github.com/MohamadParsa/marketplace/issues/132) — CV-ORDER-REV-02 | [مودال تخفیف جمع‌وجور و دو ستون](tasks/compact-coupon-modal.md) | مستقل؛ صرفاً UI |
| [#133](https://github.com/MohamadParsa/marketplace/issues/133) — CV-ORDER-REV-03 | [قرارداد سه چرخه و نگاشت سابقه](tasks/order-state-contract.md) | پاسخ ORDERREV-Q01 تا Q10 در مسیر مؤثر؛ قرارداد پیش از اصلاح داده |
| #70 موجود | [ممیزی اتمیک PlaceOrder](../next-phase/tasks/cv-next-09-atomic-place-order.md) | تحلیل موجود حفظ می‌شود؛ ورودی تغییرات تازه و failure matrix به آن افزوده شود، نه موتور موازی |
| [#134](https://github.com/MohamadParsa/marketplace/issues/134) — CV-ORDER-REV-04 | [ثبت قبل از پرداخت و حفظ مرسوله/رزرو](tasks/place-order-boundary.md) | REV-03 و خروجی ممیزی #70، Q01/Q03/Q06 |
| [#135](https://github.com/MohamadParsa/marketplace/issues/135) — CV-ORDER-REV-05 | [تلاش‌های پرداخت و تفکیک تأمین وجه/پرداخت](tasks/payment-attempt-states.md) | REV-03/04 و Q01/Q02/Q06؛ انتخاب provider واقعی گیت جدا |
| [#136](https://github.com/MohamadParsa/marketplace/issues/136) — CV-ORDER-REV-06 | [تأمین، آماده‌سازی، ارسال کددار و تحویل](tasks/order-fulfillment.md) | REV-03/04/05 و Q03/Q07/Q10؛ عدم تأمین/تقویم/ظرفیت موجود حفظ |
| [#137](https://github.com/MohamadParsa/marketplace/issues/137) — CV-ORDER-REV-07 | [پرونده کنسلی و مرجوعی روی سفارش](tasks/order-cancellation-returns.md) | REV-03، Q03/Q04/Q05؛ ثبت درخواست از اجرای مالی جدا |
| [#138](https://github.com/MohamadParsa/marketplace/issues/138) — CV-ORDER-REV-08 | [استرداد سفارش به کیف پول یا انتقال دستی](tasks/order-linked-refunds.md) | REV-07، Q08/Q09؛ مالی/کمیسیون #72 در قسمت وابسته |
| [#139](https://github.com/MohamadParsa/marketplace/issues/139) — CV-ORDER-REV-09 | [تاریخچه چرخه و نمایش وضعیت مشتری](tasks/order-history-customer-status.md) | قرارداد رویداد REV-03 و خروجی REV-04 تا08؛ اتصال پیام واقعی جدا |
| [#140](https://github.com/MohamadParsa/marketplace/issues/140) — CV-ORDER-REV-10 | [ورودی واحد سفارش و جزئیات تب‌دار](tasks/unified-order-workspace.md) | طراحی پوسته بعد REV-03 ممکن؛ پذیرش نهایی به REV-04 تا09 وابسته |

01 و02 می‌توانند مستقل انتخاب شوند. برای چرخه سفارش ابتدا 03 و #70، سپس 04 →05 →06، بعد 07 →08 و تکمیل09/10 پیشنهاد می‌شود. قسمت UI بدون تصمیم مالی نباید با رفتار مالی حدسی تکمیل اعلام شود.

## عدم هم‌پوشانی و شرط بستن

- #125 بسته می‌ماند؛ درخواست تازه Q&A تکمیل مستقل آن است. #67/#68 موتور و قرارداد کوپن‌اند؛ REV-02 فقط چیدمان است.
- #70 ممیزی عمومی تراکنش است؛ REV-04 فقط اجرای شکاف اثبات‌شده و تغییر زمان/مرز ثبت پس از تصمیم. این بررسی جای اثبات همه سناریوهای #70 نیست و آن را نمی‌بندد.
- #71 ممیزی جامع سلر و #75 ActivityLog عمومی، #77 لاگ فنی، #82 ایمیل مستقل می‌مانند؛ task چرخه سفارش مجوز ساخت دوباره آن سرویس‌ها نیست.
- #10 و #16 و #120 تا #123 تحویل‌های قبلی‌اند؛ فقط regression و اتصال تازه نیاز است. CHQ-01/#72 برای قیمت/تسویه و Q-005 برای بازگشت فیزیکی مرجع حفظ می‌شوند. Q-020 مرحله اول قبلاً حل‌شده است و دوباره پرسش باز اعلام نمی‌شود.
- بستن هر task اجرایی بعد از انتخاب صریح اجرا، پاسخ موانع مؤثر، کد، تست متناسب، شواهد، مدارک، commit/push و کامنت نتیجه مجاز است. در تحویل اولیه برنامه‌ریزی، هیچ task اجرایی انجام‌شده یا بسته نبود؛ پیشرفت بعدی پایین آمده و #95 والد صف باز می‌ماند.

## بررسی این تحویل

اعتبارسنجی انجام‌شده: ۶۴۲ مسیر نسبی در۲۵فایل مرتبط بررسی شد و لینک ناموجود نداشت؛۱۲شناسه metadata جدید در کل اسناد یکتا هستند؛۱۰شرح با task-index و۱۰سناریوی طراحی BDD تطبیق دارند. وضعیت باز، مسئول eliassarband و عبارت عدم مجوز اجرا برای هر۱۰Issue از GitHub دوباره خوانده و تأیید شد. `git diff --check` نیز موفق بود.

تست محصول، تماس درگاه، migration و تغییر سرویس محلی برای این برنامه اجرا نشد. نتیجه فنی جدول بالا بررسی سورس است و با پذیرش runtime یکی نیست. سناریوهای BDD مستندند و ادعای اجرای runner ندارند. همه تغییرات این تحویل در مسیر `docs/` هستند.


## پیشرفت اجرا

- #131: [فهرست سراسری پرسش‌ها، آزمون API/PG و مرورگر](issue-131-verification.md) با کامیت `21ebcd5` و هر۵بررسی CI موفق تکمیل و با [کامنت شواهد](https://github.com/MohamadParsa/marketplace/issues/131#issuecomment-7227102095) بسته شد.

- #132: [چیدمان دو ستونه تخفیف، اعتبارسنجی و تصاویر](issue-132-verification.md) با کامیت `ace1d0d` و هر۵بررسی CI موفق تکمیل و با [کامنت شواهد](https://github.com/MohamadParsa/marketplace/issues/132#issuecomment-7227314469) بسته شد.

- #133: [قرارداد اجرایی](../../architecture/contracts/order-lifecycle.md)، [ممیزی مرز تراکنش](place-order-audit.md) و [شواهد بررسی سند](issue-133-verification.md) آماده و تکمیل فنی شدند؛ #134 تا #140 هنوز اجراشده محسوب نمی‌شوند.

- #133 با کامیت `b73e752` و هر۵بررسی CI موفق و [کامنت شواهد](https://github.com/MohamadParsa/marketplace/issues/133#issuecomment-7227405651) بسته شد.
- #134: [ثبت پیش از تأمین وجه، وضعیت مستقل و replay پس از restart](issue-134-verification.md) تکمیل فنی و آماده CI/بستن است.

- #134 با کامیت `cf46e0b` و هر۵بررسی CI موفق و [کامنت شواهد](https://github.com/MohamadParsa/marketplace/issues/134#issuecomment-7227527986) بسته شد. ممیزی #70 هم با [تعیین مرز تحلیل/childها](https://github.com/MohamadParsa/marketplace/issues/70#issuecomment-7227529080) بسته شد.
- #135: [تلاش‌های پرداخت، وجه دیررس و عدم قطعیت مجاز](issue-135-verification.md) تکمیل فنی شد؛ منتظر CI/ثبت کامنت بستن.

## تکمیل تازه مالک — ۲۰۲۶-۱۱-۲۸

[#141 — همه کنترل‌های تاریخ شمسی در storefront و بک‌آفیس](tasks/persian-date-controls.md) پس از #140 به صف اجرا اضافه شد؛ todo و مجوز اجرا دارد. پذیرش یک‌جای مالک پس از این تکمیل است.

- #135 با کامیت `26fdef0` و هر۵بررسی CI موفق و [کامنت شواهد](https://github.com/MohamadParsa/marketplace/issues/135#issuecomment-7227893028) بسته شد.
- #136: [پردازش تا ارسال کددار و دریافت مستقل، مجوز scoped و تست/تصاویر](issue-136-verification.md) تکمیل فنی و آماده CI/بستن است.

- #136: عملیات مستقل، کد ارسال و دریافت جدا با کامیت f813e5e و CI موفق بسته شد؛ [شواهد](issue-136-verification.md).
- #137: پرونده کنسلی/مرجوعی، اقلام/تعداد، رسیدگی و برگشت موجودی مستقل از مالی؛ [شواهد تست و مرورگر](issue-137-verification.md). بستن پس از push و CI.

- #138: استرداد جزئی/کامل با محدودیت منابع و انتقال دستی مسدودشده؛ [شواهد محاسبه و PG و مرورگر](issue-138-verification.md). بستن پس از push و CI.

- #139: مرسوله‌ها، دریافت مشتری و تاریخچه مجاز صفحه‌بندی‌شده؛ [شواهد](issue-139-verification.md). بستن پس از push و CI.

- #140: فهرست واحد با query سروری و صفحه مستقل پنج‌تب؛ [شواهد مجوز/API/PG/مرورگر](issue-140-verification.md). بستن پس از push و CI.

- #141: همه ورودی‌های تاریخ/تاریخ‌وساعت، تعطیلات و ظرفیت با انتخابگر مشترک شمسی؛ [ممیزی و شواهد API/مرورگر](issue-141-verification.md). بستن پس از push و CI.
