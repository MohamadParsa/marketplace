import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  MenuItem,
  TextField,
} from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import { DataTable, PageHeader } from "../../shared/components";
import type { AttributeDefinition, Category } from "../../shared/types";

const kinds = { enum: "انتخابی", text: "متن", number: "عدد" };
const message = (e: unknown) =>
  e instanceof Error ? e.message : "تغییر ثبت نشد؛ دوباره تلاش کنید.";

export function SharedAttributesPage() {
  const { state, execute, can: may } = useWorkspace();
  const [editor, setEditor] = useState<AttributeDefinition>();
  const [search, setSearch] = useState("");
  const [limit, setLimit] = useState(
    String(state.display_settings.card_attribute_limit),
  );
  const [settingsVersion, setSettingsVersion] = useState(
    state.display_settings.version,
  );
  const [settingsKey, setSettingsKey] = useState(crypto.randomUUID());
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  useEffect(
    () => setSettingsVersion(state.display_settings.version),
    [state.display_settings.version],
  );
  const editable = may("attribute.create", "management");
  return (
    <>
      <PageHeader
        title="کتابخانه خصوصیات"
        description="یک تعریف مشترک؛ کاربرد مستقل در هر دسته، بدون ارث از والد"
        action={
          <Button
            variant="contained"
            disabled={!editable}
            onClick={() =>
              setEditor({
                id: "",
                code: "",
                name: "",
                kind: "enum",
                options: [],
                version: 0,
              })
            }
          >
            خصوصیت جدید
          </Button>
        }
      />
      <section className="panel">
        <div className="toolbar">
          <input
            className="search"
            aria-label="جست‌وجوی خصوصیت"
            placeholder="نام یا کد خصوصیت…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <DataTable
          rows={(state.attributes ?? []).filter((a) =>
            (a.name + " " + a.code)
              .toLocaleLowerCase()
              .includes(search.toLocaleLowerCase()),
          )}
          columns={[
            {
              title: "خصوصیت",
              render: (a) => (
                <>
                  <b>{a.name}</b>
                  <small dir="ltr">{a.code}</small>
                </>
              ),
            },
            {
              title: "نوع / واحد",
              render: (a) => `${kinds[a.kind]}${a.unit ? " / " + a.unit : ""}`,
            },
            {
              title: "دسته‌های مرتبط",
              render: (a) =>
                (state.categories ?? [])
                  .filter((c) =>
                    c.bindings?.some((b) => b.attribute_id === a.id),
                  )
                  .map((c) => c.name)
                  .join("، ") || "بدون انتساب",
            },
            {
              title: "عملیات",
              render: (a) => (
                <Button onClick={() => setEditor(a)}>مشاهده و ویرایش</Button>
              ),
            },
          ]}
        />
      </section>
      <section className="panel attribute-settings">
        <h2>تنظیمات نمایش کالا</h2>
        <p className="muted">
          این بخش فقط تنظیم سقف کارتِ مرتبط با این تسک است. مقدار ۳ داده آزمون
          است و انتخاب نهایی تولید محسوب نمی‌شود.
        </p>
        <form
          onSubmit={async (e) => {
            e.preventDefault();
            if (pending || !may("settings.card-limit", "management")) return;
            setPending(true);
            setError("");
            try {
              await execute({
                action: "settings.card-limit",
                target: "display",
                amount: Number(limit),
                version: settingsVersion,
                idempotency_key: settingsKey,
              });
              setSettingsVersion(settingsVersion + 1);
              setSettingsKey(crypto.randomUUID());
            } catch (e) {
              setError(message(e));
            } finally {
              setPending(false);
            }
          }}
        >
          <TextField
            size="small"
            label="حداکثر ویژگی کارت"
            type="number"
            value={limit}
            disabled={!may("settings.card-limit", "management") || pending}
            onChange={(e) => {
              setLimit(e.target.value);
              setSettingsKey(crypto.randomUUID());
            }}
            inputProps={{ min: 1, max: 20, step: 1 }}
            required
          />
          <Button
            type="submit"
            variant="contained"
            disabled={!may("settings.card-limit", "management") || pending}
          >
            ثبت سقف کارت
          </Button>
        </form>
        {error && <Alert severity="error">{error}</Alert>}
        <p>
          سقف ثبت‌شده:{" "}
          <b data-testid="saved-card-limit">
            {state.display_settings.card_attribute_limit}
          </b>{" "}
          · تغییر سقف فقط روی نمایش جاری اثر دارد.
        </p>
      </section>
      <SalesWindowSettings />
      <section className="panel attribute-settings">
        <h2>پیش‌نمایش کارت با داده سرور</h2>
        <p className="muted">
          فقط خصوصیات صریح همان دسته، با ترتیب ثبت‌شده و سقف عمومی؛ بدون تغییر
          محصول یا سفارش.
        </p>
        <div className="attribute-card-grid">
          {(state.products ?? [])
            .filter((p) => p.attributes?.length)
            .map((p) => (
              <article
                className="attribute-card"
                key={p.id}
                data-testid={`card-${p.id}`}
              >
                <span className="overline">
                  کارت نمونه ·{" "}
                  {state.categories?.find((c) => c.id === p.category_id)?.name}
                </span>
                <h3>{p.name}</h3>
                <ul>
                  {state.card_previews
                    ?.find((x) => x.product_id === p.id)
                    ?.features.map((f) => (
                      <li key={f.attribute_id}>
                        {f.name}: <strong>{f.value}</strong> {f.unit}
                      </li>
                    ))}
                </ul>
              </article>
            ))}
        </div>
      </section>
      {editor && (
        <AttributeEditor
          key={editor.id || "new"}
          attribute={editor}
          close={() => setEditor(undefined)}
        />
      )}
    </>
  );
}

function AttributeEditor({
  attribute,
  close,
}: {
  attribute: AttributeDefinition;
  close: () => void;
}) {
  const { state, execute, can: may } = useWorkspace();
  const [draft, setDraft] = useState(attribute);
  const [options, setOptions] = useState(attribute.options.join("\n"));
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  const [key, setKey] = useState(crypto.randomUUID());
  const [deleteKey] = useState(crypto.randomUUID());
  const affected = (state.categories ?? []).filter((c) =>
    c.bindings?.some((b) => b.attribute_id === attribute.id),
  );
  const used =
    affected.length > 0 ||
    (state.products ?? []).some((p) =>
      p.attributes?.some((v) => v.attribute_id === attribute.id),
    );
  const editable = may(
    attribute.id ? "attribute.update" : "attribute.create",
    "management",
  );
  const groupEditable = attribute.id
    ? may("attribute.groups.set", "management")
    : editable;
  const change = (field: keyof AttributeDefinition, value: string) => {
    setDraft({ ...draft, [field]: value });
    setKey(crypto.randomUUID());
  };
  async function save(remove = false) {
    if (pending || !(remove ? may("attribute.delete", "management") : editable))
      return;
    const values = options
      .split("\n")
      .map((x) => x.trim())
      .filter(Boolean);
    if (
      !remove &&
      draft.kind === "enum" &&
      (!values.length ||
        new Set(values.map((x) => x.toLowerCase())).size !== values.length)
    ) {
      setError("گزینه‌ها نباید خالی یا تکراری باشند.");
      return;
    }
    setPending(true);
    setError("");
    try {
      await execute(
        remove
          ? {
              action: "attribute.delete",
              target: attribute.id,
              version: attribute.version,
              idempotency_key: deleteKey,
            }
          : {
              action: attribute.id ? "attribute.update" : "attribute.create",
              target: attribute.id,
              version: attribute.version,
              attribute: {
                ...draft,
                options: draft.kind === "enum" ? values : [],
              },
              idempotency_key: key,
            },
      );
      close();
    } catch (e) {
      setError(message(e));
    } finally {
      setPending(false);
    }
  }
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
        {attribute.id ? "ویرایش خصوصیت مشترک" : "خصوصیت جدید"}
      </DialogTitle>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          void save();
        }}
      >
        <DialogContent className="attribute-form">
          <Alert severity="info">
            تغییر نام روی همه دسته‌های مرتبط اثر دارد:{" "}
            {affected.map((c) => c.name).join("، ") || "بدون انتساب"}.{" "}
            {used &&
              "نوع، واحد و گزینه‌های خصوصیتِ مورد استفاده تا تعیین سیاست تبدیل قفل‌اند."}
          </Alert>
          {error && <Alert severity="error">{error}</Alert>}
          <TextField
            label="نام خصوصیت"
            value={draft.name}
            onChange={(e) => change("name", e.target.value)}
            required
            disabled={!editable || pending}
            inputProps={{ maxLength: 80 }}
          />
          <TextField
            label="کد یکتا"
            value={draft.code}
            onChange={(e) => change("code", e.target.value)}
            required
            disabled={!editable || pending || !!attribute.id}
            inputProps={{ pattern: "[a-z][a-z0-9\\-]{0,63}", dir: "ltr" }}
            helperText="حروف کوچک انگلیسی، رقم و خط تیره؛ پس از ایجاد ثابت است."
          />
          <TextField
            select
            label="نوع خصوصیت"
            value={draft.kind}
            onChange={(e) => change("kind", e.target.value)}
            disabled={!editable || pending || used}
          >
            {Object.entries(kinds).map(([k, v]) => (
              <MenuItem key={k} value={k}>
                {v}
              </MenuItem>
            ))}
          </TextField>
          <TextField
            label="واحد (اختیاری)"
            value={draft.unit ?? ""}
            onChange={(e) => change("unit", e.target.value)}
            disabled={!editable || pending || used}
            inputProps={{ maxLength: 32 }}
          />
          {draft.kind === "enum" && (
            <TextField
              label="گزینه‌ها، هر خط یک مقدار"
              multiline
              minRows={3}
              value={options}
              onChange={(e) => {
                setOptions(e.target.value);
                setKey(crypto.randomUUID());
              }}
              disabled={!editable || pending || used}
            />
          )}
          {draft.kind === "enum" && (
            <section aria-label="گروه‌های نمایشی مقادیر">
              <h3>گروه نمایشی گارانتی</h3>
              <p>
                عنوان گروه فقط نمایش را ساده می‌کند. نام واقعی، قیمت و موجودی هر
                عرضه حفظ می‌شود؛ ذخیره روی همه محصولات مرتبط اثر دارد.
              </p>
              {(draft.groups ?? []).map((group, index) => (
                <fieldset key={group.id} disabled={!groupEditable || pending}>
                  <legend>{group.name || "گروه جدید"}</legend>
                  <TextField
                    label="عنوان گروه"
                    value={group.name}
                    inputProps={{ maxLength: 80 }}
                    onChange={(e) => {
                      setDraft({
                        ...draft,
                        groups: draft.groups!.map((g, i) =>
                          i === index ? { ...g, name: e.target.value } : g,
                        ),
                      });
                      setKey(crypto.randomUUID());
                    }}
                  />
                  <label>
                    <Checkbox
                      checked={group.active}
                      onChange={(e) => {
                        setDraft({
                          ...draft,
                          groups: draft.groups!.map((g, i) =>
                            i === index
                              ? { ...g, active: e.target.checked }
                              : g,
                          ),
                        });
                        setKey(crypto.randomUUID());
                      }}
                    />
                    فعال در نمایش
                  </label>
                  <div className="warranty-group-values">
                    {options
                      .split("\n")
                      .filter(Boolean)
                      .map((value) => (
                        <label key={value}>
                          <Checkbox
                            checked={group.values.includes(value)}
                            onChange={(e) => {
                              setDraft({
                                ...draft,
                                groups: draft.groups!.map((g, i) =>
                                  i === index
                                    ? {
                                        ...g,
                                        values: e.target.checked
                                          ? [...g.values, value]
                                          : g.values.filter((v) => v !== value),
                                      }
                                    : g,
                                ),
                              });
                              setKey(crypto.randomUUID());
                            }}
                          />
                          {value}
                        </label>
                      ))}
                  </div>
                  <Button
                    color="error"
                    onClick={() => {
                      setDraft({
                        ...draft,
                        groups: draft.groups!.filter((_, i) => i !== index),
                      });
                      setKey(crypto.randomUUID());
                    }}
                  >
                    حذف گروه نمایشی
                  </Button>
                </fieldset>
              ))}
              <Button
                disabled={
                  !groupEditable ||
                  pending ||
                  (draft.groups?.length ?? 0) >= 100
                }
                onClick={() => {
                  setDraft({
                    ...draft,
                    groups: [
                      ...(draft.groups ?? []),
                      {
                        id: "group-" + crypto.randomUUID(),
                        name: "",
                        values: [],
                        active: true,
                      },
                    ],
                  });
                  setKey(crypto.randomUUID());
                }}
              >
                افزودن گروه نمایشی
              </Button>
            </section>
          )}
        </DialogContent>
        <DialogActions>
          <Button disabled={pending} onClick={close}>
            انصراف
          </Button>
          {attribute.id && (
            <Button
              disabled={!groupEditable || pending}
              onClick={async () => {
                if (pending) return;
                setPending(true);
                setError("");
                try {
                  await execute({
                    action: "attribute.groups.set",
                    target: attribute.id,
                    version: attribute.version,
                    attribute: {
                      ...attribute,
                      options: [],
                      groups: draft.groups ?? [],
                    },
                    idempotency_key: key,
                  });
                  close();
                } catch (e) {
                  setError(message(e));
                } finally {
                  setPending(false);
                }
              }}
            >
              ثبت فقط گروه‌های نمایشی
            </Button>
          )}
          {attribute.id && (
            <Button
              color="error"
              disabled={
                !may("attribute.delete", "management") || pending || used
              }
              onClick={() => void save(true)}
            >
              حذف خصوصیت بدون استفاده
            </Button>
          )}
          <Button
            variant="contained"
            type="submit"
            disabled={!editable || pending}
          >
            ثبت خصوصیت
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}

export function CategoryBindingsDialog({
  category,
  close,
}: {
  category: Category;
  close: () => void;
}) {
  const { state, execute, can: may } = useWorkspace();
  const [original] = useState(category);
  const [bindings, setBindings] = useState(category.bindings ?? []);
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  const [key, setKey] = useState(crypto.randomUUID());
  const editable =
    may("category.bindings.set", "management") && !category.archived;
  const used = (state.products ?? []).some(
    (p) => p.category_id === category.id,
  );
  return (
    <Dialog
      open
      fullWidth
      maxWidth="md"
      onClose={() => {
        if (!pending) close();
      }}
    >
      <DialogTitle>خصوصیات دسته {category.name}</DialogTitle>
      <DialogContent>
        <Alert severity="info">
          انتساب فقط برای همین دسته است؛ خصوصیت والد خودکار اضافه نمی‌شود.{" "}
          {used &&
            "این دسته محصول دارد؛ تغییر مجموعه محورهای تنوع تا تعیین سیاست تبدیل غیرفعال است."}
        </Alert>
        {error && <Alert severity="error">{error}</Alert>}
        <div className="binding-list">
          {(state.attributes ?? []).map((a, index) => {
            const binding = bindings.find((b) => b.attribute_id === a.id);
            const originalBinding = original.bindings?.find(
              (b) => b.attribute_id === a.id,
            );
            const valueUsed = (state.products ?? []).some(
              (p) =>
                p.category_id === category.id &&
                p.attributes?.some((v) => v.attribute_id === a.id),
            );
            return (
              <div
                className="binding-row"
                key={a.id}
                data-testid={`binding-${a.id}`}
              >
                <label>
                  <Checkbox
                    checked={!!binding}
                    disabled={
                      !editable ||
                      pending ||
                      (!!binding &&
                        (valueUsed || (used && !!originalBinding?.axis)))
                    }
                    inputProps={{ "aria-label": `انتساب ${a.name}` }}
                    onChange={(_, checked) => {
                      setBindings(
                        checked
                          ? [
                              ...bindings,
                              {
                                attribute_id: a.id,
                                axis: false,
                                card: false,
                                filter: false,
                                position: index,
                              },
                            ]
                          : bindings.filter((b) => b.attribute_id !== a.id),
                      );
                      setKey(crypto.randomUUID());
                    }}
                  />
                  {a.name}
                  <small dir="ltr">{a.code}</small>
                </label>
                {(["axis", "card", "filter"] as const).map((flag) => (
                  <label key={flag}>
                    <Checkbox
                      checked={binding?.[flag] ?? false}
                      disabled={
                        !editable ||
                        pending ||
                        !binding ||
                        (flag === "axis" && (used || a.kind !== "enum"))
                      }
                      inputProps={{
                        "aria-label": `${{ axis: "محور", card: "کارت", filter: "فیلتر" }[flag]} ${a.name}`,
                      }}
                      onChange={(_, checked) => {
                        setBindings(
                          bindings.map((b) =>
                            b.attribute_id === a.id
                              ? { ...b, [flag]: checked }
                              : b,
                          ),
                        );
                        setKey(crypto.randomUUID());
                      }}
                    />
                    {
                      {
                        axis: "محور تنوع",
                        card: "نمایش کارت",
                        filter: "فیلتر",
                      }[flag]
                    }
                  </label>
                ))}
                <TextField
                  size="small"
                  label={`ترتیب ${a.name}`}
                  type="number"
                  value={binding?.position ?? ""}
                  disabled={!editable || pending || !binding}
                  inputProps={{ min: 0, max: 10000, step: 1 }}
                  onChange={(e) => {
                    setBindings(
                      bindings.map((b) =>
                        b.attribute_id === a.id
                          ? { ...b, position: Number(e.target.value) }
                          : b,
                      ),
                    );
                    setKey(crypto.randomUUID());
                  }}
                />
              </div>
            );
          })}
        </div>
        {!state.attributes?.length && (
          <p>ابتدا یک خصوصیت در کتابخانه تعریف کنید.</p>
        )}
      </DialogContent>
      <DialogActions>
        <Button disabled={pending} onClick={close}>
          انصراف
        </Button>
        <Button
          variant="contained"
          disabled={!editable || pending}
          onClick={async () => {
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
              setError(message(e));
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

function SalesWindowSettings() {
  const { state, execute, can } = useWorkspace();
  const [days, setDays] = useState(
    String(state.display_settings.sales_window_days || 30),
  );
  const [pending, setPending] = useState(false),
    [error, setError] = useState(""),
    [notice, setNotice] = useState("");
  const [key, setKey] = useState(crypto.randomUUID());
  const allowed = can("settings.sales-window");
  return (
    <section className="panel attribute-settings">
      <h2>مرتب‌سازی پرفروش‌ترین</h2>
      <p className="muted">
        تعداد خالص فروش در بازه لغزان؛ پیش‌فرض ۳۰ روز. زمان پرداخت قطعی مبناست و
        لغو/استرداد قلم از فروش کم می‌شود.
      </p>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          if (pending || !allowed) return;
          setPending(true);
          setError("");
          setNotice("");
          try {
            await execute({
              action: "settings.sales-window",
              target: "display",
              amount: Number(days),
              version: state.display_settings.version,
              idempotency_key: key,
            });
            setNotice("بازه فروش ذخیره شد.");
            setKey(crypto.randomUUID());
          } catch (e) {
            setError(message(e));
          } finally {
            setPending(false);
          }
        }}
      >
        <TextField
          label="بازه پرفروش‌ترین (روز)"
          type="number"
          size="small"
          value={days}
          onChange={(e) => {
            setDays(e.target.value);
            setKey(crypto.randomUUID());
          }}
          disabled={!allowed || pending}
          inputProps={{ min: 1, max: 3650, step: 1 }}
          required
        />
        <Button
          type="submit"
          variant="contained"
          disabled={!allowed || pending}
        >
          ثبت بازه فروش
        </Button>
      </form>
      <p>
        بازه ثبت‌شده:{" "}
        <b data-testid="saved-sales-window">
          {state.display_settings.sales_window_days || 30}
        </b>{" "}
        روز؛ تازگی آمار حداکثر ۶۰ ثانیه.
      </p>
      {error && <Alert severity="error">{error}</Alert>}
      {notice && <Alert severity="success">{notice}</Alert>}
    </section>
  );
}
