import { useEffect, useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Alert,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  Switch,
  TextField,
} from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import { DataTable, PageHeader } from "../../shared/components";
import { api } from "../../shared/api";
import type { Banner, BannerList } from "../../shared/types";

const message = (e: unknown) =>
  e instanceof Error ? e.message : "عملیات انجام نشد؛ دوباره تلاش کنید.";
type Slot = "desktop" | "mobile";

export function BannersPage() {
  const { loadBanners, execute } = useWorkspace();
  const query = useQuery({ queryKey: ["banners"], queryFn: loadBanners });
  const [selected, setSelected] = useState<string>();
  const [create, setCreate] = useState(false),
    [name, setName] = useState("");
  const [key, setKey] = useState(crypto.randomUUID()),
    [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  if (query.isPending) return <p role="status">در حال دریافت بنرها…</p>;
  if (query.isError)
    return (
      <Alert
        severity="error"
        action={<Button onClick={() => void query.refetch()}>تلاش مجدد</Button>}
      >
        {message(query.error)}
      </Alert>
    );
  const list = query.data,
    banner = list.banners.find((b) => b.id === selected);
  return (
    <>
      <PageHeader
        title="بنرهای صفحه اصلی"
        description="تصویر ثابت یا عکس کالای موجودِ دسته؛ پیش‌نویس تا انتشار روی نسخه عمومی اثر ندارد."
        action={
          <Button
            variant="contained"
            disabled={!list.can_edit}
            onClick={() => {
              setCreate(true);
              setError("");
            }}
          >
            بنر جدید
          </Button>
        }
      />
      <section className="panel">
        <DataTable
          rows={list.banners}
          empty="هنوز بنری تعریف نشده است."
          columns={[
            {
              title: "نام بنر",
              render: (b) => (
                <>
                  <b>{b.name}</b>
                  <small dir="ltr">{b.id}</small>
                </>
              ),
            },
            {
              title: "پیش‌نویس",
              render: (b) => (
                <span>
                  نسخه {b.version} ·{" "}
                  {b.draft.catalog_category_id
                    ? "تصویر از کاتالوگ"
                    : b.draft.desktop_id && b.draft.mobile_id
                      ? "دو تصویر ثبت شده"
                      : "رسانه ناقص"}
                </span>
              ),
            },
            {
              title: "عمومی",
              render: (b) =>
                b.published
                  ? `انتشار ${b.published.revision} · ${b.published.config.enabled ? "فعال" : "غیرفعال"}`
                  : "منتشر نشده",
            },
            {
              title: "عملیات",
              render: (b) => (
                <Button onClick={() => setSelected(b.id)}>
                  ویرایش و پیش‌نمایش
                </Button>
              ),
            },
          ]}
        />
      </section>
      {banner && (
        <BannerEditor
          key={banner.id}
          banner={banner}
          list={list}
          close={() => setSelected(undefined)}
        />
      )}
      {create && (
        <Dialog
          open
          fullWidth
          maxWidth="xl"
          onClose={() => {
            if (!busy) setCreate(false);
          }}
        >
          <DialogTitle>بنر جدید</DialogTitle>
          <form
            onSubmit={async (e) => {
              e.preventDefault();
              if (busy) return;
              setBusy(true);
              setError("");
              try {
                await execute({
                  action: "banner.create",
                  name,
                  idempotency_key: key,
                });
                setCreate(false);
                setName("");
                setKey(crypto.randomUUID());
              } catch (e) {
                setError(message(e));
              } finally {
                setBusy(false);
              }
            }}
          >
            <DialogContent className="banner-form">
              {error && <Alert severity="error">{error}</Alert>}
              <TextField
                label="نام بنر"
                required
                value={name}
                disabled={busy}
                inputProps={{ maxLength: 80 }}
                onChange={(e) => {
                  setName(e.target.value);
                  setKey(crypto.randomUUID());
                }}
              />
            </DialogContent>
            <DialogActions>
              <Button disabled={busy} onClick={() => setCreate(false)}>
                انصراف
              </Button>
              <Button type="submit" variant="contained" disabled={busy}>
                ساخت پیش‌نویس
              </Button>
            </DialogActions>
          </form>
        </Dialog>
      )}
    </>
  );
}

function BannerEditor({
  banner,
  list,
  close,
}: {
  banner: Banner;
  list: BannerList;
  close: () => void;
}) {
  const { state, execute, uploadBanner, loadBanners } = useWorkspace();
  const cache = useQueryClient();
  const [draft, setDraft] = useState(banner.draft),
    [saved, setSaved] = useState(banner.draft);
  const [version, setVersion] = useState(banner.version),
    [error, setError] = useState("");
  const [busy, setBusy] = useState(false),
    [notice, setNotice] = useState("");
  const [files, setFiles] = useState<Record<Slot, boolean>>({
    desktop: false,
    mobile: false,
  });
  const [saveKey, setSaveKey] = useState(crypto.randomUUID()),
    [publishKey, setPublishKey] = useState(crypto.randomUUID());
  const [removeKeys, setRemoveKeys] = useState<Record<Slot, string>>({
    desktop: crypto.randomUUID(),
    mobile: crypto.randomUUID(),
  });
  const dirty =
    draft.alt !== saved.alt ||
    draft.link !== saved.link ||
    draft.enabled !== saved.enabled ||
    draft.placement !== saved.placement ||
    draft.catalog_category_id !== saved.catalog_category_id;
  const change = (next: typeof draft) => {
    setDraft(next);
    setSaveKey(crypto.randomUUID());
    setNotice("");
  };
  const editable = list.can_edit;
  async function refresh() {
    setBusy(true);
    try {
      const latest = await loadBanners();
      cache.setQueryData(["banners"], latest);
      const b = latest.banners.find((x) => x.id === banner.id);
      if (!b) throw new Error("بنر در دسترس نیست.");
      setVersion(b.version);
      setDraft({
        ...draft,
        desktop_id: b.draft.desktop_id,
        mobile_id: b.draft.mobile_id,
      });
      setSaved(b.draft);
      setSaveKey(crypto.randomUUID());
      setPublishKey(crypto.randomUUID());
      setError("");
      setNotice(
        "نسخه تازه خوانده شد؛ متن و لینک شما حفظ شده‌اند. پیش از ثبت دوباره بازبینی کنید.",
      );
    } catch (e) {
      setError(message(e));
    } finally {
      setBusy(false);
    }
  }
  async function save() {
    if (!editable || busy) return;
    setBusy(true);
    setError("");
    setNotice("");
    try {
      await execute({
        action: "banner.save",
        target: banner.id,
        version,
        banner: draft,
        idempotency_key: saveKey,
      });
      setVersion(version + 1);
      setSaved(draft);
      setNotice("پیش‌نویس ذخیره شد؛ نسخه عمومی تغییر نکرد.");
      setPublishKey(crypto.randomUUID());
    } catch (e) {
      setError(message(e));
    } finally {
      setBusy(false);
    }
  }
  async function publish() {
    if (!editable || busy || dirty || files.desktop || files.mobile) return;
    setBusy(true);
    setError("");
    setNotice("");
    try {
      await execute({
        action: "banner.publish",
        target: banner.id,
        version,
        idempotency_key: publishKey,
      });
      setVersion(version + 1);
      setPublishKey(crypto.randomUUID());
      setNotice("نسخه تازه منتشر شد.");
    } catch (e) {
      setError(message(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <Dialog
      open
      fullWidth
      maxWidth="xl"
      onClose={() => {
        if (!busy) close();
      }}
    >
      <DialogTitle>
        {banner.name} <Chip size="small" label={`پیش‌نویس ${version}`} />
      </DialogTitle>
      <DialogContent className="banner-form">
        <Alert severity="info">
          بارگذاری تصویر و ذخیره فرم، فقط پیش‌نویس را تغییر می‌دهند. انتشار، یک
          اقدام جداگانه است.
        </Alert>
        {error && (
          <Alert severity="error">
            {error}
            <Button disabled={busy} onClick={() => void refresh()}>
              بازخوانی و بازبینی
            </Button>
          </Alert>
        )}
        {notice && <Alert severity="success">{notice}</Alert>}
        <div className="banner-fields">
          <div className="home-admin-form">
            <b>جایگاه بنر</b>
            <div className="home-admin-actions">
              {(["hero", "promo"] as const).map((value) => (
                <Button
                  key={value}
                  disabled={!editable || busy}
                  variant={
                    (draft.placement ?? "hero") === value
                      ? "contained"
                      : "outlined"
                  }
                  onClick={() =>
                    change({
                      ...draft,
                      placement: value,
                      catalog_category_id:
                        value === "hero"
                          ? undefined
                          : draft.catalog_category_id,
                    })
                  }
                >
                  {value === "hero" ? "بنر اصلی صفحه" : "بنر مکمل کنار هم"}
                </Button>
              ))}
            </div>
          </div>
          {draft.placement === "promo" && (
            <div className="home-admin-form">
              <b>منبع تصویر و مقصد</b>
              <div className="home-admin-actions">
                <Button
                  disabled={!editable || busy}
                  variant={
                    !draft.catalog_category_id ? "contained" : "outlined"
                  }
                  onClick={() =>
                    change({ ...draft, catalog_category_id: undefined })
                  }
                >
                  تصویر ثابت
                </Button>
                {state.home_settings.rails
                  .filter((r) => r.category_id)
                  .map((r) => (
                    <Button
                      key={r.key}
                      disabled={!editable || busy}
                      variant={
                        draft.catalog_category_id === r.category_id
                          ? "contained"
                          : "outlined"
                      }
                      onClick={() =>
                        change({
                          ...draft,
                          catalog_category_id: r.category_id,
                          link: "/category/" + r.category_id,
                        })
                      }
                    >
                      {state.categories?.find((c) => c.id === r.category_id)
                        ?.name ?? r.key}
                    </Button>
                  ))}
              </div>
              <small>
                در حالت داده‌محور، تصویر از اولین کالای موجود همان کروسل و مقصد
                از دسته آن گرفته می‌شود؛ بدون کالای موجود بنر نمایش داده
                نمی‌شود.
              </small>
            </div>
          )}
          <TextField
            label="متن جایگزین تصویر"
            value={draft.alt}
            disabled={!editable || busy}
            inputProps={{ maxLength: 200 }}
            onChange={(e) => change({ ...draft, alt: e.target.value })}
          />
          <TextField
            label="لینک داخلی (اختیاری)"
            value={draft.link}
            disabled={!editable || busy || !!draft.catalog_category_id}
            inputProps={{ dir: "ltr", maxLength: 2048 }}
            helperText="مانند /categories/mobile؛ بدون لینک هم می‌توان منتشر کرد."
            onChange={(e) => change({ ...draft, link: e.target.value })}
          />
        </div>
        <FormControlLabel
          label="نمایش در نسخه عمومی پس از انتشار"
          control={
            <Switch
              checked={draft.enabled}
              disabled={!editable || busy}
              onChange={(_, checked) => change({ ...draft, enabled: checked })}
            />
          }
        />
        {draft.catalog_category_id && (
          <Alert severity="info">
            تصویر و چیدمان این بنر از کاتالوگ موجود ساخته می‌شود؛ نیازی به
            بارگذاری تصویر مستقل نیست.
          </Alert>
        )}
        <div className="banner-preview-grid">
          {!draft.catalog_category_id &&
            (["desktop", "mobile"] as const).map((slot) => (
              <BannerMediaField
                key={slot}
                slot={slot}
                id={slot === "desktop" ? draft.desktop_id : draft.mobile_id}
                alt={draft.alt}
                policy={list.policy}
                disabled={!editable || busy}
                pendingFile={(value) =>
                  setFiles((prev) => ({ ...prev, [slot]: value }))
                }
                upload={async (file, key) => {
                  setBusy(true);
                  setError("");
                  setNotice("");
                  try {
                    const result = await uploadBanner({
                      id: banner.id,
                      slot,
                      version,
                      key,
                      file,
                    });
                    setVersion(result.version);
                    setDraft((prev) => ({
                      ...prev,
                      [slot === "desktop" ? "desktop_id" : "mobile_id"]:
                        result.asset_id,
                    }));
                    setNotice(
                      `تصویر ${slot === "desktop" ? "دسکتاپ" : "موبایل"} در پیش‌نویس ثبت شد.`,
                    );
                    setPublishKey(crypto.randomUUID());
                  } catch (e) {
                    setError(message(e));
                    throw e;
                  } finally {
                    setBusy(false);
                  }
                }}
                remove={async () => {
                  setBusy(true);
                  setError("");
                  try {
                    await execute({
                      action: "banner.media.remove",
                      target: banner.id,
                      version,
                      value: slot,
                      idempotency_key: removeKeys[slot],
                    });
                    setVersion(version + 1);
                    setDraft((prev) => ({
                      ...prev,
                      [slot === "desktop" ? "desktop_id" : "mobile_id"]: "",
                    }));
                    setRemoveKeys({
                      ...removeKeys,
                      [slot]: crypto.randomUUID(),
                    });
                    setPublishKey(crypto.randomUUID());
                    setNotice("تصویر از پیش‌نویس حذف شد؛ نسخه عمومی حفظ شد.");
                  } catch (e) {
                    setError(message(e));
                  } finally {
                    setBusy(false);
                  }
                }}
              />
            ))}
        </div>
        <p className="muted">
          PNG، JPEG یا WebP؛ حداکثر {list.policy.max_bytes / 1024 / 1024}{" "}
          مگابایت، هر ضلع حداکثر {list.policy.max_dimension} پیکسل. نسبت تصویر
          حفظ می‌شود؛ crop خودکار نداریم.
        </p>
        {banner.published && (
          <p>
            نسخه عمومی: {banner.published.revision} ·{" "}
            {banner.published.config.enabled ? "فعال" : "غیرفعال"}{" "}
            {banner.published.config.enabled && (
              <a
                href={api.bannerViewUrl(banner.id)}
                target="_blank"
                rel="noreferrer"
              >
                مشاهده نسخه منتشرشده
              </a>
            )}
          </p>
        )}
        {(dirty || files.desktop || files.mobile) && (
          <p role="status">
            تغییر ثبت‌نشده دارید؛ ابتدا پیش‌نویس و تصاویر انتخاب‌شده را ثبت یا
            لغو کنید.
          </p>
        )}
      </DialogContent>
      <DialogActions className="banner-actions">
        <Button disabled={busy} onClick={close}>
          بستن
        </Button>
        <Button
          disabled={!editable || busy || !dirty}
          onClick={() => void save()}
        >
          ذخیره پیش‌نویس
        </Button>
        <Button
          variant="contained"
          disabled={
            !editable ||
            busy ||
            dirty ||
            files.desktop ||
            files.mobile ||
            (!draft.catalog_category_id &&
              (!draft.desktop_id || !draft.mobile_id)) ||
            !draft.alt.trim()
          }
          onClick={() => void publish()}
        >
          انتشار نسخه
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function BannerMediaField({
  slot,
  id,
  alt,
  policy,
  disabled,
  pendingFile,
  upload,
  remove,
}: {
  slot: Slot;
  id: string;
  alt: string;
  policy: BannerList["policy"];
  disabled: boolean;
  pendingFile: (value: boolean) => void;
  upload: (file: File, key: string) => Promise<void>;
  remove: () => Promise<void>;
}) {
  const { loadBannerAsset } = useWorkspace();
  const [url, setURL] = useState("");
  const [selected, setSelected] = useState<File>();
  const [localURL, setLocalURL] = useState("");
  const [error, setError] = useState("");
  const [key, setKey] = useState(crypto.randomUUID());
  const [checking, setChecking] = useState(false);
  const attempt = useRef(0);
  const title = slot === "desktop" ? "دسکتاپ" : "موبایل";
  useEffect(() => {
    let cancelled = false,
      objectURL = "";
    setURL("");
    if (id)
      loadBannerAsset(id)
        .then((blob) => {
          objectURL = URL.createObjectURL(blob);
          if (!cancelled) setURL(objectURL);
          else URL.revokeObjectURL(objectURL);
        })
        .catch(() => {
          if (!cancelled) setError("دریافت تصویر ثبت‌شده انجام نشد.");
        });
    return () => {
      cancelled = true;
      if (objectURL) URL.revokeObjectURL(objectURL);
    };
  }, [id, loadBannerAsset]);
  useEffect(() => {
    if (!selected) {
      setLocalURL("");
      return;
    }
    const u = URL.createObjectURL(selected);
    setLocalURL(u);
    return () => URL.revokeObjectURL(u);
  }, [selected]);
  useEffect(
    () => () => {
      attempt.current++;
    },
    [],
  );
  async function choose(file?: File) {
    const current = ++attempt.current;
    if (!file) return;
    setError("");
    if (
      !policy.content_types.includes(file.type) ||
      file.size > policy.max_bytes
    ) {
      setError("قالب یا حجم فایل مجاز نیست؛ تصویر قبلی حفظ شد.");
      return;
    }
    setChecking(true);
    pendingFile(true);
    const u = URL.createObjectURL(file);
    try {
      const img = new Image();
      img.src = u;
      await img.decode();
      if (
        img.naturalWidth > policy.max_dimension ||
        img.naturalHeight > policy.max_dimension ||
        img.naturalWidth * img.naturalHeight > policy.max_pixels
      )
        throw new Error("dimensions");
      if (current === attempt.current) {
        setSelected(file);
        setKey(crypto.randomUUID());
        pendingFile(true);
      }
    } catch {
      if (current === attempt.current) {
        pendingFile(!!selected);
        setError(
          "تصویر خراب است یا ابعاد آن از حد مجاز بیشتر است؛ تصویر قبلی حفظ شد.",
        );
      }
    } finally {
      URL.revokeObjectURL(u);
      if (current === attempt.current) setChecking(false);
    }
  }
  return (
    <section
      className={`banner-media banner-media-${slot}`}
      data-testid={`banner-${slot}-preview`}
    >
      <h3>تصویر {title}</h3>
      <p className="muted">
        {slot === "desktop" ? "پیشنهاد: ۱۴۴۰ × ۴۸۰" : "پیشنهاد: ۷۵۰ × ۹۰۰"}؛
        ابعاد اجباری نیست.
      </p>
      <div className="banner-image-frame">
        {localURL || url ? (
          <img
            src={localURL || url}
            alt={alt || `پیش‌نمایش بنر ${title}`}
            onError={() => setError("تصویر قابل نمایش نیست.")}
          />
        ) : (
          <span>تصویری ثبت نشده است.</span>
        )}
      </div>
      {error && <Alert severity="error">{error}</Alert>}
      <div className="banner-media-actions">
        <Button component="label" disabled={disabled || checking}>
          انتخاب تصویر {title}
          <input
            hidden
            type="file"
            accept={policy.content_types.join(",")}
            aria-label={`فایل بنر ${title}`}
            onChange={(e) => {
              void choose(e.target.files?.[0]);
              e.target.value = "";
            }}
          />
        </Button>
        {selected && (
          <>
            <Button
              variant="outlined"
              disabled={disabled || checking}
              onClick={async () => {
                try {
                  await upload(selected, key);
                  setSelected(undefined);
                  pendingFile(false);
                } catch {
                  /* keep selected preview for retry */
                }
              }}
            >
              ثبت تصویر {title}
            </Button>
            <Button
              disabled={disabled}
              onClick={() => {
                setSelected(undefined);
                pendingFile(false);
              }}
            >
              لغو انتخاب {title}
            </Button>
          </>
        )}
        {id && !selected && (
          <Button
            color="warning"
            disabled={disabled || checking}
            onClick={() => void remove()}
          >
            حذف تصویر {title}
          </Button>
        )}
      </div>
      {checking && <p role="status">در حال بررسی تصویر…</p>}
      {selected && <small>پیش‌نمایش محلی؛ هنوز در سرور ثبت نشده است.</small>}
    </section>
  );
}
