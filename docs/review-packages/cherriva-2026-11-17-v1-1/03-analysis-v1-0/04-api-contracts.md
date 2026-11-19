# قرارداد عملیاتی API و نمونه داده

نسخه: ۱.۰ | تاریخ: ۲۰۲۶-۱۱-۱۷ | وضعیت: تحلیل و طراحی برای بازبینی تیم فنی | مالک بررسی: محصول و راهبر فنی؛ انتساب شخص هنوز انجام نشده

مسیرها و نام DTOها پیشنهاد این بسته‌اند؛ تیم باید از همین قرارداد OpenAPI تولید کند و قبل از اتصال UI در اختیار agent بگذارد. همه endpointهای نوشتن، schema محدود، مجوز، نسخه/کلید تکرار، اندازه درخواست و audit متناسب دارند. session cookie امن HttpOnly برای مرورگر پیشنهاد است؛ CSRF در عملیات مبتنی بر cookie و کنترل origin لازم است. مدل نشست مشترک/زیر‌دامنه در دفتر سؤال‌هاست.

## گروه عملیات

| گروه | مسیر پیشنهادی | ورودی اصلی / خروجی | کنترل |
| --- | --- | --- | --- |
| IAM | POST `/v1/auth/otp/request`، `/verify`؛ POST `/logout`؛ GET `/me` | mobile/challenge/code؛ session + capabilities | نرخ، expiry، یک‌بارمصرف؛ بدون لو دادن وجود حساب |
| Access | GET/POST `/v1/panel/roles`؛ PUT `/roles/{id}`؛ POST `/users/{id}/role-grants` | permission_keys، expected_version، store_scope | access.manage جدا؛ افزایش اختیار خودسرانه ممنوع |
| Stores | `/v1/panel/stores`، `/stores/{id}/policies` | owner، publication/catalog/rate policy | مالکیت، نسخه، صلاحیت |
| Catalog | `/v1/panel/categories`، `/attributes`، `/brands`، `/products`، `/variants`، `/offers` | typed definitions/content/stock؛ نسخه | create/edit/publish جدا؛ scope سلر |
| Pricing | `/v1/panel/sale-methods`، `/category-markups`، `/gold-rates`، `/gold-policies`، `/commission-rules`؛ POST `/pricing/preview` | روش، درگاه، نرخ، درصد و نسخه؛ breakdown | تغییر قانون ≠ تغییر سفارش قبلی |
| Promotion | `/v1/panel/promotions`، `/coupons`؛ POST `/v1/cart/coupon` | شرایط و کد؛ applicability/discount | مبلغ را سرور می‌سازد؛ محدودیت اتمیک |
| Content | `/v1/panel/pages`، `/pages/{id}/sections`، `/preview`، `/publish` | typed config، revision؛ نسخه عمومی | publish مستقل؛ محدودیت scope/URL |
| Discovery | GET `/v1/catalog/categories`، `/products`، `/products/{slug}`، `/search/suggest`، `/pages/{key}` | scope، filter، sort، page، query؛ کارت/فروشنده | host معتبر؛ فقط منتشر/مجاز |
| Cart | GET `/v1/cart`؛ POST `/items`، PATCH `/items/{id}`، POST `/merge` | variant_id، offer_id، quantity، version | reprice، stock، identity و idempotency |
| Checkout | POST `/v1/checkout/quotes`، `/sessions`؛ GET `/sessions/{id}` | cart_version، address، shipment_selections، sale_method؛ quote/session | مبلغ مرورگر پذیرفته نشود؛ Q جلسه باز |
| Wallet | GET `/v1/me/wallet`، `/transactions`؛ POST `/topups`، `/withdrawals` | amount_rial، destination_ref، idempotency_key | balance/lot/capacity/limit/owner |
| Payment | POST `/sessions/{id}/payments`؛ GET `/payments/{id}`؛ provider callback | gateway/method/quote؛ redirect/status | purpose و amount ثابت؛ verify server-side |
| Orders | GET `/v1/me/orders`، `/v1/panel/orders`، `/orders/{id}`؛ POST `/items/{id}/supply-decisions` | decision/quantity/reason/version | مجوز و وضعیت + DTO محدود سلر |
| Shipping | `/v1/panel/shipping-methods`، `/tariffs`، `/shipments`، `/tracking` | پوشش، هزینه، بسته و کدها | همه اقلام آماده؛ تعرفه snapshot |
| Support | `/v1/panel/cases`، `/assignments`، `/decisions` | نوع، اقلام، تأمین‌کننده، علت | وضعیت تازه؛ شماره پرونده مالی مستقل |
| Finance | `/v1/panel/refunds`، `/withdrawals/{id}/decisions`، `/settlements`، `/reconciliation` | item allocations، amount، source، version | ظرفیت و مانده؛ unknown موفق نیست |
| Engagement | `/v1/products/{id}/reviews`، `/questions`، `/me/bookmarks`، `/subscriptions` | محتوا/رضایت؛ moderation state | سیاست و ضداسپم پیش از public write |

## نمونه ساختگی درخواست و خروجی quote

```json
{"cart_version":3,"sale_method_id":"loan-demo","gateway_id":"gateway-demo","address_id":"address-demo","shipments":[{"seller_store_id":"seller-demo","method_id":"delivery-demo"}]}
```

```json
{"quote_id":"quote-demo","currency":"IRR","lines":[{"offer_id":"offer-demo","cash_before_discount_rial":1000,"cash_discount_rial":100,"markup_base_rial":900,"markup_percent":"4","markup_rule_source":"category","markup_rial":36,"item_total_rial":936}],"shipping_total_rial":50,"total_rial":986,"version":1}
```

نمونه فقط فرمول تأییدشده را نشان می‌دهد؛ بدون کمیسیون افزوده/تخفیف درگاه/مالیات نامشخص. schema نهایی اجزای باز پس از پاسخ مالک تکمیل شود؛ مثال مجوز صفرکردن آن‌ها در production نیست.

## خطا و هم‌زمانی

Envelope پیشنهادی: `code`، `message`، `field_errors`، `correlation_id`، `current_version` و `retryable`؛ داده حساس داخل خطا نباشد. 401 نشست نامعتبر؛ 403 عدم اختیار؛ 404 منبع قابل مشاهده نیست؛ 409 نسخه/مبلغ/کلید متفاوت؛ 422 ورودی یا شرط تجاری نامعتبر؛ 429 محدودیت نرخ؛ 503 وابستگی موقت. وضعیت مالی `unknown` پاسخ قابل پیگیری است و به شکست قطعی تبدیل نشود. UI برای conflict بازخوانی و تأیید انتخاب می‌خواهد؛ retry کور مالی ممنوع است.

حداقل کدها: `ACCESS_DENIED`، `STORE_SCOPE_DENIED`، `VERSION_CONFLICT`، `PRICE_CHANGED`، `GOLD_RATE_UNAVAILABLE`، `INSUFFICIENT_STOCK`، `SESSION_EXPIRED`، `PAYMENT_METHOD_INELIGIBLE`، `PAYMENT_UNKNOWN`، `COUPON_NOT_APPLICABLE`، `WITHDRAWAL_REVIEW_REQUIRED`. کد خطا مستقل از متن فارسی UI است. برای منابع خصوصی، انتخاب 403/404 بدون افشای وجود منبع در قرارداد IAM تثبیت شود.
