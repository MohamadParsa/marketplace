import { ChoiceField, MultiChoiceField } from "../../shared/option-select";
import { OptionSelect } from "../../shared/option-select";
import { useState } from "react";
import {
  Alert,
  Button,
  Checkbox,
  FormControlLabel,
  TextField,
} from "@mui/material";
import type { ShippingSchedule } from "../../../../../packages/commerce-client/types";
import { asciiDigits } from "./shipping-dates";

const weekdays: [number, string][] = [
  [6, "شنبه"],
  [0, "یکشنبه"],
  [1, "دوشنبه"],
  [2, "سه‌شنبه"],
  [3, "چهارشنبه"],
  [4, "پنج‌شنبه"],
  [5, "جمعه"],
];

export function scheduleError(schedule?: ShippingSchedule | null): string {
  if (!schedule) return "برنامهٔ تحویل هنوز تعریف نشده است.";
  if (schedule.windows.length > 200)
    return "تعداد تعریف ساعت بیش از حد مجاز است.";
  const ids = new Set<string>();
  for (const window of schedule.windows) {
    if (
      ids.has(window.id) ||
      !/^([01]\d|2[0-3]):[0-5]\d$/.test(window.start) ||
      !/^([01]\d|2[0-3]):[0-5]\d$/.test(window.end) ||
      window.start >= window.end
    )
      return "شروع و پایان هر بازه را کامل کنید؛ شروع باید پیش از پایان باشد.";
    ids.add(window.id);
  }
  for (const day of schedule.days) {
    const windows = [];
    for (const slot of day.slots) {
      const window = schedule.windows.find((w) => w.id === slot.window_id);
      if (
        !window ||
        slot.capacity == null ||
        !Number.isInteger(slot.capacity) ||
        slot.capacity < 0 ||
        slot.capacity > 1000000
      )
        return "ظرفیت صحیح و نامنفی هر بازهٔ منتسب را وارد کنید.";
      if (day.active && window.active) windows.push(window);
    }
    windows.sort((a, b) => a.start.localeCompare(b.start));
    for (let i = 1; i < windows.length; i++)
      if (windows[i].start < windows[i - 1].end)
        return "بازه‌های فعال یک روز نباید هم‌پوشان باشند.";
  }
  return "";
}

export function WeeklyScheduleEditor({
  value,
  onChange,
  disabled,
  error,
}: {
  value: ShippingSchedule;
  onChange: (value: ShippingSchedule) => void;
  disabled: boolean;
  error?: string;
}) {
  const [selected, setSelected] = useState(6),
    [copyTargets, setCopyTargets] = useState<number[]>([]);
  const day = value.days.find((d) => d.weekday === selected) ?? {
    weekday: selected,
    active: false,
    slots: [],
  };
  const label = weekdays.find(([id]) => id === selected)?.[1];
  function setDay(next: typeof day) {
    onChange({
      ...value,
      days: [...value.days.filter((d) => d.weekday !== next.weekday), next],
    });
  }
  const validation = error || scheduleError(value);
  return (
    <section className="weekly-schedule-editor" aria-label="برنامهٔ هفتگی تحویل">
      <h3>تعریف ساعت‌های تحویل</h3>
      <p>
        ساعت‌ها را یک بار تعریف و برای هر روز بازه‌ها و ظرفیت دلخواه را تعیین
        کنید. روز بدون برنامه بسته است؛ تقویم مؤثر فروشگاه و فرصت آماده‌سازی نیز
        کنترل می‌شوند.
      </p>
      {validation && <Alert severity="warning">{validation}</Alert>}
      {!value.windows.length && (
        <Alert severity="info">
          هنوز بازه‌ای تعریف نشده؛ برنامهٔ خالی گزینهٔ تحویل تولید نمی‌کند.
        </Alert>
      )}
      {value.windows.map((window, i) => (
        <div className="shipping-window-definition" key={window.id}>
          <TextField
            size="small"
            label={`شروع بازه ${i + 1}`}
            value={window.start}
            disabled={disabled}
            InputLabelProps={{ shrink: true }}
            placeholder="HH:mm"
            inputProps={{ dir: "ltr", maxLength: 5 }}
            onChange={(e) =>
              onChange({
                ...value,
                windows: value.windows.map((w) =>
                  w.id === window.id
                    ? { ...w, start: asciiDigits(e.target.value) }
                    : w,
                ),
              })
            }
          />
          <TextField
            size="small"
            label={`پایان بازه ${i + 1}`}
            value={window.end}
            disabled={disabled}
            InputLabelProps={{ shrink: true }}
            placeholder="HH:mm"
            inputProps={{ dir: "ltr", maxLength: 5 }}
            onChange={(e) =>
              onChange({
                ...value,
                windows: value.windows.map((w) =>
                  w.id === window.id
                    ? { ...w, end: asciiDigits(e.target.value) }
                    : w,
                ),
              })
            }
          />
          <FormControlLabel
            label={`بازه ${i + 1} فعال`}
            control={
              <Checkbox
                checked={window.active}
                disabled={disabled}
                onChange={(e) =>
                  onChange({
                    ...value,
                    windows: value.windows.map((w) =>
                      w.id === window.id
                        ? { ...w, active: e.target.checked }
                        : w,
                    ),
                  })
                }
              />
            }
          />
          <Button
            disabled={disabled}
            color="error"
            onClick={() =>
              onChange({
                windows: value.windows.filter((w) => w.id !== window.id),
                days: value.days.map((d) => ({
                  ...d,
                  slots: d.slots.filter((s) => s.window_id !== window.id),
                })),
              })
            }
          >
            حذف بازه {i + 1}
          </Button>
        </div>
      ))}
      <Button
        disabled={disabled || value.windows.length >= 200}
        onClick={() =>
          onChange({
            ...value,
            windows: [
              ...value.windows,
              { id: crypto.randomUUID(), start: "", end: "", active: true },
            ],
          })
        }
      >
        افزودن تعریف ساعت
      </Button>
      <h3>انتساب به روزهای هفته</h3>
      <div className="toolbar" role="group" aria-label="روز برنامهٔ تحویل">
        <OptionSelect
          label={"روز برنامهٔ تحویل"}
          multiple={false}
          options={weekdays.map(([id, name]) => ({
            id: String(id),
            name: String(name),
            selected: selected === id,
            disabled: disabled,
            onSelect: () => setSelected(id),
          }))}
        />
      </div>
      <FormControlLabel
        label={`تحویل در ${label} مجاز باشد`}
        control={
          <Checkbox
            disabled={disabled}
            checked={day.active}
            onChange={(e) => setDay({ ...day, active: e.target.checked })}
          />
        }
      />
      <MultiChoiceField
        label={"بازه‌های " + label}
        value={day.slots.map((s) => s.window_id)}
        disabled={disabled}
        options={value.windows.map((w) => ({
          id: w.id,
          name:
            (w.start || "شروع") +
            " تا " +
            (w.end || "پایان") +
            (w.active ? "" : " (غیرفعال)"),
        }))}
        onChange={(ids) =>
          setDay({
            ...day,
            slots: ids.map(
              (id) =>
                day.slots.find((s) => s.window_id === id) ?? {
                  window_id: id,
                  capacity: null,
                },
            ),
          })
        }
      />
      <div className="shipping-slot-capacities">
        {day.slots.map((slot) => {
          const window = value.windows.find((w) => w.id === slot.window_id);
          return (
            <TextField
              key={slot.window_id}
              size="small"
              type="number"
              label={`ظرفیت ${label} ${window?.start} تا ${window?.end}`}
              value={slot.capacity ?? ""}
              disabled={disabled}
              inputProps={{ min: 0, max: 1000000, step: 1 }}
              helperText="هر مرسوله یک واحد؛ صفر یعنی بسته"
              onChange={(e) =>
                setDay({
                  ...day,
                  slots: day.slots.map((s) =>
                    s.window_id === slot.window_id
                      ? {
                          ...s,
                          capacity:
                            e.target.value === ""
                              ? null
                              : Number(e.target.value),
                        }
                      : s,
                  ),
                })
              }
            />
          );
        })}
      </div>
      <details>
        <summary>کپی برنامهٔ {label} به روزهای دیگر</summary>
        <MultiChoiceField
          label="روزهای مقصد کپی"
          value={copyTargets.map(String)}
          options={weekdays
            .filter(([id]) => id !== selected)
            .map(([id, name]) => ({ id: String(id), name }))}
          disabled={disabled}
          onChange={(ids) => setCopyTargets(ids.map(Number))}
        />
        <Button
          disabled={disabled || !copyTargets.length}
          onClick={() => {
            const copies = copyTargets.map((weekday) => ({
              ...day,
              weekday,
              slots: day.slots.map((slot) => ({ ...slot })),
            }));
            onChange({
              ...value,
              days: [
                ...value.days.filter((d) => !copyTargets.includes(d.weekday)),
                ...copies,
              ],
            });
            setCopyTargets([]);
          }}
        >
          کپی به روزهای انتخاب‌شده
        </Button>
      </details>
    </section>
  );
}
