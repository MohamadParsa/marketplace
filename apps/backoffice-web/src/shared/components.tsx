import { useWorkspace } from "./workspace";
import { xlsxBlob } from "./xlsx";
import {
  useRef,
  useEffect,
  useDeferredValue,
  useMemo,
  isValidElement,
  useState,
  type ReactNode,
} from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  MenuItem,
  TextField,
} from "@mui/material";
import { PaginationControls, usePagedRows } from "./pagination";
import { labels } from "./format";
import { ApiError } from "./api";
export function Status({ value }: { value: string }) {
  const color = ["success", "published", "delivered", "confirmed"].includes(
    value,
  )
    ? "success"
    : ["failure", "rejected", "on-hold"].includes(value)
      ? "error"
      : "default";
  return (
    <Chip
      size="small"
      color={color}
      variant="outlined"
      label={labels[value] ?? value}
    />
  );
}
export interface Column<T> {
  exportValue?: (row: T) => string | number;
  exportable?: boolean;
  title: string;
  render: (row: T) => ReactNode;
}
function cellText(node: ReactNode): string {
  if (node == null || typeof node === "boolean") return "";
  if (typeof node === "string" || typeof node === "number") return String(node);
  if (Array.isArray(node)) return node.map(cellText).join(" ");
  if (
    isValidElement<{ children?: ReactNode; value?: string; label?: ReactNode }>(
      node,
    )
  ) {
    if (node.type === Status && node.props.value)
      return labels[node.props.value] ?? node.props.value;
    return cellText(node.props.children ?? node.props.label);
  }
  return "";
}
const actionColumn = (title: string) =>
  ["عملیات", "جزئیات", "رسیدگی", "تلاش مجدد"].includes(title);
const forbiddenExport =
  /(عملیات|رمز|گذرواژه|کد.?ملی|بانک|شماره.?کارت|شبا|password|national.?id|iban)/i;
export function DataTable<T extends { id: string }>({
  rows,
  columns,
  empty = "موردی برای نمایش وجود ندارد.",
}: {
  rows: T[];
  columns: Column<T>[];
  empty?: string;
}) {
  const { authToken, surface } = useWorkspace();
  const exportAbort = useRef<AbortController>();
  useEffect(() => () => exportAbort.current?.abort(), []);
  const [search, setSearch] = useState(""),
    [filters, setFilters] = useState<Record<string, string>>({}),
    [showFilters, setShowFilters] = useState(false),
    [sort, setSort] = useState<{ column: number; desc: boolean }>(),
    [selected, setSelected] = useState<string[]>([]),
    [exporting, setExporting] = useState(false),
    [error, setError] = useState("");
  const deferred = useDeferredValue(search);
  const cells = useMemo(
    () =>
      rows.map((row) => ({
        row,
        text: columns.map((c) => cellText(c.render(row))),
      })),
    [rows, columns],
  );
  const norm = (value: string) =>
    value.trim().replace(/ي/g, "ی").replace(/ك/g, "ک").toLocaleLowerCase();
  const filtered = useMemo(() => {
    const out = cells.filter(
      (x) =>
        (!deferred || norm(x.text.join(" ")).includes(norm(deferred))) &&
        Object.entries(filters).every(
          ([i, v]) => !v || norm(x.text[Number(i)]).includes(norm(v)),
        ),
    );
    if (sort)
      out.sort((a, b) => {
        const n = a.text[sort.column].localeCompare(b.text[sort.column], "fa", {
          numeric: true,
        });
        return (sort.desc ? -n : n) || a.row.id.localeCompare(b.row.id);
      });
    return out;
  }, [cells, deferred, filters, sort]);
  const paged = usePagedRows(filtered.map((x) => x.row));
  const selectedVisible = filtered.filter((x) => selected.includes(x.row.id));
  const exportRows = async (onlySelected: boolean) => {
    if (exporting) return;
    const controller = new AbortController();
    exportAbort.current = controller;
    setExporting(true);
    setError("");
    try {
      const exportColumns = columns
        .map((c, i) => ({ c, i }))
        .filter(
          ({ c }) =>
            c.exportable !== false &&
            !actionColumn(c.title) &&
            !forbiddenExport.test(c.title),
        );
      const target = onlySelected ? selectedVisible : filtered;
      if (!target.length) throw Error("ردیفی برای خروجی انتخاب نشده است.");
      if (target.length > 10000)
        throw Error(
          "حداکثر ۱۰هزار ردیف در هر خروجی مجاز است؛ فیلتر را محدودتر کنید.",
        );
      const auth = await fetch(
        (import.meta.env.VITE_API_BASE_URL ?? "") +
          `/${surface}/grids/export-authorize`,
        {
          method: "POST",
          signal: controller.signal,
          cache: "no-store",
          headers: { Authorization: `Bearer ${authToken}` },
        },
      );
      if (!auth.ok)
        throw Error(
          auth.status === 403
            ? "مجوز خروجی گرید ندارید؛ از ادمین درخواست دسترسی کنید."
            : "نشست یا ارتباط معتبر نیست؛ دوباره تلاش کنید.",
        );
      const blob = await xlsxBlob(
        exportColumns.map(({ c }) => c.title),
        target.map((x) =>
          exportColumns.map(({ c, i }) =>
            String(c.exportValue ? c.exportValue(x.row) : x.text[i]),
          ),
        ),
        controller.signal,
      );
      const url = URL.createObjectURL(blob),
        link = document.createElement("a");
      link.href = url;
      link.download = "cherriva-grid.xlsx";
      link.click();
      setTimeout(() => URL.revokeObjectURL(url), 1000);
    } catch (e) {
      setError(
        controller.signal.aborted ? "خروجی لغو شد." : (e as Error).message,
      );
    } finally {
      setExporting(false);
    }
  };
  return (
    <>
      <div className="grid-tools">
        <TextField
          size="small"
          label="جست‌وجو در جدول"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <Button
          aria-expanded={showFilters}
          onClick={() => setShowFilters((v) => !v)}
        >
          فیلتر ستون‌ها
        </Button>
        <Button
          onClick={() => {
            setSearch("");
            setFilters({});
            setSort(undefined);
            setSelected([]);
          }}
        >
          پاک‌کردن فیلترها
        </Button>
        <span aria-live="polite">
          {filtered.length.toLocaleString("fa-IR")} نتیجه ·{" "}
          {selectedVisible.length.toLocaleString("fa-IR")} انتخاب
        </span>
        <Button
          disabled={exporting || !filtered.length}
          onClick={() => void exportRows(false)}
        >
          Excel نتایج
        </Button>
        <Button
          disabled={exporting || !selectedVisible.length}
          onClick={() => void exportRows(true)}
        >
          Excel انتخاب‌شده‌ها
        </Button>
        {exporting && (
          <Button onClick={() => exportAbort.current?.abort()}>
            لغو خروجی
          </Button>
        )}
      </div>
      {showFilters && (
        <div className="grid-column-filters">
          {columns.map((c, i) =>
            actionColumn(c.title) ? null : (
              <TextField
                key={c.title}
                size="small"
                label={`فیلتر ${c.title}`}
                value={filters[i] ?? ""}
                onChange={(e) =>
                  setFilters({ ...filters, [i]: e.target.value })
                }
              />
            ),
          )}
        </div>
      )}
      {error && (
        <Alert severity="error" role="alert">
          {error}
        </Alert>
      )}
      <div className="table-scroll" tabIndex={0} aria-label="جدول اطلاعات">
        <table className="data-table">
          <thead>
            <tr>
              <th className="grid-selection">
                <input
                  type="checkbox"
                  aria-label="انتخاب ردیف‌های این صفحه"
                  checked={
                    paged.rows.length > 0 &&
                    paged.rows.every((r) => selected.includes(r.id))
                  }
                  onChange={(e) =>
                    setSelected(
                      e.target.checked
                        ? [
                            ...new Set([
                              ...selected,
                              ...paged.rows.map((r) => r.id),
                            ]),
                          ]
                        : selected.filter(
                            (id) => !paged.rows.some((r) => r.id === id),
                          ),
                    )
                  }
                />
              </th>
              {columns.map((c, i) => (
                <th
                  key={c.title}
                  className={
                    actionColumn(c.title) ? "table-actions-cell" : undefined
                  }
                  aria-sort={
                    sort?.column === i
                      ? sort.desc
                        ? "descending"
                        : "ascending"
                      : undefined
                  }
                >
                  {actionColumn(c.title) ? (
                    c.title
                  ) : (
                    <button
                      className="grid-sort"
                      onClick={() =>
                        setSort({
                          column: i,
                          desc: sort?.column === i ? !sort.desc : false,
                        })
                      }
                    >
                      {c.title}
                      {sort?.column === i ? (sort.desc ? " ↓" : " ↑") : ""}
                    </button>
                  )}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {paged.rows.map((row) => (
              <tr key={row.id}>
                <td>
                  <input
                    type="checkbox"
                    aria-label={`انتخاب ردیف ${row.id}`}
                    checked={selected.includes(row.id)}
                    onChange={(e) =>
                      setSelected(
                        e.target.checked
                          ? [...selected, row.id]
                          : selected.filter((id) => id !== row.id),
                      )
                    }
                  />
                </td>
                {columns.map((c) => (
                  <td
                    key={c.title}
                    className={
                      actionColumn(c.title) ? "table-actions-cell" : undefined
                    }
                  >
                    {actionColumn(c.title) ? (
                      <div className="table-row-actions">{c.render(row)}</div>
                    ) : (
                      c.render(row)
                    )}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
        {!filtered.length && <div className="empty">{empty}</div>}
      </div>
      <PaginationControls {...paged.controls} />
    </>
  );
}
export function PageHeader({
  title,
  description,
  action,
}: {
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <div className="page-heading">
      <div>
        <span className="overline">مدیریت بازارگاه</span>
        <h1>{title}</h1>
        <p>{description}</p>
      </div>
      {action}
    </div>
  );
}
export interface Field {
  name: string;
  label: string;
  options?: { value: string; label: string }[];
  type?: string;
  default?: string;
}
export function ActionDialog({
  title,
  fields,
  open,
  onClose,
  onSubmit,
  requestAdminCode,
}: {
  title: string;
  fields: Field[];
  open: boolean;
  onClose: () => void;
  onSubmit: (
    values: Record<string, string>,
    reauthentication?: { request_id: string; code: string },
  ) => Promise<void>;
  requestAdminCode?: () => Promise<{ request_id: string; test_code?: string }>;
}) {
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [pendingValues, setPendingValues] = useState<Record<
    string,
    string
  > | null>(null);
  const [challenge, setChallenge] = useState<{
    request_id: string;
    test_code?: string;
  } | null>(null);
  const [code, setCode] = useState("");
  async function sendAdminCode() {
    if (!requestAdminCode) return;
    setBusy(true);
    setError("");
    try {
      setChallenge(await requestAdminCode());
      setCode("");
    } catch (error) {
      setError(error instanceof Error ? error.message : "خطا در ارسال کد");
    } finally {
      setBusy(false);
    }
  }
  return (
    <Dialog
      open={open}
      onClose={busy ? undefined : onClose}
      fullWidth
      maxWidth="sm"
    >
      <Box
        component="form"
        onSubmit={async (event) => {
          event.preventDefault();
          const values = Object.fromEntries(
            new FormData(event.currentTarget).entries(),
          ) as Record<string, string>;
          setBusy(true);
          setError("");
          try {
            await onSubmit(
              pendingValues ?? values,
              challenge
                ? { request_id: challenge.request_id, code }
                : undefined,
            );
            onClose();
          } catch (error) {
            if (
              requestAdminCode &&
              error instanceof ApiError &&
              error.code === "reauthentication_required"
            ) {
              setPendingValues(values);
              await sendAdminCode();
            } else {
              setError(error instanceof Error ? error.message : "خطا در ثبت");
            }
          } finally {
            setBusy(false);
          }
        }}
      >
        <DialogTitle>{title}</DialogTitle>
        <DialogContent>
          <Alert severity="info" sx={{ mb: 2 }}>
            این عملیات روی داده آزمون اجرا و در سابقه فعالیت ثبت می‌شود.
          </Alert>
          {fields.map((field) => (
            <TextField
              key={field.name}
              name={field.name}
              label={field.label}
              type={field.type ?? "text"}
              select={!!field.options}
              required
              disabled={!!pendingValues || busy}
              fullWidth
              margin="normal"
              defaultValue={field.default ?? field.options?.[0]?.value ?? ""}
              inputProps={field.type === "number" ? { min: 0 } : undefined}
            >
              {field.options?.map((option) => (
                <MenuItem key={option.value} value={option.value}>
                  {option.label}
                </MenuItem>
              ))}
            </TextField>
          ))}
          {pendingValues && (
            <>
              <Alert severity="info" sx={{ mt: 2 }}>
                برای اعطای اختیار ادمین، کد تأیید حساب خودتان را وارد کنید.
              </Alert>
              {challenge && (
                <TextField
                  label="کد تأیید ادمین"
                  value={code}
                  onChange={(event) => setCode(event.target.value)}
                  helperText={
                    challenge.test_code
                      ? `کد آزمون: ${challenge.test_code}`
                      : undefined
                  }
                  autoComplete="one-time-code"
                  inputProps={{ inputMode: "numeric", maxLength: 6 }}
                  required
                  fullWidth
                  margin="normal"
                />
              )}
              <Button type="button" onClick={sendAdminCode} disabled={busy}>
                {challenge ? "ارسال دوباره کد" : "ارسال کد تأیید"}
              </Button>
            </>
          )}
          {error && <Alert severity="error">{error}</Alert>}
        </DialogContent>
        <DialogActions>
          <Button onClick={onClose} disabled={busy}>
            انصراف
          </Button>
          <Button
            variant="contained"
            type="submit"
            disabled={busy || (!!pendingValues && !challenge)}
          >
            {busy ? "در حال ثبت…" : "ثبت و تأیید"}
          </Button>
        </DialogActions>
      </Box>
    </Dialog>
  );
}
