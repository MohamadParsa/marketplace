import {
  useEffect,
  useLayoutEffect,
  useId,
  useRef,
  useState,
  type KeyboardEvent,
  type ComponentProps,
} from "react";
import {
  addISODays,
  asciiDigits,
  parsePersianCapacityDate,
  parsePersianDateTime,
  persianCapacityDate,
  persianDateTime,
  persianDigits,
  tehranCapacityDate,
  validISODate,
} from "./calendar";
import "./persian-date-field.css";

const months = [
  "فروردین",
  "اردیبهشت",
  "خرداد",
  "تیر",
  "مرداد",
  "شهریور",
  "مهر",
  "آبان",
  "آذر",
  "دی",
  "بهمن",
  "اسفند",
];
function yearMonth(iso: string) {
  return (persianCapacityDate(iso) || persianCapacityDate(tehranCapacityDate()))
    .split("/")
    .slice(0, 2)
    .map(Number);
}

export function PersianDateField({
  label,
  value,
  onChange,
  dateTime = false,
  required = false,
  disabled = false,
  readOnly = false,
  min,
  max,
  name,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  dateTime?: boolean;
  required?: boolean;
  disabled?: boolean;
  readOnly?: boolean;
  min?: string;
  max?: string;
  name?: string;
}) {
  const format = (date: string) =>
    persianDigits(dateTime ? persianDateTime(date) : persianCapacityDate(date));
  const parse = (text: string) =>
    dateTime ? parsePersianDateTime(text) : parsePersianCapacityDate(text);
  const id = useId(),
    input = useRef<HTMLInputElement>(null),
    dialog = useRef<HTMLDialogElement>(null),
    emitted = useRef(value);
  const [text, setText] = useState(() => format(value));
  const [view, setView] = useState(() =>
    yearMonth(
      dateTime
        ? parsePersianCapacityDate(persianDateTime(value).split(" ")[0])
        : value,
    ),
  );
  const [yearText, setYearText] = useState(() => persianDigits(view[0]));
  const pendingFocus = useRef<string>();
  useLayoutEffect(() => {
    if (pendingFocus.current) {
      dialog.current
        ?.querySelector<HTMLButtonElement>(
          `[data-day="${pendingFocus.current}"]`,
        )
        ?.focus();
      pendingFocus.current = undefined;
    }
  }, [view[0], view[1]]);
  useEffect(() => setYearText(persianDigits(view[0])), [view[0]]);
  useEffect(() => {
    if (value !== emitted.current) {
      emitted.current = value;
      setText(format(value));
    }
  }, [value, dateTime]);
  const parsed = parse(text),
    invalid = !!text && !parsed,
    range =
      !!parsed &&
      (dateTime
        ? (!!min && Date.parse(parsed) < Date.parse(min)) ||
          (!!max && Date.parse(parsed) > Date.parse(max))
        : (!!min && parsed < min) || (!!max && parsed > max));
  const error = invalid
    ? dateTime
      ? "تاریخ و ساعت شمسی معتبر وارد کنید؛ ساعت بر مبنای تهران است."
      : "تاریخ شمسی معتبر وارد کنید؛ روز این ماه را بررسی کنید."
    : range
      ? "تاریخ خارج از محدودهٔ مجاز است."
      : "";
  useEffect(() => {
    input.current?.setCustomValidity(error);
  }, [error]);
  function edit(next: string) {
    setText(persianDigits(asciiDigits(next)));
    const iso = parse(next);
    emitted.current = iso || (next.trim() ? next : "");
    onChange(emitted.current);
  }
  function open() {
    if (disabled || readOnly || input.current?.matches(":disabled")) return;
    const day = dateTime
      ? parsePersianCapacityDate(persianDateTime(value).split(" ")[0])
      : value;
    setView(yearMonth(day));
    dialog.current?.showModal();
  }
  function choose(iso: string) {
    if (disabled || readOnly) return;
    const day = persianCapacityDate(iso);
    const time = asciiDigits(text).split(/[ T]/)[1] || "00:00";
    edit(persianDigits(day + (dateTime ? " " + time : "")));
    dialog.current?.close();
    input.current?.focus();
  }
  function shift(delta: number) {
    const offset = view[0] * 12 + view[1] - 1 + delta;
    const y = Math.floor(offset / 12);
    if (y >= 1000 && y <= 9000) setView([y, (offset % 12) + 1]);
  }
  const first = parsePersianCapacityDate(`${view[0]}/${view[1]}/01`),
    offset = first ? (new Date(first + "T12:00:00Z").getUTCDay() + 1) % 7 : 0;
  const days: string[] = [];
  for (let i = 0; i < 31 && first; i++) {
    const day = addISODays(first, i);
    if (Number(persianCapacityDate(day).split("/")[1]) !== view[1]) break;
    days.push(day);
  }
  const selected = dateTime
    ? parsePersianCapacityDate(persianDateTime(value).split(" ")[0])
    : value;
  function keyboard(event: KeyboardEvent<HTMLButtonElement>, day: string) {
    const delta: Record<string, number> = {
      ArrowLeft: 1,
      ArrowRight: -1,
      ArrowUp: -7,
      ArrowDown: 7,
    };
    if (!(event.key in delta)) return;
    event.preventDefault();
    const next = addISODays(day, delta[event.key]);
    if (!dateTime && ((!!min && next < min) || (!!max && next > max))) return;
    const existing = dialog.current?.querySelector<HTMLButtonElement>(
      `[data-day="${next}"]`,
    );
    if (existing && !existing.disabled) {
      existing.focus();
      return;
    }
    pendingFocus.current = next;
    setView(yearMonth(next));
  }
  return (
    <div className="persian-date-field">
      <label htmlFor={id}>
        {label}
        {required ? " *" : ""}
      </label>
      {name && (
        <input type="hidden" name={name} value={value} disabled={disabled} />
      )}
      <div className="persian-date-input">
        <input
          ref={input}
          id={id}
          type="text"
          value={text}
          dir="ltr"
          placeholder={dateTime ? "۱۴۰۵/۰۱/۰۱ ۱۴:۳۰" : "۱۴۰۵/۰۱/۰۱"}
          required={required}
          disabled={disabled}
          readOnly={readOnly}
          aria-invalid={!!error}
          aria-describedby={`${id}-hint`}
          onChange={(e) => edit(e.target.value)}
          onBlur={() => {
            if (parsed) setText(format(parsed));
          }}
        />
        <button
          type="button"
          aria-label={`باز کردن تقویم ${label}`}
          disabled={disabled || readOnly}
          onClick={open}
        >
          تقویم
        </button>
        {!required && (
          <button
            type="button"
            aria-label={`پاک کردن ${label}`}
            disabled={disabled || readOnly || !text}
            onClick={() => edit("")}
          >
            ×
          </button>
        )}
      </div>
      <small
        id={`${id}-hint`}
        className={error ? "persian-date-error" : ""}
        role={error ? "alert" : undefined}
      >
        {error || (dateTime ? "تاریخ شمسی · ساعت تهران" : "تاریخ شمسی")}
      </small>
      <dialog
        ref={dialog}
        onClose={(event) => event.stopPropagation()}
        onCancel={(event) => event.stopPropagation()}
        onKeyDown={(event) => {
          if (event.key === "Escape") event.stopPropagation();
          if (
            event.key === "Enter" &&
            (event.target as HTMLElement).tagName === "INPUT"
          ) {
            event.preventDefault();
            event.stopPropagation();
          }
        }}
        aria-label={`تقویم شمسی ${label}`}
        className="persian-calendar"
        onClick={(e) => {
          e.stopPropagation();
          if (e.target === dialog.current) dialog.current.close();
        }}
      >
        <div className="persian-calendar-heading">
          <button type="button" onClick={() => shift(-1)} aria-label="ماه قبل">
            ‹
          </button>
          <label>
            ماه
            <select
              value={view[1]}
              onChange={(e) => setView([view[0], Number(e.target.value)])}
            >
              {months.map((month, i) => (
                <option key={month} value={i + 1}>
                  {month}
                </option>
              ))}
            </select>
          </label>
          <label>
            سال
            <input
              aria-label="سال تقویم شمسی"
              inputMode="numeric"
              dir="ltr"
              value={yearText}
              onChange={(e) => {
                setYearText(persianDigits(asciiDigits(e.target.value)));
                const raw = asciiDigits(e.target.value),
                  year = Number(raw);
                if (/^\d{4}$/.test(raw) && year >= 1000 && year <= 9000)
                  setView([year, view[1]]);
              }}
              onBlur={() => setYearText(persianDigits(view[0]))}
            />
          </label>
          <button type="button" onClick={() => shift(1)} aria-label="ماه بعد">
            ›
          </button>
        </div>
        <div
          className="persian-calendar-grid"
          role="group"
          aria-label="روزهای ماه شمسی"
        >
          {["ش", "ی", "د", "س", "چ", "پ", "ج"].map((day, index) => (
            <span key={index} aria-hidden="true">
              {day}
            </span>
          ))}
          {Array.from({ length: offset }, (_, i) => (
            <span key={`blank-${i}`} />
          ))}
          {days.map((day, i) => (
            <button
              type="button"
              key={day}
              data-day={day}
              aria-label={persianDigits(persianCapacityDate(day))}
              aria-pressed={day === selected}
              disabled={
                !dateTime && ((!!min && day < min) || (!!max && day > max))
              }
              onKeyDown={(e) => keyboard(e, day)}
              onClick={() => choose(day)}
            >
              {persianDigits(i + 1)}
            </button>
          ))}
        </div>
        <div className="persian-calendar-footer">
          <button
            type="button"
            disabled={
              !dateTime &&
              ((!!min && tehranCapacityDate() < min) ||
                (!!max && tehranCapacityDate() > max))
            }
            onClick={() => choose(tehranCapacityDate())}
          >
            امروز
          </button>
          <button type="button" onClick={() => dialog.current?.close()}>
            بستن تقویم
          </button>
        </div>
      </dialog>
    </div>
  );
}

export function PersianDateInput({
  defaultValue = "",
  ...props
}: Omit<ComponentProps<typeof PersianDateField>, "value" | "onChange"> & {
  defaultValue?: string;
}) {
  const [date, setDate] = useState(defaultValue);
  return <PersianDateField {...props} value={date} onChange={setDate} />;
}

export function PersianDatesField({
  label,
  values,
  onChange,
  disabled = false,
}: {
  label: string;
  values: string[];
  onChange: (values: string[]) => void;
  disabled?: boolean;
}) {
  const [next, setNext] = useState("");
  return (
    <section className="persian-date-list" aria-label={label}>
      <h4>{label}</h4>
      {values.map((value, index) => (
        <div key={index} className="persian-date-list-row">
          <PersianDateField
            label={`روز تعطیل ${(index + 1).toLocaleString("fa-IR")}`}
            value={value}
            disabled={disabled}
            required
            onChange={(day) =>
              onChange(values.map((old, i) => (i === index ? day : old)))
            }
          />
          <button
            type="button"
            disabled={disabled}
            onClick={() => onChange(values.filter((_, i) => i !== index))}
            aria-label={`حذف روز تعطیل ${(index + 1).toLocaleString("fa-IR")}`}
          >
            حذف
          </button>
        </div>
      ))}
      <div className="persian-date-list-row">
        <PersianDateField
          label="روز تعطیل جدید"
          value={next}
          disabled={disabled}
          onChange={setNext}
        />
        <button
          type="button"
          disabled={
            disabled ||
            !validISODate(next) ||
            values.includes(next) ||
            values.some((v) => !validISODate(v))
          }
          onClick={() => {
            onChange([...values, next].sort());
            setNext("");
          }}
        >
          افزودن روز تعطیل
        </button>
      </div>
    </section>
  );
}
