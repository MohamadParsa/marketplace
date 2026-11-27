const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const test = require("node:test");
const assert = require("node:assert/strict");
const ts = require("typescript");
const source = fs.readFileSync(
  path.join(__dirname, "../../../packages/date-controls/calendar.ts"),
  "utf8",
);
const output = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.CommonJS,
    target: ts.ScriptTarget.ES2022,
  },
}).outputText;
const helpers = {};
vm.runInNewContext(output, { exports: helpers, Intl, Date });
test("Persian dates use known calendar anchors and Persian digits", () => {
  assert.equal(helpers.persianCapacityDate("2026-11-30"), "1405/09/09");
  assert.equal(helpers.parsePersianCapacityDate("۱۴۰۵/۰۹/۰۹"), "2026-11-30");
  assert.equal(helpers.parsePersianCapacityDate("١٤٠٥/٠٩/٠٩"), "2026-11-30");
  assert.equal(helpers.parsePersianCapacityDate("1403/12/30"), "2025-03-20");
  assert.equal(helpers.parsePersianCapacityDate("1404/01/01"), "2025-03-21");
});
test("Invalid Persian dates never silently roll into the following month", () => {
  for (const value of [
    "1404/12/30",
    "1405/13/01",
    "1405/01/32",
    "1405/00/10",
    "1405/01/00",
    "wrong",
    "1405/09/09extra",
  ])
    assert.equal(helpers.parsePersianCapacityDate(value), "", value);
});
test("Business date follows Tehran across UTC midnight", () => {
  assert.equal(
    helpers.tehranCapacityDate(new Date("2026-11-26T21:00:00Z")),
    "2026-11-27",
  );
  assert.equal(
    helpers.tehranCapacityDate(new Date("2026-11-26T19:00:00Z")),
    "2026-11-26",
  );
});
