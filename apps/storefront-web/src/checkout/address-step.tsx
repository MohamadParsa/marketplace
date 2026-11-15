import { useState, type FormEvent } from "react";
import {
  destinations,
  digits,
  validateAddress,
  type Address,
  type CheckoutDraft,
} from "./model";

export function AddressStep({
  draft,
  onChange,
  onContinue,
}: {
  draft: CheckoutDraft;
  onChange: (draft: CheckoutDraft) => void;
  onContinue: () => void;
}) {
  const [editing, setEditing] = useState<Address | null>(null);
  const [error, setError] = useState("");
  function saveAddress(event: FormEvent) {
    event.preventDefault();
    if (!editing) return;
    const message = validateAddress(editing);
    if (message) {
      setError(message);
      return;
    }
    const address = {
      ...editing,
      postalCode: digits(editing.postalCode),
      phone: digits(editing.phone),
      title: editing.title.trim(),
      recipient: editing.recipient.trim(),
      street: editing.street.trim(),
    };
    onChange({
      ...draft,
      addresses: draft.addresses.some((existing) => existing.id === address.id)
        ? draft.addresses.map((existing) =>
            existing.id === address.id ? address : existing,
          )
        : [...draft.addresses, address],
      addressId: address.id,
      shipping: {},
    });
    setEditing(null);
    setError("");
  }
  function proceed(event: FormEvent) {
    event.preventDefault();
    if (
      draft.profile.firstName.trim().length < 2 ||
      draft.profile.lastName.trim().length < 2
    ) {
      setError("نام و نام خانوادگی صاحب حساب را کامل کنید.");
      return;
    }
    const selected = draft.addresses.find(
      (address) => address.id === draft.addressId,
    );
    if (!selected || validateAddress(selected)) {
      setError("یک آدرس معتبر انتخاب کنید.");
      return;
    }
    setError("");
    onContinue();
  }
  const fields: {
    key: "title" | "recipient" | "phone" | "postalCode" | "street";
    label: string;
    numeric?: boolean;
  }[] = [
    { key: "title", label: "عنوان آدرس" },
    { key: "recipient", label: "نام کامل گیرنده" },
    { key: "phone", label: "شماره تماس گیرنده", numeric: true },
    { key: "postalCode", label: "کدپستی", numeric: true },
    { key: "street", label: "نشانی کامل" },
  ];
  return (
    <>
      <div className="checkout-section-title">
        <span>۰۱</span>
        <div>
          <h2>اطلاعات و مقصد خرید</h2>
          <p>یک آدرس مشترک برای ارسال همه فروشگاه‌ها انتخاب کن.</p>
        </div>
      </div>
      <form onSubmit={proceed} noValidate>
        <div className="checkout-card">
          <h3>اطلاعات صاحب حساب</h3>
          <p className="checkout-muted">
            اطلاعات گیرنده می‌تواند متفاوت باشد. برای بررسی از اطلاعات ساختگی
            استفاده کن.
          </p>
          <div className="form-grid">
            {(["firstName", "lastName"] as const).map((key, index) => (
              <label key={key}>
                {index === 0 ? "نام صاحب حساب" : "نام خانوادگی صاحب حساب"}
                <input
                  className="text-input"
                  autoComplete="off"
                  value={draft.profile[key]}
                  onChange={(event) =>
                    onChange({
                      ...draft,
                      profile: { ...draft.profile, [key]: event.target.value },
                    })
                  }
                />
              </label>
            ))}
          </div>
        </div>
        <div className="checkout-card">
          <div className="address-heading">
            <h3>آدرس‌های من — نمونه</h3>
            <button
              type="button"
              className="text-button"
              onClick={() => {
                setError("");
                setEditing({
                  id: crypto.randomUUID(),
                  title: "",
                  province: "تهران",
                  city: "تهران",
                  street: "",
                  postalCode: "",
                  recipient: "",
                  phone: "",
                });
              }}
            >
              ＋ افزودن آدرس
            </button>
          </div>
          {draft.addresses.map((address) => (
            <div
              key={address.id}
              className={`address-card ${address.id === draft.addressId ? "selected" : ""}`}
            >
              <label className="address-option">
                <input
                  type="radio"
                  name="address"
                  checked={address.id === draft.addressId}
                  onChange={() =>
                    onChange({ ...draft, addressId: address.id, shipping: {} })
                  }
                />
                <div>
                  <strong>{address.title}</strong>
                  <p>
                    {address.province}، {address.city}، {address.street}
                  </p>
                  <small>
                    {address.recipient} · <b dir="ltr">{address.phone}</b>
                  </small>
                  <small>کدپستی: {address.postalCode}</small>
                </div>
              </label>
              <button
                type="button"
                className="text-button"
                aria-label={`ویرایش ${address.title}`}
                onClick={() => {
                  setEditing({ ...address });
                  setError("");
                }}
              >
                ویرایش
              </button>
            </div>
          ))}
        </div>
        {error && !editing && (
          <p className="form-error" role="alert">
            {error}
          </p>
        )}
        <button className="primary full" disabled={editing !== null}>
          تأیید اطلاعات و انتخاب ارسال ←
        </button>
      </form>
      {editing && (
        <section className="checkout-card address-editor" aria-label="فرم آدرس">
          <h3>
            {draft.addresses.some((address) => address.id === editing.id)
              ? "ویرایش آدرس"
              : "آدرس جدید"}
          </h3>
          <form onSubmit={saveAddress} noValidate>
            <div className="form-grid">
              {fields.slice(0, 4).map((field) => (
                <label key={field.key}>
                  {field.label}
                  <input
                    autoFocus={field.key === "title"}
                    className={`text-input ${field.numeric ? "ltr-input" : ""}`}
                    autoComplete="off"
                    inputMode={field.numeric ? "numeric" : "text"}
                    value={editing[field.key]}
                    onChange={(event) =>
                      setEditing({
                        ...editing,
                        [field.key]: field.numeric
                          ? digits(event.target.value)
                          : event.target.value,
                      })
                    }
                  />
                </label>
              ))}
              <label>
                استان
                <select
                  aria-label="استان آدرس"
                  value={editing.province}
                  onChange={(event) =>
                    setEditing({
                      ...editing,
                      province: event.target.value,
                      city: destinations[event.target.value][0],
                    })
                  }
                >
                  {Object.keys(destinations).map((province) => (
                    <option key={province}>{province}</option>
                  ))}
                </select>
              </label>
              <label>
                شهر
                <select
                  aria-label="شهر آدرس"
                  value={editing.city}
                  onChange={(event) =>
                    setEditing({ ...editing, city: event.target.value })
                  }
                >
                  {destinations[editing.province].map((city) => (
                    <option key={city}>{city}</option>
                  ))}
                </select>
              </label>
            </div>
            <label>
              نشانی کامل
              <textarea
                className="text-input"
                value={editing.street}
                onChange={(event) =>
                  setEditing({ ...editing, street: event.target.value })
                }
              />
            </label>
            {error && (
              <p className="form-error" role="alert">
                {error}
              </p>
            )}
            <div className="editor-actions">
              <button className="primary" type="submit">
                ذخیره آدرس در این مرور
              </button>
              <button
                className="text-button"
                type="button"
                onClick={() => {
                  setEditing(null);
                  setError("");
                }}
              >
                انصراف
              </button>
            </div>
          </form>
        </section>
      )}
    </>
  );
}
