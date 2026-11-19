const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
const root=path.resolve(__dirname,'..'),analysis='D:/Marketplace/15-technical-handoff/v1.0-2026-11-17';
const box={window:{}};vm.runInNewContext(fs.readFileSync(path.join(root,'data.js'),'utf8'),box);const CV=box.window.CV;
const write=(name,text)=>{const file=path.join(root,name);fs.mkdirSync(path.dirname(file),{recursive:true});fs.writeFileSync(file,text,'utf8');};
const read=n=>fs.readFileSync(path.join(analysis,n),'utf8');
const stamp='نسخه: ۰.۱ | تاریخ: ۲۰۲۶-۱۱-۱۷ | وضعیت: in-review | مالک تصمیم: الیاس؛ بازبین فنی پیشنهادی: راهبر فنی، انتساب شخص باز';
for(const [src,dest] of [['05-backoffice-screens.md','backoffice-analysis-fa.md'],['01-scope-decisions.md','scope-decisions-fa.md'],['13-open-questions.md','open-questions-fa.md']])write('references/'+dest,'> کپی مرجع تحلیل نسخه ۱.۰ مورخ ۲۰۲۶-۱۱-۱۷؛ برای خواندن مستقل بسته طراحی. لینک‌های بسته اصلی به ارجاع متنی تبدیل شده‌اند. این فایل با بازبینی طرح، تصمیم‌های تحلیل را تغییر نمی‌دهد.\n\n'+read(src).replace(/\[([^\]]+)\]\(([^)]+)\)/g,'$1 (مرجع بسته فنی: $2)'));
const specs=read('05-backoffice-screens.md').split(/\r?\n/).filter(x=>/^\| BO-\d/.test(x)).map(x=>x.split('|').slice(1,-1).map(x=>x.trim()));
const mapping=[['CV-FND-07'],['CV-IAM-04','CV-IAM-05'],['CV-IAM-03','CV-IAM-05'],['CV-IAM-06'],['CV-CAT-01'],['CV-CAT-03'],['CV-CAT-02','CV-CAT-08'],['CV-CAT-04','CV-CAT-07','CV-CAT-08'],['CV-CAT-05','CV-CAT-06'],['CV-PRC-03'],['CV-PRC-04'],['CV-PRC-01','CV-PRC-06'],['CV-PRC-02','CV-PRC-06'],['CV-PRC-05'],['CV-DSC-01','CV-DSC-02','CV-DSC-05'],['CV-CMS-01','CV-CMS-02','CV-CMS-04','CV-CMS-05'],['CV-OPS-05','CV-OPS-06'],['CV-OPS-01','CV-OPS-02'],['CV-OPS-07'],['CV-OPS-03','CV-OPS-08','CV-OPS-09'],['CV-OPS-04'],['CV-FIN-01','CV-FIN-04','CV-FIN-05','CV-FIN-08'],['CV-FIN-06','CV-FIN-08'],['CV-FND-06','CV-FIN-07','CV-INT-04'],['CV-WEB-15','CV-WEB-16'],['CV-CMS-03','CV-WEB-18']];
const flows={
dashboard:'ورود مجاز → سنجه‌های همان محدوده → صف کار → سفارش/پرونده/عرضه مرتبط → اقدام در صفحه تخصصی. نمودار، fixture بصری است و گزارش مالی محسوب نمی‌شود.',
users:'جست‌وجوی همکار → مشاهده نقش و فروشگاه → ویرایش پیش‌نویس انتساب → کنترل اختیار سمت سرور و audit در محصول. تغییر اختیار واقعی در نمونه وجود ندارد.',
roles:'انتخاب preset → مشاهده ماتریس → ساخت نام نقش سفارشی → انتخاب مجوزهای نمونه → ذخیره در حافظه → بازبینی. ماتریس نهایی تابع CHQ-06 است.',
stores:'فهرست فروشنده → پرونده → مدل کاتالوگ و نرخ → مدارک/صلاحیت → بررسی فعال‌سازی. تمام مدارک واقعی و حساب بانکی در این نمونه حذف شده‌اند.',
categories:'انتخاب گره درخت → والد/slug/ویترین → متن PLP و خصوصیات کارت → اعتبارسنجی چرخه در محصول. انتقال دسته فعال تا تعیین CHQ-09 اجرا نشود.',
brands:'انتخاب برند → نام/slug/لوگو/alt → متن و SEO → وضعیت انتشار. لوگوی جدید در نمونه upload نمی‌شود.',
attributes:'خصوصیت → نوع داده و واحد → مقادیر مجاز → فیلتر/تنوع/کارت → بررسی انتشار. تغییر مخرب نوع در محصول باید کنترل شود.',
products:'محصول → دسته/برند/SKU → محتوای مشترک → معرفی/SEO/رسانه → بررسی انتشار؛ عرضه و قیمت فرم مستقل دارند.',
offers:'عرضه همان فروشگاه → تنوع → قیمت نقدی یا مدل فرمول → موجودی کل/رزرو → آماده‌سازی → ذخیره محلی؛ موجودی زیر رزرو رد می‌شود.',
'gold-rates':'انتخاب منبع مشترک یا اختصاصی → نسخه و واحد/عیار/مأخذ/اعتبار → ثبت draft → تأیید صاحب اختیار در محصول. اجرای نرخ تا CHQ-03 مسدود است.',
'gold-policies':'انتخاب سلر → منبع نرخ → مدل قیمت → پارامترها/نسخه → breakdown → فعال‌سازی پس از CHQ-03. مقدار مجهول صفر فرض نمی‌شود.',
'sale-methods':'انتخاب روش → نوع و درگاه‌ها → مارکاپ پیش‌فرض → preview پس از تخفیف → افزودن مستقل ارسال؛ قابلیت درگاه واقعی تابع قرارداد است.',
'category-markups':'انتخاب دسته و روش → درصد شامل صفر → preview زنجیره دسته → نزدیک‌ترین قانون، وگرنه پیش‌فرض. درصدها جایگزین‌اند و جمع نمی‌شوند.',
commissions:'قرارداد → محدوده و شرایط → درصد/ثابت و شیوه اخذ → بازه مستقل قانون → preview snapshot؛ ساعت و tie-break هنوز CHQ-14 دارد.',
promotions:'نوع محصول/کوپن → هدف و سقف/بازه → منبع هزینه → محدودیت مصرف → preview؛ stacking و تخصیص مشترک تا CHQ-04 فعال نشود.',
pages:'انتخاب اصلی/gold/digi → افزودن نوع سکشن → ترتیب/نمایش/عنوان → preview → revision نمایشی. تنظیمات هر قلمرو مستقل در حافظه نمونه‌اند.',
shipping:'روش فروشگاه → پوشش/تعرفه → تقویم و زمان آماده‌سازی → برآورد → بازبینی. وعده حمل طلا تا CHQ-07 فعال نشود.',
orders:'خرید مادر → سفارش فروشنده → اقلام و snapshot → تأیید یا گزارش تأمین → آماده‌سازی → مرسوله؛ حریم مشتری قبل از تأمین رعایت شود.',
shipments:'سفارش تأمین‌شده → مقادیر باقی‌مانده → بسته/روش → رهگیری → اعزام → تحویل. ساخت واقعی بسته و تقسیم مقدار در نمونه انجام نمی‌شود.',
support:'صف پرونده → علت/قلم/مقدار/مسئول → مدارک و مهلت → رسیدگی → پیشنهاد راه‌حل → نتیجه مستقل؛ refund اکشن مالی جداست.',
assignments:'پرونده کسری → مقدار واگذاری → تأمین‌کننده جایگزین → قرارداد snapshot → تأیید → ارسال. مبلغ مشتری ثابت می‌ماند.',
finance:'انتخاب جریان شارژ/برداشت/استرداد → بررسی مبلغ/مرجع/وضعیت → شواهد → اقدام مجاز. unknown به جای موفقیت دستی، به استعلام ارجاع می‌شود.',
settlements:'فروشگاه مجاز → صورت‌حساب قلمی → سهم قرارداد/بدهی → تخصیص → انتقال → شاهد؛ عدد نمونه اثبات محاسبه یا واریز نیست.',
operations:'انتخاب تطبیق/اعلان/audit → مرجع و تلاش‌ها → جزئیات شاهد → استعلام یا پیگیری → نتیجه مستقل. export واقعی در نمونه وجود ندارد.',
moderation:'محتوای منتظر → متن/مرجع → پاسخ یا دلیل رد → پیشنهاد انتشار → کنترل سیاست و مجوز در محصول. انتشار عمومی در نمونه اجرا نمی‌شود.',
settings:'انتخاب بخش هویت/منو/فوتر/نماد/خبرنامه/ویترین → draft → بازبینی واقعیت خدمات → انتشار مستقل در محصول.'
};
const qs={roles:'CHQ-06',users:'CHQ-06',stores:'CHQ-03/06/13',categories:'CHQ-09',attributes:'CHQ-03/09',products:'CHQ-03/09',offers:'CHQ-03', 'gold-rates':'CHQ-03','gold-policies':'CHQ-03','sale-methods':'CHQ-01/11','category-markups':'CHQ-01',commissions:'CHQ-01/14',promotions:'CHQ-04',pages:'CHQ-08/12',shipping:'CHQ-07',orders:'CHQ-01/02/05',shipments:'CHQ-07',support:'CHQ-05/07',assignments:'CHQ-01/07',finance:'CHQ-01/11',settlements:'CHQ-01/11/14',operations:'CHQ-11/13',moderation:'CHQ-10',settings:'CHQ-08/10/13'};
let index=`# نقشه صفحه‌ها و فرم‌ها\n\nشناسه: CVBO-MAP-01 | ${stamp}\n\n۲۶ صفحه پنل به علاوه ورود/OTP. صفحه‌ها طراحی پیشنهادی‌اند؛ فیلدهای تحلیل که به جزئیات بیشتر نیاز دارند، در شرح صفحه و [مرز تحویل](implementation-handoff-fa.md) مشخص‌اند.\n\n| شناسه | صفحه و شرح | مسیر نمونه | تسک‌های موجود تحلیل |\n| --- | --- | --- | --- |\n`;
let imageIndex=`# تصاویر طراحی\n\nشناسه: CVBO-IMG-01 | ${stamp}\n\nتصاویر از نمونه محلی با Playwright گرفته شده‌اند؛ شواهد اجرای محصول یا درگاه نیستند. دسکتاپ: عرض ۱۴۴۰، ارتفاع viewport ۱۰۰۰ و screenshot تمام صفحه. موبایل: ۳۹۰×۸۴۴ و تمام صفحه.\n\n| صفحه | A متعادل | B فروش‌محور | C لوکس |\n| --- | --- | --- | --- |\n| ورود | [A](output/playwright/login-a.png) | [B](output/playwright/login-b.png) | [C](output/playwright/login-c.png) |\n`;
for(const [i,p] of CV.pages.entries()){
 const spec=specs.find(s=>s[0].startsWith(p.code)); const filename=p.code.toLowerCase()+'-'+p.id+'.md',img=p.code.toLowerCase()+'-'+p.id;
 const accounts=Object.entries(CV.access).filter(([r,a])=>a.includes('*')||a.includes(p.id)).map(([r])=>CV.accounts[r].label).join('، ');
 index+=`| ${p.code} | [${p.title}](screens/${filename}) | \`#${p.id}\` | ${mapping[i].join('، ')} |\n`;
 imageIndex+=`| ${p.code} · ${p.title} | [A](output/playwright/${img}-a.png) | [B](output/playwright/${img}-b.png) | [C](output/playwright/${img}-c.png) |\n`;
 write('screens/'+filename,`# ${p.code} — ${p.title}\n\nشناسه سند: CVBO-${p.code} | ${stamp}\n\n## هدف بیزینسی\n\n${p.description}. این فرم باید اقدام بعدی و محدوده مسئولیت را روشن کند؛ ظاهر فرم جای تصمیم باز یا اجازه اجرای عملیات نیست.\n\n## داده و فیلدهای تحلیل\n\n${spec?.[1]||'شمارش صف‌ها و فعالیت‌های محدوده مجاز.'}\n\n## مسیر اصلی\n\n${flows[p.id]}\n\n## دسترسی\n\nعملیات پیشنهادی رجیستری: ${spec?.[2]||'dashboard.view'}\n\nنمای نقش‌های نمونه: ${accounts}. نقش‌های نمونه تعیین نهایی اختیار نیستند. هر عملیات سمت سرور مجوز مستقل می‌خواهد؛ سلر فقط فروشگاه منتسب، نه scope ارسالی دلخواه کلاینت.\n\n## اعتبارسنجی و حالات\n\n${spec?.[3]||'سنجه فقط از داده مجاز؛ آمار ساختگی نباید گزارش فروش معرفی شود.'}\n\nloading، empty، error و forbidden از نوار بررسی قابل مشاهده‌اند. جست‌وجو و وضعیت روی فهرست‌های جدول‌دار فعال‌اند. در محصول خطای اعتبارسنجی، conflict و retry امن به فرم متصل شوند؛ این نمونه همه قراردادهای backend را اجرا نمی‌کند.\n\n## پذیرش پیشنهادی تیم\n\n- مسیر اصلی بالا با داده آزمون و نقش مجاز قابل تکمیل باشد؛ داده یا عملیات خارج scope نمایش و اجرا نشود.\n- قواعد و محدودیت‌های همین صفحه و [مرجع تحلیل](../references/backoffice-analysis-fa.md) در API و UI یکسان باشند.\n- خطا و داده خالی قابل فهم باشند؛ متن طولانی، رقم فارسی و نمای موبایل آزمون شوند.\n\n## اتصال به کار توسعه\n\n${mapping[i].join('، ')} در بسته تحلیل فنی نسخه ۱.۰. شناسه‌ها موجودند؛ در این بسته تسک جدید یا وضعیت done ساخته نشده است. حدود تعامل نمونه در [تحویل فنی](../implementation-handoff-fa.md) آمده است.\n\n${qs[p.id]?`تصمیم‌های باز مؤثر: **${qs[p.id]}**؛ [دفتر](../open-decisions-fa.md). بخش وابسته بدون پاسخ، اجرایی/مصوب فرض نشود.`:'این صفحه تصمیم مالی تازه‌ای ندارد؛ قواعد مشترک مجوز و scope همچنان لازم‌اند.'}\n\n## تصویر و مسیر\n\n[جهت A](../output/playwright/${img}-a.png) · [جهت B](../output/playwright/${img}-b.png) · [جهت C](../output/playwright/${img}-c.png)\n\nمسیر نمونه: \`#${p.id}\`. [بازگشت به نقشه](../screen-index-fa.md)\n`);
}
write('screen-index-fa.md',index+'\nفهرست نقش‌ها فقط preset نمایشی است؛ برای جزئیات به [قواعد دسترسی](references/backoffice-analysis-fa.md) و CHQ-06 رجوع کنید.\n');
imageIndex+='\n## موبایل و حالت‌های ویژه\n\n';
for(const id of ['login','dashboard','orders','pages','sale-methods','roles','finance'])imageIndex+=`- [موبایل ${id}](output/playwright/mobile-${id}.png)\n`;
for(const id of ['loading','empty','error','forbidden'])imageIndex+=`- [حالت ${id}](output/playwright/state-${id}.png)\n`;
for(const [id,title] of [['login-otp-a','مرحله OTP'],['seller-dashboard','داشبورد سلر'],['seller-order-private','حریم سفارش پیش از تأمین'],['payment-unknown','پرداخت نامشخص']])imageIndex+=`- [${title}](output/playwright/${id}.png)\n`;
write('image-index-fa.md',imageIndex);
const tasks=JSON.parse(read('task-index.json')).tasks;
const ids=new Set(mapping.flat());write('references/task-map.json',JSON.stringify({origin:'technical-handoff v1.0 / 2026-11-17',status:'read-only excerpt; not implementation progress',tasks:tasks.filter(t=>ids.has(t.id)).map(({id,title,priority,status,block})=>({id,title,priority,status,block}))},null,2));
console.log(JSON.stringify({pages:CV.pages.length,taskRefs:ids.size,missingTaskRefs:[...ids].filter(id=>!tasks.some(t=>t.id===id))}));
