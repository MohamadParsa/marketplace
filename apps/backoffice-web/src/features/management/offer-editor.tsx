import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
} from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import { ApiError } from "../../shared/api";
import { money, recordNumber } from "../../shared/format";
import type { OfferView, ProductWorkspace } from "../../shared/types";
const rial = (text: string) => {
  if (!/^\d+(\.\d)?$/.test(text)) return undefined;
  const [whole, fraction = "0"] = text.split(".");
  const value = Number(whole) * 10 + Number(fraction);
  return Number.isSafeInteger(value) && value > 0 && value <= 9000000000000000
    ? value
    : undefined;
};

export function OfferEditor({
  workspace: w,
  offer,
  close,
}: {
  workspace: ProductWorkspace;
  offer: OfferView;
  close: () => void;
}) {
  const { execute, loadProduct, surface } = useWorkspace(),
    cache = useQueryClient();
  const priceAllowed = w.permissions.price_stores.includes(offer.store_id);
  const stockAllowed = w.permissions.stock_stores.includes(offer.store_id);
  const [current, setCurrent] = useState(offer),
    [price, setPrice] = useState(String(offer.price / 10)),
    [regular, setRegular] = useState(
      offer.regular_price ? String(offer.regular_price / 10) : "",
    ),
    [stock, setStock] = useState(String(offer.stock));
  const [busy, setBusy] = useState(false),
    [error, setError] = useState(""),
    [conflict, setConflict] = useState(false),
    [key, setKey] = useState(crypto.randomUUID());
  const changed = () => {
    setKey(crypto.randomUUID());
    setError("");
  };
  return (
    <Dialog open fullWidth maxWidth="md" onClose={() => !busy && close()}>
      <DialogTitle>
        قیمت و موجودی · {w.stores.find((s) => s.id === offer.store_id)?.name}
      </DialogTitle>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          if (busy || conflict) return;
          const amount = rial(price),
            before = regular.trim() ? rial(regular) : 0,
            quantity = Number(stock);
          if (
            priceAllowed &&
            (amount === undefined ||
              before === undefined ||
              (before > 0 && before <= amount))
          ) {
            setError(
              "قیمت نهایی مثبت باشد و قیمت قبل از تخفیف، خالی یا بیشتر از قیمت نهایی باشد.",
            );
            return;
          }
          if (
            stockAllowed &&
            (!Number.isInteger(quantity) ||
              quantity < current.reserved ||
              quantity > 1000000000)
          ) {
            setError(
              `موجودی کل باید عدد صحیح و حداقل برابر رزرو (${current.reserved}) باشد.`,
            );
            return;
          }
          setBusy(true);
          setError("");
          try {
            await execute({
              action: "offer.update",
              target: offer.id,
              version: current.version,
              idempotency_key: key,
              offer: {
                ...(priceAllowed
                  ? { price: amount, regular_price: before }
                  : {}),
                ...(stockAllowed ? { stock: quantity } : {}),
              },
            });
            close();
          } catch (cause) {
            setError(
              cause instanceof Error ? cause.message : "ذخیره انجام نشد.",
            );
            if (cause instanceof ApiError && cause.status === 409)
              setConflict(true);
          } finally {
            setBusy(false);
          }
        }}
      >
        <DialogContent className="product-form">
          <p>
            {w.product.name} ·{" "}
            {w.variants
              .find((v) => v.id === offer.variant_id)
              ?.values.map((v) => v.value)
              .join(" / ") || "بدون محور"}{" "}
            · عرضهٔ {recordNumber(w, "offer", offer.id)}
          </p>
          {error && (
            <Alert severity={conflict ? "warning" : "error"}>{error}</Alert>
          )}
          {!priceAllowed && (
            <Alert severity="info">
              قیمت فقط خواندنی است؛ مجوز شما برای موجودی اعمال می‌شود.
            </Alert>
          )}
          {!stockAllowed && (
            <Alert severity="info">
              موجودی فقط خواندنی است؛ مجوز شما برای قیمت اعمال می‌شود.
            </Alert>
          )}
          <div className="commerce-grid">
            <TextField
              label="قیمت نهایی نقدی (تومان)"
              type="number"
              value={price}
              disabled={busy || !priceAllowed}
              required
              inputProps={{ min: 0.1, step: 0.1 }}
              onChange={(e) => {
                setPrice(e.target.value);
                changed();
              }}
              helperText={`جاری: ${money(current.price)}`}
            />
            <TextField
              label="قیمت قبل از تخفیف (تومان)"
              type="number"
              value={regular}
              disabled={busy || !priceAllowed}
              inputProps={{ min: 0.1, step: 0.1 }}
              onChange={(e) => {
                setRegular(e.target.value);
                changed();
              }}
              helperText="برای حذف تخفیف خالی بگذارید."
            />
            <TextField
              label="موجودی کل"
              type="number"
              value={stock}
              disabled={busy || !stockAllowed}
              required
              inputProps={{ min: current.reserved, max: 1000000000, step: 1 }}
              onChange={(e) => {
                setStock(e.target.value);
                changed();
              }}
              helperText={`رزرو: ${current.reserved.toLocaleString("fa-IR")} · آزاد فعلی: ${current.available.toLocaleString("fa-IR")}`}
            />
          </div>
          {conflict && (
            <Button
              disabled={busy}
              onClick={async () => {
                setBusy(true);
                try {
                  const latest = await loadProduct(w.product.id),
                    next = latest.offers.find((o) => o.id === offer.id);
                  if (!next) throw Error("عرضه دیگر در دسترس نیست.");
                  setCurrent(next);
                  cache.setQueryData(
                    ["product", surface, w.product.id],
                    latest,
                  );
                  setKey(crypto.randomUUID());
                  setConflict(false);
                  setError(
                    "نسخهٔ تازه دریافت شد؛ ورودی شما حفظ شده است. مقادیر جاری را بررسی و دوباره ذخیره کنید.",
                  );
                } catch (cause) {
                  setError(
                    cause instanceof Error
                      ? cause.message
                      : "بازخوانی انجام نشد.",
                  );
                } finally {
                  setBusy(false);
                }
              }}
            >
              بازخوانی و بازبینی
            </Button>
          )}
        </DialogContent>
        <DialogActions>
          <Button disabled={busy} onClick={close}>
            انصراف
          </Button>
          <Button
            type="submit"
            variant="contained"
            disabled={busy || conflict || (!priceAllowed && !stockAllowed)}
          >
            {busy ? "در حال ذخیره…" : "ذخیرهٔ قیمت و موجودی"}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}
