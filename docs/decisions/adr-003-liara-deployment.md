---
id: ADR-003
title: پلتفرم استقرار و CI/CD با Liara
status: approved
owner: product-owner
updated: 2026-11-10
related: [ADR-002, NFR-INDEX-001]
approved_by: [product-owner]
---

# پلتفرم استقرار و CI/CD با Liara

- نوع: معماری
- تاریخ تصمیم: 2026-11-10
- اسناد متأثر: [راهنمای استقرار Liara](../delivery/liara-deployment.md)، [معماری](../architecture/README.md)، [آمادگی](../delivery/readiness.md)

## زمینه و مسئله

محصول باید backend Go، frontendهای مستقل و PostgreSQL را با محیط‌های staging و production deploy کند. repository یک monorepo است و pipeline باید پیش از استقرار، آزمون و build را اجرا کند.

## گزینه‌ها

| گزینه | مزایا | هزینه / محدودیت | نتیجه |
| --- | --- | --- | --- |
| PaaS بومی Liara برای Go و React | build و deploy ساده، CLI و دامنه/SSL آماده | برخی تنظیمات پیشرفته به قابلیت‌های Liara وابسته‌اند | انتخاب شد |
| Docker روی Liara | کنترل کامل‌تر runtime و یکسانی image | نگهداری Dockerfile و پیچیدگی بیشتر برای شروع | برای نیاز مشخص آینده |
| زیرساخت مستقل | کنترل کامل topology و عملیات | هزینه و نگهداری بسیار بیشتر | فعلاً انتخاب نشد |

## تصمیم

- Liara پلتفرم استقرار اولیه است.
- backend از PaaS بومی Go و frontendها از PaaS بومی React استفاده می‌کنند؛ Docker فقط در صورت نیاز واقعی به runtime یا build سفارشی وارد می‌شود.
- PostgreSQL از DBaaS Liara و ترجیحاً در همان شبکه خصوصی برنامه تهیه می‌شود.
- CI/CD با GitHub Actions و Liara CLI انجام می‌شود تا test و build پیش از deploy gate باشند.
- برای هر محیط و هر app، Liara app مستقل و branch mapping صریح وجود دارد.
- tokenها در GitHub Secrets و app id/team id در GitHub Variables قرار می‌گیرند.
- migration پیش از start برنامه اجرا می‌شود، اما فقط با ابزار دارای idempotency و lock یا job مستقلِ کنترل‌شده.

## پیامدها

- repository به فایل‌های `liara.json` مستقل در هر app و workflowهای مسیرمحور نیاز دارد.
- پورت، نسخه Go/Node، mirror، دامنه و متغیرهای محیطی باید در قرارداد deployment ثبت شوند.
- اتصال دیتابیس عمومی برای runtime پیش‌فرض نیست و باید از شبکه خصوصی استفاده شود.
- وابستگی عملیاتی به Liara ایجاد می‌شود؛ مسیر Docker و خروج از پلتفرم باید در صورت نیاز قابل بررسی باقی بماند.

## بازنگری

افزایش نیاز به کنترل runtime، محدودیت PaaS، الزام استقرار چندناحیه‌ای یا نیاز به مقیاس‌گذاری مستقل، محرک بازنگری این تصمیم است.
