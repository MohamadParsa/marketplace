import { useEffect, useRef, useState } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
} from "@mui/material";
import { CategoryOutlined, LabelOutlined } from "@mui/icons-material";
import { api } from "../../shared/api";
import { useWorkspace } from "../../shared/workspace";

export type IconTarget = {
  kind: "category" | "brand";
  id: string;
  name: string;
  icon_id?: string;
  version: number;
};

export function CatalogIcon({
  kind,
  name,
  iconId,
}: {
  kind: IconTarget["kind"];
  name: string;
  iconId?: string;
}) {
  const [failed, setFailed] = useState(false);
  useEffect(() => setFailed(false), [iconId]);
  const noun = kind === "category" ? "دسته" : "برند";
  if (iconId && !failed)
    return (
      <img
        className="catalog-icon"
        src={api.iconUrl(iconId)}
        alt={`آیکن ${noun} ${name}`}
        onError={() => setFailed(true)}
      />
    );
  return (
    <span
      className="catalog-icon catalog-icon-default"
      role="img"
      aria-label={`آیکن پیش‌فرض ${noun} ${name}`}
    >
      {kind === "category" ? <CategoryOutlined /> : <LabelOutlined />}
    </span>
  );
}

export function CatalogIconDialog({
  target,
  close,
}: {
  target: IconTarget;
  close: () => void;
}) {
  const { state, can, uploadIcon, execute } = useWorkspace();
  const [selected, setSelected] = useState<File>();
  const [preview, setPreview] = useState("");
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  const [selectionPending, setSelectionPending] = useState(false);
  const attempt = useRef(0);
  const uploadKey = useRef(crypto.randomUUID());
  const deleteKey = useRef(crypto.randomUUID());
  const policy = state.icon_policy;
  const editable = can(`${target.kind}.icon.set`, "management") && !!policy;
  const removable = can(`${target.kind}.icon.remove`, "management") && !!policy;
  useEffect(() => {
    if (!selected) {
      setPreview("");
      return;
    }
    const url = URL.createObjectURL(selected);
    setPreview(url);
    return () => URL.revokeObjectURL(url);
  }, [selected]);
  useEffect(
    () => () => {
      attempt.current++;
    },
    [],
  );
  async function choose(file?: File) {
    const current = ++attempt.current;
    setSelectionPending(false);
    if (!file || !policy) return;
    setError("");
    if (
      !policy.content_types.includes(file.type) ||
      file.size > policy.max_bytes
    ) {
      setError(
        "فایل جدید پذیرفته نشد؛ تصویر PNG، JPEG یا WebP در محدوده حجم مجاز انتخاب کنید. تصویر قبلی حفظ شده است.",
      );
      return;
    }
    setSelectionPending(true);
    const url = URL.createObjectURL(file);
    try {
      const image = new Image();
      image.src = url;
      await image.decode();
      if (
        image.naturalWidth > policy.max_dimension ||
        image.naturalHeight > policy.max_dimension
      )
        throw new Error("dimensions");
      if (current !== attempt.current) return;
      setSelected(file);
      uploadKey.current = crypto.randomUUID();
    } catch {
      if (current === attempt.current)
        setError(
          "تصویر قابل خواندن نیست یا ابعادش از حد مجاز بیشتر است؛ تصویر قبلی حفظ شده است.",
        );
    } finally {
      URL.revokeObjectURL(url);
      if (current === attempt.current) setSelectionPending(false);
    }
  }
  async function save(remove: boolean) {
    if (
      (remove ? !removable : !editable) ||
      pending ||
      selectionPending ||
      (!remove && !selected)
    )
      return;
    setPending(true);
    setError("");
    try {
      if (remove)
        await execute({
          action: `${target.kind}.icon.remove`,
          target: target.id,
          version: target.version,
          idempotency_key: deleteKey.current,
        });
      else
        await uploadIcon({
          kind: target.kind,
          target: target.id,
          version: target.version,
          file: selected!,
          key: uploadKey.current,
        });
      close();
    } catch (e) {
      setError(
        e instanceof Error
          ? e.message
          : "ثبت آیکن انجام نشد؛ دوباره تلاش کنید.",
      );
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
        آیکن {target.kind === "category" ? "دسته" : "برند"} {target.name}
      </DialogTitle>
      <DialogContent>
        {target.kind === "brand" &&
          !(state.brands ?? []).find((b) => b.id === target.id)?.active && (
            <Alert severity="info">
              تصویر برند غیرفعال در سایت نمایش داده نمی‌شود؛ تصویر ذخیره‌شده
              محفوظ است و پس از فعال‌سازی نمایش داده می‌شود.
            </Alert>
          )}
        <div className="icon-previews">
          <div>
            <small>تصویر ثبت‌شده</small>
            <CatalogIcon
              kind={target.kind}
              name={target.name}
              iconId={target.icon_id}
            />
          </div>
          {preview && (
            <div>
              <small>پیش‌نمایش انتخاب جدید</small>
              <img
                className="icon-selection-preview"
                src={preview}
                alt={`پیش‌نمایش آیکن ${target.name}`}
              />
            </div>
          )}
        </div>
        {error && (
          <Alert severity="error" role="alert">
            {error}
          </Alert>
        )}
        {!editable && <Alert severity="info">مجوز تغییر آیکن را ندارید.</Alert>}
        <p className="muted">
          {policy
            ? `PNG، JPEG، WebP · حداکثر ${Math.round(policy.max_bytes / 1024)} کیلوبایت · هر ضلع حداکثر ${policy.max_dimension} پیکسل`
            : "سیاست رسانه دریافت نشده است."}
        </p>
        <Button
          component="label"
          variant="outlined"
          disabled={!editable || pending || selectionPending}
        >
          انتخاب تصویر
          <input
            hidden
            type="file"
            accept={policy?.content_types.join(",")}
            aria-label="انتخاب فایل آیکن"
            onChange={(e) => {
              void choose(e.target.files?.[0]);
              e.target.value = "";
            }}
          />
        </Button>
        {selectionPending && <p role="status">در حال بررسی تصویر…</p>}
      </DialogContent>
      <DialogActions className="icon-dialog-actions">
        <Button onClick={close} disabled={pending}>
          انصراف
        </Button>
        <Button
          color="warning"
          disabled={
            !removable || !target.icon_id || pending || selectionPending
          }
          onClick={() => void save(true)}
        >
          بازگشت به پیش‌فرض
        </Button>
        <Button
          variant="contained"
          disabled={!editable || !selected || pending || selectionPending}
          onClick={() => void save(false)}
        >
          {pending ? "در حال ثبت…" : "ثبت آیکن"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
