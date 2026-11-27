import { OptionSelect } from "../../shared/option-select";
import {
  Button,
  Checkbox,
  FormControlLabel,
  TextField,
  Alert,
} from "@mui/material";
import type { Business } from "../../../../../packages/commerce-client/types";
export const businessPages = [
  ["about", "about", "درباره چریوا"],
  ["contact", "contact", "تماس با ما"],
  ["shipping", "shipping_policy", "شرایط ارسال"],
  ["returns", "return_policy", "شرایط بازگشت"],
  ["privacy", "privacy", "حریم خصوصی"],
  ["terms", "terms", "قوانین خرید"],
] as const;
const kinds = [
  ["instagram", "اینستاگرام"],
  ["telegram", "تلگرام"],
  ["whatsapp", "واتساپ"],
  ["linkedin", "لینکدین"],
  ["x", "ایکس"],
  ["youtube", "یوتیوب"],
  ["aparat", "آپارات"],
  ["other", "سایر"],
] as const;
export function businessDraft(b: Business): Business {
  return {
    ...b,
    published_pages:
      b.published_pages ??
      Object.fromEntries(
        businessPages.map(([slug, key]) => [slug, !!b[key]?.trim()]),
      ),
    social_links: b.social_links ?? [],
  };
}
export function BusinessContentEditor({
  value: b,
  onChange,
}: {
  value: Business;
  onChange: (b: Business) => void;
}) {
  const links = b.social_links ?? [];
  const patch = (i: number, p: Partial<(typeof links)[number]>) =>
    onChange({
      ...b,
      social_links: links.map((l, k) => (k === i ? { ...l, ...p } : l)),
    });
  return (
    <div className="business-content-editor">
      <Alert severity="info">
        تنها صفحه دارای متن و تیک انتشار در سایت دیده می‌شود. متن پیشنهادی یا
        اطلاعات تأییدنشده را پیش‌نویس نگه دارید. نشان مجوز یا تعهد ساختگی نمایش
        داده نمی‌شود.
      </Alert>
      <div className="catalog-filter-text">
        {(
          [
            ["name", "نام کسب‌وکار"],
            ["support_phone", "شماره پشتیبانی"],
            ["support_email", "ایمیل پشتیبانی"],
          ] as const
        ).map(([key, label]) => (
          <TextField
            key={key}
            fullWidth
            size="small"
            label={label}
            value={b[key]}
            onChange={(e) => onChange({ ...b, [key]: e.target.value })}
          />
        ))}
      </div>
      {businessPages.map(([slug, key, label]) => (
        <details key={slug} className="business-page-editor">
          <summary>
            {label} — {b.published_pages?.[slug] ? "منتشرشده" : "پیش‌نویس"}
          </summary>
          <TextField
            label={`متن ${label}`}
            fullWidth
            multiline
            minRows={4}
            value={b[key]}
            inputProps={{ maxLength: 20000 }}
            onChange={(e) => onChange({ ...b, [key]: e.target.value })}
          />
          <FormControlLabel
            control={
              <Checkbox
                checked={!!b.published_pages?.[slug]}
                onChange={(e) =>
                  onChange({
                    ...b,
                    published_pages: {
                      ...b.published_pages,
                      [slug]: e.target.checked,
                    },
                  })
                }
              />
            }
            label={`انتشار صفحه ${label}`}
          />
          <div
            className="business-text-preview"
            aria-label={`پیش‌نمایش ${label}`}
          >
            {b[key] || "متن ثبت نشده است."}
          </div>
        </details>
      ))}
      <section aria-label="مدیریت شبکه‌های اجتماعی">
        <div className="toolbar">
          <h3>شبکه‌های اجتماعی</h3>
          <Button
            disabled={links.length >= 20}
            onClick={() =>
              onChange({
                ...b,
                social_links: [
                  ...links,
                  {
                    id: crypto.randomUUID(),
                    title: "",
                    kind: "other",
                    url: "",
                    position: (links.length + 1) * 10,
                    active: false,
                  },
                ],
              })
            }
          >
            افزودن شبکه اجتماعی
          </Button>
        </div>
        <p>
          نشانی کامل HTTPS؛ عدد کوچک‌تر مقدم است. شبکه غیرفعال یا بدون نشانی در
          فوتر نمایش داده نمی‌شود.
        </p>
        {!links.length && <p>شبکه‌ای تعریف نشده است.</p>}
        {links.map((l, i) => (
          <article
            key={l.id}
            className="business-social-editor"
            aria-label={`شبکه اجتماعی ${i + 1}`}
          >
            <div className="catalog-filter-text">
              <TextField
                size="small"
                label="عنوان شبکه"
                value={l.title}
                onChange={(e) => patch(i, { title: e.target.value })}
                inputProps={{ maxLength: 80 }}
              />
              <TextField
                size="small"
                label="نشانی شبکه HTTPS"
                value={l.url}
                onChange={(e) => patch(i, { url: e.target.value })}
                inputProps={{ maxLength: 2048, dir: "ltr" }}
              />
              <TextField
                size="small"
                type="number"
                label="ترتیب نمایش شبکه"
                value={l.position}
                onChange={(e) => patch(i, { position: Number(e.target.value) })}
                inputProps={{ min: 0, max: 1000000, step: 1 }}
              />
            </div>
            <div
              role="group"
              aria-label="نوع آیکن شبکه"
              className="catalog-choice-options"
            >
              <OptionSelect
                label={"نوع آیکن شبکه"}
                multiple={false}
                options={kinds.map(([key, label]) => ({
                  id: String(key),
                  name: String("" + label + ""),
                  selected: l.kind === key,
                  disabled: false,
                  onSelect: () => patch(i, { kind: key }),
                }))}
              />
            </div>
            <FormControlLabel
              control={
                <Checkbox
                  checked={l.active}
                  onChange={(e) => patch(i, { active: e.target.checked })}
                />
              }
              label="نمایش شبکه در فوتر"
            />
            <Button
              color="error"
              onClick={() =>
                onChange({
                  ...b,
                  social_links: links.filter((x) => x.id !== l.id),
                })
              }
            >
              حذف شبکه
            </Button>
          </article>
        ))}
      </section>
    </div>
  );
}
