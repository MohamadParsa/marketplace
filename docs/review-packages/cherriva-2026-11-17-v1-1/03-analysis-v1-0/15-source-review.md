# برداشت از گوشی‌شاپ و ساده‌سازی برای چریوا

نسخه: ۱.۰ | تاریخ: ۲۰۲۶-۱۱-۱۷ | وضعیت: تحلیل و طراحی برای بازبینی تیم فنی | مالک بررسی: محصول و راهبر فنی؛ انتساب شخص هنوز انجام نشده

این سند نقشه ایده است، نه دستور انتقال کد یا ادعای بررسی کامل تمام گوشی‌شاپ. گزارش تاریخی در `D:\Marketplace\11-product-flows\00-source-review.md` موجود است؛ بعضی مسیرهای آن جابه‌جا شده‌اند. در ۲۰۲۶-۱۱-۱۷ وجود و محتوای فایل‌های جدول زیر به‌صورت read-only بررسی شد؛ هیچ build/اجرا/تغییر انجام نشد. برای خواندن این بسته به سورس دسترسی لازم نیست.

| شاهد فعلی در D:\GSP_Repos\gooshishop2022 | مشاهده | انتخاب پیشنهادی چریوا |
| --- | --- | --- |
| `src/Libraries/Nop.Core/Domain/HomePage/HomePageSection.cs` و `HomePageSectionType.cs` | سکشن فراتر از Home، widget zone، هدف برند/مسیر و نوع‌های متعدد | Page+Section با چند نوع روشن؛ targeting محدود و schema معتبر |
| `src/Libraries/Nop.Core/Domain/Catalog/GoldPricingProfile.cs` | پروفایل با vendor، سود و fee درصدی | فرمول نسخه‌دار کم‌جزء؛ عدد و مبنای مالیات مستقل و تصمیم‌گیری‌شده |
| `src/Libraries/Nop.Core/Domain/GoldPricing/GoldReferencePrice.cs` | نرخ نوع طلا، ریال، اعتبار و فعال‌بودن | نرخ مشترک نسخه‌دار با واحد/عیار مشخص |
| `src/Libraries/Nop.Core/Domain/Suppliers/SupplierGoldPrice.cs` | نرخ تأمین‌کننده با بازه؛ کامنت تومان | نرخ اختصاصی سلر؛ تبدیل واحد صریح؛ کامنت سورس قرارداد پول چریوا نیست |
| `src/Libraries/Nop.Services/Catalog/GspPriceCalculationService.cs` | روش فروش و تعدیل/تخفیف هر روش و override قیمت | یک pipeline کوچک نقدی→تخفیف→markup تأییدشده؛ جزئیات مالی باز جدا |
| `src/Libraries/Nop.Core/Domain/Discounts/Discount.cs` | درصد/ثابت، سقف، بازه، کوپن، تجمع، محدودیت و زیرشاخه | subset محدود با funder و تخصیص؛ کمپین/affiliate/شرط‌های پیچیده فعلاً لازم نیست |

در گزارش تاریخی، order/refund/پرداخت/SEO/landing/کَش/host و frontend نیز بررسی شده‌اند؛ شواهد آن مربوط به تاریخ همان گزارش است و در این نوبت تمام آن فایل‌ها دوباره خوانده نشده‌اند. محدودیت افزونه قدیمی یک provider حکم عمومی درباره API امروز آن provider نیست. پذیرش واقعی، کارمزد و قابلیت refund از قرارداد جدید استخراج شود.

## پیچیدگی‌هایی که وارد چریوا نمی‌شوند

چندین موتور موازی قیمت، widget zone و layout بی‌شمار، سازگاری با کد legacy، نوع‌های کمپین سازمانی فراوان، microservice/broker از روز اول، موتور جست‌وجوی خارجی و انبارداری بزرگ. به‌جای آن Product/Variant/Offer، quote واحد، scoped permission، adapter درگاه، دفتر مالی، مرسوله مستقل و سکشن typed داریم. ساده‌سازی، حذف idempotency/محدوده داده/ثبت مالی/guard موجودی نیست.
