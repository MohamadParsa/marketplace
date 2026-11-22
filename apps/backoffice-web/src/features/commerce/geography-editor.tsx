import { useState } from "react";
import { Button, Checkbox, FormControlLabel, TextField } from "@mui/material";
import type { Province } from "../../../../../packages/commerce-client/types";
export function GeographyEditor({
  value,
  busy,
  allowed,
  save,
}: {
  value: Province[];
  busy: boolean;
  allowed: boolean;
  save: (geography: Province[]) => Promise<void>;
}) {
  const [rows, setRows] = useState(() => structuredClone(value)),
    [search, setSearch] = useState("");
  const change = (index: number, patch: Partial<Province>) =>
    setRows(rows.map((p, i) => (i === index ? { ...p, ...patch } : p)));
  return (
    <section
      className="panel commerce-settings"
      aria-label="مدیریت استان و شهر"
    >
      <h2>استان‌ها و شهرها</h2>
      <p>
        شهرها زیرمجموعه استان هستند. برای توقف انتخاب یک مقصد آن را غیرفعال
        کنید؛ آدرس و تاریخچه سفارش قبلی حذف نمی‌شوند.
      </p>
      <TextField
        size="small"
        label="جست‌وجوی استان یا شهر"
        value={search}
        onChange={(e) => setSearch(e.target.value)}
      />
      <fieldset disabled={busy || !allowed}>
        {rows.map((p, i) => (
          <details
            className="geography-province"
            key={p.id}
            open={search ? true : undefined}
            hidden={
              !!search &&
              ![p.name, ...p.cities.map((c) => c.name)].some((n) =>
                n.includes(search),
              )
            }
          >
            <summary>
              {p.name || "استان جدید"}{" "}
              <small>
                ({p.cities.length} شهر){!p.active ? " · غیرفعال" : ""}
              </small>
            </summary>
            <div className="commerce-grid">
              <TextField
                label="نام استان"
                size="small"
                required
                value={p.name}
                onChange={(e) => change(i, { name: e.target.value })}
                inputProps={{ maxLength: 100 }}
              />
              <FormControlLabel
                label="استان فعال"
                control={
                  <Checkbox
                    checked={p.active}
                    onChange={(e) => change(i, { active: e.target.checked })}
                  />
                }
              />
            </div>
            <div className="geography-cities">
              {p.cities.map((c, j) => (
                <div className="geography-city" key={c.id}>
                  <TextField
                    label="نام شهر"
                    size="small"
                    required
                    value={c.name}
                    onChange={(e) =>
                      change(i, {
                        cities: p.cities.map((x, k) =>
                          k === j ? { ...x, name: e.target.value } : x,
                        ),
                      })
                    }
                    inputProps={{ maxLength: 100 }}
                  />
                  <FormControlLabel
                    label="شهر فعال"
                    control={
                      <Checkbox
                        checked={c.active}
                        onChange={(e) =>
                          change(i, {
                            cities: p.cities.map((x, k) =>
                              k === j ? { ...x, active: e.target.checked } : x,
                            ),
                          })
                        }
                      />
                    }
                  />
                </div>
              ))}
            </div>
            <Button
              onClick={() =>
                change(i, {
                  cities: [
                    ...p.cities,
                    { id: crypto.randomUUID(), name: "", active: true },
                  ],
                })
              }
            >
              افزودن شهر به این استان
            </Button>
          </details>
        ))}
        {!rows.length && (
          <p>
            هنوز استانی ثبت نشده است؛ یک استان و حداقل یک شهر فعال اضافه کنید.
          </p>
        )}
        <div className="commerce-actions">
          <Button
            variant="outlined"
            onClick={() => {
              setRows([
                ...rows,
                { id: crypto.randomUUID(), name: "", active: true, cities: [] },
              ]);
              setSearch("");
            }}
          >
            افزودن استان
          </Button>
          <Button
            variant="contained"
            disabled={
              busy ||
              !allowed ||
              rows.some(
                (p) => !p.name.trim() || p.cities.some((c) => !c.name.trim()),
              )
            }
            onClick={() => void save(rows)}
          >
            ذخیره استان‌ها و شهرها
          </Button>
        </div>
      </fieldset>
    </section>
  );
}
