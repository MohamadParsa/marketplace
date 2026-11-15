# راهنمای استقرار روی Liara

این سند طرح اجرایی استقرار marketplace روی Liara است و باید همراه با [ADR-003](../decisions/adr-003-liara-deployment.md) خوانده شود. شناسه برنامه‌ها، تیم، دامنه، اندازه منابع و مقادیر محرمانه نمونه‌اند و باید پیش از استقرار تعیین شوند.

## نقشه سرویس‌ها

| جزء | سرویس Liara | مسیر پیشنهادی در repository | نکته استقرار |
| --- | --- | --- | --- |
| backend | PaaS Go | `apps/api` | برنامه وب روی پورت ۸۰۸۰ یا port ثبت‌شده در `liara.json` گوش می‌دهد |
| بک‌آفیس | PaaS React | `apps/backoffice-web` | خروجی Vite در `build/` قرار می‌گیرد |
| پنل فروشنده | PaaS React | `apps/seller-portal-web` | برنامه و دامنه مستقل |
| سایت مشتری | PaaS React یا Static | `apps/storefront-web` | انتخاب نهایی با نیاز SSR/SEO بازبینی شود |
| داده اصلی | DBaaS PostgreSQL | مستقل از repository | برنامه و دیتابیس در یک شبکه خصوصی قرار گیرند |

هر app قابل استقرار باید `package.json` یا `go.mod`، فایل ignore مناسب و `liara.json` مخصوص خودش را داشته باشد. در monorepo، مسیر deploy با `--path` مشخص می‌شود و نباید ریشه repository به‌عنوان context اشتباه انتخاب شود.

چون migrationهای canonical در `db/migrations` ریشه repository هستند و context سرویس `apps/api` است، workflow backend پیش از deploy آن‌ها را به `apps/api/db/migrations` کپی می‌کند. این کپی بخشی از artifact استقرار است و محل مرجع migrationها را تغییر نمی‌دهد.

## آماده‌سازی backend Go

در `apps/api` این موارد لازم است:

- `go.mod` و `go.sum` در ریشه app؛ پیش از deploy، `go mod tidy` و `go test ./...` اجرا شود.
- برنامه وب روی پورت ۸۰۸۰ اجرا شود یا پورت واقعی در `liara.json` ثبت شود.
- binding سرویس به آدرس قابل دسترسی از شبکه پلتفرم انجام شود؛ پورت hard-code شده در محیط توسعه استفاده نشود.
- لاگ برنامه به stdout/stderr نوشته شود؛ secret، token، connection string و داده شخصی در لاگ نیاید.
- endpointهای سلامت مانند `/health/live` و `/health/ready` داشته باشد. استفاده از آن‌ها به‌عنوان health check پلتفرم پس از بررسی قابلیت همان پلن تنظیم شود.

نمونه `apps/api/liara.json`:

```json
{
  "go": {
    "mainFile": "cmd/api/main.go",
    "port": 8080,
    "version": "1.23"
  }
}
```

فیلدهای `app` و `platform` در `liara.json` قرار نمی‌گیرند؛ شناسه برنامه و پلتفرم از Console یا CLI تعیین می‌شوند.

## migration و PostgreSQL

Migrationهای versioned در `db/migrations/` نگهداری می‌شوند. اجرای migration باید:

1. در محیط CI با یک دیتابیس موقت یا محیط آزمون بررسی شود.
2. در production پیش از آماده‌شدن نسخه جدید اجرا شود.
3. idempotent و قابل بازسازی باشد و در برابر اجرای هم‌زمان قفل داشته باشد.
4. در صورت شکست، deploy را fail کند و وضعیت قابل پیگیری ایجاد کند.

Liara برای Go امکان اجرای `liara_pre_start.sh` را پیش از شروع برنامه و با دسترسی به متغیرهای محیطی فراهم می‌کند. مسیر پیشنهادی فعلی:

```sh
#!/usr/bin/env sh
set -eu
go run ./cmd/migrate up
```

نام ابزار migration هنوز تصمیم‌گیری نشده است؛ دستور بالا قرارداد پیشنهادی است و باید پس از انتخاب ابزار با دستور واقعی جایگزین شود. اگر Liara برنامه را با چند نمونه اجرا کند، اجرای migration در hook هر نمونه نباید بدون lock یا job مستقل انجام شود.

قابلیت hook به نسخه و تنظیمات پلتفرم Go وابسته است؛ پیش از production، اجرای واقعی `liara_pre_start.sh` روی staging ثبت و تأیید شود.

برای PostgreSQL:

- DBaaS PostgreSQL در همان شبکه خصوصی برنامه ایجاد شود.
- مقدار `DB_URI` فقط به‌صورت secret/environment variable در Liara ثبت شود.
- اتصال از اطلاعات شبکه خصوصی انجام شود؛ دسترسی عمومی فقط برای نیاز مشخص عملیاتی فعال شود.
- connection pool با سقف‌های متناسب با پلن DBaaS تنظیم و در shutdown به‌درستی بسته شود.
- backup و restore دیتابیس پیش از production و به‌صورت دوره‌ای آزموده شود.

## آماده‌سازی frontend

برای React/Vite، Liara وجود `package.json` را تشخیص می‌دهد، `npm install` و سپس `npm run build` را اجرا می‌کند. خروجی build باید `build/` باشد؛ برای Vite در `vite.config.ts` مقدار `build.outDir` تنظیم شود.

فایل `liara.json` هر frontend فقط تنظیمات همان app را داشته باشد. در Vite، متغیرهای `VITE_*` هنگام build داخل bundle قرار می‌گیرند؛ بنابراین secret نباید با این روش یا در هیچ فایل frontend ذخیره شود و تغییر مقدار محیطی نیازمند build/deploy مجدد است. هر مقداری که به browser ارسال می‌شود عمومی محسوب می‌شود. اگر نصب یک dependency جدید با mirror پیش‌فرض Liara مشکل داشت، تنظیم `react.mirror: false` در `liara.json` طبق مستندات رسمی بررسی شود.

در monorepo، برای هر frontend یک Liara app و یک workflow مستقل داشته باشیم:

```text
apps/backoffice-web/liara.json
apps/seller-portal-web/liara.json
apps/storefront-web/liara.json
```

## CI/CD پیشنهادی

برای monorepo، GitHub Actions با Liara CLI به‌عنوان مسیر اصلی پیشنهاد می‌شود؛ این روش اجازه می‌دهد test و build پیش از deploy اجرا شوند. اتصال مستقیم GitHub در Console Liara نیز ممکن است، اما برای gateهای چند app و migration کنترل کمتری دارد.

گردش پیشنهادی:

1. Pull request: format، lint، unit test و build فقط appهای تغییرکرده.
2. push به `staging`: deploy appهای مجاز به برنامه‌های staging.
3. push به `main`: پس از موفقیت checkها، deploy production؛ migration production باید در همان مسیر کنترل‌شده اجرا شود.
4. نتیجه deploy، eventها و logهای Liara بررسی و لینک آن‌ها کنار release ثبت شود.

نمونه workflow backend:

```yaml
name: deploy-api

on:
  push:
    branches: [main]
    paths:
      - 'apps/api/**'
      - 'db/migrations/**'
      - '.github/workflows/deploy-api.yml'

jobs:
  verify-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: apps/api/go.mod
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
      - name: Verify API
        working-directory: apps/api
        run: |
          go mod tidy
          test -z "$(git status --porcelain)"
          go test ./...
          go vet ./...
      - name: Install Liara CLI
        run: npm install --global @liara/cli@9
      - name: Deploy API
        env:
          LIARA_API_TOKEN: ${{ secrets.LIARA_API_TOKEN }}
        run: >-
          liara deploy
          --app="${{ vars.LIARA_API_APP }}"
          --path=apps/api
          --platform=go
          --liara-json=apps/api/liara.json
          --api-token="$LIARA_API_TOKEN"
          --no-app-logs
```

برای frontend، همین الگو با `npm ci`، lint، test، `npm run build`، `--path=apps/backoffice-web` و `--platform=react` تکرار می‌شود. نام appها و team id باید در GitHub Variables قرار بگیرند؛ token فقط در GitHub Secrets نگهداری شود. token یا connection string نباید در `liara.json`، workflow، log یا commit قرار گیرد.

برای staging و production باید برنامه‌های Liara جدا داشته باشیم. branch به app متناظر map شود؛ یک workflow نباید بر اساس ورودی قابل‌تغییر کاربر، مقصد production را انتخاب کند. در صورت استفاده از تیم Liara، `--team-id` نیز به deploy اضافه شود.

## دامنه، SSL و شبکه

برای هر محیط دامنه جدا تعریف شود؛ نمونه‌ها:

```text
api.example.test       -> API staging
admin.example.test     -> backoffice staging
api.example.com        -> API production
admin.example.com      -> backoffice production
```

در Liara ابتدا دامنه به app اضافه و سپس رکوردهای DNS اعلام‌شده تنظیم می‌شوند. رکوردهای ALIAS/CNAME، TXT و ACME challenge باید دقیقاً مطابق Console ثبت شوند. پس از تأیید DNS، گواهی SSL از Liara صادر و تمدید آن بررسی شود. تغییر nameserver و propagation ممکن است زمان‌بر باشد؛ قبل از release، DNS و HTTPS با دامنه واقعی آزموده شوند.

## عملیات و کنترل تحویل

قبل از production این موارد باید شواهد داشته باشند:

- build تکرارپذیر و تست موفق در CI
- deploy موفق staging و smoke test API/frontend
- اتصال API به PostgreSQL شبکه خصوصی
- اجرای موفق migration روی snapshot آزمون و سناریوی rollback/restore
- بررسی log، event و restart برنامه در Liara
- بررسی CORS، cookie/secure headers و دامنه‌های مجاز
- تأیید نبود secret و داده شخصی در log، artifact و frontend bundle
- ثبت app id، team، branch mapping، نسخه Go/Node، نسخه DB و runbook دسترسی

## منابع رسمی

- [PaaS Liara](https://docs.liara.ir/paas/about/)
- [استقرار Go](https://docs.liara.ir/paas/go/quick-start/)
- [فایل liara.json برای Go](https://docs.liara.ir/paas/go/quick-start/)
- [Hookهای Go](https://docs.liara.ir/paas/go/how-tos/use-hooks/)
- [اتصال Go به PostgreSQL](https://docs.liara.ir/paas/go/how-tos/connect-to-db/postgresql/)
- [استقرار React](https://docs.liara.ir/paas/react/how-tos/deploy-app/)
- [CI/CD با GitHub](https://docs.liara.ir/paas/cicd/github/)
- [راه‌اندازی PostgreSQL](https://docs.liara.ir/dbaas/postgresql/quick-setup/)
- [اتصال دامنه](https://docs.liara.ir/paas/domains/add-domain/)
- [استقرار با Liara CLI](https://docs.liara.ir/references/cli/deploy-app/)

## lifecycle سرویس Go

ساخت وابستگی‌ها و lifecycle سرویس با Uber Fx مدیریت می‌شود. bind شدن listener در `OnStart` انجام می‌شود تا خطای پورت، startup را ناموفق کند. Fx سیگنال‌های SIGINT/SIGTERM را دریافت و `OnStop` را اجرا می‌کند؛ HTTP server حداکثر ۱۰ ثانیه برای تکمیل درخواست‌های جاری فرصت دارد و پس از timeout اتصال‌ها بسته می‌شوند. مهلت کلی توقف Fx برابر ۱۵ ثانیه است. مهلت termination در پلتفرم باید پیش از انتشار با این بودجه هماهنگ و روی staging آزموده شود؛ این مقدار به‌تنهایی تضمین زمان termination لیارا نیست.
