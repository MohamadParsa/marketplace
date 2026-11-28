const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { createRequire } = require("node:module");
const { spawnSync } = require("node:child_process");
const test = require("node:test");
const assert = require("node:assert/strict");
const ts = createRequire(path.join(process.cwd(), "package.json"))(
  "typescript",
);
const source = fs.readFileSync(path.join(__dirname, "calendar.ts"), "utf8");
const compiled = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.CommonJS,
    target: ts.ScriptTarget.ES2022,
  },
}).outputText;
const helpers = {};
vm.runInNewContext(compiled, { exports: helpers, Intl, Date });

test("Leap Esfand and Nowruz have exact known Gregorian anchors", () => {
  assert.equal(helpers.parsePersianCapacityDate("۱۴۰۳/۱۲/۳۰"), "2025-03-20");
  assert.equal(helpers.parsePersianCapacityDate("1404/01/01"), "2025-03-21");
  assert.equal(helpers.parsePersianCapacityDate("١٤٠٤/١/١"), "2025-03-21");
  for (const invalid of [
    "1404/12/30",
    "1403/12/31",
    "1404/07/31",
    "1404/00/01",
    "1404/01/00",
    "1404/13/01",
    "1404/1/1junk",
  ])
    assert.equal(helpers.parsePersianCapacityDate(invalid), "", invalid);
  for (const invalid of [
    "2025-02-30",
    "2026-11-31",
    "2026-13-01",
    "not-a-date",
  ])
    assert.equal(helpers.persianCapacityDate(invalid), "", invalid);
});
test("Every sampled date around year boundaries round-trips without local timezone arithmetic", () => {
  for (let year = 2020; year <= 2030; year++)
    for (let delta = -35; delta <= 35; delta++) {
      const iso = new Date(Date.UTC(year, 2, 21 + delta, 12))
        .toISOString()
        .slice(0, 10);
      assert.equal(
        helpers.parsePersianCapacityDate(helpers.persianCapacityDate(iso)),
        iso,
      );
    }
});
test("Datetime is Tehran wall time including historical Iran summer time", () => {
  assert.equal(
    helpers.parsePersianDateTime("۱۴۰۴/۰۱/۰۱ ۰۰:۰۰"),
    "2025-03-20T20:30:00.000Z",
  );
  assert.equal(
    helpers.parsePersianDateTime("1401/04/10 10:20"),
    "2022-07-01T05:50:00.000Z",
  );
  assert.equal(
    helpers.persianDateTime("2025-03-20T20:30:00Z"),
    "1404/01/01 00:00",
  );
  for (const instant of [
    "2026-11-27T21:00:00Z",
    "2025-03-20T20:30:12Z",
    "2022-07-01T05:50:00Z",
  ])
    assert.equal(
      helpers.parsePersianDateTime(helpers.persianDateTime(instant)),
      new Date(instant).toISOString(),
    );
  for (const text of [
    "1404/12/30 10:00",
    "1404/01/01 24:00",
    "1404/01/01 01:60",
    "1404/01/01 12:00:60",
    "1404/01/01",
    "",
  ])
    assert.equal(helpers.parsePersianDateTime(text), "", text);
  assert.equal(helpers.validISOInstant("2026-02-30T00:00:00Z"), false);
  assert.equal(helpers.persianDateTime(""), "");
});
test("Date-only and datetime are independent of device timezone", () => {
  const script = `const vm=require('node:vm');const h={};vm.runInNewContext(${JSON.stringify(compiled)},{exports:h,Intl,Date});process.stdout.write(JSON.stringify([h.parsePersianCapacityDate('1404/01/01'),h.parsePersianDateTime('1404/01/01 00:00'),h.tehranCapacityDate(new Date('2026-11-27T20:30:00Z'))]));`;
  for (const zone of [
    "UTC",
    "Asia/Tehran",
    "America/Los_Angeles",
    "Pacific/Kiritimati",
  ]) {
    const result = spawnSync(process.execPath, ["-e", script], {
      encoding: "utf8",
      env: { ...process.env, TZ: zone },
    });
    assert.equal(result.status, 0, result.stderr);
    assert.deepEqual(
      JSON.parse(result.stdout),
      ["2025-03-21", "2025-03-20T20:30:00.000Z", "2026-11-28"],
      zone,
    );
  }
});
