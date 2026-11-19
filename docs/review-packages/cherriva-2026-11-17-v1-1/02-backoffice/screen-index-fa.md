# نقشه صفحه‌ها و فرم‌ها

شناسه: CVBO-MAP-01 | نسخه: ۰.۲ | تاریخ: ۲۰۲۶-۱۱-۱۷ | وضعیت: in-review | مالک تصمیم: الیاس؛ بازبین فنی پیشنهادی: راهبر فنی، انتساب شخص باز

۲۷ صفحه پنل به علاوه ورود/OTP. صفحه‌ها طراحی پیشنهادی‌اند؛ فیلدهای تحلیل که به جزئیات بیشتر نیاز دارند، در شرح صفحه و [مرز تحویل](implementation-handoff-fa.md) مشخص‌اند.

| شناسه | صفحه و شرح | مسیر نمونه | تسک‌های موجود تحلیل |
| --- | --- | --- | --- |
| BO-01 | [نمای کلی بازارگاه](screens/bo-01-dashboard.md) | `#dashboard` | CV-FND-07 |
| BO-02 | [کاربران](screens/bo-02-users.md) | `#users` | CV-IAM-04، CV-IAM-05 |
| BO-03 | [نقش‌ها و دسترسی‌ها](screens/bo-03-roles.md) | `#roles` | CV-IAM-03، CV-IAM-05 |
| BO-04 | [فروشگاه‌ها](screens/bo-04-stores.md) | `#stores` | CV-IAM-06 |
| BO-05 | [دسته‌بندی‌ها](screens/bo-05-categories.md) | `#categories` | CV-CAT-01 |
| BO-06 | [برندها](screens/bo-06-brands.md) | `#brands` | CV-CAT-03 |
| BO-07 | [خصوصیات و تنوع‌ها](screens/bo-07-attributes.md) | `#attributes` | CV-CAT-02، CV-CAT-08 |
| BO-08 | [محصولات و محتوا](screens/bo-08-products.md) | `#products` | CV-CAT-04، CV-CAT-07، CV-CAT-08 |
| BO-09 | [عرضه و موجودی](screens/bo-09-offers.md) | `#offers` | CV-CAT-05، CV-CAT-06 |
| BO-10 | [نرخ پایه طلا](screens/bo-10-gold-rates.md) | `#gold-rates` | CV-PRC-03 |
| BO-11 | [مدل قیمت‌گذاری طلا](screens/bo-11-gold-policies.md) | `#gold-policies` | CV-PRC-04 |
| BO-12 | [روش‌های فروش](screens/bo-12-sale-methods.md) | `#sale-methods` | CV-PRC-01، CV-PRC-06 |
| BO-13 | [مارکاپ دسته‌بندی](screens/bo-13-category-markups.md) | `#category-markups` | CV-PRC-02، CV-PRC-06 |
| BO-14 | [قواعد کمیسیون](screens/bo-14-commissions.md) | `#commissions` | CV-PRC-05 |
| BO-15 | [تخفیف‌ها و کدها](screens/bo-15-promotions.md) | `#promotions` | CV-DSC-01، CV-DSC-02، CV-DSC-05 |
| BO-16 | [صفحه‌ها و سکشن‌ها](screens/bo-16-pages.md) | `#pages` | CV-CMS-01، CV-CMS-02، CV-CMS-04، CV-CMS-05 |
| BO-17 | [حمل و تقویم](screens/bo-17-shipping.md) | `#shipping` | CV-OPS-05، CV-OPS-06 |
| BO-18 | [خریدها و سفارش‌ها](screens/bo-18-orders.md) | `#orders` | CV-OPS-01، CV-OPS-02 |
| BO-19 | [مرسوله‌ها](screens/bo-19-shipments.md) | `#shipments` | CV-OPS-07 |
| BO-20 | [پشتیبانی و پرونده‌ها](screens/bo-20-support.md) | `#support` | CV-OPS-03، CV-OPS-08، CV-OPS-09 |
| BO-21 | [تأمین جایگزین](screens/bo-21-assignments.md) | `#assignments` | CV-OPS-04 |
| BO-22 | [عملیات مالی](screens/bo-22-finance.md) | `#finance` | CV-FIN-01، CV-FIN-04، CV-FIN-05، CV-FIN-08 |
| BO-23 | [تسویه فروشندگان](screens/bo-23-settlements.md) | `#settlements` | CV-FIN-06، CV-FIN-08 |
| BO-24 | [پیگیری و گزارش](screens/bo-24-operations.md) | `#operations` | CV-FND-06، CV-FIN-07، CV-INT-04 |
| BO-25 | [نظرها و پرسش‌ها](screens/bo-25-moderation.md) | `#moderation` | CV-WEB-15، CV-WEB-16 |
| BO-26 | [تنظیمات ویترین](screens/bo-26-settings.md) | `#settings` | CV-CMS-03، CV-WEB-18 |

فهرست نقش‌ها فقط preset نمایشی است؛ برای جزئیات به [قواعد دسترسی](references/backoffice-analysis-fa.md) و CHQ-06 رجوع کنید.


## اصلاحات جاری نسخه ۰.۲

[مرجع اصلاحات ۱.۱](../04-current-spec/README-fa.md) در موارد مشخص‌شده بر متن و کپی مراجع ۱.۰ مقدم است. همان User برای خرید/پنل/عضویت، ورود OTP و رمز موبایل/ایمیل و پروفایل BO-27، آیکن دسته/برند، خصوصیت مشترک با binding مستقل، اکشن سریع تنوع/عرضه، فرم typed بنر/ریل دستی/sticky و وقفه فقط در فروشگاه لحاظ شدند. قیمت/موجودی همچنان Offer است. اعتبار یکپارچگی نشست یا backend در این نمونه ادعا نمی‌شود.

مرجع بررسی جاری: [گزارش کل بسته](../VERIFICATION-fa.md). تصاویر با نام صفحه در output/playwright در این نسخه مجدداً گرفته شدند؛ تصاویر مقایسه‌ای قدیمی با توضیح در فهرست کل، سابقه‌اند. ثبت این نوبت DEC-028 / ACT-046.

| شناسه تکمیلی | شرح | مسیر | تسک |
| --- | --- | --- | --- |
| BO-27 | [پروفایل همان کاربر](screens/bo-27-my-profile.md) | `#my-profile` | CV-R2-03 |
