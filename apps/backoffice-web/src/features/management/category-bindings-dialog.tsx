import { useState, type DragEvent } from "react";
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
import { PaginationControls, usePagedRows } from "../../shared/pagination";
import type { AttributeBinding, Category } from "../../shared/types";
import { AxisConversionDialog } from "./axis-conversion-dialog";

type Role = "axis" | "card" | "filter";
const roles: { id: Role; title: string; note: string }[] = [
  {
    id: "axis",
    title: "محور تنوع",
    note: "انتخاب مشتری و ترکیب‌های قابل خرید",
  },
  { id: "card", title: "نمایش کارت", note: "ویژگی‌های روی کارت محصول" },
  { id: "filter", title: "فیلتر PLP", note: "فیلترهای فهرست محصولات" },
];
const order = (b: AttributeBinding, role: Role) =>
  b[`${role}_position`] ?? b.position;
const sameAxes = (a: AttributeBinding[], b: AttributeBinding[]) =>
  JSON.stringify(
    a
      .filter((x) => x.axis)
      .map((x) => x.attribute_id)
      .sort(),
  ) ===
  JSON.stringify(
    b
      .filter((x) => x.axis)
      .map((x) => x.attribute_id)
      .sort(),
  );

export function CategoryBindingsDialog({
  category,
  close,
}: {
  category: Category;
  close: () => void;
}) {
  const { state, execute, can } = useWorkspace();
  const [original, setOriginal] = useState(category);
  const [bindings, setBindings] = useState(category.bindings ?? []);
  const [search, setSearch] = useState("");
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  const [conversion, setConversion] = useState(false);
  const [key, setKey] = useState(crypto.randomUUID());
  const [dragged, setDragged] = useState<{ id: string; source?: Role }>();
  const [activeZone, setActiveZone] = useState<Role>();
  const editable =
    can("category.bindings.set", "management") && !original.archived;
  const attributes = state.attributes ?? [];
  const latest = state.categories?.find((c) => c.id === original.id);
  const stale = !!latest && latest.version !== original.version;
  const products = (state.products ?? []).filter(
    (p) => p.category_id === original.id,
  );
  const available = usePagedRows(
    attributes.filter((a) =>
      (a.name + " " + a.code)
        .toLocaleLowerCase()
        .includes(search.trim().toLocaleLowerCase()),
    ),
  );
  const members = (role: Role) =>
    bindings
      .filter((b) => b[role])
      .sort(
        (a, b) =>
          order(a, role) - order(b, role) ||
          a.attribute_id.localeCompare(b.attribute_id),
      );
  function change(next: AttributeBinding[]) {
    setBindings(next);
    setKey(crypto.randomUUID());
    setError("");
  }
  function insert(id: string, role: Role, before?: string) {
    if (!editable || pending) return;
    const attribute = attributes.find((a) => a.id === id);
    if (!attribute) return;
    if (role === "axis" && attribute.kind !== "enum") {
      setError("محور تنوع باید از نوع انتخابی باشد.");
      return;
    }
    if (
      !bindings.some((b) => b.attribute_id === id) &&
      bindings.length >= 100
    ) {
      setError("حداکثر ۱۰۰ خصوصیت به هر دسته متصل می‌شود.");
      return;
    }
    const ids = members(role)
      .map((b) => b.attribute_id)
      .filter((value) => value !== id);
    const at = before ? ids.indexOf(before) : -1;
    ids.splice(at < 0 ? ids.length : at, 0, id);
    const next = bindings.some((b) => b.attribute_id === id)
      ? bindings
      : [
          ...bindings,
          {
            attribute_id: id,
            axis: false,
            card: false,
            filter: false,
            position: bindings.length,
          },
        ];
    change(
      next.map((b) =>
        ids.includes(b.attribute_id)
          ? {
              ...b,
              [role]: true,
              [`${role}_position`]: ids.indexOf(b.attribute_id),
            }
          : b,
      ),
    );
  }
  function drop(event: DragEvent, role: Role, before?: string) {
    event.preventDefault();
    event.stopPropagation();
    const id =
      event.dataTransfer.getData("application/x-cherriva-attribute") ||
      dragged?.id;
    if (id && before !== id) insert(id, role, before);
    setDragged(undefined);
    setActiveZone(undefined);
  }
  function move(id: string, role: Role, direction: number) {
    const ids = members(role).map((b) => b.attribute_id),
      index = ids.indexOf(id),
      target = index + direction;
    if (target < 0 || target >= ids.length) return;
    [ids[index], ids[target]] = [ids[target], ids[index]];
    change(
      bindings.map((b) =>
        ids.includes(b.attribute_id)
          ? { ...b, [`${role}_position`]: ids.indexOf(b.attribute_id) }
          : b,
      ),
    );
  }
  const drag = (event: DragEvent, id: string, source?: Role) => {
    event.dataTransfer.setData("application/x-cherriva-attribute", id);
    event.dataTransfer.effectAllowed = "copyMove";
    setDragged({ id, source });
  };
  if (conversion)
    return (
      <AxisConversionDialog
        category={original}
        proposedBindings={bindings}
        close={() => setConversion(false)}
        completed={close}
      />
    );
  return (
    <Dialog open fullWidth maxWidth="xl" onClose={() => !pending && close()}>
      <DialogTitle>خصوصیات دسته · {original.name}</DialogTitle>
      <DialogContent className="category-zone-editor">
        <p className="muted">
          خصوصیت را از کتابخانه به هر بخش بکشید؛ یک خصوصیت می‌تواند در چند بخش
          باشد. داخل هر بخش، ترتیب مستقل است. دکمه‌های افزودن و بالا/پایین برای
          کار با صفحه‌کلید هم در دسترس‌اند.
        </p>
        {!editable && <Alert severity="info">این دسته فقط خواندنی است.</Alert>}
        {error && <Alert severity="error">{error}</Alert>}
        {stale && (
          <Alert
            severity="warning"
            action={
              <Button
                disabled={pending}
                onClick={() => {
                  setOriginal(latest!);
                  change(latest!.bindings ?? []);
                }}
              >
                دریافت نسخهٔ تازه
              </Button>
            }
          >
            دسته تغییر کرده است؛ نسخهٔ تازه را دریافت کنید. پیش‌نویس فعلی جایگزین
            می‌شود.
          </Alert>
        )}
        <div className="attribute-zones">
          <section
            className="attribute-zone attribute-library-zone"
            aria-label="کتابخانه خصوصیات"
          >
            <h3>۰ · کتابخانه خصوصیات</h3>
            <TextField
              label="جست‌وجوی خصوصیت"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
            <div className="attribute-zone-items">
              {available.rows.map((a) => (
                <article
                  key={a.id}
                  className="attribute-zone-item"
                  draggable={editable && !pending}
                  onDragStart={(e) => drag(e, a.id)}
                  onDragEnd={() => setDragged(undefined)}
                  data-attribute-id={a.id}
                >
                  <b>{a.name}</b>
                  <small>
                    {a.kind === "enum"
                      ? "انتخابی"
                      : a.kind === "number"
                        ? "عدد"
                        : "متن"}
                    {a.unit ? " · " + a.unit : ""}
                  </small>
                  <div className="attribute-zone-add-actions">
                    {roles.map((r) => (
                      <Button
                        key={r.id}
                        size="small"
                        disabled={
                          !editable ||
                          pending ||
                          (r.id === "axis" && a.kind !== "enum") ||
                          bindings.some(
                            (b) => b.attribute_id === a.id && b[r.id],
                          )
                        }
                        aria-label={`افزودن ${a.name} به ${r.title}`}
                        onClick={() => insert(a.id, r.id)}
                      >
                        + {r.title}
                      </Button>
                    ))}
                  </div>
                </article>
              ))}
              {!available.controls.total && <p>خصوصیتی پیدا نشد.</p>}
            </div>
            <PaginationControls {...available.controls} />
          </section>
          {roles.map((role, index) => (
            <section
              key={role.id}
              aria-label={role.title}
              data-zone={role.id}
              className={
                "attribute-zone" +
                (activeZone === role.id ? " drop-active" : "")
              }
              onDragOver={(e) => {
                if (editable && !pending) {
                  e.preventDefault();
                  setActiveZone(role.id);
                }
              }}
              onDrop={(e) => drop(e, role.id)}
            >
              <h3>
                {(index + 1).toLocaleString("fa-IR")} · {role.title}{" "}
                <small>
                  ({members(role.id).length.toLocaleString("fa-IR")})
                </small>
              </h3>
              <p className="muted">{role.note}</p>
              <div className="attribute-zone-items">
                {members(role.id).map((binding, position, rows) => {
                  const attribute = attributes.find(
                    (a) => a.id === binding.attribute_id,
                  );
                  return (
                    <article
                      key={binding.attribute_id}
                      data-attribute-id={binding.attribute_id}
                      className="attribute-zone-item"
                      draggable={editable && !pending}
                      onDragStart={(e) =>
                        drag(e, binding.attribute_id, role.id)
                      }
                      onDragEnd={() => {
                        setDragged(undefined);
                        setActiveZone(undefined);
                      }}
                      onDragOver={(e) =>
                        editable && !pending && e.preventDefault()
                      }
                      onDrop={(e) => drop(e, role.id, binding.attribute_id)}
                    >
                      <div className="attribute-zone-item-heading">
                        <span aria-hidden="true">⠿</span>
                        <b>{attribute?.name ?? "خصوصیت حذف‌شده"}</b>
                        <small>{(position + 1).toLocaleString("fa-IR")}</small>
                      </div>
                      <div className="attribute-zone-item-actions">
                        <Button
                          size="small"
                          disabled={!editable || pending || position === 0}
                          aria-label={`بالا بردن ${attribute?.name} در ${role.title}`}
                          onClick={() =>
                            move(binding.attribute_id, role.id, -1)
                          }
                        >
                          ↑
                        </Button>
                        <Button
                          size="small"
                          disabled={
                            !editable || pending || position === rows.length - 1
                          }
                          aria-label={`پایین بردن ${attribute?.name} در ${role.title}`}
                          onClick={() => move(binding.attribute_id, role.id, 1)}
                        >
                          ↓
                        </Button>
                        <Button
                          size="small"
                          color="error"
                          disabled={!editable || pending}
                          aria-label={`حذف ${attribute?.name} از ${role.title}`}
                          onClick={() =>
                            change(
                              bindings.map((b) =>
                                b.attribute_id === binding.attribute_id
                                  ? { ...b, [role.id]: false }
                                  : b,
                              ),
                            )
                          }
                        >
                          حذف
                        </Button>
                      </div>
                    </article>
                  );
                })}
                {!members(role.id).length && (
                  <div className="attribute-zone-empty">
                    خصوصیت را اینجا رها کنید
                  </div>
                )}
              </div>
            </section>
          ))}
        </div>
      </DialogContent>
      <DialogActions>
        <Button disabled={pending} onClick={close}>
          انصراف
        </Button>
        <Button
          variant="contained"
          disabled={!editable || pending || stale}
          onClick={async () => {
            if (pending) return;
            if (
              !sameAxes(original.bindings ?? [], bindings) &&
              products.length > 0
            ) {
              if (!can("category.axes.convert")) {
                setError(
                  "برای تغییر محور تنوعِ دارای محصول، مجوز تبدیل محور لازم است.",
                );
                return;
              }
              setConversion(true);
              return;
            }
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
            } catch (cause) {
              setError(
                cause instanceof Error ? cause.message : "ذخیره انجام نشد.",
              );
            } finally {
              setPending(false);
            }
          }}
        >
          {pending ? "در حال ذخیره…" : "ذخیرهٔ خصوصیات"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
