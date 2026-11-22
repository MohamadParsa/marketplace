"""Validate/cache referenced public images without touching either database."""
import argparse,concurrent.futures,hashlib,json,time,urllib.request
from pathlib import Path
from urllib.parse import urlsplit

def fetch(media,cache):
    url=media['url'];u=urlsplit(url)
    if u.scheme!='https' or u.netloc!='cdn.gooshishop.ir' or not u.path.startswith('/images/thumbs/'):
        return {'id':media['id'],'error':'url_not_allowed'}
    path=cache/(hashlib.sha256(url.encode()).hexdigest()+'.source')
    for attempt in range(2):
        try:
            if path.exists():data=path.read_bytes()
            else:
                req=urllib.request.Request(url,headers={'User-Agent':'Cherriva-Catalog-Snapshot/1.0'})
                with urllib.request.urlopen(req,timeout=20) as r:
                    if urlsplit(r.url).netloc!='cdn.gooshishop.ir':raise ValueError('redirect_not_allowed')
                    data=r.read(5*1024*1024+1)
                if not data or len(data)>5*1024*1024:raise ValueError('invalid_size')
            valid=data.startswith(b'\xff\xd8\xff') or data.startswith(b'\x89PNG\r\n\x1a\n') or (data[:4]==b'RIFF' and data[8:12]==b'WEBP')
            if not valid:raise ValueError('unsupported_raster')
            if not path.exists():path.write_bytes(data)
            return {'id':media['id'],'sha256':hashlib.sha256(data).hexdigest(),'bytes':len(data)}
        except Exception as e:
            if attempt==0:time.sleep(.3)
            else:return {'id':media['id'],'error':('http_'+str(e.code)) if hasattr(e,'code') else type(e).__name__}

def run():
    ap=argparse.ArgumentParser();ap.add_argument('--file',type=Path,required=True);ap.add_argument('--cache',type=Path,required=True);ap.add_argument('--report',type=Path,required=True);ap.add_argument('--workers',type=int,default=6);a=ap.parse_args()
    a.cache.mkdir(parents=True,exist_ok=True);seed=json.loads(a.file.read_text());out=[]
    with concurrent.futures.ThreadPoolExecutor(max_workers=max(1,min(a.workers,8))) as pool:
        for i,result in enumerate(pool.map(lambda m:fetch(m,a.cache),seed['media']),1):
            out.append(result)
            if i%500==0:print('Checked',i,'/',len(seed['media']),flush=True)
    a.report.write_text(json.dumps(out,separators=(',',':')))
    print(json.dumps({'checked':len(out),'available':sum('sha256' in x for x in out),'unavailable':sum('error' in x for x in out),'download_bytes':sum(x.get('bytes',0) for x in out)}))
if __name__=='__main__':run()
