export function normalizeTransferReference(value: string) {
  return value.trim().replace(/[۰-۹]/g,c=>String(c.charCodeAt(0)-1776)).replace(/[٠-٩]/g,c=>String(c.charCodeAt(0)-1632)).replace(/[a-z]/g,c=>c.toUpperCase());
}
export function transferReferenceError(value: string): string {
  const normalized=normalizeTransferReference(value);
  if(!normalized)return "کد پیگیری روی رسید انتقال بانکی را وارد کنید.";
  return /^[A-Z0-9-]{3,80}$/.test(normalized)?"":"کد پیگیری باید ۳ تا ۸۰ نویسه شامل عدد، حروف لاتین یا خط تیره باشد؛ نام شخص معتبر نیست.";
}
