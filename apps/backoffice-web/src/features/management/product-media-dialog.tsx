import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import type { Product, Variant } from "../../shared/types";
function Photo({ id }: { id: string }) {
  const { loadCatalogMedia } = useWorkspace();
  const [url, setURL] = useState("");
  useEffect(() => {
    let active = true,
      current = "";
    void loadCatalogMedia(id)
      .then((b) => {
        current = URL.createObjectURL(b);
        if (active) setURL(current);
        else URL.revokeObjectURL(current);
      })
      .catch(() => {});
    return () => {
      active = false;
      if (current) URL.revokeObjectURL(current);
    };
  }, [id, loadCatalogMedia]);
  return url ? (
    <img src={url} alt="تصویر گالری" className="home-admin-photo" />
  ) : (
    <span>تصویر در دسترس نیست</span>
  );
}
const noVariants: Variant[] = [];
export function ProductMediaDialog({
  id,
  close,
}: {
  id: string;
  close: () => void;
}) {
  const { state, execute, uploadIcon, busy, can } = useWorkspace();
  const p = (state.products ?? []).find((p) => p.id === id) as Product;
  const vs = state.variants ?? noVariants;
  const [gallery, setGallery] = useState(p.gallery_ids ?? []),
    [variants, setVariants] = useState<Record<string, string>>({}),
    [version, setVersion] = useState(p.version),
    [file, setFile] = useState<File>(),
    [preview, setPreview] = useState(""),
    [error, setError] = useState(""),
    [key, setKey] = useState(crypto.randomUUID());
  useEffect(() => {
    setGallery(p.gallery_ids ?? []);
    setVersion(p.version);
    setVariants(
      Object.fromEntries(
        vs
          .filter((v) => v.product_id === id)
          .map((v) => [v.id, v.image_id ?? ""]),
      ),
    );
    setKey(crypto.randomUUID());
  }, [p.version, id, vs, p.gallery_ids]);
  useEffect(() => {
    if (!file) {
      setPreview("");
      return;
    }
    const u = URL.createObjectURL(file);
    setPreview(u);
    return () => URL.revokeObjectURL(u);
  }, [file]);
  const dirty =
    JSON.stringify(gallery) !== JSON.stringify(p.gallery_ids ?? []) ||
    vs.some(
      (v) =>
        v.product_id === id && (variants[v.id] ?? "") !== (v.image_id ?? ""),
    );
  const all = [
    ...new Set(
      [
        p.image_id,
        ...(p.gallery_ids ?? []),
        ...vs.filter((v) => v.product_id === id).map((v) => v.image_id),
      ].filter((x): x is string => !!x),
    ),
  ];
  return (
    <Dialog open fullWidth maxWidth="md" onClose={() => !busy && close()}>
      <DialogTitle>گالری و تصویر تنوع</DialogTitle>
      <DialogContent>
        <div className="home-admin-form">
          {error && <Alert severity="error">{error}</Alert>}
          <p>
            تصویر اصلی از اکشن «تصویر اصلی» مدیریت می‌شود. حداکثر۸ تصویر تکمیلی؛
            PNG/JPEG/WebP تا۵MiB، ضلع ورودی تا۴۰۹۶ و خروجی تا۱۰۲۴ پیکسل.
          </p>
          <label>
            انتخاب تصویر گالری
            <input
              type="file"
              accept="image/png,image/jpeg,image/webp"
              disabled={busy || dirty || !can("product.icon.set", "management")}
              onChange={(e) => {
                setFile(e.target.files?.[0]);
                setKey(crypto.randomUUID());
              }}
            />
          </label>
          {preview && (
            <img
              src={preview}
              alt="پیش‌نمایش تصویر گالری"
              className="product-media-preview"
            />
          )}
          <Button
            variant="outlined"
            disabled={
              !file ||
              busy ||
              dirty ||
              gallery.length >= 8 ||
              !can("product.icon.set", "management")
            }
            onClick={async () => {
              setError("");
              if (!file) return;
              try {
                await uploadIcon({
                  kind: "product-gallery",
                  target: id,
                  version,
                  key,
                  file,
                });
                setFile(undefined);
              } catch (e) {
                setError(e instanceof Error ? e.message : "تصویر ذخیره نشد");
              }
            }}
          >
            ذخیره تصویر گالری
          </Button>
          {dirty && (
            <small>
              قبل از افزودن فایل تازه، تغییر ترتیب و تصویر تنوع را ذخیره کنید.
            </small>
          )}
          <ol className="product-media-list">
            {gallery.map((gid, i) => (
              <li key={gid}>
                <Photo id={gid} />
                <span>تصویر {i + 1}</span>
                <Button
                  disabled={busy || i === 0}
                  aria-label={`بالا بردن تصویر ${i + 1}`}
                  onClick={() => {
                    const next = [...gallery];
                    [next[i - 1], next[i]] = [next[i], next[i - 1]];
                    setGallery(next);
                    setKey(crypto.randomUUID());
                  }}
                >
                  بالاتر
                </Button>
                <Button
                  disabled={busy || i === gallery.length - 1}
                  aria-label={`پایین بردن تصویر ${i + 1}`}
                  onClick={() => {
                    const next = [...gallery];
                    [next[i + 1], next[i]] = [next[i], next[i + 1]];
                    setGallery(next);
                    setKey(crypto.randomUUID());
                  }}
                >
                  پایین‌تر
                </Button>
                <Button
                  disabled={busy}
                  aria-label={`حذف تصویر ${i + 1}`}
                  onClick={() => {
                    setGallery(gallery.filter((x) => x !== gid));
                    setKey(crypto.randomUUID());
                  }}
                >
                  حذف
                </Button>
              </li>
            ))}
          </ol>
          {vs
            .filter((v) => v.product_id === id)
            .map((v) => (
              <label key={v.id}>
                تصویر تنوع {v.values.map((x) => x.value).join(" / ") || v.sku}
                <select
                  value={variants[v.id] ?? ""}
                  disabled={busy}
                  onChange={(e) => {
                    setVariants({ ...variants, [v.id]: e.target.value });
                    setKey(crypto.randomUUID());
                  }}
                >
                  <option value="">تصویر اصلی محصول</option>
                  {all.map((gid, i) => (
                    <option key={gid} value={gid}>
                      {gid === p.image_id
                        ? "تصویر اصلی"
                        : `تصویر ثبت‌شده ${i + 1}`}
                    </option>
                  ))}
                </select>
              </label>
            ))}
          <small>
            حذف از گالری، تصویری را که هنوز به تنوع متصل است حذف نمی‌کند؛ برای
            حذف کامل، انتساب تنوع را هم پاک کنید.
          </small>
        </div>
      </DialogContent>
      <DialogActions>
        <Button disabled={busy} onClick={close}>
          بستن
        </Button>
        <Button
          variant="contained"
          disabled={
            busy || !dirty || !can("product.media.update", "management")
          }
          onClick={async () => {
            setError("");
            try {
              await execute({
                action: "product.media.update",
                target: id,
                version,
                idempotency_key: key,
                product_media: {
                  gallery_ids: gallery,
                  variant_images: variants,
                },
              });
            } catch (e) {
              setError(
                e instanceof Error
                  ? e.message
                  : "تغییرات ذخیره نشد؛ نسخه تازه را بررسی کنید",
              );
            }
          }}
        >
          ذخیره ترتیب و تصاویر تنوع
        </Button>
      </DialogActions>
    </Dialog>
  );
}
