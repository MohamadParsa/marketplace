const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const test = require("node:test");
const assert = require("node:assert/strict");
const ts = require("typescript");
const source = fs.readFileSync(
  path.join(__dirname, "../src/display-numbers.ts"),
  "utf8",
);
const output = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.CommonJS,
    target: ts.ScriptTarget.ES2020,
  },
}).outputText;
const helpers = {};
vm.runInNewContext(output, { exports: helpers });
test("Toman prices accept Persian, Arabic and ASCII digits with grouping", () => {
  for (const value of ["۳٬۱۴۹٬۰۰۰", "٣,١٤٩,٠٠٠", "3,149,000"])
    assert.equal(helpers.tomanInputMinor(value), 31490000);
  assert.equal(helpers.tomanInputMinor("۰٫۱"), 1);
  assert.equal(helpers.tomanInputMinor("12.3"), 123);
  assert.equal(helpers.tomanInputMinor("0"), 0);
});
test("Invalid and unsafe prices cannot silently round or become API filters", () => {
  for (const value of [
    "",
    "-1",
    "1.23",
    "1e6",
    "abc",
    "۱۲٫",
    "900719925474099.2",
    "Infinity",
  ])
    assert.equal(helpers.tomanInputMinor(value), null, value);
  assert.equal(
    helpers.tomanInputMinor("900719925474099.1"),
    Number.MAX_SAFE_INTEGER,
  );
});
test("Display preserves decimal precision and warranty raw values remain separate", () => {
  assert.equal(helpers.displayPriceInput("3149000.1"), "۳٬۱۴۹٬۰۰۰٫۱");
  assert.equal(helpers.displayPriceInput("3149000.1", true), "۳۱۴۹۰۰۰٫۱");
  assert.equal(helpers.displayPriceInput(""), "");
  assert.equal(helpers.attributeLabel("گارانتی", "12 ماه"), "۱۲ ماه");
  assert.equal(helpers.attributeLabel("warranty", "٦ ماه"), "۶ ماه");
  assert.equal(helpers.attributeLabel("مدل", "T13 ANC 2"), "T13 ANC 2");
});
