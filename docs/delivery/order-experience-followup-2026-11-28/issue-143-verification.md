---
id: DEL-ORDER-FOLLOW-143-VERIFICATION
title: شواهد حفظ نشست معتبر بک‌آفیس پس از رفرش و F5
status: in-review
owner: eliassarband
updated: 2026-11-28
related: [FEAT-ORDER-FOLLOW-002]
approved_by: []
implementation_status: done
---

# نتیجه #143

بازیابی نشست با GET auth/session برای بک‌آفیس و فروشنده اضافه شد؛ همان توکن و انقضای سرور حفظ و هویت/مجوز جاری دوباره بررسی می‌شوند. تصمیم فنی این برش: sessionStorage همان تب فقط token/surface/expires را نگه می‌دارد؛ رمز/OTP/نقش ذخیره نمی‌شوند و عمر نشست تمدید نمی‌شود. خروج و تعویض حساب cache و ذخیره را پاک می‌کنند؛ خطای شبکه امکان تلاش مجدد دارد. نشست رمز موقت فقط به همان الزام تغییر رمز برمی‌گردد. تست Go احراز هویت/HTTP موفق، شامل انقضا، surface دیگر، غیرفعال، AccessVersion، logout و restricted. build بک‌آفیس موفق. مرورگر fixture: ورود روی /orders، دو reload با حفظ فهرست و URL و سپس خروج و reload با بازگشت به ورود موفق بودند.

محیط بررسی ساختگی و ایزوله است؛ انتشار عمومی یا تغییر داده کسب‌وکار از این گزارش نتیجه نمی‌شود. کامیت و CI و کامنت بستن در فهرست تحویل نهایی/Issue ثبت می‌شوند.
