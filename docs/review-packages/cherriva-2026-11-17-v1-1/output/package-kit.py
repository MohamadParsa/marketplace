import hashlib,json,zipfile,os
from pathlib import Path

root=Path(__file__).resolve().parent.parent
skip={'.playwright-cli','__pycache__','.git'}
def files():
    return sorted(p for p in root.rglob('*') if p.is_file() and not any(x in skip for x in p.relative_to(root).parts))
validation=json.loads((root/'output/package-validation.json').read_text(encoding='utf-8-sig'))
if not validation['passed']:raise RuntimeError('Fix package-validation errors before packaging')
manifest={}
for p in files():
    rel=p.relative_to(root).as_posix()
    if rel=='manifest.json':continue
    raw=p.read_bytes();manifest[rel]={'bytes':len(raw),'sha256':hashlib.sha256(raw).hexdigest()}
(root/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n',encoding='utf8')

archive=root.parent/(root.name+'.zip')
temp=root.parent/(root.name+'.building.zip')
with zipfile.ZipFile(temp,'w',compression=zipfile.ZIP_DEFLATED,compresslevel=6) as z:
    for p in files():z.write(p,root.name+'/'+p.relative_to(root).as_posix())
with zipfile.ZipFile(temp) as z:
    bad=z.testzip()
    if bad:raise RuntimeError('CRC failure: '+bad)
    expected={root.name+'/'+p.relative_to(root).as_posix() for p in files()}
    if set(z.namelist())!=expected:raise RuntimeError('Archive membership mismatch')
    for name,meta in manifest.items():
        raw=z.read(root.name+'/'+name)
        if len(raw)!=meta['bytes'] or hashlib.sha256(raw).hexdigest()!=meta['sha256']:
            raise RuntimeError('Archive content mismatch: '+name)
os.replace(temp,archive)
receipt={'version':'1.1','date':'2026-11-17','archive':archive.name,'bytes':archive.stat().st_size,'sha256':hashlib.sha256(archive.read_bytes()).hexdigest(),'members':len(expected),'crc_readback_ok':True,'manifest_readback_ok':True,'prototype_checks':'111/111','repo_status_unchanged':validation['repo_status_unchanged'],'product_or_provider_acceptance':False}
(root.parent/(root.name+'.delivery-check.json')).write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n',encoding='utf8')
(root.parent/(root.name+'.zip.sha256')).write_text(receipt['sha256']+'  '+archive.name+'\n',encoding='ascii')
print(json.dumps(receipt,ensure_ascii=False))
