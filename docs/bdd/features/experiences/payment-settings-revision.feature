# language: fa
# مرجع: docs/delivery/payment-settings-revision-2026-11-28.md
@payment-settings-revision
ویژگی: تنظیمات ثابت و کشویی درگاه و credential اختصاصی

  @acceptance=FEAT-BO-GATEWAY-SETTINGS-REVISION-001-AC01
  سناریو: OAuth بدون Merchant اجباری
    با فرض ادمین مجاز و اطلاعات کلاینت ساختگی
    هنگامی که Client ID فارسی و secret بلند بدون Merchant ذخیره و بعد ویرایش شوند
    آنگاه حروف و فاصله secret حفظ و مقدار خالی موجب حذف credential قبلی نشود

  @acceptance=FEAT-BO-GATEWAY-SETTINGS-REVISION-001-AC02
  سناریو: schema و سوییچ پایدار
    با فرض provider پشتیبانی‌شده بدون رکورد و schema آزمایشی فقط API key
    هنگامی که تنظیمات خوانده و وضعیت مجاز تغییر و سرور دوباره باز شود
    آنگاه تنظیمات همیشه موجود و برای provider یکتا باشد و credential خام یا OTP ذخیره‌شدنی نباشد و روش‌های پرداخت فعلی حفظ شوند

  @acceptance=FEAT-BO-GATEWAY-SETTINGS-REVISION-001-AC03
  سناریو: شروع بسته و تغییر بدون build
    با فرض صفحه روش‌های پرداخت در دسکتاپ یا موبایل
    هنگامی که صفحه بارگذاری و سپس تنظیمات باز و ذخیره شوند
    آنگاه همه بخش‌ها در شروع بسته و دکمه افزودن غایب باشد و سوییچ و credential پس از refresh حفظ شوند و فرم سرریز نکند

  @acceptance=FEAT-BO-GATEWAY-SETTINGS-REVISION-001-AC04
  سناریو: شرح وابسته به عملیات و حفظ سابقه
    با فرض تنظیمات قبلی دارای Merchant و شرح ثابت
    هنگامی که OAuth اضافه و شرح خرید یا شارژ ساخته شود
    آنگاه credential قبلی حفظ و شرح از همان عملیات ساخته شود و تنظیمات به‌تنهایی پرداخت بیرونی اجرا نکنند
