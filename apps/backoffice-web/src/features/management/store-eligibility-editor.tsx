import { CreateButton } from "../../shared/create-button";
import { ChoiceField, MultiChoiceField } from "../../shared/option-select";
import { OptionSelect } from "../../shared/option-select";
import { useState } from "react";
import {
  Alert,
  Button,
  Checkbox,
  Chip,
  LinearProgress,
  TextField,
} from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import { CategoryTree, categoryPath } from "./category-tree";
import type {
  Store,
  StoreEligibility,
  StoreSaleMethodChoice,
} from "../../shared/types";

type Input = {
  version: number;
  category_mode: "all" | "selected";
  allowed_categories: string[];
  denied_categories: string[];
  methods: { method_id: string; display_name: string }[];
};
export function StoreEligibilityEditor({
  store,
  methods,
  allowed,
  save,
  saved,
  locked,
  onBusyChange,
}: {
  store: Store & { version: number };
  methods: StoreSaleMethodChoice[];
  allowed: boolean;
  save: (input: Input, key: string) => Promise<Store & { version: number }>;
  saved: (row: Store & { version: number }) => void;
  locked: boolean;
  onBusyChange: (busy: boolean) => void;
}) {
  const { state } = useWorkspace(),
    categories = state.categories ?? [],
    policy = store.eligibility;
  const [mode, setMode] = useState<"all" | "selected">(
      policy?.category_mode ?? "all",
    ),
    [allow, setAllow] = useState(policy?.allowed_categories ?? []),
    [deny, setDeny] = useState(policy?.denied_categories ?? []),
    [enabled, setEnabled] = useState(
      policy?.methods.map((m) => m.method_id) ?? methods.map((m) => m.id),
    ),
    [aliases, setAliases] = useState<Record<string, string>>(
      Object.fromEntries(
        (policy?.methods ?? []).map((m) => [m.method_id, m.display_name ?? ""]),
      ),
    ),
    [category, setCategory] = useState(""),
    [preview, setPreview] = useState(""),
    [busy, setBusy] = useState(false),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [key, setKey] = useState(crypto.randomUUID()),
    [dirty, setDirty] = useState(!policy);
  const changed = () => {
    setDirty(true);
    setKey(crypto.randomUUID());
    setNotice("");
  };
  const previewCategory = () => {
    if (!category) return undefined;
    let current = category,
      permit = mode === "all";
    const seen = new Set<string>();
    while (current) {
      if (seen.has(current)) return false;
      seen.add(current);
      if (deny.includes(current)) return false;
      if (allow.includes(current)) permit = true;
      const c = categories.find((c) => c.id === current);
      if (!c) return false;
      current = c.parent_id ?? "";
    }
    return permit;
  };

  async function submit() {
    if (busy || locked || !allowed) return;
    setBusy(true);
    onBusyChange(true);
    setError("");
    setNotice("");
    try {
      const next = await save(
        {
          version: store.version,
          category_mode: mode,
          allowed_categories: allow,
          denied_categories: deny,
          methods: enabled.map((id) => ({
            method_id: id,
            display_name: aliases[id] ?? "",
          })),
        },
        key,
      );
      const result = next.eligibility as StoreEligibility;
      setMode(result.category_mode);
      setAllow(result.allowed_categories);
      setDeny(result.denied_categories);
      setEnabled(result.methods.map((m) => m.method_id));
      setAliases(
        Object.fromEntries(
          result.methods.map((m) => [m.method_id, m.display_name ?? ""]),
        ),
      );
      setKey(crypto.randomUUID());
      setDirty(false);
      setNotice("مجوزهای فروشگاه ذخیره شد.");
      saved(next);
    } catch (e) {
      setError(e instanceof Error ? e.message : "ذخیره مجوز انجام نشد.");
    } finally {
      setBusy(false);
      onBusyChange(false);
    }
  }
  return (
    <section
      className="store-detail-box store-policy-editor"
      aria-label="مجوزهای فروشگاه"
    >
      <h3>دسته‌ها و روش‌های فروش مجاز</h3>
      {!allowed && (
        <Alert severity="info">
          مجوز مدیریت سیاست فروشگاه ندارید؛ فقط مشاهده می‌کنید.
        </Alert>
      )}
      {!policy && (
        <Alert severity="info">
          این فروشگاه هنوز سیاست اختصاصی ندارد؛ مجوز قبلی تمام دسته‌ها و روش‌ها
          تا ثبت این فرم حفظ می‌شود.
        </Alert>
      )}
      {error && <Alert severity="error">{error}</Alert>}
      {notice && <Alert severity="success">{notice}</Alert>}
      {busy && <LinearProgress aria-label="ذخیره مجوز فروشگاه" />}
      <fieldset disabled={busy || locked || !allowed} className="axis-inputs">
        <legend>محدوده دسته‌ها</legend>
        <div className="toolbar" role="group" aria-label="حالت دسته‌های مجاز">
          <OptionSelect
            label={"حالت دسته‌های مجاز"}
            multiple={false}
            options={(
              [
                ["all", "همه دسته‌ها"],
                ["selected", "فقط دسته‌های انتخابی"],
              ] as const
            ).map(([id, name]) => ({
              id: String(id),
              name: String("" + name + ""),
              selected: mode === id,
              disabled: false,
              onSelect: () => {
                setMode(id);
                changed();
              },
            }))}
          />
        </div>
        <p>
          انتخاب هر دسته شامل زیرشاخه‌های فعلی و آینده است؛ ممنوعیت همیشه بر
          مجوز مقدم است.
        </p>
        <CategoryTree
          categories={categories}
          value={category}
          onChange={setCategory}
          allowRoot={false}
        />
        <div className="toolbar">
          <CreateButton
            disabled={
              !category || mode !== "selected" || allow.includes(category)
            }
            onClick={() => {
              setAllow((old) => [...old, category]);
              changed();
            }}
          >
            افزودن به دسته‌های مجاز
          </CreateButton>
          <CreateButton
            disabled={!category || deny.includes(category)}
            onClick={() => {
              setDeny((old) => [...old, category]);
              changed();
            }}
          >
            افزودن به دسته‌های ممنوع
          </CreateButton>
        </div>
        <h4>دسته‌های مجاز</h4>
        {mode === "all" ? (
          <p>همه دسته‌ها، به‌جز شاخه‌های ممنوع پایین.</p>
        ) : (
          <MultiChoiceField
            label="دسته‌های مجاز"
            value={allow}
            disabled={busy || locked || !allowed}
            options={categories.map((c) => ({
              id: c.id,
              name: categoryPath(categories, c.id),
            }))}
            onChange={(ids) => {
              setAllow(ids);
              changed();
            }}
          />
        )}
        {mode === "selected" && !allow.length && (
          <Alert severity="warning">
            هنوز دسته مجازی انتخاب نشده؛ این سیاست خرید تازه را برای همه دسته‌ها
            می‌بندد.
          </Alert>
        )}
        <h4>دسته‌های ممنوع</h4>
        {
          <MultiChoiceField
            label="دسته‌های ممنوع"
            value={deny}
            disabled={busy || locked || !allowed}
            options={categories.map((c) => ({
              id: c.id,
              name: categoryPath(categories, c.id),
            }))}
            onChange={(ids) => {
              setDeny(ids);
              changed();
            }}
          />
        }
        {category && (
          <Alert severity={previewCategory() ? "success" : "warning"}>
            نتیجه این پیش‌نویس برای «{categoryPath(categories, category)}»:{" "}
            {previewCategory() ? "دسته مجاز است" : "دسته مجاز نیست"}؛ روش فروش
            نیز جدا بررسی می‌شود.
          </Alert>
        )}
        <h4>روش‌های فروش</h4>
        <p>
          در PDP فعلی فقط قیمت نقدی و نام اصلی نمایش داده می‌شود. نام اختصاصی پس
          از انتخاب روش در مسیر خرید اعمال می‌شود؛ هویت فروشگاه تغییر نمی‌کند.
        </p>
        <MultiChoiceField
          label="روش‌های فروش مجاز"
          value={enabled}
          options={methods.map((m) => ({ id: m.id, name: m.name }))}
          disabled={busy || locked || !allowed}
          onChange={(ids) => {
            setEnabled(ids);
            changed();
          }}
        />
        {methods.map((m) => (
          <div className="store-policy-method" key={m.id}>
            <></>
            <TextField
              size="small"
              label={`نام نمایشی فروشگاه برای ${m.name}`}
              value={aliases[m.id] ?? ""}
              disabled={!enabled.includes(m.id)}
              helperText="اختیاری؛ نام خالی از نام اصلی استفاده می‌کند."
              inputProps={{ maxLength: 160 }}
              onChange={(e) => {
                setAliases((old) => ({ ...old, [m.id]: e.target.value }));
                changed();
              }}
            />
          </div>
        ))}
        {enabled
          .filter((id) => !methods.some((m) => m.id === id))
          .map((id) => (
            <Alert
              key={id}
              severity="warning"
              action={
                <Button
                  onClick={() => {
                    setEnabled((old) => old.filter((x) => x !== id));
                    changed();
                  }}
                >
                  حذف از سیاست
                </Button>
              }
            >
              یک روش ذخیره‌شده دیگر در تنظیمات روش‌های فروش موجود نیست.
            </Alert>
          ))}
        {!enabled.includes("cash") && (
          <Alert severity="warning">
            مجوز نقدی انتخاب نشده است؛ نمایش و خرید نقدی فعلی این فروشگاه در PDP
            غیرفعال می‌شود. نمایش روش‌محور PDP تسک آینده است.
          </Alert>
        )}
        {!enabled.length && (
          <Alert severity="warning">
            هیچ روش فروشی مجاز نیست؛ عرضه‌ها حذف نمی‌شوند ولی خرید و ثبت/ویرایش
            عرضه تازه مجاز نخواهد بود.
          </Alert>
        )}
        <h4>پیش‌نمایش نام</h4>
        <ChoiceField
          label="پیش‌نمایش نام فروشگاه"
          value={preview}
          clearLabel="پیش از انتخاب روش"
          options={methods
            .filter((m) => enabled.includes(m.id))
            .map((m) => ({ id: m.id, name: m.name }))}
          onChange={setPreview}
        />
        <p className="store-policy-name-preview">
          {preview && enabled.includes(preview)
            ? aliases[preview]?.trim() || store.name
            : store.name}
        </p>
        <p className="muted">
          تغییر سیاست، پیش‌فاکتور قبلی را نیازمند بررسی دوباره می‌کند؛ سفارش‌های
          قطعی و دفتر مالی بازنویسی نمی‌شوند.
        </p>
        <Button
          variant="contained"
          disabled={!dirty || busy || locked || !allowed}
          onClick={() => void submit()}
        >
          ذخیره مجوزهای فروشگاه
        </Button>
      </fieldset>
    </section>
  );
}
