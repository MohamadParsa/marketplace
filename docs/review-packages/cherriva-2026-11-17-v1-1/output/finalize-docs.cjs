const fs=require('fs'),path=require('path');
const write=(f,s)=>{fs.mkdirSync(path.dirname(f),{recursive:true});fs.writeFileSync(f,s.trim()+'\n','utf8');};
const header='نسخه: ۱.۱ | تاریخ: ۲۰۲۶-۱۱-۱۷ | وضعیت سند: in-review | مالک تصمیم: الیاس؛ بازبین فنی: راهبر فنی، انتساب فرد باز';
const rows=[
 ['01','single-user','هویت واحد و عضویت چندمنظوره','P0','W0',['CV-FND-04'],['CV-IAM-04','CV-IAM-05','CV-IAM-07'],'identity-auth-fa.md','خریدار و همکار پنل و عضو فروشگاه، بدون ایجاد حساب تکراری از یک هویت استفاده کنند.','User مرجع هویت؛ RoleAssignment و StoreMembership رابطه مستقل؛ buyer/customer فقط کاربرد User. Store و آدرس گیرنده مستقل باقی بمانند. موبایل نرمال یکتا؛ سیاست ایمیل در کار بعد. اگر داده قدیمی واقعی وجود داشت، گزارش تکراری‌ها قبل از merge، بدون ادغام خودکار.','User، RoleAssignment، StoreMembership؛ GET /me و GET /admin/users/:id، DTO بدون credential.','users.view/edit و members.manage در scope مجاز؛ داشتن کاربرد خریدار، مجوز پنل ایجاد نمی‌کند.','شناسه ناموجود 404؛ دسترسی خارج scope 403 یا 404 طبق قرارداد مشترک؛ تعارض نسخه409.',['USR-001 در خرید و پنل و عضویت S1 همان شناسه بماند.','تعویض نما/فروشگاه، User جدید یا گسترش scope نسازد.','کاربر سلر با تغییر store_id به داده S2 دسترسی پیدا نکند.'],'CHQ-06 فقط ماتریس ارتقای اختیار/انتساب؛ اصل هویت واحد روشن است.'],
 ['02','phone-otp-link','ورود نخست با موبایل و اتصال به همان هویت','P0','W0',['CV-R2-01','CV-IAM-01','CV-IAM-02'],['CV-IAM-01','CV-IAM-02'],'identity-auth-fa.md','شروع ورود ساده بماند و تأیید دوباره یک شماره، حساب تازه نسازد.','OTP تأییدشده به normalized_mobile همان User resolve شود؛ ساخت User و مصرف challenge اتمیک/تکرارامن؛ داده هویت سایت و پنل مشترک. نشست و scope از سرور؛ نمایش انتخاب نقش در mock وارد محصول نشود.','POST /auth/otp/request و /verify؛ AuthSession شامل user_id و مجوز/عضویت مجاز؛ سیاست cookie پس از CHQ-08.','مسیر عمومی rate-limited؛ اعطای نقش هرگز از verify ورودی کلاینت انجام نشود.','OTP_INVALID، OTP_EXPIRED، TOO_MANY_ATTEMPTS؛ پاسخ عمومی بدون افشای وجود حساب.',['شماره فارسی/لاتین یکسان به یک User resolve شود.','دو verify هم‌زمان نه User و نه مصرف challenge تکراری بسازند.','OTP منقضی/مصرف‌شده نشست ندهد؛ لاگ کد خام نداشته باشد.'],'اتصال پیامک واقعی CV-INT-04؛ پارامترهای جاری OTP از تسک اصلی، تغییرشان در این اصلاح درخواست نشده است.'],
 ['03','profile-email-password','پروفایل و ورود با ایمیل یا موبایل و یک رمز','P0','W0',['CV-R2-02'],['CV-IAM-02','CV-IAM-07','CV-WEB-07'],'identity-auth-fa.md','کاربر پس از ورود اولیه، ایمیل و رمز خود را اضافه کند؛ دو رمز یا حساب جدا نسازد.','یک PasswordCredential برای User؛ endpoint ورود identifier+password، normalize موبایل/ایمیل؛ hash امن و عدم بازگشت رمز. ایمیل اختیاری؛ تأیید مالکیت ایمیل قبل از login پیشنهاد امنیتی قابل انتخاب است. پروفایل از سایت و پنل یک قرارداد مصرف کند.','PATCH /me؛ POST /me/email/verify و /me/password؛ POST /auth/password پیشنهادی. unique email پس از سیاست انتخاب‌شده؛ ابطال نشست‌های لازم بعد تغییر credential.','فقط مالک همان User؛ ویرایش نقش یا store_membership در self-profile ممنوع.','INVALID_CREDENTIALS عمومی؛ EMAIL_CONFLICT، EMAIL_UNVERIFIED، PASSWORD_POLICY؛ پیام فیلد و حفظ فرم.',['موبایل+رمز و ایمیل+همان رمز به user_id یکسان وارد شوند.','قبل از تعریف رمز، password login شکست قابل فهم داشته باشد؛ OTP باقی بماند.','ایمیل خالی مجاز؛ تکراری/تأییدنشده طبق سیاست منتخب؛ هیچ رمز در log/DTO نباشد.'],'سیاست بازیابی/تغییر موبایل و جزئیات password/session در open-items-fa.md؛ انتخاب سیاست قبل از انتشار مسیر مربوط.'],
 ['04','category-brand-media','آیکن دسته و برند با پیش‌فرض','P1','W1',['CV-CAT-01','CV-CAT-03','CV-CAT-07'],['CV-CAT-01','CV-CAT-03','CV-CAT-07'],'catalog-workspace-fa.md','دسته و برند از پنل نشان قابل تغییر داشته باشند و بدون upload هم خوانا دیده شوند.','media_id اختیاری، alt و پیش‌فرض داخلی ثابت؛ حذف/فایل خراب به fallback قابل پیش‌بینی. component مشترک preview/upload/remove، اعتبارسنجی نوع/حجم و reference پیش از حذف asset. سقف فایل در mock قرارداد نهایی نیست.','Category.icon_media_id و Brand.logo_media_id؛ media adapter آزمون؛ PATCH موجودیت نسخه‌دار.','categories.edit یا brands.edit؛ سلر به‌صورت پیش‌فرض مدیر کاتالوگ عمومی نیست؛ مجوز نهایی تابع CHQ-06.','MEDIA_TYPE_UNSUPPORTED، MEDIA_TOO_LARGE، MEDIA_NOT_READY؛ خطا کنار upload؛ حفظ مقدار قبلی در شکست.',['پیش‌فرض در نبود تصویر و پس از حذف نمایش داده شود.','تصویر موبایل/دسکتاپ سایت و پنل کشیده یا مخدوش نشود؛ alt معنا داشته باشد.','فایل نامعتبر و URL/script دلخواه پذیرفته نشود؛ draft منتشرنشده public نشود.'],'ابعاد و حجم نهایی قالب به انتخاب جهت طراحی وابسته است؛ storage واقعی CV-INT-05.'],
 ['05','shared-attribute-bindings','کتابخانه خصوصیت و انتساب مستقل به دسته','P0','W1',['CV-CAT-01','CV-CAT-02'],['CV-CAT-02','CV-CAT-05','CV-CAT-08'],'catalog-workspace-fa.md','تعریف تکراری خصوصیت کم شود؛ هر دسته تعیین کند چه چیزی انتخاب مشتری و چه چیزی اطلاعات کارت است.','AttributeDefinition مشترک با type/unit/options؛ CategoryAttribute با is_variant_axis، show_on_card، is_filterable و order مستقل. مثال موبایل رنگ/گارانتی axis، حافظه/رم/دوربین card. تعداد محور یا ویژگی از مثال ثابت نشود. تغییر محور فعال، preview اثر و version بخواهد؛ ترکیب Variant باید معتبر باشد.','GET /attributes و /categories/:id/attributes؛ PUT binding نسخه‌دار؛ attribute_id مشترک؛ ارث خودکار از والد تا پاسخ CHQ-09 اجرا نشود.','attributes.manage و categories.edit؛ اثر تغییر تعریف مشترک در دسته‌های دیگر قابل مشاهده باشد.','ATTRIBUTE_IN_USE، OPTION_IN_USE، VERSION_CONFLICT؛ حذف مخرب مانع یا گردش بازبینی لازم.',['همان color_id در موبایل و هدفون استفاده شود؛ تغییر binding موبایل روی هدفون اثر نگذارد.','PDP چیپ رنگ/گارانتی و کارت حافظه/رم/دوربین را از binding بگیرند.','تعریف تکراری و ترکیب تنوع نامعتبر رد شود؛ تغییر محور، سفارش snapshot را عوض نکند.'],'CHQ-09 ارث و تغییر دسته فعال؛ تعداد/overflow قالب CHQ-12.'],
 ['06','product-workspace','اکشن سریع تنوع، قیمت و موجودی در محصول','P0','W1',['CV-R2-05','CV-CAT-04','CV-CAT-06'],['CV-CAT-04','CV-CAT-05','CV-CAT-06'],'catalog-workspace-fa.md','کاربر پنل از ردیف محصول یا جزئیات، سریع به همان فرم عملیاتی برسد.','سه tab مشخصات/تنوع/عرضه‌ها؛ یک component برای هر فرم و ورودی عمیق با product_id. Variant ترکیب مشخصات؛ Offer کلید فروشنده+تنوع و محل قیمت/stock. فیلتر scope روی query و mutation؛ محصول مشترک به سلر حق ویرایش محتوای عمومی نمی‌دهد.','GET /products/:id/variants و /offers?product_id؛ PATCH /offers/:id با version و موجودی کل/رزرو؛ ذخیره گروهی یا ردیفی صریح باشد.','products.edit، variants.edit، offers.edit و stock.edit مستقل؛ مخفی‌کردن اکشن جای کنترل API نیست.','STOCK_BELOW_RESERVED، PRICE_INVALID، OFFER_CONFLICT؛ خطای هر ردیف؛ موفقیت جزئی شفاف و بدون از دست رفتن draft.',['هر دو entry point یک فرم و همان شناسه تنوع/عرضه را باز کنند.','تغییر رنگ/گارانتی به عرضه صحیح برسد؛ price روی Product نوشته نشود.','موجودی کمتر از رزرو و ویرایش عرضه فروشگاه دیگر رد شود.'],'قیمت فرمولی طلا تا CHQ-03 فقط پیکربندی draft؛ race واقعی نیازمند تست سرور است.'],
 ['07','responsive-banner-section','سکشن بنر با تصویر دسکتاپ و موبایل','P1','W2',['CV-CMS-01','CV-CMS-02','CV-R2-04'],['CV-CMS-01','CV-CMS-02'],'sections-fa.md','بنر روی موبایل crop نامناسب نگیرد و فرم آن فیلد نامرتبط ریل محصول نداشته باشد.','discriminated schema برای hero/banner-grid؛ desktop_media_id و mobile_media_id و alt/link مستقل هر اسلاید. preview دو اندازه؛ draft/revision و publish موجود حفظ شود. script arbitrary در link یا متن ممنوع؛ template نسبت تصویر را اعلام کند.','Section.type و payload معتبر برای همان type؛ picture/source با breakpoint قالب؛ GET page revision و publish اتمیک.','pages.edit و pages.publish مستقل و scope main/gold/digi؛ upload مجوز رسانه.','DESKTOP_IMAGE_REQUIRED، MOBILE_IMAGE_REQUIRED، LINK_INVALID، REVISION_CONFLICT؛ انتشار ناقص متوقف شود.',['فرم بنر دو preview مستقل دارد؛ تصویر تغییر یک اندازه دیگری را overwrite نکند.','عرض390 و1440 تصویر درست را نشان دهند؛ alt و لینک قابل لمس باشند.','ویرایش draft، صفحه منتشرشده یا کش عمومی را تا publish عوض نکند.'],'اندازه نمونه پیشنهادی؛ نسبت/سقف محتوا نهایی CHQ-12؛ asset واقعی مستقل.'],
 ['08','manual-product-rail','کروسل اختصاصی با انتخاب و ترتیب دستی کالا','P1','W2',['CV-CMS-01','CV-CMS-02','CV-CAT-09','CV-PRC-07'],['CV-CMS-01','CV-CMS-02','CV-WEB-03'],'sections-fa.md','بازاریابی بدون کدنویسی، کالاهای منتخب هر ویترین را بچیند.','product_ids ordered unique؛ picker جست‌وجوی DB و فیلتر scope؛ محصول انتخاب شود نه قیمت ثابت یا offer قفل‌شده. renderer از عرضه نماینده معتبر قیمت بخواند؛ ناموجود/خارج scope/unpublished پس از publish حذف امن شود. افزودن/حذف/بالا/پایین در draft.','Section.product-rail source=manual یا منبع محدود؛ query picker صفحه‌بندی‌شده؛ title/more_link و product_ids در payload.','pages.edit/publish؛ query فقط کاتالوگ مجاز همان ویترین؛ انتخاب ID دلخواه از API کنترل شود.','PRODUCT_NOT_ELIGIBLE، DUPLICATE_PRODUCT، SECTION_EMPTY، REVISION_CONFLICT؛ وضعیت کالای حذف‌شده واضح.',['انتخاب دستی، حذف و جابه‌جایی بعد بازخوانی draft ترتیب را حفظ کنند.','کالای دیجیتال در gold پذیرفته نشود؛ قیمت از offer معتبر بروز شود.','ناموجود شدن محصول باعث لینک خرید غلط یا کارت بدون قیمت معتبر نشود.'],'رفتار کمتر از۵ کالا CHQ-12؛ سیاست render empty پیش از انتشار تعیین شود.'],
 ['09','sticky-banner-section','سکشن استیکی با کنترل نمایش و بستن','P1','W2',['CV-R2-07'],['CV-CMS-01','CV-CMS-02','CV-WEB-03'],'sections-fa.md','کمپین برجسته دیده شود بدون پوشاندن مسیر خرید و دکمه‌ها.','sticky-banner با desktop/mobile media، alt/link، position، dismissible و enabled؛ پیش‌فرض غیرفعال. renderer فقط در جای مجاز قالب، offset برای bottom navigation و sticky CTA. دوام بستن در session یا مدت کمپین پیشنهاد UX است و باید انتخاب شود.','payload typed؛ scope و revision؛ محدودیت تعداد هم‌زمان و صفحه‌های مجاز از قالب.','pages.edit/publish با scope؛ محتوای بدون مجوز منتشر نشود.','STICKY_POSITION_UNSUPPORTED، SECTION_CONFLICT؛ کنترل بستن قابل دسترس با keyboard و نام.',['در390 و1440 CTA و هدر/فیلتر پوشانده نشوند.','بستن banner کار کند؛ focus گم نشود؛ enabled=false چیزی render نکند.','checkout حساس طبق سیاست منتخب از مزاحمت banner مصون بماند.'],'دامنه مجاز صفحه/دوام بستن و تعداد هم‌زمان پیشنهاد است، در open-items ثبت شده.'],
 ['10','seller-supply-lead','وقفه تأمین ثابت فروشنده و snapshot سفارش','P0','W1',['CV-IAM-06','CV-CAT-06','CV-OPS-01'],['CV-IAM-06','CV-CAT-06','CV-OPS-05','CV-OPS-06'],'seller-lead-time-fa.md','هر فروشنده یک وقفه ثابت داشته باشد؛ مشتری آماده‌سازی را با زمان حمل اشتباه نگیرد.','Store.supply_lead_time تنها منبع جاری؛ حذف lead override در Product/Category/Offer. read model از فروشگاه join شود؛ order_line snapshot مقدار وقت خرید را حفظ کند. اگر داده قبلی اختلاف داشت، گزارش conflict و انتخاب مسئول؛ migration مقدار دلخواه انتخاب نکند.','PATCH /stores/:id برای مقدار؛ Offer DTO فیلد read-only مشتق؛ quote/order snapshot شامل supply_lead و حمل مستقل.','stores.edit یا permission خاص فروشنده طبق CHQ-06؛ تغییر موجودی به معنی اختیار تغییر SLA نیست.','SUPPLY_LEAD_INVALID؛ ورود فیلد override به Offer رد شود؛ stale version409.',['دو offer یک فروشنده همیشه مقدار یکسان جاری نشان دهند.','تغییر Store روی quote جدید اثر بگذارد؛ سفارش پرداخت‌شده قبلی ثابت بماند.','صفحه حمل، وقفه و transit را جدا نشان دهد؛ زمان تقویم بدون CHQ-07 اختراع نشود.'],'واحد روز کاری در نمونه است؛ تقویم، cut-off و ترکیب وعده حمل CHQ-07 پیش از وعده واقعی لازم‌اند.'],
 ['11','storefront-alignment','همگام‌سازی سایت با قراردادهای اصلاح‌شده','P0','W3',['CV-R2-03','CV-R2-05','CV-R2-06','CV-R2-08','CV-R2-09','CV-R2-10'],['CV-WEB-02','CV-WEB-03','CV-WEB-04','CV-WEB-05','CV-WEB-06','CV-WEB-07','CV-WEB-08'],'storefront-impact-fa.md','کاربر سایت همان تنظیمات و قواعد پنل را با زبان ساده ببیند؛ طرح جدید رفتار خرید قبلی را حفظ کند.','مصرف قرارداد مشترک auth/profile، آیکن fallback، binding کارت/چیپ، انتخاب offer و lead فروشنده. سبد با offer_id و variant_id، قیمت/stock server-authoritative؛ banner/rail published. سه جهت بصری فقط theme/layout هستند؛ منطق مشترک باشد.','DTOهای عمومی بدون اطلاعات خصوصی؛ افزودن به سبد idempotent طبق قرارداد قبلی؛ هیچ محاسبه قطعی قیمت در browser.','public read scoped؛ account/cart owner-only؛ سرور بررسی session/store و انتخاب offer.','NO_MATCHING_OFFER، OFFER_UNAVAILABLE، QUOTE_CHANGED، SESSION_EXPIRED؛ متن واضح و اقدام اصلاح.',['گارانتی۲۴ پس از انتخاب در قیمت، لیست seller، سبد و checkout یکسان دیده شود.','کارت بدون انتخاب تنوع به PDP مستقل برود؛ CTA خرید همان استایل مرجع حفظ شود.','درخت PLP و برگ بدون زیرشاخه، سه جهت بصری و موبایل مرور شوند؛ قوانین طلا حفظ شوند.'],'بعد از پذیرش پنل/سلر آزمون آغاز شود؛ host/session CHQ-08 و قیمت طلا CHQ-03 فقط مسیر وابسته را می‌بندند.'],
 ['12','revision-acceptance','پذیرش اصلاحات، ردیابی و آماده‌سازی تحویل','P0','W3',['CV-R2-01','CV-R2-02','CV-R2-03','CV-R2-04','CV-R2-05','CV-R2-06','CV-R2-07','CV-R2-08','CV-R2-09','CV-R2-10','CV-R2-11'],['CV-QA-01','CV-QA-04','CV-QA-06'],'README-fa.md','تیم بداند چه چیزی طراحی شده، چه چیزی ساخته/آزموده شده و چه چیزی هنوز تصمیم لازم دارد.','ماتریس rule→spec→task→screen→evidence؛ contract tests برای scope/credential/lead و browser برای binding/section/checkout. گزارش mock، integration، browser و provider جدا؛ اصلاح تسک اصلی به‌جای دوباره‌شمردن قابلیت‌ها.','test fixtures بدون اطلاعات شخصی؛ snapshot قرارداد و گزارش قابل بازتولید، هیچ provider واقعی در این task لازم نیست.','بازبین مشخص برای محصول/فنی؛ test runner مجوز production یا انتقال وجه ندارد.','شکست تست مانع همان مسیر؛ broken link/شناسه تکراری/گم‌شدن نیازمندی رد تحویل.',['تمام ACهای اصلاحی به شاهد پیوند داشته باشند؛ هیچ done بدون شاهد ثبت نشود.','تست مثبت/منفی مستقیم API برای هویت واحد و scope فروشگاه وجود داشته باشد.','اسناد current و archive تفکیک، لینک/شناسه و task graph معتبر و طرح انتخابی ثبت شده باشند.'],'قبولی مرورگر نمونه موجود جای آزمون محصول را نمی‌گیرد؛ اتصال واقعی گیت جدا دارد.']
];
const tasks=rows.map(([n,slug,title,priority,wave,deps,amends,spec,business,technical,data,permission,errors,acceptance,block])=>({id:'CV-R2-'+n,file:'tasks/cv-r2-'+n+'-'+slug+'.md',title,priority,wave,status:'todo',owner:'unassigned',dependencies:deps,amends,spec,business,technical,data,permission,errors,acceptance,block}));
for(const t of tasks){write('04-current-spec/'+t.file,`---
id: ${t.id}
document-status: in-review
implementation-status: todo
priority: ${t.priority}
wave: ${t.wave}
owner: unassigned
updated: 2026-11-17
---

# ${t.id} — ${t.title}

## هدف بیزینسی

${t.business}

این کار اصلاح/تکمیل تسک‌های ${t.amends.join('، ')} است؛ قابلیت مستقل تکراری یا پیاده‌سازی انجام‌شده محسوب نشود. قواعد صریح را از [مشخصات جاری](../${t.spec}) و راه‌حل پیشنهادی را از متن زیر بخوانید.

## دامنه و راه‌حل فنی پیشنهادی

${t.technical}

## داده و قرارداد پیشنهادی

${t.data}

نام endpoint و schema در این سند پیشنهاد قرارداد است؛ در kickoff با OpenAPI پروژه مقصد هماهنگ شود و بدون نسخه‌گذاری با مسیر قدیمی موازی نشود.

## دسترسی و محدوده

${t.permission}

## خطا و حالات رابط

${t.errors}

حالت loading، خالی، خطای شبکه و تلاش مجدد امن، عدم دسترسی و conflict متناسب با همان عملیات لازم‌اند. خطا نباید draft معتبر کاربر را حذف یا موفقیت ساختگی اعلام کند.

## وابستگی و مانع دقیق

- وابستگی‌ها: ${t.dependencies.map(id=>id.startsWith('CV-R2')?'['+id+']('+path.basename(tasks.find(x=>x.id===id).file)+')':'['+id+'](../../03-analysis-v1-0/tasks/'+id.toLowerCase()+'.md)').join('، ')}.
- ${t.block}
- مسئول فردی و موعد تعیین نشده‌اند؛ اندازه‌بندی در kickoff پس از بررسی سورس مقصد. اولویت ${t.priority}؛ موج ${t.wave}.

## معیار پذیرش

${t.acceptance.map((a,i)=>'- '+t.id+'-AC0'+(i+1)+': '+a).join('\n')}

## خروجی قابل تحویل تیم

قرارداد/کد محدود به این دامنه، migration در صورت نیاز با گزارش اثر داده، بررسی معنادار ACها و گزارش فایل‌ها/نتیجه/محدودیت. ابتدا پنل و سلر با داده آزمون؛ فلو سایت پس از گیت مرحله اول. بررسی mock و اتصال واقعی جدا گزارش شوند؛ این فایل مجوز شروع خودکار یا انتشار نیست.

[فهرست اصلاحات و ترتیب اجرا](../task-index-fa.md) · [راهنمای شروع](../../START-HERE-fa.md)
`);}
write('04-current-spec/task-index.json',JSON.stringify({version:'1.1',date:'2026-11-17',baseline_tasks:91,amendment_count:12,note:'Amendments to existing tasks; not 103 independent features. All implementation states todo.',tasks},null,2));
write('04-current-spec/task-index-fa.md',`# کارهای اصلاحی و ترتیب شروع

${header} | شناسه: CV-R2-TASKS

این ۱۲ کار، اصلاح ۹۱ تسک پایه‌اند؛ ابتدا در backlog تیم با تسک متأثر ادغام شوند. ۱۰۳ قابلیت مستقل یا ۱۰۳ تسک ساخته‌شده محصول ادعا نمی‌شود. وضعیت اجرای همه todo و مسئول فردی unassigned است. W3 سایت بعد از قبولی تمام دامنه پنل/سلر با داده آزمون است؛ P0 به معنی حذف این گیت نیست.

| کار | اولویت / موج | اصلاح تسک‌های پایه | وابستگی |
| --- | --- | --- | --- |
${tasks.map(t=>'| ['+t.id+' — '+t.title+']('+t.file+') | '+t.priority+' / '+t.wave+' | '+t.amends.join('، ')+' | '+t.dependencies.join('، ')+' |').join('\n')}

## مسیر پیشنهادی

W0: قرارداد پایه و هویت → OTP مشترک → پروفایل و رمز. W1: کاتالوگ/خصوصیت/عرضه و منبع وقفه فروشنده. W2: فرم typed بنر، ریل دستی و sticky. در کنار این اصلاحات، تمام سفارش/تأمین/ارسال/پشتیبانی/واگذاری/مالی/تسویه پایه با آداپتر آزمون ادامه دارند و از مرحله اول حذف نشده‌اند. W3: پس از پذیرش پنل، مصرف این قراردادها در سایت و پذیرش نهایی اصلاحات.

\`\`\`mermaid
flowchart LR
  I[هویت واحد] --> O[OTP و نشست]
  O --> P[پروفایل و رمز]
  A[خصوصیت مشترک] --> C[فضای محصول و عرضه]
  M[رسانه و fallback] --> B[بنر responsive]
  B --> S[استیکی]
  R[ریل دستی] --> W[همگام‌سازی سایت]
  P --> W
  C --> W
  S --> W
  L[وقفه فروشنده] --> W
  W --> Q[پذیرش با شاهد]
\`\`\`

گراف فقط اصلاحات را نشان می‌دهد؛ وابستگی تمام کارهای پایه در [task-index پایه](../03-analysis-v1-0/task-index.json) و [برنامه تحویل](../03-analysis-v1-0/12-delivery-plan.md) است. نسخه ماشین‌خوان اصلاحات: [JSON](task-index.json).
`);
write('04-current-spec/open-items-fa.md',`# تصمیم‌ها و ابهام‌های باقی‌مانده

${header} | شناسه: CV-R2-OPEN

[دفتر ۱۴ پرسش پایه](../03-analysis-v1-0/13-open-questions.md) حفظ می‌شود. پاسخ تازه کاربر فقط اصل هویت مشترک و مشخصات کاتالوگ/سکشن/وقفه را روشن کرده است؛ همه CHQ-06 یا CHQ-12 خودکار بسته نشده‌اند.

| شناسه تکمیلی | آنچه روشن است | آنچه نیازمند تصمیم است | مانع دقیق |
| --- | --- | --- | --- |
| R2-Q01 | موبایل نخست، ایمیل اختیاری، یک رمز مشترک | تأیید مالکیت و یکتایی ایمیل؛ بازیابی رمز؛ تغییر موبایل و اثر بر نشست‌ها | انتشار credential واقعی؛ طراحی/fixture مستقل قابل بررسی |
| R2-Q02 | همان User می‌تواند چند کاربرد و عضویت داشته باشد | اتحاد/انتخاب نقش‌ها، ارتقای اختیار، اعضای چندفروشگاهی و محافظ آخرین مدیر | CHQ-06؛ اعطای دسترسی نهایی |
| R2-Q03 | Attribute مشترک و binding هر دسته؛ مثال رنگ/گارانتی | ارث از والد، تغییر محور فعال، overflow کارت و محور | CHQ-09/12؛ تغییر مخرب و قالب نهایی |
| R2-Q04 | دو تصویر بنر، ریل دستی، sticky | نسبت/حجم نهایی، حداقل ریل، مدت بستن sticky، صفحه/موقعیت مجاز و تعداد هم‌زمان | انتشار قالب نهایی؛ فرم draft قابل اجرا |
| R2-Q05 | وقفه فقط سطح فروشنده | تقویم روز کاری، cut-off، زمان حمل و ترکیب وعده دقیق | CHQ-07؛ وعده واقعی |
| R2-Q06 | هویت منطقی مشترک سایت و پنل | host و اشتراک نشست/cookie و login destination | CHQ-08؛ استقرار و session بین دامنه‌ها |

این‌ها پرسش اجباری تازه در گفتگو نیستند؛ فهرست تصمیم قابل پیگیری kickoff هستند. پاسخ شامل تصمیم‌گیر، تاریخ، دامنه، مثال و اسناد/تسک‌های متأثر باشد. بدون پاسخ، مسیر وابسته متوقف و کار روشن مستقل ادامه یابد. پیشنهاد امنیتی «تأیید ایمیل» در دیزاین دیده می‌شود و به نام کاربر تصویب نشده است.
`);
write('04-current-spec/traceability-fa.md',`# ردیابی اصلاحات به طرح و تسک

${header} | شناسه: CV-R2-TRACE

| قاعده | شرح و قرارداد | دیزاین پنل | دیزاین سایت | کار اصلاحی |
| --- | --- | --- | --- | --- |
| R2-01/02 | [هویت](identity-auth-fa.md) | BO-02/03/27 و login | login و profile | CV-R2-01/02/03 |
| R2-03 | [رسانه](catalog-workspace-fa.md) | BO-05/06 | category و brand rail | CV-R2-04 |
| R2-04 | [خصوصیت](catalog-workspace-fa.md) | BO-07 و preview | کارت و چیپ PDP | CV-R2-05/11 |
| R2-05 | [فضای محصول](catalog-workspace-fa.md) | BO-08/09 | offer/variant selection | CV-R2-06 |
| R2-06 | [سکشن](sections-fa.md) | BO-16 و سه فرم اختصاصی | home و حالت sticky | CV-R2-07/08/09 |
| R2-07 | [وقفه](seller-lead-time-fa.md) | BO-04/09/17 | PDP/cart/shipping | CV-R2-10 |
| R2-08 | [اثر روی سایت](storefront-impact-fa.md) | [نقشه پنل](../02-backoffice/screen-index-fa.md) | [گالری سایت](../01-storefront/index.html) | CV-R2-11/12 |

شواهد همین نمونه در [گزارش بررسی](../VERIFICATION-fa.md)، تصویرهای جاری در [فهرست تصاویر](../IMAGE-INDEX-fa.md) و تمام ACها در [تسک‌ها](task-index-fa.md) هستند. وضعیت پیاده‌سازی محصول از وضعیت تهیه سند و مرور نمونه مستقل است.
`);

// Refresh the active copy of backoffice docs. Baseline source and archive remain unchanged.
const bo='02-backoffice/';
for(const f of ['README-fa.md','flows-fa.md','implementation-handoff-fa.md','screen-index-fa.md','design-guide-fa.md','verification-fa.md','open-decisions-fa.md','review-log-fa.md','image-index-fa.md']){
 let s=fs.readFileSync(bo+f,'utf8').replaceAll('نسخه: ۰.۱','نسخه: ۰.۲').replaceAll('۲۶ صفحه','۲۷ صفحه').replaceAll('۲۶ مسیر','۲۷ مسیر');
 s=s.replace('مدیر `09000000001`، سلر `09000000006`','همه نماهای دسترسی با همان هویت `USR-001` و موبایل `09000000001`').replace('پنج حساب نمونه','پنج نمای دسترسی از همان هویت');
 if(f==='flows-fa.md')s=s.replace('موجودی/آماده‌سازی/قیمت','موجودی/قیمت و وقفه خواندنی فروشنده');
 if(f==='implementation-handoff-fa.md')s=s.replace('ویرایش محلی قیمت/عنوان/موجودی/آماده‌سازی','ویرایش محلی قیمت/عنوان/موجودی؛ وقفه خواندنی از فروشگاه').replace('خریدار، فروشگاه و حساب کاربر یکی فرض نشوند','خریدار همان User است؛ فروشگاه و گیرنده سفارش مستقل‌اند');
 s+='\n\n## اصلاحات جاری نسخه ۰.۲\n\n[مرجع اصلاحات ۱.۱](../04-current-spec/README-fa.md) در موارد مشخص‌شده بر متن و کپی مراجع ۱.۰ مقدم است. همان User برای خرید/پنل/عضویت، ورود OTP و رمز موبایل/ایمیل و پروفایل BO-27، آیکن دسته/برند، خصوصیت مشترک با binding مستقل، اکشن سریع تنوع/عرضه، فرم typed بنر/ریل دستی/sticky و وقفه فقط در فروشگاه لحاظ شدند. قیمت/موجودی همچنان Offer است. اعتبار یکپارچگی نشست یا backend در این نمونه ادعا نمی‌شود.\n\nمرجع بررسی جاری: [گزارش کل بسته](../VERIFICATION-fa.md). تصاویر با نام صفحه در output/playwright در این نسخه مجدداً گرفته شدند؛ تصاویر مقایسه‌ای قدیمی با توضیح در فهرست کل، سابقه‌اند. ثبت این نوبت DEC-028 / ACT-046.\n';
 if(f==='screen-index-fa.md')s+='\n| شناسه تکمیلی | شرح | مسیر | تسک |\n| --- | --- | --- | --- |\n| BO-27 | [پروفایل همان کاربر](screens/bo-27-my-profile.md) | `#my-profile` | CV-R2-03 |\n';
 write(bo+f,s);
}
const corrections={
 'bo-02-users':'کاربر و مشتری دو هویت جدا نیستند. فهرست کاربران، کاربرد خریدار/بک‌آفیس و عضویت فروشگاه را برای همان User نشان می‌دهد؛ کاربرد خریدار مجوز پنل ایجاد نمی‌کند. CV-R2-01/02/03.',
 'bo-03-roles':'اصل چند کاربرد برای یک User روشن شد؛ CHQ-06 فقط برای جزئیات اعطای نقش، ارتقای اختیار، union/scope و محافظ آخرین مدیر باز است. presetهای UI الزام امنیتی نیستند.',
 'bo-04-stores':'وقفه تأمین ثابت فقط در Store تعریف/ویرایش می‌شود؛ موجودی یا دسته/محصول/عرضه override ندارند. زمان حمل/تعرفه مستقل و snapshot سفارش قبلی ثابت است. CV-R2-10.',
 'bo-05-categories':'آپلود آیکن اختیاری و بازگشت به پیش‌فرض؛ انتساب خصوصیت از کتابخانه مشترک و flagهای axis/card/filter مستقل برای همین دسته. تغییر موبایل به هدفون نشت نمی‌کند. CV-R2-04/05.',
 'bo-06-brands':'آیکن/لوگوی آپلودی، متن جایگزین و fallback داخلی؛ حذف فایل یا شکست دریافت، نام برند و پیش‌فرض را حفظ کند. upload این نمونه فقط preview محلی است. CV-R2-04.',
 'bo-07-attributes':'تعریف مشترک خصوصیت بین چند دسته؛ binding مستقل در هر دسته. مثال موبایل: رنگ و گارانتی axis؛ حافظه/رم/دوربین card. این مثال سقف تعداد ثابت ایجاد نمی‌کند. preview از binding جاری ساخته می‌شود. CV-R2-05.',
 'bo-08-products':'از گرید و جزئیات، اکشن‌های مشخصات/تنوع/قیمت و موجودی به همان فرم مشترک برسند. Product محتوا، Variant ترکیب، Offer قیمت و موجودی فروشنده است. CV-R2-06.',
 'bo-09-offers':'وقفه عرضه فقط خواندنی از Store؛ فیلد editable آماده‌سازی در Offer حذف شده است. کنترل موجودی کمتر از رزرو پابرجاست. رنگ/گارانتی نمونه تنوع و قیمت/stock روی عرضه‌اند. CV-R2-06/10.',
 'bo-16-pages':'فرم بر اساس type عوض می‌شود: بنر دو تصویر desktop/mobile، ریل دستی با picker و ترتیب، sticky با تصویر و جای نمایش/بستن. draft هر scope مستقل و انتشار شبیه‌سازی‌شده است. CV-R2-07/08/09.',
 'bo-17-shipping':'زمان حمل و تعرفه متعلق به روش ارسال‌اند؛ وقفه تأمین فقط از فروشنده خوانده می‌شود و این فرم حق override کالا/دسته ندارد. تقویم و cut-off CHQ-07. CV-R2-10.'
};
for(const [name,note] of Object.entries(corrections)){let s=fs.readFileSync(bo+'screens/'+name+'.md','utf8').replace('نسخه: ۰.۱','نسخه: ۰.۲');if(name==='bo-09-offers')s=s.replace('قیمت، موجودی و زمان آماده‌سازی هر تنوع از هر فروشنده.','قیمت و موجودی هر تنوع از هر فروشنده؛ وقفه خواندنی از فروشگاه.').replace('آماده‌سازی و فعال‌بودن','وقفه خواندنی فروشگاه و فعال‌بودن').replace('→ آماده‌سازی → ذخیره محلی','→ مشاهده وقفه فروشگاه → ذخیره محلی');write(bo+'screens/'+name+'.md',s+'\n\n## اصلاح مشخصات جاری\n\n'+note+'\n\n[مرجع جاری](../../04-current-spec/README-fa.md) · [کارهای اصلاحی](../../04-current-spec/task-index-fa.md)\n');}
write(bo+'screens/bo-27-my-profile.md',`# BO-27 — پروفایل همان کاربر

شناسه: CVBO-BO-27 | نسخه: ۰.۲ | تاریخ: ۲۰۲۶-۱۱-۱۷ | وضعیت: in-review | مالک تصمیم: الیاس

هدف: افزودن ایمیل و یک رمز برای همان User که خریدار/همکار/عضو فروشگاه است. موبایل تأییدشده در نمونه خواندنی؛ تغییر موبایل نیازمند مسیر تأیید مستقل است. نام و ایمیل در فرم اطلاعات؛ رمز و تکرار در فرم جدا. ایمیل اختیاری؛ تأیید مالکیت ایمیل پیشنهاد طراحی امنیتی است.

کاربر فقط اطلاعات خودش را تغییر دهد؛ self-profile مجوز مدیریت نقش/عضویت ندارد. error فیلدی، رمز نامطابق، ایمیل تکراری یا تأییدنشده و نشست منقضی باید روشن باشند. password policy نمونه نهایی نیست. یک رمز برای موبایل/ایمیل به همان user_id برسد و در log/DTO بازگردانده نشود.

در نمونه، ذخیره ایمیل → تأیید کد000000 → ثبت رمز ساختگی → خروج → password login با موبایل یا ایمیل در همان تب قابل مرور است؛ Reload حافظه را پاک می‌کند. سایت و پنل storage/session مشترک واقعی ندارند و فقط fixture هویت USR-001 یکسان است.

[قرارداد هویت](../../04-current-spec/identity-auth-fa.md) · [تسک CV-R2-03](../../04-current-spec/tasks/cv-r2-03-profile-email-password.md)

[جهت A](../output/playwright/bo-27-my-profile-a.png) · [جهت B](../output/playwright/bo-27-my-profile-b.png) · [جهت C](../output/playwright/bo-27-my-profile-c.png) · [نقشه پنل](../screen-index-fa.md)
`);
// Site docs keep previous details and point to precise new rules.
let siteRead=fs.readFileSync('01-storefront/README-fa.md','utf8').replace('نسخه ۰٫۳ — ۲۰۲۶-۱۱-۱۶','نسخه ۰٫۴ — ۲۰۲۶-۱۱-۱۷');siteRead+='\n\n## اصلاحات ۰.۴ و تصاویر جاری\n\n[اثر اصلاحات بر سایت](../04-current-spec/storefront-impact-fa.md): ورود و پروفایل همان User، محورهای رنگ/گارانتی، ویژگی‌های حافظه/رم/دوربین کارت، وقفه فروشنده و حالت بنر استیکی. دکمه مرجع خرید و سه جهت قبلی حفظ شده‌اند. [تصاویر تازه](../IMAGE-INDEX-fa.md) در images/revision-v0-4 مرجع نسخه جاری‌اند؛ تصویرهای مسیر images بیرون از این زیرپوشه و گزارش قدیمی سابقه نسخه۰.۳ هستند.\n\nشبیه‌سازی سایت و پنل storage و session مشترک واقعی ندارد؛ یک fixture User مشترک مدل می‌شود. فرم ایمیل/رمز سایت فقط همان صفحه را تغییر می‌دهد و بازخوانی یا ناوبری حافظه را پاک می‌کند. [بررسی جاری](../VERIFICATION-fa.md) و [نقطه شروع تیم](../START-HERE-fa.md).\n';write('01-storefront/README-fa.md',siteRead);
write('README-fa.md',`# بسته یکپارچه طراحی و شروع پیاده‌سازی چریوا

شناسه: CV-KIT-20261117 | ${header}

این پوشه خروجی مستقل و قابل ارسال برای تیم است: دیزاین سایت نسخه۰.۴، بک‌آفیس/سلر نسخه۰.۲، بسته تحلیل۱.۰ با ۹۱ تسک پایه و اصلاحات جاری۱.۱ با ۱۲ کار تکمیلی. هیچ فایل محصول در D:\\Cherriva برای تهیه این خروجی تغییر داده نشده است. همه طراحی‌ها سه جهت **متعادل و شیک، فروش‌محور، لوکس و تصویری** دارند؛ انتخاب نهایی باز است.

**ابتدا [START-HERE-fa.md](START-HERE-fa.md) را بخوانید.** برای مشاهده طرح‌ها [index.html](index.html) را باز کنید. تمام assetها محلی‌اند. در صورت محدودیت مرورگر: در همین پوشه \`node serve.cjs\` و سپس http://127.0.0.1:4190/ ؛ توقف با Ctrl+C. این فقط سرور نمایش طرح است.

| پوشه | کاربرد و نسخه |
| --- | --- |
| [01-storefront](01-storefront/README-fa.md) | سایت ۰.۴؛ Home/PLP/PDP/cart/login/address/shipping/payment/result/profile؛ سه جهت و موبایل |
| [02-backoffice](02-backoffice/README-fa.md) | پنل مشترک۰.۲؛ ۲۷ صفحه و login؛ نمای مدیر/سلر و سه جهت |
| [03-analysis-v1-0](03-analysis-v1-0/README.md) | snapshot بسته تحلیل اولیه امروز؛ ۹۱ تسک، API، داده، مالی، قیمت، عملیات و گیت‌ها؛ با اصلاحات جاری خوانده شود |
| [04-current-spec](04-current-spec/README-fa.md) | نیازمندی‌های تازه، ۱۲ اصلاح تسک، ردیابی و ابهام باقی‌مانده؛ مرجع تعارض‌های مشخص‌شده |
| [IMAGE-INDEX](IMAGE-INDEX-fa.md) | تصاویر تازه و تفکیک تصاویر تاریخی |
| [VERIFICATION](VERIFICATION-fa.md) | بررسی نمونه/بسته و حدود نتیجه؛ به معنی تست محصول نیست |

نه تسک‌های محصول done شده‌اند و نه اتصال پیامک/پرداخت/برداشت واقعی ساخته شده است. قاعده صریح کاربر، پیشنهاد فنی و تصمیم باز در اسناد جدا هستند. واحدها و نرخ fixture تصمیم مالی تازه نیستند. کل پوشه و فایل ZIP همراه آن برای ارسال آماده‌اند؛ لازم نیست فایل‌های دیزاین را جدا جابه‌جا کنید.

ثبت مرکزی: DEC-028 / ACT-046. [دفتر نظر همین بسته](REVIEW-fa.md) برای اصلاحات بعدی و [فهرست تمام فایل‌ها](manifest.json) برای کنترل تحویل.
`);
write('START-HERE-fa.md',`# از اینجا شروع کنید

${header} | شناسه: CV-KIT-START

این بسته طراحی و تحلیل است؛ شروع پیاده‌سازی در پروژه مقصد با درخواست/دامنه مشخص تیم و رعایت AGENTS همان پروژه انجام شود. تهیه این پوشه مجوز تغییر ریپوی موجود، انتشار یا اتصال مالی نیست.

## ۱. مرجع معتبر هر موضوع

1. [قواعد پایه و دامنه](03-analysis-v1-0/01-scope-decisions.md) و [پرسش‌های پایه](03-analysis-v1-0/13-open-questions.md).
2. [اصلاحات جاری](04-current-spec/README-fa.md) و [ابهام تکمیلی](04-current-spec/open-items-fa.md). این اصلاحات فقط در موضوع‌های صریح جدول R2 روی متن قبلی اعمال می‌شوند؛ جدیدبودن بقیه فایل‌ها به‌تنهایی مرجعیت نمی‌دهد.
3. [معماری پیشنهادی](03-analysis-v1-0/02-architecture.md)، [مدل داده](03-analysis-v1-0/03-data-model.md)، [API](03-analysis-v1-0/04-api-contracts.md) و [گذار وضعیت‌ها](03-analysis-v1-0/18-state-transitions.md)؛ schema هویت/خصوصیت/وقفه با مشخصات جاری اصلاح شود.
4. [۹۱ تسک پایه](03-analysis-v1-0/task-index.json) + [۱۲ اصلاح قابل ادغام](04-current-spec/task-index-fa.md)؛ وضعیت todo. دامنه پایه کامل است، ۱۲ کار اخیر جای بقیه عملیات/مالی را نمی‌گیرد.

## ۲. دیزاین‌ها

[هاب طرح‌ها](index.html) → [سایت](01-storefront/index.html) یا [پنل](02-backoffice/index.html). A متعادل و شیک، B فروش‌محور، C لوکس و تصویری. پنل همان فرم‌های مشترک با مجوز و scope فروشگاه دارد. نوار نقش/حالت ابزار بازبینی است؛ داخل محصول منتشر نشود.

پنل: ۲۷ صفحه به‌علاوه ورود؛ [فهرست](02-backoffice/screen-index-fa.md). از خصوصیات و دسته → محصول/تنوع/عرضه → سکشن‌ها → سفارش و مالی مرور کنید. سایت: ده صفحه اصلی و حالت‌های طلا/اقساط/خطا/انتخاب‌ها؛ [راهنمای مسیر خرید](01-storefront/checkout-handoff-fa.md). دکمه افزودن به سبد همان استایل مرجع قبل/بعد افزودن را حفظ کرده است. [تصاویر جاری](IMAGE-INDEX-fa.md).

ورود نمونه پنل: **09000000001** و کد **000000**؛ همین USR-001 در نمای خریدار/مدیر/سلر. از پروفایل ایمیل operator@example.test، تأیید نمونه و یک رمز ساختگی ثبت کنید؛ بعد از خروج در همان تب، با موبایل یا ایمیل و همان رمز وارد شوید. Reload اطلاعات را پاک می‌کند. نمونه‌های سایت و پنل نشست واقعی مشترک ندارند. رمز واقعی وارد نکنید.

## ۳. ترتیب شروع اجرا

- گام صفر: مسئول هر حوزه و طرح منتخب ثبت؛ readiness هر مسیر و سؤال مسدودکننده بررسی شود. قرارداد واحد پول/گردکردن و امنیت از نمونه JS استنتاج نشود.
- مرحله۱: بک‌آفیس و سلر با داده آزمون، **تمام حوزه‌ها**: هویت/دسترسی، کاتالوگ/عرضه، قیمت/تخفیف، سکشن، سفارش، تأمین، چندمرسوله، پشتیبانی، واگذاری، کیف پول/پرداخت/برداشت/استرداد، تطبیق و تسویه. بخش وابسته به CHQ در حد draft/adapter روشن بماند؛ تصمیم مالی ساخته نشود.
- مرحله۲: بعد از پذیرش تمام دامنه مرحله۱، سایت از قراردادهای مشترک و همین دیزاین استفاده کند. اصلاحات فرانت حذف نشده‌اند؛ ترتیب اجرا عقب‌تر است.
- گیت مستقل اتصال: sandbox/provider و سپس انتشار واقعی با قرارداد/مجوز، شاهد فنی و اجازه صریح. mock success یا HTTP200 به معنی قبولی provider نیست.

[برنامه تحویل تفصیلی](03-analysis-v1-0/12-delivery-plan.md) و [کیفیت/عملیات](03-analysis-v1-0/11-quality-operations.md) همراه این ترتیب خوانده شوند. زمان و بودجه واقعی این بسته تعیین نمی‌کند.

## ۴. چه چیزهایی از اصلاحات نهایی‌اند و چه چیزهایی بازند؟

اصل یک User، موبایل نخست و ایمیل/یک رمز بعدی، آیکن upload+fallback، خصوصیت مشترک، رنگ/گارانتی در نمونه PDP، اکشن سریع، فرم typed سکشن و وقفه فقط فروشنده صریحاً اعلام شده‌اند. پیاده‌سازی schema/API، تأیید ایمیل، اندازه فایل، رفتار sticky و انتخاب A/B/C پیشنهاد/بازبینی‌اند. طلا/مالی مرکب/تخفیف ترکیبی/حمل واقعی/ارتقای نقش/host و الزامات انتشار هنوز CHQ دارند. [ردیابی](04-current-spec/traceability-fa.md).

## ۵. ورودی کار برای تیم یا AI agent

\`\`\`text
کار: CV-... به همراه اصلاح CV-R2-... مربوط؛ فقط scope همان کار.
قرارداد: مشخصات جاری + تحلیل پایه با اصلاحات صریح؛ تصمیم باز را حدس نزن.
پروژه مقصد و اجازه تغییر: توسط تیم تعیین شود؛ این پوشه طرح مستقل است.
وابستگی‌ها و پاسخ سؤال مانع: با شاهد مشخص شوند.
تحویل: کد/قرارداد محدود، تست معنادار، گزارش رفتار/فایل/نتیجه/ابهام.
mock، integration، browser، sandbox و production را جدا گزارش کن.
\`\`\`

معیار پایان هر کار: ACها با شاهد، مجوز و scope در سرور، خطا و حالت خالی، conflict/تکرار امن متناسب، سند و ردیابی همگام. [راهنمای agent](03-analysis-v1-0/16-agent-delivery.md) و [AGENTS تحویل](AGENTS.md).
`);
write('AGENTS.md',`# راهنمای تیم و agent برای این بسته

این پوشه خروجی مستقل طراحی/تحلیل و در وضعیت in-review است. محصول اجراشده نیست. ابتدا START-HERE-fa.md، سپس 04-current-spec/README-fa.md و سند/تسک مرتبط را بخوانید. 03-analysis-v1-0 سابقه پایه است؛ اصلاحات R2 فقط در موارد صریح روی آن اعمال می‌شوند. هیچ سؤال مالی/امنیتی را از fixture یا ظاهر حل نکنید.

زبان توضیح فارسی، فایل و شناسه انگلیسی kebab-case؛ UTF-8. وضعیت سند و اجرای task جدا؛ approved فقط با تصمیم صریح مالک. User خریدار/ادمین/عضو سلر یک هویت؛ Store جدا. وقفه فقط فروشنده؛ قیمت/stock روی Offer. OTP و شارژ با پرداخت سفارش یکی نیستند. clock کمیسیون و جلسه خرید مستقل‌اند.

تحویل فعلی مجوز ویرایش D:\\Cherriva یا GooshiShop، نصب dependency، اتصال provider، انتقال وجه یا deploy نیست. در کار پیاده‌سازی، اجازه مشخص و AGENTS مقصد ملاک است. تقدم: تمام پنل/سلر با داده آزمون، سپس سایت؛ اتصال واقعی گیت جدا. دارایی‌ها محلی و اطلاعات ساختگی‌اند. رمز واقعی یا اطلاعات شخصی در نمونه نگذارید.

اصلاح آتی: نسخه و REVIEW و task/traceability و فهرست تصویر را همگام کنید؛ قبل از جایگزینی baseline، snapshot نگه دارید. تصاویر/validation قدیمی را شاهد نسخه جدید معرفی نکنید. معیارهای تحویل و محدودیت‌ها در VERIFICATION-fa.md هستند.
`);
write('REVIEW-fa.md',`# دفتر بازبینی بسته یکپارچه

${header} | شناسه: CV-KIT-REVIEW

ساخت بسته انجام شده؛ انتخاب نهایی بصری، پذیرش قراردادها، انتساب کارها و اجرای محصول بازند. فقط متن صریح الیاس در R2 به عنوان قاعده اعلام‌شده ثبت شده است.

| شناسه نظر | تاریخ | نویسنده واقعی | صفحه/تسک | نظر و دلیل | پاسخ/تصمیم/وضعیت |
| --- | --- | --- | --- | --- | --- |
| — | — | — | — | هنوز نظر تیم در این فایل ثبت نشده است | باز |

برای نظر، شناسه صفحه BO یا page/direction/state سایت، تصویر، رفتار فعلی، پیشنهاد و دلیل را ثبت کنید. نظر حل‌شده حذف نشود؛ به تصمیم جایگزین و نسخه بعدی لینک دهید. امضا یا تأیید اشخاص ساخته نشود.
`);
write('index.html',`<!doctype html><html lang="fa" dir="rtl"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Cherriva · بسته طراحی و شروع</title><style>@font-face{font-family:v;src:url('02-backoffice/assets/vazirmatn-regular.woff2')}@font-face{font-family:v;src:url('02-backoffice/assets/vazirmatn-bold.woff2');font-weight:700}*{box-sizing:border-box}body{margin:0;background:#f7f4f3;color:#271c22;font-family:v,Tahoma;line-height:1.9}header{background:#8e1642;color:white;padding:52px max(6vw,24px)}header img{width:140px;background:white;border-radius:16px;padding:12px}h1{font-size:clamp(24px,4vw,44px);margin:20px 0 6px}header p{max-width:800px;color:#ffe4ed}main{max-width:1240px;margin:auto;padding:30px 24px}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(270px,1fr));gap:20px}article{background:white;border:1px solid #e8dce2;border-radius:22px;padding:26px}h2{margin:5px 0;font-size:22px}a{color:#961842;text-decoration:none}a:hover{text-decoration:underline}.go{display:inline-block;background:#a51b4b;color:white;border-radius:12px;padding:12px 20px;margin-top:14px}.badge{font-size:12px;color:#7a5264}.links{display:flex;gap:18px;flex-wrap:wrap;margin:26px 0}.note{background:#fff1dc;border-radius:16px;padding:20px;margin-top:24px}footer{font-size:13px;color:#775f6b;margin-top:24px}</style><header><img src="02-backoffice/assets/logo.svg" alt="چریوا"><h1>از طراحی تا شروع ساخت چریوا</h1><p>یک بسته برای سایت، بک‌آفیس و پنل فروشنده. سه جهت بصری برای مقایسه، قراردادهای روشن و کارهای قابل تحویل.</p><span>۱۷ نوامبر ۲۰۲۶ · نسخه ۱.۱ · در حال بازبینی</span></header><main><div class="grid"><article><div class="badge">دیزاین سایت · نسخه۰.۴</div><h2>تجربه خرید چریوا</h2><p>خانه، دسته‌بندی، محصول، سبد و مسیر خرید تا نتیجه پرداخت؛ ورود و پروفایل با اصلاحات جدید.</p><a class="go" href="01-storefront/index.html">بازکردن سه طرح سایت ←</a></article><article><div class="badge">بک‌آفیس و سلر · نسخه۰.۲</div><h2>پنل مشترک با دسترسی متفاوت</h2><p>۲۷ صفحه؛ هویت، کاتالوگ، قیمت، سکشن‌ها، سفارش، تأمین، ارسال و مالی. همان کاربر، چند کاربرد.</p><a class="go" href="02-backoffice/index.html">بازکردن طراحی پنل ←</a></article><article><div class="badge">تحلیل و برنامه پیاده‌سازی · نسخه۱.۱</div><h2>راهنمای تیم فنی</h2><p>۹۱ تسک پایه و ۱۲ اصلاح قابل ادغام؛ قواعد بیزینسی، داده، API، مجوز، خطا، وابستگی و پذیرش.</p><a class="go" href="START-HERE-fa.md">راهنمای شروع ←</a></article></div><nav class="links"><a href="README-fa.md">معرفی بسته</a><a href="04-current-spec/README-fa.md">اصلاحات جاری</a><a href="04-current-spec/task-index-fa.md">تسک‌های اصلاحی</a><a href="03-analysis-v1-0/README.md">تحلیل پایه امروز</a><a href="IMAGE-INDEX-fa.md">همه تصاویر</a><a href="VERIFICATION-fa.md">گزارش بررسی</a><a href="REVIEW-fa.md">ثبت نظر</a></nav><div class="note">ترتیب اجرا: تمام بک‌آفیس و سلر با داده آزمون، سپس سایت مشتری. اتصال واقعی پرداخت، پیامک و انتشار، پذیرش مستقل دارند. این نمونه‌ها حساب یا تراکنش واقعی ایجاد نمی‌کنند.</div><footer>متعادل و شیک · فروش‌محور · لوکس و تصویری — انتخاب نهایی هنوز ثبت نشده است. برای خواندن MD از ویرایشگر Markdown تیم استفاده کنید.</footer></main></html>`);
console.log(JSON.stringify({amendments:tasks.length,filesWritten:'current specs, tasks, active design docs and root guide'}));
