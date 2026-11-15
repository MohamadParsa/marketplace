import { products, stores, type CartLine } from "../catalog";

export const DEMO_PHONE = "09000000003";
export const DEMO_CODE = "123456";
export const DEMO_WALLET = 200000;
export const CHECKOUT_DURATION = 10 * 60 * 1000;
export const digits = (value: string) =>
  value
    .replace(/[۰-۹]/g, (digit) => String("۰۱۲۳۴۵۶۷۸۹".indexOf(digit)))
    .replace(/[٠-٩]/g, (digit) => String("٠١٢٣٤٥٦٧٨٩".indexOf(digit)));
export type Address = {
  id: string;
  title: string;
  province: string;
  city: string;
  street: string;
  postalCode: string;
  recipient: string;
  phone: string;
};
export type Profile = { firstName: string; lastName: string };
export type CheckoutDraft = {
  profile: Profile;
  addresses: Address[];
  addressId: string;
  shipping: Record<string, string>;
};
export const initialDraft: CheckoutDraft = {
  profile: { firstName: "", lastName: "" },
  addresses: [
    {
      id: "sample-home",
      title: "خانه نمونه",
      province: "تهران",
      city: "تهران",
      street: "خیابان نمونه، کوچه آزمایش، پلاک ۱، واحد ۲",
      postalCode: "1111111111",
      recipient: "سارا نمونه",
      phone: DEMO_PHONE,
    },
    {
      id: "sample-work",
      title: "محل کار نمونه",
      province: "البرز",
      city: "کرج",
      street: "بلوار نمونه، ساختمان آزمایشی، طبقه اول",
      postalCode: "2222222222",
      recipient: "علی نمونه",
      phone: "09000000004",
    },
  ],
  addressId: "sample-home",
  shipping: {},
};
export const destinations: Record<string, string[]> = {
  تهران: ["تهران"],
  البرز: ["کرج"],
  فارس: ["شیراز"],
};
export function validateAddress(address: Address) {
  if (
    !address.title.trim() ||
    address.recipient.trim().length < 3 ||
    address.street.trim().length < 10
  )
    return "عنوان، نام کامل گیرنده و نشانی کامل را وارد کنید.";
  if (!destinations[address.province]?.includes(address.city))
    return "استان و شهر را انتخاب کنید.";
  if (!/^\d{10}$/.test(digits(address.postalCode)))
    return "کدپستی باید ۱۰ رقم باشد.";
  if (!/^09\d{9}$/.test(digits(address.phone)))
    return "شماره تماس گیرنده باید ۱۱ رقم و با ۰۹ شروع شود.";
  return "";
}
export type ShippingOption = {
  id: string;
  title: string;
  cost: number;
  transit: string;
  arrival: string;
};
// این تعرفه‌ها و بازه‌ها فقط fixture رابط‌اند؛ تقویم یا تعرفه عملیاتی نیستند.
export function shippingOptions(
  storeId: string,
  city: string,
): ShippingOption[] {
  if (city !== "تهران" && city !== "کرج") return [];
  const preparation = storeId === "cherry" ? 2 : 1;
  return [
    {
      id: "post",
      title: "پست — نمونه",
      cost:
        (city === "تهران" ? 65000 : 85000) + (storeId === "pearl" ? 10000 : 0),
      transit: "۲ تا ۴ روز کاری",
      arrival: `${preparation + 2} تا ${preparation + 4} روز کاری پس از پرداخت`,
    },
    ...(city === "تهران"
      ? [
          {
            id: "courier",
            title: "پیک — نمونه",
            cost: storeId === "cherry" ? 110000 : 95000,
            transit: "۱ روز کاری",
            arrival: `${preparation + 1} روز کاری پس از پرداخت`,
          },
        ]
      : []),
  ];
}
export function cartGroups(cart: CartLine[]) {
  return stores
    .filter((store) => cart.some((line) => line.storeId === store.id))
    .map((store) => ({
      ...store,
      lines: cart.filter((line) => line.storeId === store.id),
    }));
}
export function itemsTotal(cart: CartLine[]) {
  return cart.reduce(
    (total, line) =>
      total +
      (products.find((product) => product.id === line.productId)!.price +
        stores.find((store) => store.id === line.storeId)!.extra) *
        line.quantity,
    0,
  );
}
export function validCart(cart: CartLine[]) {
  return (
    cart.length > 0 &&
    cart.every(
      (line) =>
        Number.isInteger(line.quantity) &&
        line.quantity > 0 &&
        products.some(
          (product) =>
            product.id === line.productId &&
            product.colors.includes(line.color),
        ) &&
        stores.some((store) => store.id === line.storeId),
    ) &&
    products.every(
      (product) =>
        cart
          .filter((line) => line.productId === product.id)
          .reduce((quantity, line) => quantity + line.quantity, 0) <=
        product.stock,
    )
  );
}
export function selectedShipping(cart: CartLine[], draft: CheckoutDraft) {
  const address = draft.addresses.find(
    (address) => address.id === draft.addressId,
  );
  return cartGroups(cart).map((store) => ({
    store,
    option: shippingOptions(store.id, address?.city ?? "").find(
      (option) => option.id === draft.shipping[store.id],
    ),
  }));
}

export type PaymentSession = {
  expiresAt: number;
  subtotal: number;
  shipping: number;
  total: number;
  address: Address;
};
