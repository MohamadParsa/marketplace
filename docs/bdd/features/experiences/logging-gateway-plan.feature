# language: fa
# مرجع: docs/delivery/logging-gateway-plan-2026-11-28/README.md
# این فایل سناریوهای طراحی آینده است؛ runner اجرا نشده است.
@logging-gateway-plan
ویژگی: ثبت محلی رویدادهای مهم پیش از اتصال درگاه

  @acceptance=FEAT-LOG-FOUNDATION-001-AC01 @acceptance=FEAT-LOG-FOUNDATION-001-AC02 @acceptance=FEAT-LOG-FOUNDATION-001-AC03 @acceptance=FEAT-LOG-FOUNDATION-001-AC04
  سناریو: قرارداد مشترک لاگ، شناسه پیگیری و حذف داده حساس
    با فرض داده ساختگی و پیش‌شرط‌های تسک CV-LOG-01
    هنگامی که قابلیت در نوبت اجرای انتخاب‌شده با حالت موفق و شکست بررسی شود
    آنگاه یک عملیات ساختگی با یک operation_id در درخواست، سرویس و کار پس‌زمینه دنبال شود؛ request_id هر درخواست مستقل و provider_attempt_id تلاش بیرونی را جدا کند.
    و نوع‌های technical/integration/business/security/behavior از سطح debug/info/warn/error و نتیجه started/succeeded/rejected/failed/pending/unknown/cancelled جدا باشند؛ زمان رخداد و ثبت، نسخه و service/environment/surface مشخص باشند.
    و در تست تو در تو و خطا، password/OTP/token/Authorization/cookie/client_secret/API-key/DSN/کارت کامل و بدنه خصوصی در هیچ مقصد یا fallback ظاهر نشوند؛ داده بزرگ و log injection کنترل شوند.
    و فیلدهای مشتری/مهمان/فروشگاه از context معتبر سرور باشند؛ داده مرورگر trusted تلقی نشود و logger مالک ledger یا مجوز نباشد.

  @acceptance=FEAT-LOG-LOCAL-SINK-001-AC01 @acceptance=FEAT-LOG-LOCAL-SINK-001-AC02 @acceptance=FEAT-LOG-LOCAL-SINK-001-AC03 @acceptance=FEAT-LOG-LOCAL-SINK-001-AC04
  سناریو: ذخیره محلی لاگ، چرخش فایل و امکان تغییر مقصد
    با فرض داده ساختگی و پیش‌شرط‌های تسک CV-LOG-02
    هنگامی که قابلیت در نوبت اجرای انتخاب‌شده با حالت موفق و شکست بررسی شود
    آنگاه رویداد معتبر در فایل محلی قابل خواندن و پس از restart موجود باشد؛ مسیر پیش‌فرض خصوصی و از HTTP/استاتیک غیرقابل دریافت باشد.
    و چرخش، انقضا و سقف مصرف با داده ساختگی آزموده شوند؛ دو API روی یک دستگاه خط‌ها را خراب نکنند و هر رویداد مبدأ پردازه/سرویس داشته باشد.
    و دیسک پر، مجوز نامعتبر، صف پر و shutdown بدون نشت secret/بن‌بست و حذف خاموش داده مالی آزمایش شوند؛ سیاست شکست audit از telemetry جدا و در NEXT-Q09 تعیین شود.
    و با sink آزمایشی دوم، ثبت‌کننده‌های کسب‌وکار بدون تغییر کار کنند؛ schema ثابت بماند و راهنمای مشاهده/پاکسازی محدود و شاخص dropped/write-failure تحویل شود.

  @acceptance=FEAT-LOG-OPERATIONS-001-AC01 @acceptance=FEAT-LOG-OPERATIONS-001-AC02 @acceptance=FEAT-LOG-OPERATIONS-001-AC03 @acceptance=FEAT-LOG-OPERATIONS-001-AC04
  سناریو: پوشش مسیر خرید، عملیات حساس و تماس سرویس‌های بیرونی
    با فرض داده ساختگی و پیش‌شرط‌های تسک CV-LOG-03
    هنگامی که قابلیت در نوبت اجرای انتخاب‌شده با حالت موفق و شکست بررسی شود
    آنگاه برای مسیر موفق، رد اعتبارسنجی، مجوز، conflict، timeout و نتیجه نامشخص رویداد معنایی درست با operation مشترک ثبت شود؛ کلیک یا HTTP200 معادل موفقیت مالی نباشد.
    و هر تماس واقعی به provider ثبت مستقل شروع/پایان یا شروع بدون پایان قابل بازیابی، زمان و attempt داشته باشد؛ درخواست/پاسخ فقط فیلدهای مجاز و نه dump خام باشند.
    و نمونه خرید از PlaceOrder تا credit/debit یا شکست با شناسه خرید/سفارش/تلاش و منشأ خطا دنبال شود؛ replay با رخداد replay از اثر تجاری تازه جدا شود.
    و پوشش اکشن‌های آماده درگاه (initialize/token/refresh/activation/request/redirect/callback/verify/inquiry/retry/reconciliation و موارد مالی آینده در صورت پیاده‌سازی) در ماتریس ثبت شود؛ هیچ تماس فعال بدون instrumentation پذیرفته نشود.

  @acceptance=FEAT-PAY-GATEWAY-CALLBACK-001-AC01 @acceptance=FEAT-PAY-GATEWAY-CALLBACK-001-AC02 @acceptance=FEAT-PAY-GATEWAY-CALLBACK-001-AC03 @acceptance=FEAT-PAY-GATEWAY-CALLBACK-001-AC04
  سناریو: ساخت callback هنگام پرداخت با دامنه اصلی سایت
    با فرض داده ساختگی و پیش‌شرط‌های تسک CV-PAY-GW-04
    هنگامی که قابلیت در نوبت اجرای انتخاب‌شده با حالت موفق و شکست بررسی شود
    آنگاه در توسعه/استیج/پروداکشن ساخت URL بدون تغییر کد و در لحظه request انجام شود؛ در محیط عمومی origin سایت HTTPS باشد و API/backoffice origin جای آن ننشیند.
    و تغییر Host/Forwarded headers یا query مقصد callback را عوض نکند؛ origin نامعتبر موجب خطای مشخص پیش از تماس بیرونی شود.
    و تلاش آغازشده با callback/تنظیمات خودش بازیابی شود و تغییر دامنه/تنظیم جاری مسیر تلاش قبلی را بازنویسی نکند؛ handler روی دامنه سایت واقعاً route شود.
    و آزمون config/URL و request fixture دامنه درست را اثبات کند؛ log عمومی حاوی token/query حساس نباشد.

  @acceptance=FEAT-PAY-GATEWAY-AUTH-001-AC01 @acceptance=FEAT-PAY-GATEWAY-AUTH-001-AC02 @acceptance=FEAT-PAY-GATEWAY-AUTH-001-AC03 @acceptance=FEAT-PAY-GATEWAY-AUTH-001-AC04
  سناریو: احراز حساب پذیرنده و چرخه فعال‌سازی زرین‌پال
    با فرض داده ساختگی و پیش‌شرط‌های تسک CV-PAY-GW-05
    هنگامی که قابلیت در نوبت اجرای انتخاب‌شده با حالت موفق و شکست بررسی شود
    آنگاه اولین اتصال در پنل و با اقدام ادمین انجام شود؛ روشن‌کردن سوییچ با اتصال معتبر OTP جدید نسازد و خطا/انصراف عرضه درگاه ناقص به مشتری ایجاد نکند.
    و OTP فقط موقت، توکن‌ها رمزنگاری‌شده و فاقد readback باشند؛ refresh با یک نتیجه سازگار در رقابت پردازه‌ها و بدون چاپ credential انجام شود.
    و provider-operation تعیین کند چه credential/قابلیتی لازم است؛ Client ID/Secret جای Merchant لازم V4 فرض نشوند و همه درگاه‌ها مجبور به OAuth نباشند.
    و تمام مراحل تلاش/نتیجه/تمدید/خطا با شناسه پیگیری پالایش‌شده ثبت و نقش مجاز/نسخه/تغییر credential و قطع اتصال آزموده شوند.

  @acceptance=FEAT-PAY-GATEWAY-ADAPTER-001-AC01 @acceptance=FEAT-PAY-GATEWAY-ADAPTER-001-AC02 @acceptance=FEAT-PAY-GATEWAY-ADAPTER-001-AC03 @acceptance=FEAT-PAY-GATEWAY-ADAPTER-001-AC04
  سناریو: اتصال request و verify زرین‌پال با تلاش نسخه‌دار
    با فرض داده ساختگی و پیش‌شرط‌های تسک CV-PAY-GW-06
    هنگامی که قابلیت در نوبت اجرای انتخاب‌شده با حالت موفق و شکست بررسی شود
    آنگاه request ساختگی با مبلغ تبدیل‌شده دقیق، شرح خرید/شارژ و callback دامنه سایت ساخته شود؛ authority پیش از redirect ثبت و request تکراری تلاش ناخواسته دوم نسازد.
    و callback با correlation سروری بدون نیاز به نشست مرورگر به تلاش درست برسد؛ verify با authority/مبلغ/اتصال ثبت‌شده انجام شود.
    و کد100/101، رفض/لغو، پاسخ نامعتبر، timeout و تغییر تنظیمات وسط خرید پوشش داده شوند؛101 مجوز credit دوباره نباشد.
    و تمام اکشن‌ها و تماس‌ها لاگ نوع integration/technical با outcome درست و بدون secret داشته باشند؛ Sandbox به دفتر واقعی متصل نشود.

  @acceptance=FEAT-PAY-GATEWAY-CHECKOUT-001-AC01 @acceptance=FEAT-PAY-GATEWAY-CHECKOUT-001-AC02 @acceptance=FEAT-PAY-GATEWAY-CHECKOUT-001-AC03 @acceptance=FEAT-PAY-GATEWAY-CHECKOUT-001-AC04
  سناریو: هماهنگ‌سازی پرداخت درگاه با کیف پول و سفارش مشتری
    با فرض داده ساختگی و پیش‌شرط‌های تسک CV-PAY-GW-07
    هنگامی که قابلیت در نوبت اجرای انتخاب‌شده با حالت موفق و شکست بررسی شود
    آنگاه callback تکراری/هم‌زمان و restart فقط یک credit و حداکثر یک debit سفارش بسازند؛ ledger مرجع باشد نه log.
    و خرید معتبر با وجه تازه و اعتبار انتخابی به مبلغ درست برسد؛ مصرف خودکار یا دستی فقط پس از پاسخ صریح GATEWAY-Q01 اجرا شود.
    و شارژ مستقل سفارش ایجاد/مصرف نکند؛ نتیجه دیررس اعتبار را حفظ و سفارش منقضی را احیا نکند؛ سفارش صفر بدون redirect به درگاه طبق مسیر فعلی پرداخت شود.
    و مشتری وضعیت موفق/ناموفق/در انتظار نتیجه و مرجع مجاز خودش را ببیند؛ admin/store scope و ثبت رویدادهای مالی/خرید آزموده شوند.

  @acceptance=FEAT-PAY-GATEWAY-RECOVERY-001-AC01 @acceptance=FEAT-PAY-GATEWAY-RECOVERY-001-AC02 @acceptance=FEAT-PAY-GATEWAY-RECOVERY-001-AC03 @acceptance=FEAT-PAY-GATEWAY-RECOVERY-001-AC04
  سناریو: بازیابی نتیجه نامشخص و retry امن پرداخت
    با فرض داده ساختگی و پیش‌شرط‌های تسک CV-PAY-GW-08
    هنگامی که قابلیت در نوبت اجرای انتخاب‌شده با حالت موفق و شکست بررسی شود
    آنگاه قطع مرورگر، callback نرسیده، timeout verify و شکست DB پس از verify تا نتیجه قابل توضیح و اثر مالی یکتا بازیابی شوند.
    و inquiry فقط وضعیت را گزارش دهد؛ PAID/VERIFIED/IN_BANK/FAILED/REVERSED مطابق قرارداد تفسیر و نهایی‌سازی با کنترل مالی انجام شود.
    و آستانه retry و توقف/ارجاع بررسی با attempt، علت، زمان و correlation در لاگ مشخص باشند؛ دو worker یک تلاش را هم‌زمان مؤثر اجرا نکنند.
    و تغییر/غیرفعال‌سازی درگاه مانع پرداخت جدید شود ولی پیگیری تلاش قبلی با نسخه اتصال خودش ممکن بماند؛ نتیجه ناشناخته موفق یا شکست قطعی جعل نشود.

  @acceptance=FEAT-PAY-GATEWAY-ACCEPTANCE-001-AC01 @acceptance=FEAT-PAY-GATEWAY-ACCEPTANCE-001-AC02 @acceptance=FEAT-PAY-GATEWAY-ACCEPTANCE-001-AC03 @acceptance=FEAT-PAY-GATEWAY-ACCEPTANCE-001-AC04
  سناریو: پذیرش یکپارچه زرین‌پال و راهنمای فعال‌سازی
    با فرض داده ساختگی و پیش‌شرط‌های تسک CV-PAY-GW-09
    هنگامی که قابلیت در نوبت اجرای انتخاب‌شده با حالت موفق و شکست بررسی شود
    آنگاه ماتریس خرید/اعتبار/کوپن/مبلغ صفر/شارژ مستقل/لغو/دیررس/تکرار/عدم بازگشت، با trace قابل پیگیری و داده ساختگی کامل باشد.
    و برای همه تماس‌ها و اکشن‌های auth/payment/recovery، زمان/تلاش/نتیجه بدون اطلاعات حساس در ذخیره محلی جست‌وجو شود.
    و Sandbox فقط در محیط ایزوله و پس از انتخاب اجرا؛ تست واقعی فقط با شرایط GATEWAY-Q02/Q03 و دستور جدا، به همراه ثبت دامنه callback در پذیرنده.
    و راهنمای اپراتور و محدودیت‌های باقی‌مانده، commit/push/CI و پذیرش مالک ثبت شوند؛ شروع پردازش و استرداد فعلی با اتصال جدید سازگار بمانند.
