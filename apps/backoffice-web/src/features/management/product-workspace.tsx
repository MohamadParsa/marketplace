import { CreateButton } from "../../shared/create-button";
import { OptionSelect } from "../../shared/option-select";
import { OfferEditor } from "./offer-editor";
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
import { money, recordNumber, variantCode } from "../../shared/format";
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
const variantName = (v?: Variant) =>
  (v?.values ?? []).map((x) => x.value).join(" / ") || "بدون محور";
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
    <Dialog open fullWidth maxWidth="xl" onClose={close}>
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
  const [mediaOpen, setMediaOpen] = useState(false);
  const [tab, setTab] = useState(initialTab);
  const [editor, setEditor] = useState<
    "variant" | "offer" | { offer: OfferView }
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
          <small dir="ltr">{w.product.public_code || "کد در حال تخصیص"}</small>
          <small>
            {w.product.active === false
              ? "غیرفعال؛ خرید تازه مجاز نیست"
              : "فعال"}
          </small>
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
      </Tabs>
      {tab === 0 && (
        <section className="product-workspace-section">
          <div className="product-workspace-toolbar">
            <p className="muted">
              محورها: {w.axes.map((a) => a.name).join("، ") || "بدون محور تنوع"}
            </p>
            <CreateButton
              variant="contained"
              disabled={!w.permissions.create_variant}
              onClick={() => setEditor("variant")}
            >
              افزودن تنوع
            </CreateButton>
          </div>
          <DataTable
            rows={w.variants}
            columns={[
              {
                title: "ترکیب ویژگی‌ها",
                render: (v) => <b>{variantName(v)}</b>,
              },
              { title: "SKU", render: (v) => <span dir="ltr">{variantCode(w, v.id)}</span> },
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
            <CreateButton
              variant="contained"
              disabled={
                !w.permissions.create_offer_stores.length || !w.variants.length
              }
              onClick={() => setEditor("offer")}
            >
              افزودن عرضه
            </CreateButton>
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
                        : "تنوع در دسترس نیست"}
                    </small>
                    <small>عرضهٔ {recordNumber(w, "offer", o.id)}</small>
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
                render: (o) =>
                  `${o.lead_days} روز ${o.lead_day_basis === "calendar" ? "تقویمی" : "کاری"} · از فروشنده`,
              },
              {
                title: "عملیات",
                render: (o) => (
                  <>
                    <Button disabled={!o.store_active || (!w.permissions.price_stores.includes(o.store_id) && !w.permissions.stock_stores.includes(o.store_id))} onClick={() => setEditor({ offer: o })}>قیمت و موجودی</Button>
                  </>
                ),
              },
            ]}
          />
        </section>
      )}
      {tab === 2 && <VariantSelection key={id} workspace={w} />}
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
  const [selected, setSelected] = useState<Record<string, string>>({});
  const [version, setVersion] = useState(w.product.variant_version);
  const [key, setKey] = useState(crypto.randomUUID());
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  return (
    <Dialog
      open
      fullWidth
      maxWidth="xl"
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
                sku: "",
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
            فروشنده ثبت می‌شود. کد SKU به‌صورت خودکار با قالب chr و عدد ساخته می‌شود.
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
      maxWidth="xl"
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
                {variantName(v)} · {variantCode(w, v.id)}
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

function VariantSelection({ workspace: w }: { workspace: ProductWorkspace }) {
  const { resolveVariant, surface } = useWorkspace();
  const [manual, setManual] = useState<Record<string, string>>();
  const best = [...w.offers].filter(o => o.store_active && !o.sale_restricted && o.available > 0 && o.price > 0 && w.variants.some(v => v.id === o.variant_id)).sort((a, b) => a.price - b.price || a.lead_days - b.lead_days || a.id.localeCompare(b.id))[0];
  const automatic = w.variants.find(v => v.id === best?.variant_id) ?? w.variants[0];
  const selection = manual ?? Object.fromEntries(automatic?.values.map(v => [v.attribute_id, v.value]) ?? []);
  const setSelection = setManual;
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
        قیمت نهایی خرید نیست. در حالت خودکار، ارزان‌ترین عرضهٔ موجود و سپس تأمین زودتر انتخاب می‌شود.
      </Alert>
      <Button onClick={() => setManual(undefined)}>انتخاب خودکار بر اساس قیمت و تأمین</Button>
      {w.axes.map((a) => (
        <div className="variant-chip-group" key={a.id}>
          <div>
            <OptionSelect
              label={a.name}
              multiple={false}
              options={a.options.map((option) => ({
                id: String(option),
                name: String(option),
                selected: selection[a.id] === option,
                disabled: false,
                onSelect: () => setSelection({ ...selection, [a.id]: option }),
              }))}
            />
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
            <small dir="ltr">{variantCode(w, query.data.variant.id)}</small>
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
              {
                title: "وقفه تأمین",
                render: (o) =>
                  `${o.lead_days} روز ${o.lead_day_basis === "calendar" ? "تقویمی" : "کاری"}`,
              },
            ]}
          />
        </div>
      )}
    </section>
  );
}
