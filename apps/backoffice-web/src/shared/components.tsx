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
import { labels } from "./format";
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
  return (
    <div className="table-scroll">
      <table>
        <thead>
          <tr>
            {columns.map((c) => (
              <th key={c.title}>{c.title}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.id}>
              {columns.map((c) => (
                <td key={c.title}>{c.render(row)}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      {rows.length === 0 && <div className="empty">{empty}</div>}
    </div>
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
}: {
  title: string;
  fields: Field[];
  open: boolean;
  onClose: () => void;
  onSubmit: (values: Record<string, string>) => Promise<void>;
}) {
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
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
            await onSubmit(values);
            onClose();
          } catch (error) {
            setError(error instanceof Error ? error.message : "خطا در ثبت");
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
          {error && <Alert severity="error">{error}</Alert>}
        </DialogContent>
        <DialogActions>
          <Button onClick={onClose} disabled={busy}>
            انصراف
          </Button>
          <Button variant="contained" type="submit" disabled={busy}>
            {busy ? "در حال ثبت…" : "ثبت و تأیید"}
          </Button>
        </DialogActions>
      </Box>
    </Dialog>
  );
}
