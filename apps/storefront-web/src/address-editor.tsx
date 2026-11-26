import { AddressMap } from "./address-map";
import { useEffect, useState } from "react";
import type {
  Address,
  CustomerInfo,
  Province,
} from "../../../packages/commerce-client/types";
import { ToggleSwitch } from "./toggle-switch";
import { call, useCommerce } from "./commerce";
import { BusyRegion } from "./busy-region";

export function AddressEditor({
  initial,
  geography,
  busy,
  error,
  onSave,
  allowOtherRecipient = false,
}: {
  allowOtherRecipient?: boolean;
  initial: Address;
  geography: Province[];
  busy: boolean;
  error: string;
  onSave: (address: Address) => void;
}) {
  const { token, profile, refreshProfile } = useCommerce();
  const [draft, setDraft] = useState<Address>(() => ({
    ...initial,
    recipient_self: initial.recipient_self ?? !initial.id,
  }));
  const [selfProfile, setSelfProfile] = useState<CustomerInfo | null>(profile);
  useEffect(() => {
    setSelfProfile(profile);
  }, [profile]);
  const [profileNeeded, setProfileNeeded] = useState(false);
  const [profileBusy, setProfileBusy] = useState(false),
    [profileError, setProfileError] = useState("");
  const [profileName, setProfileName] = useState(
    profile?.name === "مشتری" ? "" : (profile?.name ?? ""),
  );
  const [nationalID, setNationalID] = useState(profile?.national_id ?? "");
  const ready =
    !!selfProfile?.name &&
    selfProfile.name !== "مشتری" &&
    !!selfProfile.national_id;
  const province = geography.find((p) => p.active && p.name === draft.province);
  const cityValid = !!province?.cities.some(
    (c) => c.active && c.name === draft.city,
  );
  const update = (key: keyof Address, value: string) =>
    setDraft((d) => ({ ...d, [key]: value }));
  const field = (
    key: keyof Address,
    label: string,
    required = false,
    maxLength = 100,
    numeric = false,
  ) => (
    <label
      className={
        "buy-field" +
        (["title", "full_address", "street", "postal_code"].includes(key)
          ? " address-field-wide"
          : "")
      }
      key={key}
    >
      {label}
      {required ? " *" : ""}
      <input
        name={key}
        value={String(draft[key] ?? "")}
        required={required}
        maxLength={maxLength}
        inputMode={numeric ? "numeric" : undefined}
        dir={numeric ? "ltr" : undefined}
        onChange={(e) => update(key, e.target.value)}
      />
    </label>
  );
  return (
    <form
      className="address-editor"
      onSubmit={(e) => {
        e.preventDefault();
        if (draft.recipient_self && !ready) {
          setProfileNeeded(true);
          return;
        }
        const {
          id,
          version,
          title,
          full_address,
          house_number,
          unit,
          recipient_self,
          recipient_national_id,
          recipient,
          phone,
          province,
          city,
          street,
          postal_code,
          latitude,
          longitude,
        } = draft;
        onSave({
          id,
          version,
          title,
          full_address,
          house_number,
          unit,
          recipient_self,
          recipient_national_id,
          recipient,
          phone,
          province,
          city,
          street,
          postal_code,
          latitude,
          longitude,
        });
      }}
    >
      <fieldset className="buy-fields" disabled={busy || profileBusy}>
        {field("title", "عنوان آدرس (مثلاً خانه)")}
        <AddressMap
          latitude={draft.latitude}
          longitude={draft.longitude}
          disabled={busy || profileBusy}
          onPick={(point) =>
            setDraft((d) => ({
              ...d,
              latitude: point?.lat,
              longitude: point?.lng,
            }))
          }
          onSuggestion={(suggestion) =>
            setDraft((d) => ({
              ...d,
              full_address: suggestion.full_address,
              ...(suggestion.geography_matched
                ? { province: suggestion.province, city: suggestion.city }
                : {}),
            }))
          }
        />
        {field("full_address", "نشانی کامل قابل ویرایش", false, 1500)}

        <label className="buy-field">
          استان *
          <select
            name="province"
            required
            value={province ? draft.province : ""}
            onChange={(e) =>
              setDraft((d) => ({ ...d, province: e.target.value, city: "" }))
            }
          >
            <option value="">انتخاب استان</option>
            {geography
              .filter((p) => p.active)
              .map((p) => (
                <option key={p.id} value={p.name}>
                  {p.name}
                </option>
              ))}
          </select>
        </label>
        <label className="buy-field">
          شهر *
          <select
            name="city"
            required
            disabled={!province}
            value={cityValid ? draft.city : ""}
            onChange={(e) => update("city", e.target.value)}
          >
            <option value="">انتخاب شهر</option>
            {province?.cities
              .filter((c) => c.active)
              .map((c) => (
                <option key={c.id} value={c.name}>
                  {c.name}
                </option>
              ))}
          </select>
        </label>
        {field("street", "آدرس (خیابان و کوچه)", true, 1000)}
        {field("house_number", "پلاک", true, 32)}
        {field("unit", "واحد", false, 32)}
        {field("postal_code", "کد پستی", true, 10, true)}
        {allowOtherRecipient && (
          <section className="address-recipient">
            <div className="buy-row">
              <ToggleSwitch
                checked={!!draft.recipient_self}
                label="تحویل‌گیرنده خودم هستم"
                onChange={(v) => {
                  setDraft((d) => ({ ...d, recipient_self: v }));
                  setProfileNeeded(false);
                }}
              />
              <b>تحویل‌گیرنده خودم هستم</b>
            </div>
            {!draft.recipient_self && (
              <div className="address-other-recipient">
                {field("recipient", "نام و نام خانوادگی گیرنده", true)}
                {field(
                  "recipient_national_id",
                  "کد ملی گیرنده",
                  true,
                  10,
                  true,
                )}
                {field("phone", "شماره موبایل گیرنده", true, 11, true)}
              </div>
            )}
          </section>
        )}
        {(profileNeeded || !selfProfile?.national_id) &&
          draft.recipient_self &&
          !ready && (
            <div className="address-profile-completion">
              <p>
                نام و نام خانوادگی و کد ملی را یک بار در پروفایل تکمیل کنید.
              </p>
              <label className="buy-field">
                نام و نام خانوادگی *
                <input
                  value={profileName}
                  onChange={(e) => setProfileName(e.target.value)}
                  maxLength={100}
                />
              </label>
              <label className="buy-field">
                کد ملی *
                <input
                  value={nationalID}
                  onChange={(e) => setNationalID(e.target.value)}
                  maxLength={10}
                  inputMode="numeric"
                  dir="ltr"
                />
              </label>
              <BusyRegion busy={profileBusy}>
                <button
                  type="button"
                  className="buy-secondary"
                  disabled={
                    profileBusy || !profileName.trim() || !nationalID.trim()
                  }
                  onClick={() => {
                    setProfileBusy(true);
                    setProfileError("");
                    void call(
                      "/storefront/account/profile",
                      token,
                      { name: profileName, national_id: nationalID },
                      "PATCH",
                    )
                      .then(() => refreshProfile())
                      .then(setSelfProfile)
                      .catch((e) => setProfileError((e as Error).message))
                      .finally(() => setProfileBusy(false));
                  }}
                >
                  ثبت اطلاعات در پروفایل
                </button>
              </BusyRegion>
              {profileError && (
                <p className="buy-alert" role="alert">
                  {profileError}
                </p>
              )}
            </div>
          )}
        {!geography.length && (
          <p className="buy-alert">فهرست استان و شهر هنوز تنظیم نشده است.</p>
        )}
        {error && (
          <p className="buy-alert" role="alert">
            {error}
          </p>
        )}
        <button className="buy-primary" disabled={!province || !cityValid}>
          ذخیره آدرس
        </button>
      </fieldset>
    </form>
  );
}
