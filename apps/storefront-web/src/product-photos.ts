import type { Product } from "./catalog";

type Crop = "detail" | "handle" | "rim";
type ProductPhoto = { src: string; label: string; color: string; crop?: Crop };

export function productPhoto(product: Product, color: string): string {
  switch (product.image) {
    case "cup":
      return color === "کرم"
        ? "/images/cup-cream.webp"
        : "/images/cup-ivory.webp";
    case "headphones":
      return color === "نقره‌ای"
        ? "/images/headphones-silver.webp"
        : "/images/headphones-black.webp";
    case "tea":
      return "/images/tea-leaves.webp";
    case "watch":
      return color === "مشکی"
        ? "/images/watch.svg"
        : "/images/watch-silver.webp";
    default:
      return `/images/${product.image}.svg`;
  }
}

export function productPhotos(product: Product): ProductPhoto[] {
  const photos: ProductPhoto[] = product.colors.flatMap((color) => {
    const src = productPhoto(product, color);
    return [
      { src, label: color, color },
      { src, label: `نمای نزدیک ${color}`, color, crop: "detail" as const },
    ];
  });
  if (product.image === "cup" && product.colors.length > 1) {
    photos.push(
      { ...photos[0], label: "جزئیات دسته", crop: "handle" },
      { ...photos[2], label: "جزئیات لبه", crop: "rim" },
    );
  }
  return photos;
}
