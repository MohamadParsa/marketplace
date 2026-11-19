/* Design fixtures only. No live provider, customer data, or authentication. */
window.CV = {
  pages: [
    ['dashboard','نمای کلی بازارگاه','یک نگاه روشن به امروز، کارهای در انتظار و جریان بازارگاه','نمای کلی','dashboard','BO-01',''],
    ['users','کاربران','حساب‌ها، وضعیت دسترسی و محدوده فعالیت هر همکار','مدیریت دسترسی','users','BO-02','کاربر جدید'],
    ['roles','نقش‌ها و دسترسی‌ها','نقش قابل تعریف، مجوز هر عملیات و محدوده مستقل فروشگاه','مدیریت دسترسی','shield','BO-03','نقش جدید'],
    ['stores','فروشگاه‌ها','فروشندگان منتخب و سیاست عرضه، انتشار و نرخ هر فروشگاه','کاتالوگ و فروشندگان','store','BO-04','فروشگاه جدید'],
    ['categories','دسته‌بندی‌ها','یک ساختار درختی برای تمام ویترین‌های چریوا','کاتالوگ و فروشندگان','category','BO-05','دسته جدید'],
    ['brands','برندها','برند، محتوای صفحه و نمایش آن در ویترین دیجیتال','کاتالوگ و فروشندگان','award','BO-06','برند جدید'],
    ['attributes','خصوصیات و تنوع‌ها','فیلترهای مرتبط دسته و ویژگی‌های مهم کارت محصول','کاتالوگ و فروشندگان','sliders','BO-07','خصوصیت جدید'],
    ['products','محصولات و محتوا','کاتالوگ مشترک، نسخه محتوا و صف بررسی انتشار','کاتالوگ و فروشندگان','box','BO-08','محصول جدید'],
    ['offers','عرضه و موجودی','قیمت، موجودی و زمان آماده‌سازی هر تنوع از هر فروشنده','کاتالوگ و فروشندگان','layers','BO-09','عرضه جدید'],
    ['gold-rates','نرخ پایه طلا','نرخ مشترک چریوا و نرخ اختصاصی فروشندگان، با نسخه و اعتبار','قیمت و فروش','gold','BO-10','ثبت نرخ جدید'],
    ['gold-policies','مدل قیمت‌گذاری طلا','انتخاب منبع نرخ و پیش‌نمایش اجزای قیمت برای هر سلر','قیمت و فروش','calculator','BO-11','سیاست جدید'],
    ['sale-methods','روش‌های فروش','فروش نقدی، اقساطی و اعتباری با درگاه و درصد مستقل','قیمت و فروش','card','BO-12','روش فروش جدید'],
    ['category-markups','مارکاپ دسته‌بندی','نزدیک‌ترین قانون دسته، جایگزین درصد پیش‌فرض روش فروش','قیمت و فروش','percent','BO-13','قانون جدید'],
    ['commissions','قواعد کمیسیون','درآمد چریوا؛ مستقل از مارکاپ روش فروش و هزینه تخفیف','قیمت و فروش','receipt','BO-14','قانون کمیسیون'],
    ['promotions','تخفیف‌ها و کدها','پیشنهادهای معتبر و منبع هزینه هر تخفیف','قیمت و فروش','tag','BO-15','تخفیف جدید'],
    ['pages','صفحه‌ها و سکشن‌ها','محتوای قابل چیدمان، پیش‌نمایش و انتشار مستقل','محتوا و ویترین','layout','BO-16','افزودن سکشن'],
    ['shipping','حمل و تقویم','پوشش مقصد، تعرفه فروشگاه و روزهای کاری آماده‌سازی','عملیات و پشتیبانی','truck','BO-17','روش ارسال جدید'],
    ['orders','خریدها و سفارش‌ها','یک خرید، سفارش‌های مستقل فروشندگان و یک تاریخچه روشن','عملیات و پشتیبانی','receipt','BO-18',''],
    ['shipments','مرسوله‌ها','اقلام آماده، بسته‌ها و رهگیری اعزام هر فروشنده','عملیات و پشتیبانی','truck','BO-19','مرسوله جدید'],
    ['support','پشتیبانی و پرونده‌ها','عدم تأمین، نرسیدن و کسری؛ با مسئول و مهلت رسیدگی','عملیات و پشتیبانی','support','BO-20',''],
    ['assignments','تأمین جایگزین','واگذاری مقدار مشخص، با حفظ مبلغ پرداختی مشتری','عملیات و پشتیبانی','swap','BO-21',''],
    ['finance','عملیات مالی','شارژ، برداشت، استرداد و دفتر مالی؛ هر جریان مستقل','مالی و پیگیری','wallet','BO-22',''],
    ['settlements','تسویه فروشندگان','تخصیص قلمی، مانده قابل تسویه و بدهی‌های قابل پیگیری','مالی و پیگیری','bank','BO-23','پیش‌نمایش تسویه'],
    ['operations','پیگیری و گزارش','تراکنش‌های نامشخص، اعلان‌ها و ردپای عملیات','مالی و پیگیری','activity','BO-24',''],
    ['moderation','نظرها و پرسش‌ها','محتوای کاربران و پاسخ‌های قابل انتشار','محتوا و ویترین','chat','BO-25',''],
    ['settings','تنظیمات ویترین','هویت، منوها، تماس، مجوزها و خدمات قابل نمایش','محتوا و ویترین','settings','BO-26','ذخیره پیش‌نویس']
  ].map(([id,title,description,group,icon,code,primary])=>({id,title,description,group,icon,code,primary})),
  accounts: {
    admin:{name:'مدیر نمونه',label:'مدیر کل',initial:'م',phone:'09000000001',store:null},
    seller:{name:'گالری آوای طلا',label:'فروشنده',initial:'آ',phone:'09000000006',store:'S1'},
    finance:{name:'همکار مالی نمونه',label:'مالی',initial:'ف',phone:'09000000004',store:null},
    support:{name:'همکار پشتیبانی نمونه',label:'پشتیبانی',initial:'پ',phone:'09000000003',store:null},
    viewer:{name:'مشاهده‌گر نمونه',label:'فقط مشاهده',initial:'ن',phone:'09000000005',store:null}
  },
  access:{
    admin:['*'],
    seller:['dashboard','stores','offers','shipping','orders','shipments','assignments','settlements'],
    finance:['dashboard','gold-rates','gold-policies','sale-methods','category-markups','commissions','promotions','finance','settlements','operations'],
    support:['dashboard','stores','products','orders','shipments','support','assignments','operations','moderation'],
    viewer:['dashboard','stores','categories','brands','attributes','products','offers','orders','shipments','operations']
  },
  stores:[
    {id:'S1',name:'گالری آوای طلا',category:'طلا و سکه',code:'SEL-001',owner:'حساب سلر ۰۱',active:true,mode:'کاتالوگ مشترک',rate:'مشترک چریوا',offers:3,color:'gold'},
    {id:'S2',name:'دیجیتال آبان',category:'موبایل و لوازم جانبی',code:'SEL-002',owner:'حساب سلر ۰۲',active:true,mode:'کاتالوگ مشترک',rate:'—',offers:3,color:'violet'},
    {id:'S3',name:'زرین‌قطعه نمونه',category:'طلای فیزیکی',code:'SEL-003',owner:'حساب سلر ۰۳',active:true,mode:'کاتالوگ مستقل',rate:'اختصاصی فروشنده',offers:2,color:'gold'},
    {id:'S4',name:'تأمین کالای چریوا',category:'چنددسته‌ای',code:'SEL-004',owner:'حساب تأمین نمونه',active:false,mode:'کاتالوگ مشترک',rate:'—',offers:0,color:'rose'}
  ],
  products:[
    {id:'P1',name:'شمش طلای ۱ گرمی',cat:'طلا / شمش',brand:'برند نمونه طلا',sku:'GLD-BAR-1',img:'gold-bar.jpg',status:'published',variants:2,offers:2,kind:'gold'},
    {id:'P2',name:'آیفون ۱۶، حافظه ۱۲۸ گیگابایت',cat:'دیجیتال / موبایل',brand:'Apple',sku:'APL-IP16-128',img:'iphone.jpg',status:'published',variants:3,offers:1,kind:'digital'},
    {id:'P3',name:'طلای پارسیان ۵۰۰ سوت',cat:'طلا / پارسیان',brand:'برند نمونه طلا',sku:'GLD-PRS-500',img:'parsian.png',status:'pending',variants:1,offers:1,kind:'gold'},
    {id:'P4',name:'گلکسی S25، حافظه ۲۵۶ گیگابایت',cat:'دیجیتال / موبایل',brand:'Samsung',sku:'SMS-S25-256',img:'samsung.jpg',status:'published',variants:2,offers:1,kind:'digital'},
    {id:'P5',name:'ایرپاد پرو، نسل دوم',cat:'دیجیتال / هدفون',brand:'Apple',sku:'APL-AIRP-2',img:'airpods.jpg',status:'pending',variants:1,offers:1,kind:'digital'},
    {id:'P6',name:'قطعه طلای آب‌شده ۵ گرمی',cat:'طلا / آب‌شده',brand:'فروشنده نمونه',sku:'GLD-MLT-5',img:'gold-bar.jpg',status:'draft',variants:1,offers:1,kind:'gold'},
    {id:'P7',name:'شیائومی ۱۵، حافظه ۲۵۶ گیگابایت',cat:'دیجیتال / موبایل',brand:'Xiaomi',sku:'XMI-15-256',img:'xiaomi.png',status:'published',variants:2,offers:1,kind:'digital'},
    {id:'P8',name:'سکه پارسیان ۱ گرمی',cat:'طلا / پارسیان',brand:'برند نمونه طلا',sku:'GLD-PRS-1',img:'parsian.png',status:'published',variants:1,offers:1,kind:'gold'}
  ],
  offers:[
    {id:'O1',product:'P1',store:'S1',variant:'۱ گرم · عیار ۹۹۵',cash:8950000,stock:12,reserved:2,lead:'۱ روز کاری',mode:'فرمول طلا',status:'active'},
    {id:'O2',product:'P1',store:'S3',variant:'۱ گرم · عیار ۹۹۵',cash:8980000,stock:6,reserved:1,lead:'۱ روز کاری',mode:'فرمول طلا',status:'active'},
    {id:'O3',product:'P2',store:'S2',variant:'۱۲۸ گیگ · مشکی',cash:78900000,stock:4,reserved:1,lead:'۲ روز کاری',mode:'قیمت ثابت',status:'active'},
    {id:'O4',product:'P3',store:'S1',variant:'۵۰۰ سوت · عیار ۷۵۰',cash:4560000,stock:8,reserved:0,lead:'۱ روز کاری',mode:'قیمت ثابت',status:'pending'},
    {id:'O5',product:'P4',store:'S2',variant:'۲۵۶ گیگ · نقره‌ای',cash:65400000,stock:3,reserved:1,lead:'۲ روز کاری',mode:'قیمت ثابت',status:'active'},
    {id:'O6',product:'P5',store:'S2',variant:'سفید · USB-C',cash:16400000,stock:5,reserved:0,lead:'۲ روز کاری',mode:'قیمت ثابت',status:'pending'},
    {id:'O7',product:'P6',store:'S3',variant:'۵ گرم · عیار نمونه',cash:null,stock:2,reserved:0,lead:'۱ روز کاری',mode:'فرمول طلا',status:'draft'},
    {id:'O8',product:'P8',store:'S1',variant:'۱ گرم · عیار ۷۵۰',cash:9120000,stock:7,reserved:1,lead:'۱ روز کاری',mode:'قیمت ثابت',status:'active'}
  ],
  orders:[
    {id:'ORD-1048-1',purchase:'CH-1048',store:'S1',product:'P1',city:'تهران',qty:2,amount:17900000,payment:'paid',status:'supply-pending',method:'نقدی',time:'۱۴:۰۵',supply:false},
    {id:'ORD-1048-2',purchase:'CH-1048',store:'S2',product:'P2',city:'تهران',qty:1,amount:78900000,payment:'paid',status:'preparing',method:'نقدی',time:'۱۴:۰۵',supply:true},
    {id:'ORD-1047-1',purchase:'CH-1047',store:'S3',product:'P1',city:'شیراز',qty:1,amount:9518800,payment:'paid',status:'supply-issue',method:'اعتباری',time:'۱۳:۴۲',supply:false},
    {id:'ORD-1046-1',purchase:'CH-1046',store:'S2',product:'P4',city:'اصفهان',qty:1,amount:65400000,payment:'paid',status:'dispatched',method:'نقدی',time:'۱۲:۵۰',supply:true},
    {id:'ORD-1045-1',purchase:'CH-1045',store:'S1',product:'P8',city:'مشهد',qty:1,amount:9120000,payment:'paid',status:'delivered',method:'نقدی',time:'۱۲:۱۷',supply:true},
    {id:'ORD-1044-1',purchase:'CH-1044',store:'S2',product:'P2',city:'کرج',qty:1,amount:78900000,payment:'paid',status:'supply-pending',method:'اقساطی',time:'۱۱:۵۵',supply:false}
  ],
  cases:[
    {id:'CASE-031',order:'ORD-1047-1',store:'S3',reason:'عدم تأمین',title:'نیاز به تعیین تکلیف شمش ۱ گرمی',deadline:'۳۵ دقیقه باقی‌مانده',status:'open',owner:'بدون مسئول'},
    {id:'CASE-030',order:'ORD-1044-1',store:'S2',reason:'کسری تعداد',title:'بررسی مقدار قابل تأمین',deadline:'در حال رسیدگی',status:'in-progress',owner:'پشتیبانی نمونه'},
    {id:'CASE-029',order:'ORD-1045-1',store:'S1',reason:'پیگیری دریافت',title:'دریافت مرسوله تأیید شد',deadline:'بسته‌شده',status:'closed',owner:'پشتیبانی نمونه'}
  ],
  transfers:[
    {id:'PAY-2081',type:'شارژ کیف پول',source:'مشتری آزمایشی ۰۱',amount:18000000,status:'unknown',when:'۱۴:۰۸',ref:'DEMO-TXN-2081'},
    {id:'PAY-2080',type:'شارژ کیف پول',source:'مشتری آزمایشی ۰۲',amount:79000000,status:'success',when:'۱۳:۵۸',ref:'DEMO-TXN-2080'},
    {id:'WDR-0092',type:'برداشت موجودی',source:'مشتری آزمایشی ۰۳',amount:2500000,status:'review',when:'۱۳:۵۲',ref:'در انتظار بررسی'},
    {id:'RFD-0038',type:'استرداد سفارش',source:'CH-1043',amount:9120000,status:'processing',when:'۱۳:۴۵',ref:'DEMO-RFD-0038'},
    {id:'SET-0021',type:'تسویه فروشنده',source:'گالری آوای طلا',amount:28000000,status:'unknown',when:'۱۳:۳۰',ref:'DEMO-SET-0021'},
    {id:'PAY-2079',type:'شارژ کیف پول',source:'مشتری آزمایشی ۰۴',amount:5000000,status:'failed',when:'۱۳:۲۰',ref:'DEMO-TXN-2079'}
  ],
  notifications:[{id:'N1',kind:'اعلان خرید',target:'CH-1048',status:'success',attempts:1},{id:'N2',kind:'عدم تأمین',target:'CASE-031',status:'failed',attempts:3},{id:'N3',kind:'اعلان ارسال',target:'ORD-1046-1',status:'pending',attempts:1}],
  audit:[
    {action:'نرخ مشترک طلا به‌روزرسانی شد',who:'مدیر نمونه',ref:'RATE-034 · نسخه جدید',time:'۱۴:۰۰',store:null,icon:'gold'},
    {action:'سفارش وارد آماده‌سازی شد',who:'دیجیتال آبان',ref:'ORD-1048-2',time:'۱۳:۵۸',store:'S2',icon:'box'},
    {action:'پرونده عدم تأمین ایجاد شد',who:'سیستم نمونه',ref:'CASE-031',time:'۱۳:۴۲',store:'S3',icon:'support'},
    {action:'دریافت مرسوله تأیید شد',who:'پشتیبانی نمونه',ref:'ORD-1045-1',time:'۱۳:۳۰',store:'S1',icon:'check'},
    {action:'عرضه شمش به‌روزرسانی شد',who:'گالری آوای طلا',ref:'O1 · موجودی ۱۲',time:'۱۳:۱۵',store:'S1',icon:'layers'}
  ],
  roles:[{id:'admin',name:'مدیر کل',desc:'دسترسی سراسری بازارگاه',count:62},{id:'catalog',name:'مدیر کاتالوگ',desc:'محصول، دسته و انتشار محتوا',count:18},{id:'seller',name:'فروشنده منتخب',desc:'عملیات فروشگاه منتسب',count:12},{id:'finance',name:'همکار مالی',desc:'مشاهده و اقدامات مالی مجاز',count:15},{id:'support',name:'پشتیبانی',desc:'سفارش و پرونده‌های مجاز',count:10},{id:'viewer',name:'مشاهده‌گر',desc:'بدون اختیار نوشتن',count:8}],
  methods:[{id:'cash',name:'فروش نقدی',provider:'درگاه نقدی نمونه',percent:0,icon:'wallet',status:'active'},{id:'installment',name:'فروش اقساطی',provider:'اسنپ‌پی · نمونه قرارداد',percent:3,icon:'calendar',status:'draft'},{id:'loan',name:'فروش با وام',provider:'دیجی‌پی · مثال طراحی',percent:4,icon:'bank',status:'draft'}],
  markups:[{id:'M1',category:'طلا',path:'چریوا / طلا',method:'اقساطی',percent:6},{id:'M2',category:'شمش',path:'چریوا / طلا / شمش',method:'اقساطی',percent:5},{id:'M3',category:'پارسیان',path:'چریوا / طلا / پارسیان',method:'اقساطی',percent:0}],
  sections:[{id:'SC1',name:'بنر اصلی',type:'hero',desc:'تصویر دسکتاپ و موبایل · یک مقصد',icon:'image',enabled:true},{id:'SC2',name:'دسته‌های دارای موجودی',type:'category-icons',desc:'منبع خودکار · فقط دسته قابل خرید',icon:'category',enabled:true},{id:'SC3',name:'پیشنهادهای ویژه',type:'product-rail',desc:'تخفیف معتبر · حداکثر ۱۰ محصول',icon:'tag',enabled:true},{id:'SC4',name:'منتخب دیجیتال',type:'product-rail',desc:'دسته دیجیتال · مشاهده همه',icon:'mobile',enabled:true},{id:'SC5',name:'منتخب طلا',type:'product-rail',desc:'دسته طلا · مشاهده همه',icon:'gold',enabled:true},{id:'SC6',name:'برندهای دیجیتال',type:'brand-rail',desc:'اپل، سامسونگ و شیائومی',icon:'award',enabled:true}],
  publications:{main:3,gold:1,digi:2},
  dailySales:[{day:'چهارشنبه',value:39},{day:'پنج‌شنبه',value:51},{day:'جمعه',value:29},{day:'شنبه',value:68},{day:'یکشنبه',value:61},{day:'دوشنبه',value:88},{day:'سه‌شنبه',value:72}],
  permissionGroups:[['کاتالوگ و عرضه',['مشاهده محصول','ویرایش محصول','انتشار محصول','ویرایش موجودی']],['سفارش و ارسال',['مشاهده سفارش','تأیید تأمین','ثبت اعزام','رسیدگی پرونده']],['مالی',['مشاهده صورتحساب','بررسی برداشت','اجرای استرداد','اجرای تسویه']],['محتوا و دسترسی',['ویرایش سکشن','انتشار صفحه','مشاهده کاربران','مدیریت نقش‌ها']]],
  status:{published:['منتشرشده','green'],pending:['در انتظار بررسی','amber'],draft:['پیش‌نویس','gray'],active:['فعال','green'],inactive:['غیرفعال','gray'],'supply-pending':['منتظر تأمین','amber'],preparing:['آماده‌سازی','violet'],'supply-issue':['کسری تأمین','rose'],dispatched:['ارسال‌شده','violet'],delivered:['تحویل‌شده','green'],paid:['پرداخت قطعی','green'],open:['باز','amber'],'in-progress':['در حال رسیدگی','violet'],closed:['بسته‌شده','gray'],unknown:['در حال بررسی','amber'],success:['موفق قطعی','green'],failed:['ناموفق','red'],review:['نیازمند تأیید','rose'],processing:['در حال پردازش','violet']}
};
