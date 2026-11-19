import re,json,subprocess,hashlib
from pathlib import Path
from urllib.parse import unquote,urlsplit
from PIL import Image

root=Path(__file__).resolve().parent.parent
skip={'.playwright-cli','__pycache__','.git'}
files=[p for p in root.rglob('*') if p.is_file() and not any(x in skip for x in p.relative_to(root).parts)]
report={'version':'1.1','date':'2026-11-17','scope':'standalone kit; no product/backend/provider tests','errors':[]}
links=0;text_count=0
for p in files:
    if p.suffix.lower() not in {'.md','.json','.html','.js','.cjs','.css','.py','.txt','.svg'}:continue
    try:t=p.read_text(encoding='utf-8-sig')
    except UnicodeError as e:report['errors'].append({'kind':'utf8','file':str(p.relative_to(root)),'error':str(e)});continue
    text_count+=1
    if '\ufffd' in t:report['errors'].append({'kind':'replacement-character','file':str(p.relative_to(root))})
    if p.suffix!='.md':continue
    # Code examples can contain intentional placeholder URLs; check actual prose links.
    clean=re.sub(r'```[^\n]*\n[\s\S]*?```','',t)
    for target in re.findall(r'\]\(([^)]+)\)',clean):
        target=target.strip()
        if target.startswith('<') and target.endswith('>'):target=target[1:-1]
        if re.match(r'^(https?:|mailto:|#|data:|app:|codex:)',target):continue
        rel=unquote(target.split('#')[0].split('?')[0])
        if not rel:continue
        links+=1;dest=(p.parent/rel).resolve()
        if not dest.is_relative_to(root):report['errors'].append({'kind':'nonportable-link','file':str(p.relative_to(root)),'target':target})
        elif not dest.exists():report['errors'].append({'kind':'broken-link','file':str(p.relative_to(root)),'target':target})
    if t.count('```')%2:report['errors'].append({'kind':'unbalanced-fence','file':str(p.relative_to(root))})

baseline=json.loads((root/'03-analysis-v1-0/task-index.json').read_text(encoding='utf-8-sig'))['tasks']
current=json.loads((root/'04-current-spec/task-index.json').read_text(encoding='utf-8-sig'))['tasks']
alltasks=baseline+current
ids=[x['id'] for x in alltasks];index={x['id']:x for x in alltasks}
if len(ids)!=len(set(ids)):report['errors'].append({'kind':'duplicate-task-id'})
for group,prefix in [(baseline,'03-analysis-v1-0'),(current,'04-current-spec')]:
    for task in group:
        p=root/prefix/task['file']
        if not p.exists():report['errors'].append({'kind':'missing-task-file','task':task['id']})
        if task.get('status')!='todo':report['errors'].append({'kind':'unexpected-status','task':task['id']})
        for dep in task['dependencies']:
            if dep not in index:report['errors'].append({'kind':'unknown-dependency','task':task['id'],'dependency':dep})
            elif index[dep]['wave']>task['wave']:report['errors'].append({'kind':'wave-order','task':task['id'],'dependency':dep})
        for target in task.get('amends',[]):
            if target not in {x['id'] for x in baseline}:report['errors'].append({'kind':'unknown-amendment','task':task['id'],'amends':target})
visiting=set();visited=set()
def walk(i):
    if i in visiting:report['errors'].append({'kind':'dependency-cycle','task':i});return
    if i in visited:return
    visiting.add(i)
    for j in index[i]['dependencies']:
        if j in index:walk(j)
    visiting.remove(i);visited.add(i)
for i in index:walk(i)

js_files=[p for p in files if p.suffix in {'.js','.cjs'}]
for p in js_files:
    r=subprocess.run(['node','--check',str(p)],capture_output=True,text=True,encoding='utf8')
    if r.returncode:report['errors'].append({'kind':'js-syntax','file':str(p.relative_to(root)),'detail':r.stderr[:300]})

asset_count=0
for p in files:
    if p.suffix not in {'.html','.css'}:continue
    t=p.read_text(encoding='utf-8-sig')
    refs=re.findall(r'(?:src|href)=["\']([^"\']+)["\']',t) if p.suffix=='.html' else re.findall(r'url\(["\']?([^"\')]+)',t)
    for ref in refs:
        if re.match(r'^(https?:|data:|#|mailto:)',ref):continue
        rel=unquote(ref.split('?')[0].split('#')[0])
        if not rel:continue
        # Preview code may build paths inside templates. Static asset links only.
        if '${' in rel:continue
        dest=(p.parent/rel).resolve();asset_count+=1
        if not dest.is_relative_to(root) or not dest.exists():report['errors'].append({'kind':'missing-static-asset','file':str(p.relative_to(root)),'target':ref})

images=[p for p in files if p.suffix.lower() in {'.png','.jpg','.jpeg','.webp'}]
for p in images:
    try:
        with Image.open(p) as img:img.verify()
    except Exception as e:report['errors'].append({'kind':'image-read','file':str(p.relative_to(root)),'detail':str(e)})
before=(root/'output/repo-status-before.txt').read_text(encoding='utf-8-sig').splitlines()
after=(root/'output/repo-status-after.txt').read_text(encoding='utf-8-sig').splitlines()
same=before==after
if not same:report['errors'].append({'kind':'repo-status-difference','before':len(before),'after':len(after)})
report.update({'markdown_local_links':links,'utf8_text_files':text_count,'static_html_css_references':asset_count,'images_readable':len(images),'js_syntax_files':len(js_files),'baseline_tasks':len(baseline),'amendments':len(current),'task_graph_acyclic':not any(e['kind']=='dependency-cycle' for e in report['errors']),'repo_status_unchanged':same,'packaged_file_count_before_manifest':len(files),'passed':not report['errors']})
(root/'output/package-validation.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n',encoding='utf8')
print(json.dumps(report,ensure_ascii=False))
