/* Run from repository root: node tests/unit/grid-xlsx.cjs */
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('../../apps/backoffice-web/node_modules/typescript');
const source = fs.readFileSync(path.join(__dirname, '../../apps/backoffice-web/src/shared/xlsx.ts'), 'utf8');
const compiled = ts.transpileModule(source, {compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText;
const target = {exports:{}};
new Function('exports','module',compiled)(target.exports,target);
const {xlsxBlob} = target.exports;
(async()=>{
 const checks=[];
 const check=(name)=>checks.push({id:name,pass:true});
 const blob=await xlsxBlob(['نام','مقدار'],[['<آزمون & "نمونه">','=SUM(1,1)'],['نام\u0000مجاز','@external']]);
 const raw=Buffer.from(await blob.arrayBuffer());
 assert.equal(raw.readUInt32LE(0),0x04034b50);
 const text=raw.toString('utf8');
 assert.ok(text.includes('&lt;آزمون &amp; &quot;نمونه&quot;&gt;'));
 assert.ok(text.includes('=SUM(1,1)'));assert.ok(!text.includes('<f>'));assert.ok(text.includes('t="inlineStr"'));
 assert.ok(text.includes('ناممجاز'));check('xml-escaped-inline-text-without-formula');
 assert.equal(blob.type,'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet');check('real-ooxml-zip-mime');
 for(const [name,headers,rows] of [
  ['row-limit',['x'],Array.from({length:10001},()=>['x'])],
  ['column-limit',Array(101).fill('x'),[]],
  ['shape-validation',['x'],[['x','y']]],
  ['cell-limit',['x'],[['x'.repeat(32768)]]],
  ['size-limit',Array(100).fill('x'),Array.from({length:2},()=>Array(100).fill('x'.repeat(30000)))]
 ]){await assert.rejects(()=>xlsxBlob(headers,rows));check(name)}
 const maximum=await xlsxBlob(["x"],Array.from({length:10000},(_,i)=>[String(i)]));assert.ok(maximum.size>10000);check("ten-thousand-row-boundary");
 const abort=new AbortController();
 const pending=xlsxBlob(['x'],Array.from({length:10000},()=>['آزمون']),abort.signal);
 setTimeout(()=>abort.abort(),1);await assert.rejects(()=>pending,/لغو/);check('cancel-during-generation');
 console.log(JSON.stringify({checks,passed:checks.length}));
})().catch(e=>{console.error(e);process.exitCode=1;});
