---
id: VERIFY-BUSINESS-CONTENT-20261126
title: تحویل مستقل فوتر و محتوای عمومی — #87
status: in-review
owner: eliassarband
updated: 2026-11-26
related: [FEAT-NEXT-026, API-BUSINESS-CONTENT-001]
approved_by: []
implementation_status: done
execution_authorized: true
---

# نتیجه #87

صفحات عمومی با وضعیت انتشار مستقل و شبکه‌های اجتماعی با عنوان/آیکن داخلی/HTTPS/ترتیب/active در فرم تنظیمات موجود مدیریت می‌شوند. فوتر مشترک فقط صفحه دارای متن عمومی و شبکه فعال معتبر را نشان می‌دهد؛ URL خالی/خطرناک یا شبکه غیرفعال و متن پیش‌نویس منتشر نمی‌شود. API عمومی مستقل و reader کوچک PostgreSQL، کش۵دقیقه و ETag با invalidation تراکنشی و startup warmup اضافه شد. همان مجوز، version و idempotency تنظیمات موجود برقرار است.

طبق پاسخ صریح مالک، [متن‌های پیشنهادی](issue-87/content-drafts-fa.md) آماده و **انتشار نهایی از پنل با خودشان** است. متن واقعی، شماره/ایمیل/شبکه واقعی، نماد یا تعهد فرضی در داده اصلی اضافه نشد. تکمیل قابلیت مدیریت/انتشار از انتشار محتوای واقعی و استقرار سایت جداست.

[سورس](https://github.com/MohamadParsa/marketplace/commit/0c80dcf7c2e03c2b5a883a680bdf8b7918c9b39e) · [قرارداد](../../architecture/contracts/business-content.md) · [تسک](tasks/cv-next-26-footer-business-content.md) · [شواهد ماشین‌خوان](issue-87-checks.json).

| پذیرش | شاهد |
| --- | --- |
| AC01 لینک فعال با مقصد واقعی | انتشار درباره ساختگی از پنل، رفتن به /info/about و نمایش متن؛ حذف لینک تماس/مرجوعی بدون انتشار در مرورگر |
| AC02 مدیریت/انتشار امن | متن جدید پیش‌نویس می‌ماند؛ رد true بدون متن؛ پیش‌نویس از /business و /commerce/config حذف، در پنل پس از restart محفوظ است |
| AC03 شبکه و حفاظت | ترتیب۱۰/۲۰، active و حذف عمومی پس از غیرفعال‌سازی، target/rel امن و fallback؛ 401/403، URL خطرناک، duplicate ID، نسخه قدیمی و تکرار؛ cache و PostgreSQL/restart |
| AC04 تجربه و یکپارچگی | موبایل پنل/فوتر، عدم overflow، فرم حفظ‌شده پس از400، خانه/PLP/PDP، نبود placeholder با شبکه صفر، منع viewer و صفر runtime pageerror |

## بررسی‌ها

- `go test -json ./...` با PostgreSQL schema جدا و Redis تست: **۳۸۹تست/زیرتست موفق،۴skip اختیاری،صفرشکست**. شواهد قابلیت در `apps/api/internal/bootstrap/business_content_test.go` است.
- `go vet ./...` و build هر دو app موفق؛ هشدار شناخته‌شده chunk بزرگ پنل باقی است؛ dependency جدید نداریم. race محلی به‌دلیل نبود toolchain/CGO اجرا نشد و نتیجه CI در رسید جدا ثبت می‌شود.
- `tests/e2e/business-content.cjs`: **۱۶کنترل مرورگر موفق**، روی schema جدا و پورت8097/5199؛ محتوا، لینک‌ها و ایمیل screenshot ساختگی‌اند. URL نامعتبر عمداً400 گرفت، بنابراین ادعای نبود خطای شبکه نداریم.
- نسخه اصلی در `D:\Cherriva` با Start-Cherriva.ps1 -NoBuild و Redis/warmup به‌روز شد. مقایسه **۲۸بخش داده** پس از restart برابر baseline است؛ کاربر/رمز/نقش، کاتالوگ، رسانه، سفارش، بانک، روش‌های پرداخت/ارسال، کیف پول و دفتر محفوظ‌اند. فقط projection قابل بازسازی در startup تکمیل می‌شود؛ SQL جایگزینی داده اصلی و اتصال مبدأ نداریم.
- reader فوتر در تست cache hit/miss Snapshot خصوصی نمی‌گیرد؛ تغییر محتوا کش فوتر را عوض می‌کند. ETag بدنه خانه/PLP/PDP ممکن است ثابت بماند، چون فوتر از endpoint مستقل تازه دریافت می‌شود.

## تصاویر

[فوتر دسکتاپ](../../../output/playwright/issue-87/footer-desktop.png) · [صفحه درباره منتشرشده آزمون](../../../output/playwright/issue-87/about-published.png) · [ویرایشگر موبایل](../../../output/playwright/issue-87/editor-mobile.png) · [فوتر موبایل](../../../output/playwright/issue-87/footer-mobile.png).

## محدودیت و ادامه

صفحه منتشرنشده در SPA پیام عدم دسترسی دارد و لینک فعالش در فوتر نیست؛ رفتار HTTP/SEO کامل صفحات در #74 است. شکل URL کنترل می‌شود و حساب/دسترس‌پذیری واقعی شبکه احراز نشده است. نبود نگاشت انتشار در داده قدیمی، متن غیرخالی قبلی را طبق رفتار قدیم عمومی نگه می‌دارد؛ صفحه تازه در فرم پیش‌نویس است. سند in-review؛ main/Liara ادغام یا منتشر نشده‌اند. #119 جداگانه پس از این تحویل اجرا می‌شود.

## رسید نهایی

سورس و مدارک بالا روی شاخه PR #106 پوش شدند؛ [کامنت پایان](https://github.com/MohamadParsa/marketplace/issues/87#issuecomment-7198121596) ثبت و #87 completed بسته شد. هر۵job [CI](https://github.com/MohamadParsa/marketplace/actions/runs/45861713576) با سورس یکسان، شامل race/vet، موفق‌اند. [رسید](issue-87-delivery.json). انتشار محتوای واقعی و استقرار جداست.
