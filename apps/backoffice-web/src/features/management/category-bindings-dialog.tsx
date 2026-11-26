import { useMemo, useState } from "react";
import {
  Alert,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
} from "@mui/material";
import { Link } from "react-router-dom";
import { useWorkspace } from "../../shared/workspace";
import { PaginationControls, usePagedRows } from "../../shared/pagination";
import type { AttributeDefinition, Category } from "../../shared/types";

const kinds = { enum: "انتخابی", text: "متن", number: "عدد" };
const roleTitles = {
  axis: "محور تنوع",
  card: "نمایش کارت",
  filter: "فیلتر PLP",
};
export function CategoryBindingsDialog({
  category,
  close,
}: {
  category: Category;
  close: () => void;
}) {
  const { state, execute, can } = useWorkspace();
  const [original, setOriginal] = useState(category),
    [bindings, setBindings] = useState(category.bindings ?? []),
    [search, setSearch] = useState(""),
    [error, setError] = useState(""),
    [pending, setPending] = useState(false),
    [key, setKey] = useState(crypto.randomUUID());
  const attributes = state.attributes ?? [],
    editable = can("category.bindings.set", "management") && !original.archived;
  const products = useMemo(
    () => (state.products ?? []).filter((p) => p.category_id === original.id),
    [state.products, original.id],
  );
  const impacts = useMemo(() => {
    const map = new Map<string, typeof products>();
    for (const product of products)
      for (const value of product.attributes ?? [])
        map.set(value.attribute_id, [
          ...(map.get(value.attribute_id) ?? []),
          product,
        ]);
    return map;
  }, [products]);
  const latest = state.categories?.find((c) => c.id === original.id),
    stale = latest && latest.version !== original.version;
  const matches = (a: AttributeDefinition) =>
    (a.name + " " + a.code + " " + kinds[a.kind])
      .toLocaleLowerCase()
      .includes(search.trim().toLocaleLowerCase());
  const assigned = bindings
    .map((b) => ({
      binding: b,
      attribute: attributes.find((a) => a.id === b.attribute_id),
    }))
    .filter(
      (
        v,
      ): v is {
        binding: (typeof bindings)[number];
        attribute: AttributeDefinition;
      } => !!v.attribute,
    )
    .sort(
      (a, b) =>
        a.binding.position - b.binding.position ||
        a.attribute.name.localeCompare(b.attribute.name, "fa"),
    );
  const available = usePagedRows(
    attributes.filter(
      (a) => !bindings.some((b) => b.attribute_id === a.id) && matches(a),
    ),
  );
  const change = (
    id: string,
    field: "axis" | "card" | "filter" | "position",
    value: boolean | number,
  ) => {
    setBindings((old) =>
      old.map((b) => (b.attribute_id === id ? { ...b, [field]: value } : b)),
    );
    setKey(crypto.randomUUID());
  };
  const add = (a: AttributeDefinition) => {
    setBindings((old) => [
      ...old,
      {
        attribute_id: a.id,
        axis: false,
        card: false,
        filter: false,
        position: Math.min(
          10000,
          Math.max(-1, ...old.map((b) => b.position)) + 1,
        ),
      },
    ]);
    setKey(crypto.randomUUID());
  };
  return (
    <Dialog
      open
      fullWidth
      maxWidth="lg"
      onClose={() => {
        if (!pending) close();
      }}
    >
      <DialogTitle>خصوصیات دسته {original.name}</DialogTitle>
      <DialogContent className="category-bindings-editor">
        <Alert severity="info">
          خصوصیت یک تعریف مشترک در کتابخانه دارد؛ نقش و ترتیب آن فقط برای همین
          دسته است و از والد به ارث نمی‌رسد. ترتیب عددی کوچک‌تر جلوتر است؛ یک
          ترتیب مشترک در هر سه نقش اعمال می‌شود.
        </Alert>
        {!editable && (
          <Alert severity="warning">
            {original.archived
              ? "دسته بایگانی شده است؛ فقط می‌توانید تنظیمات را ببینید."
              : "مجوز تغییر خصوصیات این دسته را ندارید؛ فرم فقط خواندنی است."}
          </Alert>
        )}
        {stale && (
          <Alert
            severity="warning"
            action={
              <Button
                disabled={pending}
                onClick={() => {
                  if (latest) {
                    setOriginal(latest);
                    setBindings(latest.bindings ?? []);
                    setKey(crypto.randomUUID());
                    setError("");
                  }
                }}
              >
                بارگیری نسخه تازه
              </Button>
            }
          >
            تنظیمات دسته در جای دیگری تغییر کرده است؛ پیش از ثبت، نسخه تازه را
            بازبینی کنید. بارگیری، پیش‌نویس فعلی را جایگزین می‌کند.
          </Alert>
        )}
        {error && <Alert severity="error">{error}</Alert>}
        <TextField
          size="small"
          label="جست‌وجوی خصوصیات دسته"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <section className="binding-assigned" aria-label="خصوصیات منتسب">
          <h3>خصوصیات منتسب ({bindings.length.toLocaleString("fa-IR")})</h3>
          {assigned
            .filter((x) => matches(x.attribute))
            .map(({ attribute: a, binding: b }) => {
              const originalBinding = original.bindings?.find(
                (v) => v.attribute_id === a.id,
              );
              const valueUsers = impacts.get(a.id) ?? [];
              const axisLocked = products.length > 0;
              const removalLocked =
                valueUsers.length > 0 ||
                (axisLocked && !!originalBinding?.axis);
              const removeReason = !editable
                ? "مجوز حذف انتساب ندارید."
                : valueUsers.length
                  ? `${valueUsers.length.toLocaleString("fa-IR")} محصول مقدار این خصوصیت را دارد؛ حذف بدون تبدیل داده مجاز نیست.`
                  : axisLocked && originalBinding?.axis
                    ? `این محور در دسته دارای ${products.length.toLocaleString("fa-IR")} محصول استفاده شده؛ تغییر آن نیازمند تبدیل کنترل‌شده است.`
                    : "حذف انتساب آزاد است؛ تعریف مشترک در کتابخانه باقی می‌ماند.";
              const axisReason = !editable
                ? "مجوز تغییر محور ندارید."
                : a.kind !== "enum"
                  ? "محور تنوع فقط برای خصوصیت انتخابی قابل استفاده است."
                  : axisLocked
                    ? "این دسته محصول دارد؛ افزودن یا حذف محور از مسیر تبدیل کنترل‌شده انجام می‌شود."
                    : "این محور انتخاب مشتری در PDP و ساخت تنوع را مشخص می‌کند.";
              const dependent = originalBinding?.axis ? products : valueUsers;
              return (
                <article
                  className="binding-assigned-row"
                  key={a.id}
                  data-testid={`binding-${a.id}`}
                >
                  <div className="binding-row-heading">
                    <div>
                      <b>{a.name}</b>
                      <small>
                        {kinds[a.kind]}
                        {a.unit ? ` · ${a.unit}` : ""}
                      </small>
                    </div>
                    <Button
                      color="error"
                      disabled={!editable || pending || removalLocked}
                      aria-label={`حذف انتساب ${a.name}`}
                      onClick={() => {
                        setBindings((old) =>
                          old.filter((v) => v.attribute_id !== a.id),
                        );
                        setKey(crypto.randomUUID());
                      }}
                    >
                      حذف انتساب
                    </Button>
                  </div>
                  <p
                    className={removalLocked ? "binding-lock-reason" : "muted"}
                  >
                    {removeReason}
                  </p>
                  <div className="binding-role-controls">
                    {(["axis", "card", "filter"] as const).map((flag) => (
                      <label key={flag}>
                        <Checkbox
                          checked={b[flag]}
                          disabled={
                            !editable ||
                            pending ||
                            (flag === "axis" &&
                              (axisLocked || a.kind !== "enum"))
                          }
                          inputProps={{
                            "aria-label": `${roleTitles[flag]} ${a.name}`,
                            "aria-describedby":
                              flag === "axis" ? `axis-note-${a.id}` : undefined,
                          }}
                          onChange={(_, checked) => change(a.id, flag, checked)}
                        />
                        {roleTitles[flag]}
                      </label>
                    ))}
                    <TextField
                      size="small"
                      type="number"
                      label={`ترتیب ${a.name}`}
                      value={b.position}
                      inputProps={{ min: 0, max: 10000, step: 1 }}
                      disabled={!editable || pending}
                      onChange={(e) =>
                        change(a.id, "position", Number(e.target.value))
                      }
                    />
                  </div>
                  <p className="binding-axis-note" id={`axis-note-${a.id}`}>
                    {axisReason}
                  </p>
                  {!!dependent.length && (
                    <details className="binding-dependencies">
                      <summary>
                        {dependent.length.toLocaleString("fa-IR")} کالای وابسته
                        و بررسی اثر
                      </summary>
                      <ul>
                        {dependent.slice(0, 10).map((p) => (
                          <li key={p.id}>
                            <Link
                              to={`/products/${encodeURIComponent(p.id)}`}
                              onClick={close}
                            >
                              {p.name}
                            </Link>
                          </li>
                        ))}
                      </ul>
                      <Link
                        to={`/products?category=${encodeURIComponent(original.id)}`}
                        onClick={close}
                      >
                        مشاهده همه محصولات این دسته
                      </Link>
                      {axisLocked && (
                        <p>
                          تبدیل محور مصرف‌شده در{" "}
                          <a
                            href="https://github.com/MohamadParsa/marketplace/issues/109"
                            target="_blank"
                            rel="noreferrer"
                          >
                            تسک ۱۰۹
                          </a>{" "}
                          پیگیری می‌شود؛ هنوز عملیات تبدیل از این فرم اجرا
                          نمی‌شود.
                        </p>
                      )}
                    </details>
                  )}
                </article>
              );
            })}
          {!assigned.filter((x) => matches(x.attribute)).length && (
            <p>خصوصیت منتسبی با این جست‌وجو پیدا نشد.</p>
          )}
        </section>
        <section
          className="binding-role-previews"
          aria-label="پیش‌نمایش نقش خصوصیات"
        >
          {(["axis", "card", "filter"] as const).map((flag) => (
            <div key={flag}>
              <h4>{roleTitles[flag]}</h4>
              <p className="muted">
                {flag === "axis"
                  ? "انتخاب تنوع در PDP"
                  : flag === "card"
                    ? `تا ${state.display_settings.card_attribute_limit.toLocaleString("fa-IR")} ویژگی دارای مقدار روی کارت`
                    : "فیلترهای PLP همین دسته"}
              </p>
              <div className="binding-preview-chips">
                {assigned
                  .filter((x) => x.binding[flag])
                  .map((x) => (
                    <span key={x.attribute.id}>{x.attribute.name}</span>
                  ))}
                {!assigned.some((x) => x.binding[flag]) && (
                  <small>هیچ خصوصیتی انتخاب نشده است.</small>
                )}
              </div>
            </div>
          ))}
        </section>
        <section className="binding-available" aria-label="خصوصیات قابل افزودن">
          <h3>خصوصیات قابل افزودن</h3>
          <p>
            از همین کتابخانه انتخاب کنید؛ تعریف جدید یا تکراری برای فرزند ساخته
            نمی‌شود.
          </p>
          <div className="binding-available-list">
            {available.rows.map((a) => (
              <div key={a.id}>
                <div>
                  <b>{a.name}</b>
                  <small>
                    {kinds[a.kind]}
                    {a.unit ? ` · ${a.unit}` : ""}
                  </small>
                </div>
                <Button
                  disabled={!editable || pending || bindings.length >= 100}
                  aria-label={`افزودن خصوصیت ${a.name}`}
                  onClick={() => add(a)}
                >
                  افزودن
                </Button>
              </div>
            ))}
          </div>
          {!available.controls.total && (
            <p>خصوصیت قابل افزودنی با این جست‌وجو پیدا نشد.</p>
          )}
          <PaginationControls {...available.controls} />
        </section>
      </DialogContent>
      <DialogActions>
        <Button disabled={pending} onClick={close}>
          انصراف
        </Button>
        <Button
          variant="contained"
          disabled={
            !editable ||
            pending ||
            bindings.some(
              (b) =>
                !Number.isInteger(b.position) ||
                b.position < 0 ||
                b.position > 10000,
            )
          }
          onClick={async () => {
            if (pending) return;
            setPending(true);
            setError("");
            try {
              await execute({
                action: "category.bindings.set",
                target: original.id,
                version: original.version,
                bindings,
                idempotency_key: key,
              });
              close();
            } catch (e) {
              setError(
                e instanceof Error
                  ? e.message
                  : "ثبت انجام نشد؛ دوباره تلاش کنید.",
              );
            } finally {
              setPending(false);
            }
          }}
        >
          ثبت انتساب‌ها
        </Button>
      </DialogActions>
    </Dialog>
  );
}
