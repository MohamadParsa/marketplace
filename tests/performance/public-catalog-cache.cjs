// Local-only, read-only comparison. Tokens and response bodies are never written.
const fs = require('node:fs');
const cp = require('node:child_process');
const {performance} = require('node:perf_hooks');
const mode=process.argv[2]||'warm';
const out=process.argv[3];
const origin=process.env.CACHE_BENCH_ORIGIN||'http://127.0.0.1:5197';
if(!/^http:\/\/(127\.0\.0\.1|localhost):\d+$/.test(origin))throw Error('loopback review only');
const paths=['/storefront/home','/storefront/catalog?category=gsp-cat-1','/storefront/products/gsp-product-37354','/catalog/icons/gsp-media-44593'];
async function measure(path,headers={}){
 const records=[];
 for(let i=0;i<30;i++){
  if(mode==='cold'){
   // Only disposable Cherriva public DTOs on the isolated local Redis are removed.
   const command=process.platform==='win32'?'wsl.exe':'redis-cli';
   const prefix=process.platform==='win32'?['--exec','redis-cli']:[];
   const keys=cp.execFileSync(command,[...prefix,'-h','127.0.0.1','-p','56379','--scan','--pattern','cherriva:public:v2:*'],{encoding:'utf8'}).trim().split(/\r?\n/).filter(Boolean);
   if(keys.length)cp.execFileSync(command,[...prefix,'-h','127.0.0.1','-p','56379','unlink',...keys]);
  }
  const start=performance.now();
  try{const r=await fetch(origin+path,{headers,signal:AbortSignal.timeout(20000)});const b=await r.arrayBuffer();records.push({ms:performance.now()-start,status:r.status,bytes:b.byteLength});}
  catch(e){records.push({ms:performance.now()-start,status:0,error:e.name});}
 }
 const ok=records.filter(r=>r.status===200).map(r=>r.ms).sort((a,b)=>a-b);
 return{path,samples:records.length,errors:records.filter(r=>r.status!==200).length,p50:ok[Math.ceil(ok.length*.5)-1],p95:ok[Math.ceil(ok.length*.95)-1],bytes:records[0].bytes,records};
}
(async()=>{const result={mode,concurrency:1,date:'2026-11-23',results:[]};for(const p of paths)result.results.push(await measure(p));if(out)fs.writeFileSync(out,JSON.stringify(result,null,2));console.log(JSON.stringify({...result,results:result.results.map(({records,...r})=>r)}));})().catch(e=>{console.error(e.name);process.exitCode=1});
