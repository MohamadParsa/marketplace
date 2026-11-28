# صف اصلاحات ورود، سفارش و مرسوله — ۲۰۲۶-۱۱-۲۸

[اصلاح بازبینی بعدی مالک و بررسی پشتیبانی، تأمین جایگزین و مالی — #150](storefront-presentation-and-legacy-menu-review.md). هشت تحویل اولیه بسته می‌مانند؛ بازخورد تازه مستقل ثبت شد.

مالک: @eliassarband · وضعیت اسناد: `in-review` · وضعیت اجرا: `done` · `execution_authorized: true` · والد صف: [#95](https://github.com/MohamadParsa/marketplace/issues/95).

برنامه اولیه پاسخ به «اینا را تسک کن» بود؛ مالک سپس اجرای هر۸تسک را صریحاً انتخاب کرد. نیازمندی‌های روشن در [قواعد مشترک](../../domains/orders/experience-followup-2026-11-28.md) و جزئیات مستقل در شرح‌ها ثبت می‌شوند. #131 تا #141 تحویل‌های قبلی و بسته می‌مانند؛ این صف اصلاحات حاصل بازبینی مالک است. نتیجه اجرای انتخاب‌شده در گزارش‌های شواهد هر Issue و جدول تحویل نهایی ثبت است.

## نتیجه بررسی سورس و هم‌پوشانی

مبنای بررسی: `a8d0d998b730db02900a3493fecc2c2e8b104cf6` روی `codex/local-parsa-integration-20261124`، PR [#106](https://github.com/MohamadParsa/marketplace/pull/106). بررسی read-only بوده و بازتولید runtime یا اصلاح سفارش واقعی ادعا نمی‌شود.

| موضوع | وضعیت فعلی و شکاف |
| --- | --- |
| ورود | login-page.tsx مرحله identifier و method جدا دارد؛ ادامه حذف و دو روش مستقیم عرضه شوند. |
| نشست | SessionApp در app.tsx با نشست null شروع می‌شود و restore ندارد؛ Authenticate سرور expiry/surface/active/access-version را کنترل می‌کند. روش امن بازیابی باید پیش از کد انتخاب شود. |
| سفارش | order-workspace.tsx و order_workspace.go فهرست سروری دارند؛ فیلتر موبایل مستقل و عمل انتهای ردیف ندارند. |
| کیف پول | commerce_checkout.go درخواست بیش از q.Total را رد می‌کند؛ سقف نیاز قبل از اعتبارسنجی کفایت اعتبار اعمال شود. |
| ثبت و قطعیت | AcceptQuote و ConfirmPurchase دو مرحله‌اند؛ مسیر بدون مانده در ثبت خودکار شود و کنترل‌های مالی موجود بازاستفاده شوند. |
| مرسوله | عملیات scoped قبلاً موجود است؛ فهرست مستقل برای هر دو surface و ورود پس از شروع همان گروه لازم است. |
| حساب مشتری | order-shipments.tsx و order-history-panel.tsx مبنای نمایش موجودند؛ جزئیات مالی روشن و گراف مرحله‌ها توسعه داده شوند. |

عنوان و شرح Issueهای موجود بررسی شدند؛ برای این هشت خروجی Issue بازِ هم‌دامنه پیدا نشد. #71 ممیزی گسترده دسترسی فروشندگان است؛ کنترل محدوده همین صف الزامی است و جای آن ممیزی را نمی‌گیرد. #75/#77 تاریخچه جامع/لاگ فنی و #82 اعلان ایمیل مالک جدا دارند؛ این صف تنها تاریخچه و کنترل لازم سفارش را استفاده می‌کند. پرداخت/عملیات #134 تا #140 بازنویسی کامل نمی‌شوند.

## ترتیب پیشنهادی

| ترتیب / Issue | خروجی و شرح | وابستگی |
| --- | --- | --- |
| [CV-ORD-FOLLOW-01 / #142](https://github.com/MohamadParsa/marketplace/issues/142) | [ورود مستقیم با رمز یا OTP پس از شماره موبایل](tasks/direct-login-methods.md) | مستقل |
| [CV-ORD-FOLLOW-02 / #143](https://github.com/MohamadParsa/marketplace/issues/143) | [حفظ نشست معتبر بک‌آفیس پس از رفرش و F5](tasks/restore-valid-backoffice-session.md) | مستقل |
| [CV-ORD-FOLLOW-03 / #144](https://github.com/MohamadParsa/marketplace/issues/144) | [فیلتر موبایل مشتری و دکمه مشاهده در انتهای ردیف سفارش](tasks/orders-mobile-filter-view-action.md) | مستقل |
| [CV-ORD-FOLLOW-04 / #145](https://github.com/MohamadParsa/marketplace/issues/145) | [اعمال خودکار کیف پول فقط تا مبلغ لازم سفارش](tasks/cap-applied-wallet-credit.md) | مستقل |
| [CV-ORD-FOLLOW-05 / #146](https://github.com/MohamadParsa/marketplace/issues/146) | [ثبت خودکار پرداخت سفارش با مانده پرداخت صفر](tasks/automatic-covered-order-payment.md) | CV-ORD-FOLLOW-04 |
| [CV-ORD-FOLLOW-06 / #147](https://github.com/MohamadParsa/marketplace/issues/147) | [منوی مرسوله‌ها با صف پردازش و دسترسی ادمین و فروشگاه](tasks/scoped-shipments-workspace.md) | مستقل |
| [CV-ORD-FOLLOW-07 / #148](https://github.com/MohamadParsa/marketplace/issues/148) | [ارتفاع یکدست توضیح عملیات و راهنمای کوتاه مراحل مرسوله](tasks/shipment-operation-form-guidance.md) | مستقل |
| [CV-ORD-FOLLOW-08 / #149](https://github.com/MohamadParsa/marketplace/issues/149) | [جزئیات سفارش مشتری با تفکیک مالی و روند گرافیکی مرسوله‌ها](tasks/customer-order-detail-progress.md) | مستقل |

مالک اجرای همین ترتیب را انتخاب کرد. 01،02،03 و04 مستقل‌اند؛ 05 به04 وابسته است. 06 از قرارداد و عملیات قبلی استفاده می‌کند و برای اثبات کامل سناریوی paid خودکار با05 تست یکپارچه می‌خواهد. 07 در فرم فعلی مستقل است و در صف06 نیز استفاده شود. 08 از APIهای موجود قابل طراحی است و پذیرش یکپارچه نهایی آن خروجی05 تا07 را نیز پوشش دهد.

## آمادگی و پرسش مؤثر

دامنه، مالک، مجوز، ورودی/خطا، پذیرش و شاهد بستن هر تسک مشخص‌اند. [معیار آمادگی](../readiness.md) برای هر برش انتخاب‌شده بررسی شد؛ وضعیت in-review اسناد از تکمیل اجرا جدا می‌ماند. طراحی نشست همان تب با اعتبارسنجی سرور و بازاستفاده از قطعیت مالی در شواهد #143/#146 ثبت شده‌اند. [ORDERFOLLOW-Q01](../../product/open-questions.md#order-experience-followup-20261128) با پاسخ صریح مالک حل و در #147 اجرا شد. جزئیات پیشنهادها در سند قواعد از تصمیم‌های صریح جدا هستند.

## پذیرش کلی و شرط پایان

- تست متناسب هر تغییر و سناریوهای منفی مجوز، مالکیت و هم‌زمانی لازم است. مسیر مالی باید با PostgreSQL و شکست ذخیره/پاسخ نامعلوم اثبات شود؛ UI صرفاً از رنگ یا متن ساختگی نتیجه نسازد.
- کنترل‌های تاریخ شمسی و ارقام فارسی، فیلتر بسته اولیه، جست‌وجوی primary، label یکسان و RTL در همه فرم‌های تازه حفظ شوند.
- هر Issue فقط بعد از کد و تست موفق، شواهد، هماهنگی مستندات، commit/push و کامنت مناسب بسته شود؛ بستن هر Issue در نتیجه اجرای خودش و با شواهد CI انجام می‌شود.
- [فهرست ماشین‌خوان](task-index.json)، [BDD طراحی](../../bdd/features/experiences/order-experience-followup.feature) و [ردیابی](../../governance/traceability.md#order-experience-followup-20261128) مرجع ارتباط‌اند. BDD در این نوبت اجرا نشده است.

## بررسی تحویل برنامه‌ریزی

اعتبارسنجی انجام شد: ۵۳۰ مسیر نسبی در۱۸فایل مرتبط بررسی شدند و لینک ناموجود نداشتند؛۹شناسه جدید مشخصات در کل اسناد یکتا هستند؛۲۶معیار پذیرش با۲۶سناریوی طراحی و۸ردیف task-index تطبیق دارند. وضعیت باز، انتساب eliassarband و execution_authorized=false برای هر۸Issue از GitHub دوباره خوانده و تأیید شد؛ #95 با حفظ صف قبلی تکمیل شد. `git diff --check` موفق است.

تمام تغییرات مخزن در docs هستند. کد محصول، سرویس محلی و داده کسب‌وکار این تحویل تغییر نکردند؛ تست/ساخت محصول اجرا نشد و سناریوهای BDD صرفاً طراحی هستند.

## شروع اجرا

مالک اجرای هر۸تسک، تست و تکمیل گیت و بستن با شواهد را انتخاب کرد. ORDERFOLLOW-Q01 با انتخاب صریح پیشنهاد پاسخ گرفت: لغوشده پس از شروع در تاریخچه و فیلتر لغوشده می‌ماند و صف پیش‌فرض فقط فعال‌هاست. گزارش برنامه‌ریزی بالا تاریخی است؛ پیشرفت جاری در task-index و گزارش مستقل هر Issue ثبت می‌شود.

## نتیجه اجرا

هر۸کار با تست و push جداگانه و هر۵check موفق CI، با کامنت شواهد بسته شدند. [تحویل نهایی و اجرای محلی](final-local-review.md).

| Issue و شواهد | کامیت | کامنت بستن / CI |
| --- | --- | --- |
| [#142](issue-142-verification.md) | [4049af9](https://github.com/MohamadParsa/marketplace/commit/4049af9ab74e32a0dd741336d1e25a168776d7fe) | [کامنت](https://github.com/MohamadParsa/marketplace/issues/142#issuecomment-7235847266) · [CI موفق](https://github.com/MohamadParsa/marketplace/actions/runs/46120293895/job/140903109206) |
| [#143](issue-143-verification.md) | [c1e5f9e](https://github.com/MohamadParsa/marketplace/commit/c1e5f9ec9b45581ac5effab2d7d2b418432a3c1b) | [کامنت](https://github.com/MohamadParsa/marketplace/issues/143#issuecomment-7235895777) · [CI موفق](https://github.com/MohamadParsa/marketplace/actions/runs/46120618838/job/140904088124) |
| [#144](issue-144-verification.md) | [c321435](https://github.com/MohamadParsa/marketplace/commit/c321435d746e41a17add984ca07045e3f02bb021) | [کامنت](https://github.com/MohamadParsa/marketplace/issues/144#issuecomment-7235967160) · [CI موفق](https://github.com/MohamadParsa/marketplace/actions/runs/46120916826/job/140905001954) |
| [#145](issue-145-verification.md) | [ac5ec92](https://github.com/MohamadParsa/marketplace/commit/ac5ec92fa14d5098121386bdbd4b4436e5c1c18e) | [کامنت](https://github.com/MohamadParsa/marketplace/issues/145#issuecomment-7235967780) · [CI موفق](https://github.com/MohamadParsa/marketplace/actions/runs/46121071170/job/140905477576) |
| [#146](issue-146-verification.md) | [7aee4ed](https://github.com/MohamadParsa/marketplace/commit/7aee4ed4455ca2aee9900cea95ee79036afdebb3) | [کامنت](https://github.com/MohamadParsa/marketplace/issues/146#issuecomment-7236082941) · [CI موفق](https://github.com/MohamadParsa/marketplace/actions/runs/46121979106/job/140908260258) |
| [#147](issue-147-verification.md) | [d277959](https://github.com/MohamadParsa/marketplace/commit/d277959ea49d22a3a4276b31a0dac6ddedbd686d) | [کامنت](https://github.com/MohamadParsa/marketplace/issues/147#issuecomment-7236100818) · [CI موفق](https://github.com/MohamadParsa/marketplace/actions/runs/46122064767/job/140908518379) |
| [#148](issue-148-verification.md) | [c5d5ba2](https://github.com/MohamadParsa/marketplace/commit/c5d5ba26fc4a8a7522ec7bd1e8e163c8bcf41396) | [کامنت](https://github.com/MohamadParsa/marketplace/issues/148#issuecomment-7236158092) · [CI موفق](https://github.com/MohamadParsa/marketplace/actions/runs/46122361727/job/140909427173) |
| [#149](issue-149-verification.md) | [ec759db](https://github.com/MohamadParsa/marketplace/commit/ec759db393118c78fc4af356b14115633b70eb00) | [کامنت](https://github.com/MohamadParsa/marketplace/issues/149#issuecomment-7236252808) · [CI موفق](https://github.com/MohamadParsa/marketplace/actions/runs/46123223630/job/140912028637) |

برای #146، کامیت اصلی19b7d3f و اصلاح انتظار تست هم‌زمانی7aee4ed هر دو پوش شده‌اند؛ CI نهایی آن سبز است. وضعیت سند in-review، تکمیل اجرا و پذیرش دستی مالک سه موضوع مستقل‌اند. PR #106 باز است و merge/deploy نشده است.
