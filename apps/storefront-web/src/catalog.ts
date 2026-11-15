// داده ساختگی نسخه بررسی؛ مبلغ نمایشی تومان و بدون اتصال مالی است.
export type Category = "all" | "digital" | "home" | "tea" | "gold";
export type Product = {
  id: string;
  name: string;
  subtitle: string;
  category: Category;
  image: string;
  price: number;
  sales: number;
  stock: number;
  colors: string[];
  description: string;
};
export const categories: { id: Category; name: string; icon: string }[] = [
  { id: "all", name: "همه کالاها", icon: "◈" },
  { id: "digital", name: "دنیای دیجیتال", icon: "⌘" },
  { id: "home", name: "خانه و زندگی", icon: "⌂" },
  { id: "tea", name: "چای و قهوه", icon: "♧" },
  { id: "gold", name: "طلا و زیورآلات", icon: "◇" },
];
export const products: Product[] = [
  {
    id: "headphones",
    name: "هدفون بی‌سیم استودیو",
    subtitle: "صدایی برای لحظه‌های خودت",
    category: "digital",
    image: "headphones",
    price: 2450000,
    sales: 95,
    stock: 8,
    colors: ["مشکی", "نقره‌ای"],
    description:
      "هدفون دورگوشی سبک برای موسیقی و استفاده روزمره. این کالا و مشخصات آن صرفاً نمونه طراحی هستند.",
  },
  {
    id: "cup",
    name: "ماگ سرامیکی مینیمال",
    subtitle: "شروع گرم یک روز تازه",
    category: "home",
    image: "cup",
    price: 385000,
    sales: 89,
    stock: 12,
    colors: ["سفید", "کرم"],
    description:
      "فرم ساده و دوست‌داشتنی برای میز کار و گوشه دنج خانه. تصویر و مشخصات در این نسخه نمایشی‌اند.",
  },
  {
    id: "tea",
    name: "چای دست‌چین بهاره",
    subtitle: "عطر یک مکث دل‌نشین · ۲۵۰ گرم",
    category: "tea",
    image: "tea",
    price: 295000,
    sales: 86,
    stock: 20,
    colors: ["بسته ۲۵۰ گرمی"],
    description:
      "یک انتخاب نمونه از دنیای چای، برای بررسی تجربه انتخاب کالا و فروشگاه. ادعای تجاری یا کیفیت واقعی مطرح نیست.",
  },
  {
    id: "watch",
    name: "ساعت مچی کلاسیک",
    subtitle: "سادگی، همیشه به‌وقت",
    category: "home",
    image: "watch",
    price: 1850000,
    sales: 80,
    stock: 5,
    colors: ["سفید", "مشکی"],
    description:
      "ساعت نمونه با طراحی ساده و بند قابل انتخاب. فروشگاه‌ها و قیمت‌های این صفحه ساختگی‌اند.",
  },
  {
    id: "coffee",
    name: "قهوه ترکیبی روزانه",
    subtitle: "برای دوست‌داران صبح · ۲۵۰ گرم",
    category: "tea",
    image: "coffee",
    price: 465000,
    sales: 72,
    stock: 16,
    colors: ["آسیاب‌شده", "دانه کامل"],
    description:
      "قهوه نمونه برای مرور کاتالوگ و آزمودن انتخاب ترکیب. در این نسخه امکان ثبت سفارش واقعی وجود ندارد.",
  },
  {
    id: "cup-pair",
    name: "ست دو عددی فنجان",
    subtitle: "یک گفت‌وگوی گرم دونفره",
    category: "home",
    image: "cup",
    price: 690000,
    sales: 65,
    stock: 7,
    colors: ["سفید", "کرم"],
    description:
      "مجموعه نمونه دو فنجان برای خانه. تعداد، قیمت و فروشگاه فقط برای بررسی رابط کاربری هستند.",
  },
  {
    id: "headphones-pro",
    name: "هدفون استودیو پلاس",
    subtitle: "همراه روزهای پرموسیقی",
    category: "digital",
    image: "headphones",
    price: 3250000,
    sales: 58,
    stock: 4,
    colors: ["مشکی", "نقره‌ای"],
    description:
      "مدل دوم هدفون در کاتالوگ نمونه، برای آزمایش فیلتر و مقایسه قیمت.",
  },
  {
    id: "watch-silver",
    name: "ساعت نقره‌ای مینیمال",
    subtitle: "جزئیات کوچک، حس متفاوت",
    category: "home",
    image: "watch",
    price: 2150000,
    sales: 99,
    stock: 0,
    colors: ["نقره‌ای"],
    description:
      "کالای ناموجودِ نمونه؛ برای بررسی نمایش ناموجودها در انتهای فهرست.",
  },
];
export const stores = [
  { id: "cherry", name: "فروشگاه آلبالو", extra: 0, preparation: "۲ روز کاری" },
  {
    id: "pearl",
    name: "فروشگاه مروارید",
    extra: 45000,
    preparation: "۱ روز کاری",
  },
];
export const money = (value: number) =>
  new Intl.NumberFormat("fa-IR").format(value);
export type CartLine = {
  productId: string;
  color: string;
  storeId: string;
  quantity: number;
};
export const lineKey = (
  line: Pick<CartLine, "productId" | "color" | "storeId">,
) => `${line.productId}/${line.color}/${line.storeId}`;
