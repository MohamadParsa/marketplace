# تحویل و بررسی #96 تا #99

تاریخ:۲۰۲۶-۱۱-۲۵ | مالک تصمیم: الیاس | وضعیت اجرای چهار تسک: done | وضعیت سند: in-review | ادغام/استقرار: انجام نشده

سورس: `35f9c59e9e3b11a7398d7f537451f60014859b36` روی `codex/local-parsa-integration-20261124`. دستور مالک، تکمیل، کامیت/پوش و بستن همین چهار تسک است. تغییرات یکپارچه قبلی پارسا و #102/#103/#104 در تاریخچه همین شاخه حفظ‌اند؛ PR #94 نسخه قدیمی‌تر بسته کوپن/کیف پول است و جایگزین این نسخه ترکیبی نیست.

## نتیجه قابل مشاهده

- #96: هدر باکس‌های مستقل کوپن و کیف پول، باز/بسته‌شدن با کیبورد، حالت applied شامل خلاصه و انصراف، پیام روشن خطا و هم‌راستایی فرم فراهم‌اند. ورودی پس از خطا حفظ می‌شود؛ تغییر معتبر، quote و رضایت قبلی را باطل می‌کند. میانبرهای قدیمی کیف پول حذف شدند.
- #97: مصرف ابتدا فقط‌خرید و سپس قابل‌برداشت، FIFO هر نوع؛ hold/spend/return منبع‌دار و حفظ محدودیت در خرید→لغو اجرا شد. برداشت مستقیم یا تبدیل غیرمستقیم اعتبار محدود به نقد رد می‌شود. وجه مخصوص همان خرید جدا می‌ماند.
- #98: شارژ از ردیف User مشترک با مبلغ/عنوان/دلیل و تیک پیش‌فرض خاموش؛ مجوز قابل انتساب، نقش مالی/ادمین، کنترل مبلغ، ثبت عامل/مقصد و recovery نتیجه نامعلوم بدون شارژ دوم اجرا شد. برای شارژ، واریز بانکی یا رسید جعلی ساخته نمی‌شود.
- #99: تصویر ثابت با fallback، نام/تنوع/تعداد/قیمت هر قلم، سود مستقیم مستقل از کوپن و حذف سطرهای صفر اجرا شد. خلاصه سفارش از snapshot خوانده می‌شود و تغییر بعدی قیمت آن را عوض نمی‌کند.

[قرارداد قواعد مصوب و API/مهاجرت](../../architecture/contracts/wallet-credit-policy.md) مرجع جزئیات است. فایل‌های شرح هر کار از [فهرست همین پوشه](README.md) در دسترس‌اند.

## نگاشت پذیرش به شاهد

| پذیرش‌ها | تست/شاهد اجرا | نتیجه |
| --- | --- | --- |
| FEAT-REV-001-AC01/02 | مرورگر: collapsed/applied، کیبورد، حفظ فرم، alignment و390px | موفق |
| FEAT-REV-001-AC03/04 | مرورگر: یک preview در دوبار کلیک، حذف adjustment و ابطال quote، کد اشتباه/منقضی، سهمیه و شبکه و موجودی ناکافی | موفق؛ سهمیه و شبکه در UI با fault injection، قواعد واقعی سهمیه در Go |
| FEAT-REV-002-AC01/02 | `TestWalletCreditFourSourcesFIFOAndCancellationRestriction` و چرخه سفارش مرورگر | مصرف چهار منبع و برگشت محدودیت؛ رد برداشت پس از لغو |
| FEAT-REV-002-AC03 | `TestWalletCreditLegacyBalanceHoldsAndExpiry`، `TestWalletCreditLegacyConfirmedOrderCancellationKeepsCash` و تست PostgreSQL | مانده و سفارش قدیمی نقدی باقی ماندند؛ دفتر قبلی بازنویسی نشد |
| FEAT-REV-002-AC04 | `TestWalletCreditConcurrentPurchaseAndWithdrawal`، تست retry و expiry | فقط یک رزرو از مانده مشترک؛ آزادسازی بدون تغییر نوع یا بدهکارشدن دوباره |
| FEAT-REV-003-AC01/02 | تست سرویس، HTTP و مجوز قابل انتساب/ابطال آن | هر دو نوع، مبلغ/فیلد نامعتبر، نشست/مجوز و کلید متعارض بررسی شدند؛۱۲retry هم‌زمان یک credit ساخت |
| FEAT-REV-003-AC03/04 | PostgreSQL reopen/قیود immutable و مرورگر lost-success-reply و موبایل | نتیجه ثبت گم‌شده با خواندن همان کلید بازیابی شد؛ منشأ و محدودیت پابرجا ماند |
| FEAT-REV-004-AC01 | مرورگر دو قلم/دو فروشنده، تنوع/تعداد۲، تصویر خراب و قاب ثابت در دسکتاپ/موبایل | موفق |
| FEAT-REV-004-AC02/03 | `TestOrderSavingsSeparateFromCouponAndImmutableAfterConfirmation` و مرورگر | طلا/دیجیتال و تعداد، مرجع نامعتبر، کوپن مستقل، سطر صفر و مبلغ مثبت بررسی شدند |
| FEAT-REV-004-AC04 | همان تست snapshot و خواندن سفارش پس از تغییر قیمت‌ها | خلاصه و سود خرید قطعی ثابت ماند |

تست‌ها: [سرویس و قواعد مالی](../../../apps/api/internal/application/backoffice/commerce_wallet_policy_test.go)، [HTTP و PostgreSQL](../../../apps/api/internal/bootstrap/wallet_credit_policy_test.go)، [مرورگر](../../../tests/e2e/issues-96-99.cjs)، [نتایج ماشین‌خوان](issues-96-99-checks.json). [۱۶ سناریوی BDD](../../bdd/features/experiences/checkout-wallet-policy.feature) مستندند؛ runner مستقل Gherkin اجرا نشده است.

## بررسی‌های نهایی

- `go test -count=1 -json ./...`: **۲۸۸ تست و زیرسناریو موفق، صفر شکست**؛ تست PostgreSQL و Redis با dependency آزمون اجرا شد.
- PostgreSQL برای اجرای نهایی یک cluster تازه روی loopback و پورت55525 بود؛ هر تست schema جدا داشت. cluster پس از تست به‌صورت graceful متوقف شد. بانک بررسی اصلی و داده مالی کاربران آن تغییر نکردند.
- دو skip: `TestCatalogSnapshotPostgresRoundTrip` که بسته snapshot بیرونی می‌خواهد، و `TestAdjustmentsBrowserHarness` که فقط با opt-in جدا اجرا می‌شود. harness مرورگر جداگانه با موفقیت استفاده شد.
- `go vet ./...`: موفق. `go test -race` محلی اجرا نشد؛ CGO/کامپایلر C این Windows فراهم نیست و CI لینوکس آن را اجرا می‌کند.
- lint و build سایت و بک‌آفیس موفق‌اند؛ هشدار قبلی اندازه bundle/annotation در بک‌آفیس مانع build نیست.
- **۳۰/۳۰ بررسی Playwright موفق، بدون pageerror**. خطاهای شبکه عمداً برای تصویر، پیام سهمیه، قطع ارتباط و پاسخ گم‌شده اعمال شدند؛ ادعای console/network بدون خطا نداریم.

در نمونه مرورگر: کالاها۳۰۰٬۰۰۰، ارسال۲۰٬۰۰۰، سود مستقیم۷۰٬۰۰۰، کوپن۳۰٬۰۰۰ و مبلغ نهایی۲۹۰٬۰۰۰ تومان است. انتخاب کیف پول۲۵۰٬۰۰۰ و وجه جدید۴۰٬۰۰۰ جدا ثبت شدند. لغو،۵۰٬۰۰۰ فقط‌خرید و۲۴۰٬۰۰۰ قابل‌برداشت برگرداند؛ هیچ وجه واقعی منتقل نشد.

## تصاویر

- [بازیابی شارژ ثبت‌شده پس از گم‌شدن پاسخ](../../../output/playwright/issues-96-99/admin-credit-recovered-desktop.png)
- [فرم شارژ ادمین در موبایل](../../../output/playwright/issues-96-99/admin-credit-mobile.png)
- [پرداخت با کوپن و اعتبار، دسکتاپ](../../../output/playwright/issues-96-99/checkout-applied-desktop.png)
- [پرداخت با مبلغ‌های اعمال‌شده، موبایل](../../../output/playwright/issues-96-99/checkout-applied-mobile.png)
- [بازگشت فرم ویرایش اعتبار](../../../output/playwright/issues-96-99/checkout-editable-mobile.png)
- [منابع اعتبار سفارش](../../../output/playwright/issues-96-99/order-source-breakdown-mobile.png)
- [نوع و منشأ اعتبار مشتری](../../../output/playwright/issues-96-99/wallet-source-policy-mobile.png)

## تکرار مرورگر

پس از build دو frontend، از `apps/api` و بدون اتصال به دیتابیس بررسی اجرا کنید:

~~~powershell
$env:CHERRIVA_ADJUSTMENTS_BROWSER='1'
$env:CHERRIVA_WALLET_POLICY_BROWSER='1'
go test ./internal/bootstrap -run '^TestAdjustmentsBrowserHarness$' -count=1 -v
~~~

سپس از ریشه ریپو:

~~~powershell
npx --yes --package @playwright/cli playwright-cli -s=cherriva-wallet-final open http://localhost:8096
npx --yes --package @playwright/cli playwright-cli -s=cherriva-wallet-final run-code --filename tests/e2e/issues-96-99.cjs
~~~

هر بار اجرای سناریو، harness حافظه‌ای تازه لازم دارد؛ داده را تغییر می‌دهد و پس از پایان فرآیند از بین می‌رود. پورت‌های8096 و5196 با نسخه بررسی8098/5197 جدا هستند.

## مرز انتشار

Migration `000010-wallet-credit-policy` افزایشی و همراه سورس است؛ اجرای آن روی محیط بررسی اصلی یا سرور در این تحویل انجام نشده است. rollback پس از ثبت اطلاعات مالی صریحاً متوقف می‌شود. مبلغ/محدودیت/منشأ منبع در PostgreSQL immutable و spent افزایشی‌اند؛ تست حذف، تغییر محدودیت، کاهش spent و rollback دارای داده موفق بود.

این گزارش مجوز یا شاهد استقرار، پیامک/بانک واقعی، merge به main یا تکمیل #105/#12 نیست. درخواست جاری اجازه پوش همین شاخه را می‌دهد؛ نتیجه کامیت/PR/CI و بستن Issueها در گیت جدا ثبت می‌شود. اسناد تاریخی و manifest بسته طراحی تغییر نکرده‌اند.

## نتیجه تحویل روی گیت — ۲۰۲۶-۱۱-۲۵

شاخه پوش شد و [PR #106](https://github.com/MohamadParsa/marketplace/pull/106) باز است. چهار Issue #96/#97/#98/#99 با دلیل completed و کامنت شواهد بسته شدند؛ [رسید لینک‌ها و وضعیت‌ها](issues-96-99-delivery.json). [CI نسخه کد و شواهد](https://github.com/MohamadParsa/marketplace/actions/runs/45556932708) روی `b724a694f089beb06afcc6294d8cf62352ec87b3` در هر پنج job موفق شد؛ `go test -race` لینوکس و vet هم موفق‌اند. کامیت بعدی فقط رسید تحویل و وضعیت اسناد را ثبت می‌کند و تغییر اجرایی ندارد. #95 و #105 باز مانده‌اند؛ ادغام و استقرار انجام نشده است.
