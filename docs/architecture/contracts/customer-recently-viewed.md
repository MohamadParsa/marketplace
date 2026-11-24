---
id: API-CUSTOMER-RECENTLY-VIEWED-001
title: سابقه خصوصی محصولات اخیراً دیده‌شده مشتری
status: in-review
owner: eliassarband
updated: 2026-11-23
related: [FEAT-NEXT-003, API-PRODUCT-DETAIL-001, API-CUSTOMER-WISHLIST-001]
approved_by: []
---

# قرارداد محلی #64

این قابلیت فهرست آخرین مشاهده محصول است؛ منبع رویداد خام Analytics یا شمارنده عمومی بازدید/پرفروش‌ترین نیست. [گزارش اجرا](../../delivery/next-phase/issue-64-verification.md) شواهد و مرز انتشار را ثبت می‌کند. تغییرات محلی‌اند؛ وضعیت سند از done محلی و وضعیت GitHub جداست.

## قواعد تأییدشده مالک

| شناسه | قاعده |
| --- | --- |
| API-CUSTOMER-RECENTLY-VIEWED-001-R01 | فقط نمایش موفق PDP برای کاربر واردشده ثبت شود؛ مهمان سابقه محلی یا ادغام بعد از ورود ندارد. |
| API-CUSTOMER-RECENTLY-VIEWED-001-R02 | حداکثر۵۰محصول متفاوت از۳۰روز اخیر؛ بازدید تکراری رکورد جدید نسازد و محصول را به ابتدا بیاورد. |
| API-CUSTOMER-RECENTLY-VIEWED-001-R03 | حذف تکی و پاک‌کردن کل سابقه همان کاربر ممکن باشد. |
| API-CUSTOMER-RECENTLY-VIEWED-001-R04 | صفحه «بازدیدهای اخیر» در حساب و منوی پروفایل؛ در این نسخه ریل PDP یا آیکن اضافه هدر لازم نیست. |
| API-CUSTOMER-RECENTLY-VIEWED-001-R05 | محصول پنهان/حذف‌شده همانند علاقه‌مندی یک ردیف عمومی قابل حذف، بدون نام/عکس/قیمت/لینک دارد؛ ناموجود عمومی با مشخصات مجاز و بدون قیمت قابل خرید باقی است. |

منبع، درخواست اولیه و سه پاسخ صریح مالک در نوبت #64 است. سیاست wishlist در [قرارداد خودش](customer-wishlist.md) می‌ماند؛ پاک‌کردن history آن را حذف نمی‌کند.

## مدل و ثبت بازدید

[RecentlyViewedEntry](../../../apps/api/internal/domain/customer/recently_viewed.go) شامل user_id، product_id و viewed_at سمت سرور است. همان payload پایدار Workspace، با فیلد مستقل recently_viewed، استفاده می‌شود؛ جدول/وابستگی/migration تازه ندارد. کلید منطقی user/product است و تنوع، فروشنده، تعداد refresh یا انتخاب رنگ ردیف اضافه نمی‌سازند. نام/تصویر/قیمت قدیمی snapshot نمی‌شوند.

[Service](../../../apps/api/internal/application/backoffice/recently_viewed.go) در همان Update سریال PG/mutex حافظه، active/access_version و surface customer را دوباره کنترل می‌کند. افزودن فقط برای محصول عمومی طبق publicProduct مجاز است، از جمله محصول عمومی ناموجود. آخرین رکورد همان user/product جایگزین، viewed_at از ساعت سرور و مقدار جدید در ابتدا قرار می‌گیرد. مرتب‌سازی پایدار نزولی timestamp است؛ در ساعت برابر، ترتیب آرایه پایدار (آخرین مشاهده commitشده در ابتدا) حفظ می‌شود. این تساوی نیز مستقل تست شده است.

[hook PDP](../../../apps/storefront-web/src/recently-viewed.tsx) فقط پس از پایان موفق دریافت/نمایش محصول و وجود token فراخوانی می‌شود. فهرست/خانه، GET عمومی یا warmup اثری بر سابقه ندارند. query/انتخاب تنوع همان PDP بازدید جدیدی نمی‌سازد؛ refresh یا بازکردن دوباره همان صفحه مشاهده تازه است، ولی همان ردیف را به‌روز می‌کند. keepalive اجازه می‌دهد خروج سریع از صفحه، درخواست شروع‌شده بازدید را بی‌دلیل قطع نکند. خطای ثبت سابقه نمایش موفق PDP را به صفحه خطا تبدیل نمی‌کند؛ تلاش خودکار مکرر در همان نمایش نداریم و بازدید واقعی بعدی دوباره تلاش می‌کند.

این POST یک اعلام بازدید از کلاینت معتبر و با اعتبارسنجی visibility سرور است؛ اثبات انسانی/ضدتقلب یا آمار خام Analytics ادعا نمی‌شود. #76 مالک تحلیل رفتار مستقل است.

## سقف و انقضا

بازه۳۰روز، مدت لغزان `۳۰×۲۴ساعت` بر مبنای زمان سرور است؛ زمان مساوی cutoff منقضی محسوب می‌شود. GET همیشه انقضا/حداکثر۵۰ را پیش از صفحه‌بندی اعمال می‌کند و هیچ مورد قدیمی را نمایش نمی‌دهد. نوشتن مشاهده، داده منقضی/تکراری و اقلام بیش از سقف همان حساب را فشرده می‌کند؛ حساب دیگر برای این عمل حذف نمی‌شود.

برای اینکه داده حساب غیرفعال هم واقعاً پاک شود، [چرخه مستقل cleanup](../../../apps/api/internal/bootstrap/recently_viewed_cleanup.go) در سرور پایدار یک‌بار startup و سپس ساعتی، برای همه حساب‌ها PurgeRecentlyViewed را اجرا می‌کند؛ Redis/فعال‌بودن warmup شرط آن نیست. داده منقضی در خروجی فوراً پنهان است و پاک‌سازی فیزیکی در اجرای سالم سرویس می‌تواند تا نوبت ساعتی تأخیر داشته باشد. با توقف برنامه، job بعدی startup آن را انجام می‌دهد. خطای cleanup بدون اطلاعات شخصی log می‌شود؛ فقط تعداد حذف aggregate گزارش می‌شود. پاک‌سازی در backupهای تاریخی این job نیست.

اگر تغییری لازم نباشد، Purge فقط Snapshot می‌خواند و Update ایجاد نمی‌کند. اگر بین خواندن و Update نویسنده دیگری عمل کند، فشرده‌سازی روی **داده تازه داخل قفل تراکنش** انجام می‌شود، نه نسخه قدیمی Snapshot. Cleanup به موجودی، دفتر، بانک، سفارش، علاقه‌مندی و نقش دست نمی‌زند. داده حافظه‌ای آزمون با restart ذاتاً بازنشانی می‌شود؛ ماندگاری و job runtime در محیط پایدار PostgreSQL بررسی شده‌اند.

## HTTP، مالکیت و خطا

تمام endpointها Bearer با surface=customer، پاسخ no-store و Vary:Authorization دارند؛ ETag یا Redis عمومی برای داده شخصی وجود ندارد. actor از نشست گرفته می‌شود؛ query/body نمی‌تواند کاربر یا زمان ثبت را تعیین کند. شناسه مسیر، ID داخلی محصول برگشتی PDP است؛ URL عمومی کارت را همان ProductURL می‌سازد.

| عملیات | ورودی | خروجی |
| --- | --- | --- |
| GET /storefront/account/recently-viewed | page پیش‌فرض۱، page_size پیش‌فرض۲۴/حداکثر۴۸؛ page در۱تا۱۰۰۰۰۰۰ | items، total، page، page_size، max_items=50، retention_days=30 |
| POST /storefront/account/recently-viewed/{id} | بدون body | recorded:true پس از ذخیره؛ upsert، بدون ردیف تکراری |
| DELETE /storefront/account/recently-viewed/{id} | بدون body | removed؛ فقط مورد همان حساب، غیبت رکورد موفق و بی‌اثر است |
| DELETE /storefront/account/recently-viewed | بدون body | removed؛ پاک‌کردن کل سابقه همان حساب، بدون اثر بر علاقه‌مندی یا حساب دیگر |

هر item شامل product_id، viewed_at معتبر RFC3339 و product از نوع ListingCard یا null است. total ردیف‌های غیرقابل‌دسترسی خود مشتری را هم می‌شمارد؛ metadata محصول غیرعمومی در null نیست. نام/تصویر/ویژگی/قیمت از همان publicProduct/listingCard و قواعد دسته/اجداد/برند فعلی ساخته می‌شوند، نه داده قبلی بازدید.

۴۰۰ ورودی/body/query ناشناخته، timestamp/user_id دلخواه، صفحه/اندازه نامعتبر؛۴۰۱ نشست/سطح اشتباه؛۴۰۳ actor غیرفعال یا نسخه قدیمی؛۴۰۴ ثبت محصول غیرعمومی/حذف‌شده؛۵۰۳ خطای زیرساخت. POST تکراری زمان را به‌روز می‌کند و فقط تعداد رکورد را idempotent نگه می‌دارد؛ DELETE در برابر retry idempotent است. دو عملیات متضاد هم‌زمان به ترتیب commit اعمال می‌شوند؛ پاک‌کردن سابقه توقف ثبت بازدیدهای بعدی نیست.

سابقه از خروجی workspace پنل و projection عمومی حذف می‌شود؛ داده خصوصی از cache stamp و گرم‌کردن عمومی خارج است. خواندن فهرست user دیگر یا دیدن تاریخچه مشتری در پنل فروشنده/ادمین جزو این قابلیت نیست.

## UI و خطا

[صفحه خصوصی](../../../apps/storefront-web/src/recently-viewed-page.tsx) در `/account/recently-viewed` با لینک حساب/پروفایل و همان کارت محصول/قالب A است؛ صفحه‌بندی، تاریخ آخرین مشاهده با تقویم فارسی و timezone تهران، حذف تکی، حالت empty/guest/error/retry و تأیید پاک‌کردن همه دارد. حذف کل نیازمند تأیید صریح همان فرم است؛ cancel داده را نگه می‌دارد و خطا، فرم و داده را از بین نمی‌برد. حذف آخرین مورد صفحه بعدی، به آخرین صفحه معتبر برمی‌گردد.

موجودی/قیمت هنگام مشاهده و خرید مجدداً از کاتالوگ کنترل می‌شوند؛ سابقه رزرو خرید نیست. پاسخ دیررس پس از unmount/تعویض صفحه یا حساب نمایش داده نمی‌شود. logout و تعویض حساب از CommerceProvider/WishlistSession فعلی استفاده می‌کنند؛ فهرست شخصی حساب قبلی نمی‌ماند. loading از BusyRegion مشترک است و صفحه خصوصی با CommerceShell، noindex/nofollow می‌گیرد. داده history در localStorage ذخیره نمی‌شود.

## سازگاری و شواهد

نبود recently_viewed در workspace قدیمی معادل فهرست خالی است؛ خواننده جدید سایر فیلدها را حفظ می‌کند. مانند wishlist، writer قدیمیِ ناآشنا با فیلد تازه برای نوشتن هم‌زمان/rollback مناسب نیست؛ reader/writer سازگار باید با هم منتشر شوند. کد/اسناد هنوز فقط لوکال‌اند.

[تست دامنه/سرویس](../../../apps/api/internal/application/backoffice/recently_viewed_test.go) · [تست HTTP/PG](../../../apps/api/internal/bootstrap/recently_viewed_test.go) · [تست startup/shutdown مستقل از Redis](../../../apps/api/internal/bootstrap/recently_viewed_cleanup_test.go) · [تست مرورگر](../../../tests/e2e/customer-recently-viewed.cjs). BDD در [سناریوهای مستند](../../bdd/features/experiences/customer-recently-viewed.feature) است و runner مستقل آن ادعا نمی‌شود.
