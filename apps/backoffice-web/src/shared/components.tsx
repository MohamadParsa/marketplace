import { useState, type ReactNode } from "react";
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
  title: string;
  render: (row: T) => ReactNode;
}
export function DataTable<T extends { id: string }>({
  rows,
  columns,
  empty = "موردی برای نمایش وجود ندارد.",
}: {
  rows: T[];
  columns: Column<T>[];
  empty?: string;
}) {
  const paged = usePagedRows(rows);
  return (
    <>
      <div className="table-scroll" tabIndex={0} aria-label="جدول اطلاعات">
        <table className="data-table">
          <colgroup>
            {columns.map((c, i) => (
              <col
                key={c.title}
                style={{
                  width:
                    c.title === "عملیات"
                      ? "290px"
                      : i === 0
                        ? "32%"
                        : undefined,
                }}
              />
            ))}
          </colgroup>
          <thead>
            <tr>
              {columns.map((c) => (
                <th
                  key={c.title}
                  className={
                    c.title === "عملیات" ? "table-actions-cell" : undefined
                  }
                >
                  {c.title}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {paged.rows.map((row) => (
              <tr key={row.id}>
                {columns.map((c) => (
                  <td
                    key={c.title}
                    className={
                      c.title === "عملیات" ? "table-actions-cell" : undefined
                    }
                  >
                    {c.title === "عملیات" ? (
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
        {rows.length === 0 && <div className="empty">{empty}</div>}
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
