# مستندات BDD بازارگاه

این پوشه مرجع سناریوهای رفتاری بازارگاه است. سناریوها با قالب Gherkin فارسی در فایل‌های `.feature` ثبت شده‌اند و هر سناریو با شناسه پذیرش موجود در سند حوزه یا تجربه پیوند دارد. این اسناد در وضعیت مستندات هستند و به‌تنهایی مجوز ساخت یا اتصال به سرویس واقعی نیستند.

## قواعد نگارش و ردیابی

- هر فایل یک `Feature` و هر پذیرش یک `Scenario` با همان شناسه `DOM-*` یا `FEAT-*` دارد.
- `Given` فقط پیش‌شرط و داده آزمون را آماده می‌کند؛ `When` یک محرک قابل مشاهده است؛ `Then` نتیجه قابل مشاهده و قابل بررسی را بیان می‌کند.
- داده‌ها ساختگی‌اند و در [داده آزمون](test-data.md) تعریف می‌شوند؛ شماره، کلید، درگاه، پیامک و انتقال مالی واقعی وارد سناریو نمی‌شود.
- قواعد قطعی از سند حوزه/تجربه خوانده می‌شوند. طراحی پیشنهادی، پرسش باز و قرارداد انتخاب‌نشده نباید به‌عنوان نتیجه قطعی سناریو نوشته شود.
- هر سناریوی مسدودشده باید در سند مرجع به `Q-*` لینک شود. تا رفع مانع، نتیجه آن «قابل اجرا نیست» است، نه قبولی.
- اجرای آینده باید برای هر سناریو شواهدی شامل شناسه اجرا، داده ورودی، نتیجه و لینک گزارش ثبت کند؛ وضعیت سند با وضعیت اجرای آزمون یکی نیست.

## پوشش فعلی

در این نسخه ۲۱۸ سناریو ثبت شده است: ۱۶۶ سناریوی پذیرش پایه، ۱۱ سناریوی تصمیم/جریان ترکیبی،۳ سناریوی خصوصیات مشترک،۳ سناریوی تنوع/عرضه و۲ سناریوی بنر دو اندازه و۹ سناریوی مستند تکمیل محتوا/SEO و۲۴ سناریوی انتشار اولیه؛ اجرای فنی برش خانه در [گزارش #27](../delivery/initial-launch/issue-27-verification.md) و بقیه مراحل در کارهای خودشان ردیابی می‌شوند. اجرای مستقل برش‌ها در [گزارش Issue#5](../delivery/review-v1-1/issue-5-verification.md) و [گزارش Issue#6](../delivery/review-v1-1/issue-6-verification.md) ثبت می‌شود؛ runner Gherkin مستقلی اجرا نشده است.

| بخش | فایل | سناریو | مرجع |
| --- | --- | ---: | --- |
| انتشار اولیه چهار برش (CV-LCH-01 تا04) | [initial-launch.feature](features/experiences/initial-launch.feature) | 24 | [برنامه](../delivery/initial-launch/README.md) |
| محتوای دسته/محصول و SEO (CV-R2-16/17) | [catalog-content-seo.feature](features/experiences/catalog-content-seo.feature) | 9 | [مرجع](../domains/catalog/content-and-seo.md) |
| بنر دو اندازه (CV-R2-07) | [features/experiences/responsive-banners.feature](features/experiences/responsive-banners.feature) | 2 | [شرح کار](../delivery/review-v1-1/tasks/cv-r2-07-responsive-banner.md) |
| تنوع و عرضه محصول (CV-R2-06) | [features/experiences/product-workspace.feature](features/experiences/product-workspace.feature) | 3 | [شرح کار](../delivery/review-v1-1/tasks/cv-r2-06-product-workspace.md) |
| خصوصیات مشترک و کارت (CV-R2-05) | [features/experiences/shared-attributes.feature](features/experiences/shared-attributes.feature) | 3 | [شرح کار](../delivery/review-v1-1/tasks/cv-r2-05-shared-attributes.md) |
| کاتالوگ، فروشگاه و موجودی (DOM-CATALOG-001) | [`features/domains/catalog.feature`](features/domains/catalog.feature) | 15 | [`domains/catalog/README.md`](../domains/catalog/README.md) |
| سبد، رزرو موجودی و جلسه پرداخت (DOM-CHECKOUT-001) | [`features/domains/checkout.feature`](features/domains/checkout.feature) | 12 | [`domains/checkout/README.md`](../domains/checkout/README.md) |
| حساب و هویت (DOM-IDENTITY-001) | [`features/domains/identity.feature`](features/domains/identity.feature) | 11 | [`domains/identity/README.md`](../domains/identity/README.md) |
| اطلاع‌رسانی پیامکی (DOM-NOTIFICATIONS-001) | [`features/domains/notifications.feature`](features/domains/notifications.feature) | 5 | [`domains/notifications/README.md`](../domains/notifications/README.md) |
| سفارش، تأمین و رسیدگی به عدم تأمین (DOM-ORDERS-001) | [`features/domains/orders.feature`](features/domains/orders.feature) | 13 | [`domains/orders/README.md`](../domains/orders/README.md) |
| پرداخت، کیف پول، برداشت و بازپرداخت (DOM-PAYMENTS-001) | [`features/domains/payments.feature`](features/domains/payments.feature) | 15 | [`domains/payments/README.md`](../domains/payments/README.md) |
| قیمت‌گذاری و کمیسیون (DOM-PRICING-001) | [`features/domains/pricing.feature`](features/domains/pricing.feature) | 13 | [`domains/pricing/README.md`](../domains/pricing/README.md) |
| تسویه فروشنده (DOM-SETTLEMENTS-001) | [`features/domains/settlements.feature`](features/domains/settlements.feature) | 6 | [`domains/settlements/README.md`](../domains/settlements/README.md) |
| آماده‌سازی و ارسال (DOM-SHIPPING-001) | [`features/domains/shipping.feature`](features/domains/shipping.feature) | 19 | [`domains/shipping/README.md`](../domains/shipping/README.md) |
| عملیات مالی و تسویه (FEAT-BO-003) | [`features/experiences/backoffice-finance.feature`](features/experiences/backoffice-finance.feature) | 6 | [`experiences/backoffice/finance.md`](../experiences/backoffice/finance.md) |
| مدیریت فروشگاه، کاتالوگ و سیاست‌ها (FEAT-BO-001) | [`features/experiences/backoffice-management.feature`](features/experiences/backoffice-management.feature) | 8 | [`experiences/backoffice/management.md`](../experiences/backoffice/management.md) |
| رسیدگی عملیاتی سفارش (FEAT-BO-002) | [`features/experiences/backoffice-support.feature`](features/experiences/backoffice-support.feature) | 8 | [`experiences/backoffice/support.md`](../experiences/backoffice/support.md) |
| مدیریت عرضه فروشگاه (FEAT-SP-001) | [`features/experiences/seller-portal-catalog.feature`](features/experiences/seller-portal-catalog.feature) | 7 | [`experiences/seller-portal/catalog.md`](../experiences/seller-portal/catalog.md) |
| تأیید تأمین و ارسال فروشنده (FEAT-SP-002) | [`features/experiences/seller-portal-fulfillment.feature`](features/experiences/seller-portal-fulfillment.feature) | 5 | [`experiences/seller-portal/fulfillment.md`](../experiences/seller-portal/fulfillment.md) |
| سبد تا نتیجه پرداخت (FEAT-SF-002) | [`features/experiences/storefront-checkout.feature`](features/experiences/storefront-checkout.feature) | 5 | [`experiences/storefront/checkout.md`](../experiences/storefront/checkout.md) |
| کشف کالا و انتخاب فروشگاه (FEAT-SF-001) | [`features/experiences/storefront-discovery.feature`](features/experiences/storefront-discovery.feature) | 12 | [`experiences/storefront/discovery.md`](../experiences/storefront/discovery.md) |
| پیگیری خرید و لغو در عدم تأمین (FEAT-SF-003) | [`features/experiences/storefront-orders.feature`](features/experiences/storefront-orders.feature) | 5 | [`experiences/storefront/orders.md`](../experiences/storefront/orders.md) |
| کیف پول مشتری (FEAT-SF-004) | [`features/experiences/storefront-wallet.feature`](features/experiences/storefront-wallet.feature) | 4 | [`experiences/storefront/wallet.md`](../experiences/storefront/wallet.md) |
| جریان‌های ترکیبی بحرانی | [`features/cross-domain-critical-flows.feature`](features/cross-domain-critical-flows.feature) | 8 | [`../governance/traceability.md`](../governance/traceability.md) |

## قرارداد آماده‌سازی آزمون

پیش از اجرای هر فایل، داده پایه از `test-data.md` بارگذاری و شناسه محیط آزمون ثبت شود. هر سناریو باید در برابر داده‌های خالی، خطای اعتبارسنجی، رد مجوز، تکرار درخواست و نتیجه نامشخص، در صورت ارتباط با مرجع خود، مسیر جدا داشته باشد. اگر مرجع سناریو این حالت را پوشش نداده است، آن را با تصمیم حدسی به سناریو اضافه نکنید؛ خلأ را در `docs/product/open-questions.md` ثبت کنید.

## وضعیت و مالکیت

این خروجی بخشی از مستندات `in-review` است. مالک قواعد همان مالک سند مرجع است و مالک اجرای آینده باید هنگام آماده‌سازی برش تحویل تعیین شود. هیچ سناریویی در این پوشه `approved` یا `passed` تلقی نمی‌شود مگر اینکه مالک تصمیم و شواهد اجرای مستقل ثبت شده باشد.

اجرای برش بنر در [گزارش Issue#7](../delivery/review-v1-1/issue-7-verification.md) ثبت شده است؛ وجود فایل Gherkin به معنی اجرای runner مستقل نیست.


در گیت انتشار اولیه فقط دامنه ADR-007 و سناریوهای CV-LCH مربوط پذیرفته شوند؛ سناریوهای BNPL/برداشت/اتصال درگاه در محصول کامل شرط این گیت نیستند. سناریوی قدیمی BNPL مستقیم (DOM-PAYMENTS-001-AC03) پس از تصمیم جدید معوق و نیازمند بازنگری قرارداد است.

شواهد API/دیتابیس/مرورگر خرید در [گزارش #30](../delivery/initial-launch/issue-30-verification.md) ثبت شد؛ تعداد سناریوهای مستند۲۱۸ باقی است. تگ مسدود LCH-Q02/03/04 حذف شد؛ AC07 محیط عمومی همچنان نیازمند شاهد خارجی است.
