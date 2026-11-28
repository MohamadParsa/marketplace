const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),test=require('node:test'),assert=require('node:assert/strict'),ts=require('typescript');
const helpers={};
vm.runInNewContext(ts.transpileModule(fs.readFileSync(path.join(__dirname,'../../../packages/commerce-client/order-progress.ts'),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2020}}).outputText,{exports:helpers});
const created='2026-11-28T07:00:00Z',at='2026-11-28T08:00:00Z';
const states=steps=>steps.map(s=>s.state).join(',');
test('paid placement does not imply processing or completion',()=>{
 assert.equal(states(helpers.purchaseProgress('placed',created,[{action:'order.paid',order_status:'placed',at}])),'current,pending,pending');
});
test('processing uses the first actual event and preserves pending completion',()=>{
 const steps=helpers.purchaseProgress('processing',created,[{action:'fulfillment.start',order_status:'processing',at},{action:'fulfillment.correct',order_status:'processing',at:'2026-11-29T08:00:00Z'}]);
 assert.equal(states(steps),'done,current,pending');assert.equal(steps[1].at,at);assert.equal(steps[2].at,undefined);
});
test('cancelled order only marks evidenced stages, not a successful end',()=>{
 const steps=helpers.purchaseProgress('cancelled',created,[{action:'fulfillment.start',order_status:'processing',at}]);
 assert.equal(states(steps),'done,done,pending');
});
test('two shipments can independently be shipped and awaiting supply',()=>{
 assert.equal(states(helpers.shipmentProgress('shipped',[])),'done,done,done,current,pending');
 assert.equal(states(helpers.shipmentProgress('awaiting-supply',[])),'current,pending,pending,pending,pending');
});
test('hold and cancellation never fabricate subsequent fulfillment or timestamps',()=>{
 for(const status of ['on-hold','cancelled']){
  const steps=helpers.shipmentProgress(status,[{action:'fulfillment.start',at},{action:'fulfillment.supply',at:'2026-11-28T09:00:00Z'}]);
  assert.equal(states(steps),'done,done,pending,pending,pending');assert.equal(steps[4].at,undefined);
 }
});
test('legacy and unknown states keep unknown dates, payment wait keeps stages pending',()=>{
 assert.equal(states(helpers.shipmentProgress('awaiting-payment',[])),'pending,pending,pending,pending,pending');
 assert.equal(states(helpers.shipmentProgress('legacy-status',[])),'pending,pending,pending,pending,pending');
 assert.equal(helpers.purchaseProgress('placed','bad',[])[0].at,undefined);
 assert.equal(helpers.shipmentProgress('supplied',[{action:'fulfillment.supply',at:'bad'}])[1].at,undefined);
});
