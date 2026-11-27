import { useRef, useState } from "react";
import DOMPurify from "dompurify";
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
import type { CatalogInput, ProductWorkspace } from "../../shared/types";
import {
  ProductContentFields,
  emptyProductContent,
} from "./product-content-fields";
const states: Record<string, string> = {
  pending: "در انتظار بررسی",
  rejected: "نیازمند اصلاح",
  approved: "تأیید شده",
  "applied-direct": "مستقیم اعمال شد",
};
export function ProductContentReview({
  workspace: w,
}: {
  workspace: ProductWorkspace;
}) {
  const { surface, execute, busy, state } = useWorkspace();
  const [edit, setEdit] = useState(false),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [reason, setReason] = useState("");
  const revision = w.product.content_revision;
  const owner = state.stores?.find((s) => s.id === w.product.owner_store_id);
  const run = async (action: string) => {
    setError("");
    try {
      await execute({
        action,
        target: w.product.id,
        version: w.product.version,
        reason,
        idempotency_key: crypto.randomUUID(),
      });
      setNotice("نتیجه بررسی ذخیره شد.");
    } catch (e) {
      setError((e as Error).message);
    }
  };
  if (!w.permissions.edit_content && !revision) return null;
  return (
    <section className="content-review-panel">
      <div className="product-workspace-toolbar">
        {w.permissions.edit_content && (
          <Button
            variant="outlined"
            disabled={busy}
            onClick={() => setEdit(true)}
          >
            ویرایش محتوای فروشگاه
          </Button>
        )}
        {owner && (
          <small>
            سیاست محتوای فروشگاه:{" "}
            {owner.content_policy === "direct"
              ? "انتشار مستقیم"
              : "نیازمند تأیید بک‌آفیس"}
          </small>
        )}
      </div>
      {notice && <Alert severity="success">{notice}</Alert>}
      {error && <Alert severity="error">{error}</Alert>}
      {revision && (
        <>
          <h3>ویرایش فروشنده: {states[revision.state] ?? revision.state}</h3>
          {revision.state === "pending" && (
            <Alert severity="info">
              نسخه تأییدشده قبلی در سایت نمایش داده می‌شود. این ویرایش قیمت،
              موجودی و رزرو عرضه‌ها را تغییر نمی‌دهد.
            </Alert>
          )}
          {revision.reason && <p>نتیجه بررسی: {revision.reason}</p>}
          <div className="content-review-compare">
            <article>
              <h4>نسخه جاری</h4>
              <b>{w.product.name}</b>
              <p>
                {w.product.content?.short_description || w.product.description}
              </p>
              <div
                dangerouslySetInnerHTML={{
                  __html: DOMPurify.sanitize(
                    w.product.content?.full_description ?? "",
                  ),
                }}
              />
            </article>
            <article>
              <h4>نسخه ارسالی فروشنده</h4>
              <b>{revision.name}</b>
              <p dir="auto">{revision.english_name}</p>
              <p>
                {revision.content?.short_description || revision.description}
              </p>
              <div
                dangerouslySetInnerHTML={{
                  __html: DOMPurify.sanitize(
                    revision.content?.full_description ?? "",
                  ),
                }}
              />
              {revision.content?.advantages?.length ? (
                <p>مزایا: {revision.content.advantages.join("، ")}</p>
              ) : null}
              {revision.content?.disadvantages?.length ? (
                <p>معایب: {revision.content.disadvantages.join("، ")}</p>
              ) : null}
              <details>
                <summary>اطلاعات تکمیلی ارسالی</summary>
                <p>
                  برند:{" "}
                  {state.brands?.find((b) => b.id === revision.brand_id)
                    ?.name || "بدون برند"}
                </p>
                {revision.attributes?.map((a) => (
                  <p key={a.attribute_id}>
                    {
                      state.attributes?.find((x) => x.id === a.attribute_id)
                        ?.name
                    }
                    : {a.value}
                  </p>
                ))}
                <p>عنوان متا: {revision.content?.seo.meta_title}</p>
                <p>کلمات متا: {revision.content?.seo.meta_keywords}</p>
                <p>شرح متا: {revision.content?.seo.meta_description}</p>
                <p>وزن حمل: {revision.content?.shipping.weight ?? "—"} گرم</p>
                <p>
                  ابعاد میلی‌متر: {revision.content?.shipping.length ?? "—"} ×{" "}
                  {revision.content?.shipping.width ?? "—"} ×{" "}
                  {revision.content?.shipping.height ?? "—"}
                </p>
              </details>
            </article>
          </div>
          {surface === "backoffice" &&
            w.permissions.review_content &&
            revision.state === "pending" && (
              <>
                <TextField
                  label="یادداشت بررسی / دلیل رد"
                  fullWidth
                  size="small"
                  multiline
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  inputProps={{ maxLength: 1000 }}
                />
                <div className="product-workspace-toolbar">
                  <Button
                    variant="contained"
                    disabled={busy}
                    onClick={() => void run("product.content.approve")}
                  >
                    تأیید و اعمال ویرایش فروشنده
                  </Button>
                  <Button
                    color="error"
                    disabled={busy || !reason.trim()}
                    onClick={() => void run("product.content.reject")}
                  >
                    رد ویرایش با دلیل
                  </Button>
                </div>
                <small>
                  اگر محتوای مرجع پس از ارسال تغییر کرده باشد، تأیید نسخه قدیمی
                  رد می‌شود و فروشنده باید دوباره بررسی و ارسال کند.
                </small>
              </>
            )}
        </>
      )}
      {edit && <ContentEditDialog workspace={w} close={() => setEdit(false)} />}
    </section>
  );
}
function ContentEditDialog({
  workspace: w,
  close,
}: {
  workspace: ProductWorkspace;
  close: () => void;
}) {
  const { execute, busy, state } = useWorkspace(),
    revision = w.product.content_revision;
  const useRevision =
    revision && (revision.state === "pending" || revision.state === "rejected");
  const source = useRevision ? revision : w.product;
  const [input, setInput] = useState<CatalogInput>({
    name: source.name,
    english_name: source.english_name ?? "",
    description: source.description ?? "",
    category_id: w.product.category_id,
    brand_id: source.brand_id ?? "",
    parent_id: "",
    published: false,
    content: source.content ?? emptyProductContent(),
    attributes: source.attributes ?? [],
  });
  const version = useRef(w.product.version),
    key = useRef(crypto.randomUUID());
  const [error, setError] = useState("");
  return (
    <Dialog
      open
      fullWidth
      maxWidth="lg"
      onClose={() => {
        if (!busy) close();
      }}
    >
      <DialogTitle>محتوای محصول مستقل فروشگاه</DialogTitle>
      <DialogContent>
        <Alert severity="info">
          فقط محتوای محصول مستقل خودتان قابل ارسال است. تغییر قیمت و موجودی از
          فرم عرضه انجام می‌شود. سیاست بررسی فروشگاه را بک‌آفیس تعیین می‌کند.
        </Alert>
        <form
          id="seller-content-edit"
          className="home-admin-form"
          onSubmit={(e) => {
            e.preventDefault();
            setError("");
            void execute({
              action: "product.content.submit",
              target: w.product.id,
              version: version.current,
              catalog: input,
              idempotency_key: key.current,
            })
              .then(close)
              .catch((e) => setError(e.message));
          }}
        >
          <fieldset disabled={busy}>
            <label>
              نام فارسی
              <input
                required
                maxLength={160}
                value={input.name}
                onChange={(e) => setInput({ ...input, name: e.target.value })}
              />
            </label>
            <label>
              نام انگلیسی
              <input
                maxLength={160}
                value={input.english_name}
                onChange={(e) =>
                  setInput({ ...input, english_name: e.target.value })
                }
              />
            </label>
            <label>
              برند
              <select
                value={input.brand_id}
                onChange={(e) =>
                  setInput({ ...input, brand_id: e.target.value })
                }
              >
                <option value="">بدون برند</option>
                {state.brands
                  ?.filter((b) => b.active || b.id === input.brand_id)
                  .map((b) => (
                    <option key={b.id} value={b.id}>
                      {b.name}
                    </option>
                  ))}
              </select>
            </label>
            <ProductContentFields
              value={input.content!}
              onChange={(content) => setInput({ ...input, content })}
            />
            {error && <Alert severity="error">{error}</Alert>}
          </fieldset>
        </form>
      </DialogContent>
      <DialogActions>
        <Button disabled={busy} onClick={close}>
          انصراف
        </Button>
        <Button
          type="submit"
          form="seller-content-edit"
          variant="contained"
          disabled={busy}
        >
          ذخیره / ارسال محتوا
        </Button>
      </DialogActions>
    </Dialog>
  );
}
