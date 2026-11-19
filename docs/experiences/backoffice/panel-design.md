# طرح بک‌آفیس و پنل فروشنده

وضعیت طرح `in-review` است. این راهنما به طرح نسخه ۰.۲ بسته بازبینی و مشخصات جاری پنل‌ها پیوند می‌دهد. شناسه‌های BO در جدول شناسه صفحه داخل بسته‌اند. رودمپ و گزارش نسخه بررسی نظام شماره‌گذاری جدا دارند و برای تطبیق باید از نام صفحه و شناسه قابلیت استفاده کرد. [راهنمای بسته](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/README-fa.md) · [رودمپ](../../delivery/roadmap.md) · [گزارش اجرا](../../delivery/stage-1-progress.md)

## چیدمان و حالت‌های مشترک

- طرح سه جهت دارد. A متعادل با کارت روشن و تراکم متوسط، B فروش‌محور با ردیف‌های متراکم‌تر، و C تصویری با زمینه گرم و فاصله بیشتر است. طراح A را پیشنهاد کرده و انتخاب طرح پنل باز است. نسخه ۰.۲ از لوگوی چریوا، رنگ گیلاسی و وزیرمتن استفاده می‌کند. [مقایسه تصویری](../../review-packages/cherriva-2026-11-17-v1-1/output/playwright/backoffice-three-directions.png) · [Q-028](../../product/open-questions.md#q-028)
- منوی راست بخش‌ها را بر اساس کار گروه‌بندی می‌کند. بالای صفحه مسیر جاری، جست‌وجوی بخش‌ها و حساب دیده می‌شود. در موبایل، منو کشویی و فرم جزئیات تمام‌عرض است و جدول داخل قاب خودش پیمایش می‌شود.
- الگوی پیشنهادی فهرست شامل عنوان، اقدام اصلی، فیلتر، جدول و شمارش نتیجه است. جزئیات در پنل کشویی باز می‌شود تا زمینه فهرست حفظ شود. حالت خالی، دریافت، خطای شبکه و عدم دسترسی پیام‌های جدا دارند. تعارض نسخه و حفظ تغییر ذخیره‌نشده در مشخصات تحویل برای بررسی آمده‌اند. [راهنمای طراحی](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/design-guide-fa.md) · [مرز تحویل](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/implementation-handoff-fa.md)
- پنل فروشنده فرم‌های مرتبط را با داده فروشگاه خودش می‌بیند. مجوز عملیات و محدوده داده در سامانه اعمال می‌شود. ورود پیامکی، انتخاب نقش و فرم رمز در این پیش‌نمایش با داده ساختگی شبیه‌سازی می‌شوند. وضعیت نقش سفارشی و هویت در دفتر پرسش‌ها آمده است. [نقش‌ها](../../product/roles.md) · [اختلاف‌های پنل](../../product/open-questions.md#اختلافهای-بسته-بازبینی-پنل)

## نقشه ۲۷ صفحه و ورود

شرح هر صفحه فیلدها، مسیر، دسترسی و تسک بسته را دارد. تصاویر A/B/C برای مقایسه در فهرست تصاویر آمده‌اند. [فهرست تصاویر و فرم‌ها](../../review-packages/cherriva-2026-11-17-v1-1/IMAGE-INDEX-fa.md) · [نقشه مبدأ و تسک‌ها](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screen-index-fa.md)

| صفحه در بسته | موضوع مرور | مرجع در مستندات اصلی |
| --- | --- | --- |
| ورود | شماره، کد، تغییر شماره و حالت خطا | [هویت](../../domains/identity/README.md) |
| BO-01 | [نمای کلی بازارگاه](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-01-dashboard.md) | [نقش‌ها](../../product/roles.md) |
| BO-02 | [کاربران](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-02-users.md) | [نقش‌ها](../../product/roles.md) · [هویت](../../domains/identity/README.md) |
| BO-03 | [نقش‌ها و دسترسی‌ها](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-03-roles.md) | [نقش‌ها](../../product/roles.md) |
| BO-04 | [فروشگاه‌ها](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-04-stores.md) | [FEAT-BO-001](management.md) |
| BO-05 | [دسته‌بندی‌ها](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-05-categories.md) | [FEAT-BO-001](management.md) |
| BO-06 | [برندها](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-06-brands.md) | [FEAT-BO-001](management.md) |
| BO-07 | [خصوصیات و تنوع‌ها](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-07-attributes.md) | [FEAT-BO-001](management.md) · [FEAT-SP-001](../seller-portal/catalog.md) |
| BO-08 | [محصولات و محتوا](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-08-products.md) | [FEAT-BO-001](management.md) · [FEAT-SP-001](../seller-portal/catalog.md) |
| BO-09 | [عرضه و موجودی](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-09-offers.md) | [FEAT-BO-001](management.md) · [FEAT-SP-001](../seller-portal/catalog.md) |
| BO-10 | [نرخ پایه طلا](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-10-gold-rates.md) | [قیمت‌گذاری](../../domains/pricing/README.md) · [پرسش‌های طرح](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/open-decisions-fa.md) |
| BO-11 | [مدل قیمت‌گذاری طلا](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-11-gold-policies.md) | [قیمت‌گذاری](../../domains/pricing/README.md) · [پرسش‌های طرح](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/open-decisions-fa.md) |
| BO-12 | [روش‌های فروش](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-12-sale-methods.md) | [قیمت‌گذاری](../../domains/pricing/README.md) · [پرسش‌های طرح](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/open-decisions-fa.md) |
| BO-13 | [مارکاپ دسته‌بندی](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-13-category-markups.md) | [قیمت‌گذاری](../../domains/pricing/README.md) · [پرسش‌های طرح](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/open-decisions-fa.md) |
| BO-14 | [قواعد کمیسیون](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-14-commissions.md) | [FEAT-BO-001](management.md) · [قیمت‌گذاری](../../domains/pricing/README.md) |
| BO-15 | [تخفیف‌ها و کدها](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-15-promotions.md) | [پرسش‌های طرح](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/open-decisions-fa.md) |
| BO-16 | [صفحه‌ها و سکشن‌ها](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-16-pages.md) | [FEAT-BO-001](management.md) · [تجربه سایت](../storefront/discovery.md) |
| BO-17 | [حمل و تقویم](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-17-shipping.md) | [FEAT-BO-001](management.md) · [FEAT-SP-002](../seller-portal/fulfillment.md) |
| BO-18 | [خریدها و سفارش‌ها](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-18-orders.md) | [FEAT-BO-002](support.md) · [FEAT-SP-002](../seller-portal/fulfillment.md) |
| BO-19 | [مرسوله‌ها](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-19-shipments.md) | [FEAT-BO-002](support.md) · [FEAT-SP-002](../seller-portal/fulfillment.md) |
| BO-20 | [پشتیبانی و پرونده‌ها](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-20-support.md) | [FEAT-BO-002](support.md) |
| BO-21 | [تأمین جایگزین](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-21-assignments.md) | [FEAT-BO-002](support.md) · [FEAT-SP-002](../seller-portal/fulfillment.md) |
| BO-22 | [عملیات مالی](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-22-finance.md) | [FEAT-BO-003](finance.md) |
| BO-23 | [تسویه فروشندگان](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-23-settlements.md) | [FEAT-BO-003](finance.md) |
| BO-24 | [پیگیری و گزارش](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-24-operations.md) | [FEAT-BO-002](support.md) · [FEAT-BO-003](finance.md) |
| BO-25 | [نظرها و پرسش‌ها](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-25-moderation.md) | [پرسش‌های طرح](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/open-decisions-fa.md) |
| BO-26 | [تنظیمات ویترین](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-26-settings.md) | [تجربه سایت](../storefront/discovery.md) · [پرسش‌های طرح](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/open-decisions-fa.md) |
| BO-27 | [پروفایل همان کاربر](../../review-packages/cherriva-2026-11-17-v1-1/02-backoffice/screens/bo-27-my-profile.md) | [هویت](../../domains/identity/README.md) |

## شواهد پیش‌نمایش

گزارش بسته مرور صفحه‌ها را در عرض ۳۹۰ و ۱۴۴۰ و کنترل‌های تعامل نمونه ثبت کرده است. تغییرهای نمونه در حافظه مرورگرند. پیشرفت رابط متصل به API، وضعیت استقرار و آزمون‌های آن در گزارش اجرای مرحله ۱ پیگیری می‌شوند. [گزارش کنترل بسته](../../review-packages/cherriva-2026-11-17-v1-1/VERIFICATION-fa.md) · [گزارش اجرای مرحله ۱](../../delivery/stage-1-progress.md)
