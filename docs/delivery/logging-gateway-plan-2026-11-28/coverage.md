# ماتریس پوشش اجرای لاگ

| مسیر اجراشده | ثبت و نتیجه | منبع |
| --- | --- | --- |
| HTTP حساس/سایت | شروع و پایان با request/operation، route قالبی، status و مدت؛ نه body/path/query خام. خطا/رد/نتیجه نامشخص جداست. | [middleware](../../../apps/api/internal/transport/http/operation_logging.go) |
| login/مجوز | نوع security؛ actor فقط پس از احراز سروری. کد و توکن و cookie ثبت نمی‌شوند. | [auth](../../../apps/api/internal/application/auth/service.go) و middleware |
| سبد، quote، پذیرش سفارش، credit/debit، کوپن، تنظیمات، مراحل مرسوله، درخواست/استرداد | technical شروع/نتیجه و replay با context و مرجع مجاز پاسخ؛ موفقیت HTTP با پرداخت قطعی یکی نیست. | [mutateCommerce](../../../apps/api/internal/application/backoffice/commerce_base.go) و [مرجع/طبقه‌بندی](../../../apps/api/internal/application/backoffice/operation_logging.go) |
| پیش‌نمایش خرید و command مدیریت | نتیجه typed عملیات با context مشترک | [checkout](../../../apps/api/internal/application/backoffice/commerce_checkout.go) و [service](../../../apps/api/internal/application/backoffice/service.go) |
| پیامک OTP کاوه‌نگار | هر Send: شروع/نتیجه/مدت/status/code؛ HTTP200 با نتیجه رد provider موفق نیست. شماره/OTP/کلید/بدنه حذف‌اند. | [adapter](../../../apps/api/internal/adapters/kavenegar/notifier.go) |
| reverse نشان | هر Lookup: شروع/نتیجه/مدت/status و timeout؛ مختصات/نشانی/کلید/پاسخ خام حذف‌اند. | [adapter](../../../apps/api/internal/adapters/neshan/reverse.go) |
| lifecycle، کش و cleanup موجود | slog از همان Recorder پالایش‌شده؛ متن/exception آزاد ذخیره نمی‌شود. | [entrypoint](../../../apps/api/cmd/api/main.go) و [handler slog](../../../apps/api/internal/application/operationlog/slog.go) |
| gateway initialize/token/refresh/activation/request/redirect/callback/verify/inquiry/retry/reconciliation و مالی بعدی | **معوق و اجرا نشده**؛ قرارداد BeginExternal و schema آماده است، هیچ ادعای تماس/پوشش بانک واقعی نیست. | #158..163 |

audit کسب‌وکار (#75)، رفتار سروری و قیف (#76) و پذیرش نهایی (#77) در [گزارش اجرا](execution.md) جدا تکمیل می‌شوند؛ این ماتریس ادعای ثبت هر رویداد فرضی یا اتصال بیرونی تازه نیست. همه تست‌های نوشتنی روی fixture/مسیر موقت‌اند.
