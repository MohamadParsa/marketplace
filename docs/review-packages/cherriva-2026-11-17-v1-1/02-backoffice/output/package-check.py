import hashlib,json,re,zipfile
from pathlib import Path

root=Path(__file__).resolve().parent.parent
archive=root.parent/'cherriva-backoffice-design-v0.1.zip'
ignored={'.playwright-cli'}
def paths():
    return [p for p in root.rglob('*') if p.is_file() and not any(x in ignored for x in p.relative_to(root).parts)]
broken=[];links=0;ids=[]
for p in paths():
    if p.suffix!='.md':continue
    text=p.read_text(encoding='utf-8-sig')
    ids+=re.findall(r'شناسه(?: سند)?:\s*(CVBO-[A-Z0-9-]+)',text)
    for target in re.findall(r'\]\(([^)]+)\)',text):
        if re.match(r'^(https?:|mailto:|#)',target):continue
        dest=target.split('#')[0]
        if not dest:continue
        links+=1
        if not (p.parent/dest).exists():broken.append({'file':str(p.relative_to(root)),'target':target})
duplicates=sorted({x for x in ids if ids.count(x)>1})
before=(root/'output/repo-status-before.txt').read_text(encoding='utf-8-sig')
after=(root/'output/repo-status-after.txt').read_text(encoding='utf-8-sig')
qa=json.loads((root/'output/verification.json').read_text(encoding='utf-8-sig'))
capture=json.loads((root/'output/capture.json').read_text(encoding='utf-8-sig'))
result={'version':'0.1','date':'2026-11-17','status':'design-in-review','markdown_links_checked':links,'broken_links':broken,'duplicate_document_ids':duplicates,'screen_spec_count':len(list((root/'screens').glob('*.md'))),'png_count':len(list((root/'output/playwright').glob('*.png'))),'browser_checks_passed':qa['passed'],'browser_checks_total':qa['total'],'browser_page_errors':qa['errors'],'capture_page_errors':capture['errors'],'capture_overflow_cases':[x for x in capture['screens'] if x['overflow']],'repo_status_equal':before==after,'security_or_financial_acceptance':False}
if broken or duplicates or not result['repo_status_equal'] or qa['passed']!=qa['total'] or result['capture_overflow_cases'] or qa['errors'] or capture['errors']:
    print(json.dumps(result,ensure_ascii=False));raise SystemExit('Package checks failed')
(root/'bundle-checks.json').write_text(json.dumps(result,ensure_ascii=False,indent=2),encoding='utf-8')
manifest=[]
for p in paths():
    if p.name in ('manifest.json','bundle-checks.json'):continue
    manifest.append({'path':p.relative_to(root).as_posix(),'bytes':p.stat().st_size,'sha256':hashlib.sha256(p.read_bytes()).hexdigest()})
(root/'manifest.json').write_text(json.dumps({'version':'0.1','files':manifest},ensure_ascii=False,indent=2),encoding='utf-8')
with zipfile.ZipFile(archive,'w',zipfile.ZIP_DEFLATED,compresslevel=6) as z:
    for p in paths():z.write(p,p.relative_to(root).as_posix())
with zipfile.ZipFile(archive) as z:
    assert z.testzip() is None
    assert all(not n.startswith(('/', '..')) and ':' not in n for n in z.namelist())
    names=set(z.namelist())
    for entry in manifest:assert hashlib.sha256(z.read(entry['path'])).hexdigest()==entry['sha256']
digest=hashlib.sha256(archive.read_bytes()).hexdigest()
archive.with_suffix('.sha256').write_text(digest+'  '+archive.name+'\n',encoding='ascii')
print(json.dumps({'checks':result,'zip':str(archive),'zip_bytes':archive.stat().st_size,'zip_sha256':digest,'manifest_files':len(manifest)},ensure_ascii=False))
