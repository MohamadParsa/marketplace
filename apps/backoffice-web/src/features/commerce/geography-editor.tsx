import { useState } from "react";
import { Button, Checkbox, FormControlLabel, TextField } from "@mui/material";
import type { Province } from "../../../../../packages/commerce-client/types";

const cityPageSize = 50;
const searchKey = (text: string) =>
  text
    .trim()
    .replace(/ي/g, "ی")
    .replace(/ك/g, "ک")
    .replace(/[\s\u200c]+/g, "");

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
    [search, setSearch] = useState(""),
    [expanded, setExpanded] = useState<Record<string, boolean>>({}),
    [limits, setLimits] = useState<Record<string, number>>({});
  const query = searchKey(search);
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
        onChange={(e) => {
          setSearch(e.target.value);
          setLimits({});
        }}
      />
      <fieldset disabled={busy || !allowed}>
        {rows.map((p, i) => {
          const provinceMatches = searchKey(p.name).includes(query);
          const cities = p.cities
            .map((city, index) => ({ city, index }))
            .filter(
              ({ city }) =>
                provinceMatches || searchKey(city.name).includes(query),
            );
          if (query && !provinceMatches && !cities.length) return null;
          const isOpen = !!query || !!expanded[p.id];
          const limit = limits[p.id] ?? cityPageSize;
          return (
            <details
              className="geography-province"
              key={p.id}
              open={isOpen}
              onToggle={(event) => {
                if (!query) {
                  const open = event.currentTarget.open;
                  setExpanded((previous) => ({ ...previous, [p.id]: open }));
                }
              }}
            >
              <summary>
                {p.name || "استان جدید"}{" "}
                <small>
                  ({p.cities.length} شهر){!p.active ? " · غیرفعال" : ""}
                </small>
              </summary>
              {isOpen && (
                <>
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
                          onChange={(e) =>
                            change(i, { active: e.target.checked })
                          }
                        />
                      }
                    />
                  </div>
                  <div className="geography-cities">
                    {cities.slice(0, limit).map(({ city: c, index: j }) => (
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
                                    k === j
                                      ? { ...x, active: e.target.checked }
                                      : x,
                                  ),
                                })
                              }
                            />
                          }
                        />
                      </div>
                    ))}
                  </div>
                  {cities.length > limit && (
                    <Button
                      onClick={() =>
                        setLimits((previous) => ({
                          ...previous,
                          [p.id]: limit + cityPageSize,
                        }))
                      }
                    >
                      نمایش شهرهای بیشتر (
                      {Math.min(limit, cities.length).toLocaleString("fa-IR")}{" "}
                      از {cities.length.toLocaleString("fa-IR")})
                    </Button>
                  )}
                  <Button
                    onClick={() => {
                      change(i, {
                        cities: [
                          ...p.cities,
                          { id: crypto.randomUUID(), name: "", active: true },
                        ],
                      });
                      setSearch("");
                      setExpanded((previous) => ({
                        ...previous,
                        [p.id]: true,
                      }));
                      setLimits((previous) => ({
                        ...previous,
                        [p.id]: p.cities.length + 1,
                      }));
                    }}
                  >
                    افزودن شهر به این استان
                  </Button>
                </>
              )}
            </details>
          );
        })}
        {query &&
          !rows.some(
            (p) =>
              searchKey(p.name).includes(query) ||
              p.cities.some((c) => searchKey(c.name).includes(query)),
          ) && <p>استان یا شهری با این نام پیدا نشد.</p>}
        {!rows.length && (
          <p>
            هنوز استانی ثبت نشده است؛ یک استان و حداقل یک شهر فعال اضافه کنید.
          </p>
        )}
        <div className="commerce-actions">
          <Button
            variant="outlined"
            onClick={() => {
              const id = crypto.randomUUID();
              setRows([...rows, { id, name: "", active: true, cities: [] }]);
              setSearch("");
              setExpanded((previous) => ({ ...previous, [id]: true }));
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
