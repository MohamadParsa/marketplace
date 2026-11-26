import { ProductQA } from "./product-qa";
import { ProductContentReview } from "./product-content-review";
import { ProductMediaDialog } from "./product-media-dialog";
import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import {
  Alert,
  Button,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  MenuItem,
  Tab,
  Tabs,
  TextField,
} from "@mui/material";
import { DataTable, PageHeader } from "../../shared/components";
import { money } from "../../shared/format";
import { ApiError } from "../../shared/api";
import { useWorkspace } from "../../shared/workspace";
import type {
  Command,
  OfferView,
  ProductWorkspace,
  Variant,
} from "../../shared/types";

const errorMessage = (e: unknown) =>
  e instanceof Error
    ? e.message
    : "ارتباط با سرور برقرار نشد؛ دوباره تلاش کنید.";
const variantName = (v: Variant) =>
  v.values.map((x) => x.value).join(" / ") || "بدون محور";
const tomanToRial = (input: string) => {
  if (!/^\d+(\.\d)?$/.test(input)) return undefined;
  const [whole, fraction = "0"] = input.split(".");
  const rial = Number(whole) * 10 + Number(fraction);
  return Number.isSafeInteger(rial) && rial > 0 && rial <= 9000000000000000
    ? rial
    : undefined;
};

export function ProductQuickDialog({
  id,
  initialTab,
  close,
}: {
  id: string;
  initialTab: number;
  close: () => void;
}) {
  return (
    <Dialog open fullWidth maxWidth="lg" onClose={close}>
      <DialogTitle>مدیریت تنوع و عرضه</DialogTitle>
      <DialogContent>
        <ProductPanel id={id} initialTab={initialTab} />
      </DialogContent>
      <DialogActions>
        <Button onClick={close}>بستن</Button>
      </DialogActions>
    </Dialog>
  );
}

export function ProductDetailsPage() {
  const { id } = useParams();
  return (
    <>
      <PageHeader
        title="جزئیات محصول"
        description="محتوای مرجع محصول از تنوع‌ها و عرضه فروشندگان مستقل است."
        action={
          <Button component={Link} to="/products">
            بازگشت به محصولات
          </Button>
        }
      />
      <section className="panel product-detail">
        <ProductPanel id={id ?? ""} initialTab={0} />
      </section>
    </>
  );
}

function ProductPanel({ id, initialTab }: { id: string; initialTab: number }) {
  const { loadProduct, surface, can, role } = useWorkspace();
  const canManageQA = surface === "backoffice" && role === "super-admin";
  const [mediaOpen, setMediaOpen] = useState(false);
  const [tab, setTab] = useState(initialTab);
  const [editor, setEditor] = useState<
    "variant" | "offer" | { offer: OfferView; field: "price" | "stock" }
  >();
  const query = useQuery({
    queryKey: ["product", surface, id],
    queryFn: () => loadProduct(id),
  });
  if (query.isPending)
    return (
      <div className="loading">
        <CircularProgress />
        <p>در حال دریافت محصول…</p>
      </div>
    );
  if (query.isError)
    return (
      <Alert
        severity="error"
        action={<Button onClick={() => void query.refetch()}>تلاش مجدد</Button>}
      >
        {errorMessage(query.error)}
      </Alert>
    );
  const w = query.data;
  return (
    <div className="product-workspace" data-testid={`product-workspace-${id}`}>
      {mediaOpen && (
        <ProductMediaDialog id={id} close={() => setMediaOpen(false)} />
      )}
      {surface === "backoffice" && (
        <div className="product-workspace-toolbar">
          <Button
            disabled={!can("catalog.product.update", "management")}
            component={Link}
            to={"/products?product=" + encodeURIComponent(id)}
          >
            ویرایش محتوای محصول
          </Button>
          <Button
            disabled={!can("product.media.update", "management")}
            onClick={() => setMediaOpen(true)}
          >
            گالری و تصویر تنوع
          </Button>
        </div>
      )}
      <div className="product-workspace-heading">
        <div>
          <h2>{w.product.name}</h2>
          <small dir="ltr">{w.product.id}</small>
        </div>
        <Chip
          label={w.product.model === "shared" ? "محصول مشترک" : "محصول مستقل"}
        />
      </div>
      <ProductContentReview workspace={w} />
      <Tabs
        value={tab}
        onChange={(_, value: number) => setTab(value)}
        variant="scrollable"
        scrollButtons="auto"
        aria-label="بخش‌های محصول"
      >
        <Tab label="تنوع‌ها" />
        <Tab label="قیمت و موجودی" />
        <Tab label="پیش‌نمایش انتخاب" />
        {canManageQA && <Tab label="پرسش و پاسخ" />}
      </Tabs>
      {tab === 0 && (
        <section className="product-workspace-section">
          <div className="product-workspace-toolbar">
            <p className="muted">
              محورها: {w.axes.map((a) => a.name).join("، ") || "بدون محور تنوع"}
            </p>
            <Button
              variant="contained"
              disabled={!w.permissions.create_variant}
              onClick={() => setEditor("variant")}
            >
              افزودن تنوع
            </Button>
          </div>
          <DataTable
            rows={w.variants}
            columns={[
              {
                title: "ترکیب ویژگی‌ها",
                render: (v) => <b>{variantName(v)}</b>,
              },
              {
                title: "شناسه تنوع",
                render: (v) => <small dir="ltr">{v.id}</small>,
              },
              { title: "SKU", render: (v) => <span dir="ltr">{v.sku}</span> },
              {
                title: "عرضه‌های قابل مشاهده",
                render: (v) =>
                  w.offers.filter((o) => o.variant_id === v.id).length,
              },
            ]}
          />
          <p className="muted">
            ویرایش مخرب محور یا ترکیب استفاده‌شده در این مرحله فعال نیست.
          </p>
        </section>
      )}
      {tab === 1 && (
        <section className="product-workspace-section">
          <div className="product-workspace-toolbar">
            <p className="muted">
              قیمت پایه نقدی و موجودی برای هر فروشنده و تنوع مستقل است.
            </p>
            <Button
              variant="contained"
              disabled={
                !w.permissions.create_offer_stores.length || !w.variants.length
              }
              onClick={() => setEditor("offer")}
            >
              افزودن عرضه
            </Button>
          </div>
          <DataTable
            rows={w.offers}
            columns={[
              {
                title: "فروشنده / تنوع",
                render: (o) => (
                  <>
                    <b>{w.stores.find((s) => s.id === o.store_id)?.name}</b>
                    <small>
                      {w.variants.find((v) => v.id === o.variant_id)
                        ? variantName(
                            w.variants.find((v) => v.id === o.variant_id)!,
                          )
                        : o.variant_id}
                    </small>
                    <small dir="ltr">{o.id}</small>
                  </>
                ),
              },
              { title: "قیمت نهایی نقدی", render: (o) => money(o.price) },
              {
                title: "موجودی کل / رزرو / آزاد",
                render: (o) => (
                  <span>
                    {o.stock} / {o.reserved} / <b>{o.available}</b>
                    {o.sale_restricted && (
                      <small>برای خرید نقدی مجاز نیست</small>
                    )}
                  </span>
                ),
              },
              {
                title: "وقفه تأمین",
                render: (o) => `${o.lead_days} روز کاری · از فروشنده`,
              },
              {
                title: "عملیات",
                render: (o) => (
                  <>
                    <Button
                      disabled={
                        !o.store_active ||
                        !w.permissions.price_stores.includes(o.store_id)
                      }
                      onClick={() => setEditor({ offer: o, field: "price" })}
                    >
                      ویرایش قیمت
                    </Button>
                    <Button
                      disabled={
                        !o.store_active ||
                        !w.permissions.stock_stores.includes(o.store_id)
                      }
                      onClick={() => setEditor({ offer: o, field: "stock" })}
                    >
                      ویرایش موجودی
                    </Button>
                  </>
                ),
              },
            ]}
          />
        </section>
      )}
      {tab === 2 && <VariantSelection key={id} workspace={w} />}
      {tab === 3 && canManageQA && <ProductQA key={id} productId={id} />}
      {editor === "variant" && (
        <VariantEditor workspace={w} close={() => setEditor(undefined)} />
      )}
      {editor === "offer" && (
        <OfferCreator workspace={w} close={() => setEditor(undefined)} />
      )}
      {typeof editor === "object" && (
        <OfferEditor
          workspace={w}
          offer={editor.offer}
          field={editor.field}
          close={() => setEditor(undefined)}
        />
      )}
    </div>
  );
}

function VariantEditor({
  workspace: w,
  close,
}: {
  workspace: ProductWorkspace;
  close: () => void;
}) {
  const { execute, loadProduct, surface } = useWorkspace();
  const cache = useQueryClient();
  const [sku, setSKU] = useState("");
  const [selected, setSelected] = useState<Record<string, string>>({});
  const [version, setVersion] = useState(w.product.variant_version);
  const [key, setKey] = useState(crypto.randomUUID());
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  return (
    <Dialog
      open
      fullWidth
      maxWidth="sm"
      onClose={() => {
        if (!pending) close();
      }}
    >
      <DialogTitle>افزودن تنوع محصول</DialogTitle>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          if (pending) return;
          setError("");
          setPending(true);
          try {
            await execute({
              action: "variant.create",
              target: w.product.id,
              version,
              idempotency_key: key,
              variant: {
                product_id: w.product.id,
                sku,
                values: w.axes.map((a) => ({
                  attribute_id: a.id,
                  value: selected[a.id] ?? "",
                })),
              },
            });
            close();
          } catch (e) {
            setError(errorMessage(e));
          } finally {
            setPending(false);
          }
        }}
      >
        <DialogContent className="product-form">
          <Alert severity="info">
            هر ترکیب برای محصول فقط یک بار ساخته می‌شود؛ قیمت در فرم عرضه
            فروشنده ثبت می‌شود.
          </Alert>
          {error && (
            <Alert severity="error">
              {error}
              <Button
                disabled={pending}
                onClick={async () => {
                  setPending(true);
                  try {
                    const latest = await loadProduct(w.product.id);
                    cache.setQueryData(
                      ["product", surface, w.product.id],
                      latest,
                    );
                    setVersion(latest.product.variant_version);
                    setKey(crypto.randomUUID());
                    setError(
                      "نسخه تازه دریافت شد؛ انتخاب‌های شما حفظ شده‌اند، بررسی و دوباره ثبت کنید.",
                    );
                  } catch (e) {
                    setError(errorMessage(e));
                  } finally {
                    setPending(false);
                  }
                }}
              >
                بازبینی نسخه فعلی
              </Button>
            </Alert>
          )}
          <TextField
            label="SKU تنوع"
            required
            value={sku}
            disabled={pending}
            inputProps={{
              pattern: "[A-Za-z0-9][A-Za-z0-9._\\-]{0,63}",
              dir: "ltr",
            }}
            onChange={(e) => {
              setSKU(e.target.value);
              setKey(crypto.randomUUID());
            }}
          />
          {w.axes.map((a) => (
            <TextField
              select
              key={a.id}
              label={a.name}
              required
              value={selected[a.id] ?? ""}
              disabled={pending}
              onChange={(e) => {
                setSelected({ ...selected, [a.id]: e.target.value });
                setKey(crypto.randomUUID());
              }}
            >
              <MenuItem value="">انتخاب کنید</MenuItem>
              {a.options.map((o) => (
                <MenuItem key={o} value={o}>
                  {o}
                </MenuItem>
              ))}
            </TextField>
          ))}
        </DialogContent>
        <DialogActions>
          <Button disabled={pending} onClick={close}>
            انصراف
          </Button>
          <Button type="submit" variant="contained" disabled={pending}>
            ثبت تنوع
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}

function OfferCreator({
  workspace: w,
  close,
}: {
  workspace: ProductWorkspace;
  close: () => void;
}) {
  const { execute } = useWorkspace();
  const [variant, setVariant] = useState(w.variants[0]?.id ?? ""),
    [store, setStore] = useState(w.permissions.create_offer_stores[0] ?? "");
  const [price, setPrice] = useState(""),
    [stock, setStock] = useState("0"),
    [pending, setPending] = useState(false),
    [error, setError] = useState("");
  const [key, setKey] = useState(crypto.randomUUID());
  const change = (setter: (v: string) => void, value: string) => {
    setter(value);
    setKey(crypto.randomUUID());
  };
  return (
    <Dialog
      open
      fullWidth
      maxWidth="sm"
      onClose={() => {
        if (!pending) close();
      }}
    >
      <DialogTitle>افزودن عرضه فروشنده</DialogTitle>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          if (pending) return;
          const rial = tomanToRial(price);
          if (rial === undefined) {
            setError("قیمت مثبت با حداکثر یک رقم اعشار تومان وارد کنید.");
            return;
          }
          setPending(true);
          setError("");
          try {
            await execute({
              action: "offer.create",
              target: w.product.id,
              version: w.product.variant_version,
              idempotency_key: key,
              offer: {
                product_id: w.product.id,
                variant_id: variant,
                store_id: store,
                price: rial,
                stock: Number(stock),
              },
            });
            close();
          } catch (e) {
            setError(errorMessage(e));
          } finally {
            setPending(false);
          }
        }}
      >
        <DialogContent className="product-form">
          {error && <Alert severity="error">{error}</Alert>}
          <TextField
            select
            label="تنوع عرضه"
            value={variant}
            required
            disabled={pending}
            onChange={(e) => change(setVariant, e.target.value)}
          >
            {w.variants.map((v) => (
              <MenuItem key={v.id} value={v.id}>
                {variantName(v)} · {v.sku}
              </MenuItem>
            ))}
          </TextField>
          <TextField
            select
            label="فروشنده عرضه"
            value={store}
            required
            disabled={pending}
            onChange={(e) => change(setStore, e.target.value)}
          >
            {w.stores
              .filter((s) => w.permissions.create_offer_stores.includes(s.id))
              .map((s) => (
                <MenuItem key={s.id} value={s.id}>
                  {s.name}
                </MenuItem>
              ))}
          </TextField>
          <TextField
            label="قیمت پایه نقدی (تومان)"
            value={price}
            type="number"
            required
            inputProps={{ min: 0.1, max: 900000000000000, step: 0.1 }}
            disabled={pending}
            onChange={(e) => change(setPrice, e.target.value)}
          />
          <TextField
            label="موجودی کل"
            value={stock}
            type="number"
            required
            inputProps={{ min: 0, max: 1000000000, step: 1 }}
            disabled={pending}
            onChange={(e) => change(setStock, e.target.value)}
          />
          <p>
            وقفه تأمین: {w.stores.find((s) => s.id === store)?.lead_days} روز
            کاری؛ از تنظیم فروشنده، فقط خواندنی.
          </p>
        </DialogContent>
        <DialogActions>
          <Button disabled={pending} onClick={close}>
            انصراف
          </Button>
          <Button type="submit" variant="contained" disabled={pending}>
            ثبت عرضه
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}

function OfferEditor({
  workspace: w,
  offer,
  field,
  close,
}: {
  workspace: ProductWorkspace;
  offer: OfferView;
  field: "price" | "stock";
  close: () => void;
}) {
  const { execute, loadProduct, surface } = useWorkspace();
  const cache = useQueryClient();
  const [current, setCurrent] = useState(offer),
    [value, setValue] = useState(
      String(field === "price" ? offer.price / 10 : offer.stock),
    );
  const [key, setKey] = useState(crypto.randomUUID()),
    [pending, setPending] = useState(false),
    [error, setError] = useState("");
  const [conflict, setConflict] = useState(false);
  const price = field === "price";
  const label = price ? "قیمت نهایی نقدی (تومان)" : "موجودی کل";
  const [regular, setRegular] = useState(
    offer.regular_price ? String(offer.regular_price / 10) : "",
  );
  return (
    <Dialog
      open
      fullWidth
      maxWidth="sm"
      onClose={() => {
        if (!pending) close();
      }}
    >
      <DialogTitle>
        {price ? "ویرایش قیمت عرضه" : "ویرایش موجودی عرضه"}
      </DialogTitle>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          if (pending || conflict) return;
          const amount = price ? tomanToRial(value) : Number(value);
          if (
            amount === undefined ||
            (!price &&
              (!Number.isInteger(amount) ||
                amount < current.reserved ||
                amount > 1000000000))
          ) {
            setError(
              price
                ? "قیمت مثبت با حداکثر یک رقم اعشار تومان وارد کنید."
                : `موجودی کل نمی‌تواند کمتر از رزرو (${current.reserved}) باشد.`,
            );
            return;
          }
          const beforeDiscount =
            regular.trim() === "" ? 0 : tomanToRial(regular);
          if (
            price &&
            (beforeDiscount === undefined ||
              (beforeDiscount > 0 && beforeDiscount <= amount))
          ) {
            setError(
              "قیمت قبل از تخفیف باید بیشتر از قیمت نهایی باشد؛ برای حذف تخفیف فیلد را خالی کنید.",
            );
            return;
          }
          const command: Command = {
            action: price ? "offer.price.update" : "offer.stock.update",
            target: offer.id,
            version: current.version,
            idempotency_key: key,
            offer: price
              ? { price: amount, regular_price: beforeDiscount }
              : { stock: amount },
          };
          setError("");
          setPending(true);
          try {
            await execute(command);
            close();
          } catch (e) {
            setError(errorMessage(e));
            if (e instanceof ApiError && e.status === 409) setConflict(true);
          } finally {
            setPending(false);
          }
        }}
      >
        <DialogContent className="product-form">
          <p>
            <b>{w.stores.find((s) => s.id === offer.store_id)?.name}</b> ·{" "}
            {variantName(w.variants.find((v) => v.id === offer.variant_id)!)}
            <small dir="ltr">{offer.id}</small>
          </p>
          <div className="offer-stock-summary">
            <span>کل: {current.stock}</span>
            <span>رزروشده: {current.reserved}</span>
            <span>قابل تخصیص: {current.available}</span>
          </div>
          {error && <Alert severity="error">{error}</Alert>}
          {conflict && (
            <Button
              disabled={pending}
              onClick={async () => {
                setPending(true);
                try {
                  const latest = await loadProduct(w.product.id);
                  const next = latest.offers.find((o) => o.id === offer.id);
                  if (!next) throw new Error("عرضه دیگر در دسترس نیست.");
                  setCurrent(next);
                  cache.setQueryData(
                    ["product", surface, w.product.id],
                    latest,
                  );
                  setKey(crypto.randomUUID());
                  setConflict(false);
                  setError(
                    "نسخه تازه دریافت شد؛ ورودی شما حفظ شده است. مقدار جاری را بررسی و دوباره ثبت کنید.",
                  );
                } catch (e) {
                  setError(errorMessage(e));
                } finally {
                  setPending(false);
                }
              }}
            >
              بازخوانی و بازبینی
            </Button>
          )}
          {price && (
            <TextField
              label="قیمت قبل از تخفیف (تومان)"
              type="number"
              value={regular}
              disabled={pending}
              onChange={(e) => {
                setRegular(e.target.value);
                setKey(crypto.randomUUID());
              }}
              inputProps={{ min: 0.1, step: 0.1, max: 900000000000000 }}
              helperText="اختیاری؛ قیمت نهایی همان مبلغ قابل پرداخت مشتری است. خالی یعنی بدون تخفیف."
            />
          )}
          <TextField
            label={label}
            type="number"
            required
            value={value}
            disabled={pending}
            inputProps={{
              min: price ? 0.1 : current.reserved,
              max: price ? 900000000000000 : 1000000000,
              step: price ? 0.1 : 1,
            }}
            onChange={(e) => {
              setValue(e.target.value);
              setKey(crypto.randomUUID());
            }}
            helperText={
              price
                ? `قیمت جاری: ${money(current.price)}`
                : `رزرو قابل ویرایش نیست؛ موجودی باید حداقل ${current.reserved} باشد.`
            }
          />
          <p className="muted">
            وقفه تأمین: {current.lead_days} روز کاری؛ از فروشنده، فقط خواندنی.
          </p>
        </DialogContent>
        <DialogActions>
          <Button disabled={pending} onClick={close}>
            انصراف
          </Button>
          <Button
            variant="contained"
            type="submit"
            disabled={pending || conflict}
          >
            ثبت تغییر عرضه
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}

function VariantSelection({ workspace: w }: { workspace: ProductWorkspace }) {
  const { resolveVariant, surface } = useWorkspace();
  const [selection, setSelection] = useState<Record<string, string>>(() =>
    Object.fromEntries(
      w.variants[0]?.values.map((v) => [v.attribute_id, v.value]) ?? [],
    ),
  );
  const complete = w.axes.every((a) => !!selection[a.id]);
  const values = w.axes.map((a) => ({
    attribute_id: a.id,
    value: selection[a.id] ?? "",
  }));
  const query = useQuery({
    queryKey: ["variant-resolution", surface, w.product.id, values],
    queryFn: () => resolveVariant(w.product.id, values),
    enabled: complete,
  });
  return (
    <section className="product-workspace-section">
      <Alert severity="info">
        پیش‌نمایش انتخاب تنوع با داده جاری عرضه‌ها؛ این بخش ثبت سفارش یا محاسبه
        قیمت نهایی خرید نیست.
      </Alert>
      {w.axes.map((a) => (
        <div className="variant-chip-group" key={a.id}>
          <b>{a.name}</b>
          <div>
            {a.options.map((option) => (
              <Chip
                key={option}
                label={option}
                component="button"
                clickable
                aria-label={`${a.name}: ${option}`}
                aria-pressed={selection[a.id] === option}
                color={selection[a.id] === option ? "primary" : "default"}
                onClick={() => setSelection({ ...selection, [a.id]: option })}
              />
            ))}
          </div>
        </div>
      ))}
      {!complete && <p>مقادیر همه محورهای تنوع را انتخاب کنید.</p>}
      {query.isFetching && <p role="status">در حال دریافت انتخاب…</p>}
      {query.isError && (
        <Alert
          severity={
            query.error instanceof ApiError && query.error.status === 404
              ? "info"
              : "error"
          }
        >
          {query.error instanceof ApiError && query.error.status === 404
            ? "برای این ترکیب هنوز تنوعی تعریف نشده است."
            : errorMessage(query.error)}
        </Alert>
      )}
      {query.data && (
        <div data-testid="resolved-variant">
          <p>
            تنوع انتخاب‌شده: <b>{variantName(query.data.variant)}</b>
            <small dir="ltr">{query.data.variant.id}</small>
          </p>
          <DataTable
            rows={query.data.offers}
            empty="این تنوع عرضه قابل تخصیص در محدوده دسترسی شما ندارد."
            columns={[
              {
                title: "فروشنده",
                render: (o) => w.stores.find((s) => s.id === o.store_id)?.name,
              },
              { title: "قیمت نهایی نقدی", render: (o) => money(o.price) },
              { title: "قابل تخصیص", render: (o) => o.available },
              { title: "وقفه تأمین", render: (o) => `${o.lead_days} روز کاری` },
            ]}
          />
        </div>
      )}
    </section>
  );
}
