const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),test=require('node:test'),assert=require('node:assert/strict');
const ts=require('node:module').createRequire(path.join(process.cwd(),'package.json'))('typescript');
const helpers={};
vm.runInNewContext(ts.transpileModule(fs.readFileSync(path.join(__dirname,'transfer-reference.ts'),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2020}}).outputText,{exports:helpers});
test('receipt reference normalizes digits and preserves leading zeros',()=>{
 assert.equal(helpers.normalizeTransferReference(' ۰۰۰۱۲۳۴۵۶ '),'000123456');
 assert.equal(helpers.normalizeTransferReference('ab-١٢٣'),'AB-123');
 for(const value of ['۰۰۰۱۲۳۴۵۶','AB-123','123'])assert.equal(helpers.transferReferenceError(value),'');
});
test('names and unsupported receipt formats receive field guidance',()=>{
 for(const value of ['','مدیر سایت','12','AB 123','ß12','A'.repeat(81)])assert.notEqual(helpers.transferReferenceError(value),'');
});
