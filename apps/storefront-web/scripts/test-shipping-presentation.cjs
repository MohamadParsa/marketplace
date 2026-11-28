const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),test=require('node:test'),assert=require('node:assert/strict'),ts=require('typescript');
const helpers={};
vm.runInNewContext(ts.transpileModule(fs.readFileSync(path.join(__dirname,'../../../packages/commerce-client/shipping-presentation.ts'),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2020}}).outputText,{exports:helpers});
test('Persian delivery date explicitly orders weekday day month year',()=>{
 assert.equal(helpers.shippingDate('2026-12-07'),'دوشنبه ۱۶ آذر ۱۴۰۵');
 assert.equal(helpers.shippingDate('2025-03-21'),'جمعه ۱ فروردین ۱۴۰۴');
 assert.doesNotMatch(helpers.shippingDate('2026-12-07'),/[,،\u200e\u200f\u061c]/);
});
test('slot and range promises share the same readable date order',()=>{
 const slot={capacity_day:'2026-12-07',slot:{start_at:'2026-12-07T10:30:00Z',end_at:'2026-12-07T12:30:00Z'}};
 assert.equal(helpers.deliveryPromise(slot,'2026-12-07','2026-12-07'),'تحویل دوشنبه ۱۶ آذر ۱۴۰۵، ساعت ۱۴:۰۰ تا ۱۶:۰۰');
 assert.equal(helpers.deliveryPromise(undefined,'2026-12-07','2026-12-08'),'تحویل حدودی از دوشنبه ۱۶ آذر ۱۴۰۵ تا سه‌شنبه ۱۷ آذر ۱۴۰۵');
});
test('missing and invalid calendar dates do not invent a delivery day',()=>{
 for(const date of ['', '2026-02-30','2026-99-01','invalid']) assert.equal(helpers.shippingDate(date),'—');
});
