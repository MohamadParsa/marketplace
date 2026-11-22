"""Build a public-catalog seed from SELECT-only extraction files. No database credentials."""
import argparse
import collections
import datetime as dt
import hashlib
import html
from html.parser import HTMLParser
import json
import math
from pathlib import Path
import re
from urllib.parse import quote

class Text(HTMLParser):
    def __init__(self):
        super().__init__(); self.parts=[]; self.hidden=0
    def handle_starttag(self,tag,attrs):
        if tag in ('script','style'): self.hidden+=1
        if tag in ('p','br','li','div','tr','h2','h3'): self.parts.append('\n')
    def handle_endtag(self,tag):
        if tag in ('script','style'): self.hidden=max(0,self.hidden-1)
        if tag in ('p','li','div','tr'): self.parts.append('\n')
    def handle_data(self,data):
        if not self.hidden: self.parts.append(data)
def text(value):
    p=Text(); p.feed(value or '');return '\n'.join(' '.join(s.split()) for s in ''.join(p.parts).splitlines() if s.strip())
def clean(s): return (s or '').replace('\x00','').strip()
def identity(kind,value): return f'gsp-{kind}-{value}'
def money(value): return int(math.floor(float(value or 0)+0.5))
def utc(value):
    return (value[:26]+'Z') if value else None

def build(raw,output,media_checks=None):
    read=lambda name:json.loads((raw/(name+'.json')).read_text(encoding='utf-8-sig'))
    config={x['Name']:x['Value'] for x in read('public-config')};currencies={str(x['Id']):x for x in read('currencies')}
    currency=currencies.get(config.get('currencysettings.primarystorecurrencyid',''),{})
    if currency.get('CurrencyCode')!='IRT':raise ValueError('Source currency must be verified IRT; add explicit conversion for a different source')
    rawcats=read('categories');rawbrands=read('brands');rawproducts=read('products')
    cats={x['Id']:x for x in rawcats};brands={x['Id']:x for x in rawbrands}
    flags={x['Id']:x for x in read('product-public-flags')};
    for p in rawproducts:p.update({k:v for k,v in flags.get(p['Id'],{}).items() if k!='Id'})
    inventory_maps=read('product-inventory-maps')
    products={x['Id']:x for x in rawproducts};values={x['Id']:x for x in read('inventory-values')}
    pictures={x['Id']:x for x in read('pictures')}
    spacs=read('selected-spacs');spacmap={x['Id']:x for x in spacs}
    bypac=collections.defaultdict(list)
    for s in spacs:bypac[s['ProductAttributeCombinationId']].append(s)
    pacs=[x for x in read('eligible-pacs') if x['ProductId'] in products and x['StockAvailabilityTypeId']==20 and x['StockQuantity']>0 and bypac[x['Id']]]
    pacproducts=collections.defaultdict(list)
    for x in pacs:pacproducts[x['ProductId']].append(x)
    photos=collections.defaultdict(list)
    for x in read('product-pictures'):photos[x['ProductId']].append(x['PictureId'])
    def published(i,seen=()):
        if i in seen or i not in cats:return False
        c=cats[i];return c['Published'] and not c['Deleted'] and (not c['ParentCategoryId'] or published(c['ParentCategoryId'],seen+(i,)))
    now=dt.datetime.now(dt.timezone.utc)
    def available(p):
        for field,started in [('AvailableStartDateTimeUtc',True),('AvailableEndDateTimeUtc',False)]:
            if p.get(field):
                d=dt.datetime.fromisoformat(p[field]).replace(tzinfo=dt.timezone.utc)
                if (started and d>now) or (not started and d<now):return False
        return not p.get('AvailableForPreOrder') and not p.get('DisableBuyButton') and not p.get('CallForPrice')
    eligible=[p for p in rawproducts if available(p)]
    eligible=[p for p in eligible if p['Id'] in pacproducts and photos[p['Id']] and published(p['MainCategoryId']) and brands.get(p['MainManufacturerId'],{}).get('Published')]
    selected=set()
    for field in ['MainManufacturerId','MainCategoryId']:
        groups=collections.defaultdict(list)
        for p in eligible:groups[p[field]].append(p)
        for rows in groups.values():
            rows.sort(key=lambda p:(p['UpdatedOnUtc'],p['Id']),reverse=True)
            selected.update(p['Id'] for p in rows[:20])
    chosen=[products[i] for i in sorted(selected)]
    cpacs=[x for x in pacs if x['ProductId'] in selected]
    specs=[x for x in read('specs') if x['ProductId'] in selected]
    specsbyproduct=collections.defaultdict(list)
    for s in specs:specsbyproduct[s['ProductId']].append(s)
    issues=[]
    def bounded(value,n,where,plain=False):
        value=text(value) if plain else clean(value)
        if len(value)>n:issues.append({'field':where,'action':'truncated_for_current_model','original_length':len(value),'limit':n})
        return value[:n]
    def seo(x,where):
        return {k:bounded(x.get(src),n,where+'.'+k,True) for k,src,n in [('meta_title','MetaTitle',160),('meta_keywords','MetaKeywords',500),('meta_description','MetaDescription',400)]}
    usedmedia=set()
    def image(i):
        if not i or i not in pictures:return ''
        usedmedia.add(i);return identity('media',i)
    def option(i):
        v=values.get(i);return bounded(v['Name'],80,'inventory_value.'+str(i),True) if v else 'نامشخص در مبدأ'
    axisfields={'color':'ColorProductInventoryAttributeValueId','guarantee':'GuaranteeProductInventoryAttributeValueId','weight':'WeightProductInventoryAttributeValueId'}
    axisnames={'color':'رنگ','guarantee':'گارانتی','weight':'وزن طلا'}
    axes={k:{'id':identity('axis',k),'code':'gsp-axis-'+k,'name':name,'kind':'enum','unit':'','options':sorted({option(p[axisfields[k]]) for p in cpacs}),'version':1} for k,name in axisnames.items()}
    activecats={p['MainCategoryId'] for p in chosen}
    cataxes={i:(['weight','guarantee'] if any(p['MainCategoryId']==i and p['PricingTypeId']==1 for p in chosen) else ['color','guarantee']) for i in activecats}
    # All public specification values are retained in source_records; visible fields are mapped to current scalar attributes.
    defs={};catbindings=collections.defaultdict(dict);pattrs={}
    for p in chosen:
        grouped=collections.defaultdict(list)
        for s in specsbyproduct[p['Id']]:
            if s['ShowOnProductPage']:grouped[s['AttributeId']].append(s)
        attrs=[]
        ranked=[]
        for sid,rows in sorted(grouped.items(),key=lambda kv:min(r['DisplayOrder'] for r in kv[1])):
            s=rows[0];aid=identity('spec',sid)
            defs[aid]={'id':aid,'code':f'gsp-spec-{sid}','name':bounded(s['AttributeName'],80,'spec.'+str(sid),True),'kind':'text','unit':bounded(s['Unit'],32,'spec-unit.'+str(sid),True),'options':[],'version':1}
            vals=[]
            for r in rows:
                v=text(r['CustomValue'] if r['AttributeTypeId']!=0 else r['OptionName'])
                if v and v not in vals:vals.append(v)
            val=bounded('، '.join(vals),160,f'product.{p["Id"]}.spec.{sid}')
            if not val:continue
            attrs.append({'attribute_id':aid,'value':val})
            rank=(0 if any(r['IsPin'] for r in rows) else 1, min(r['DisplayOrder'] for r in rows),sid)
            ranked.append((rank,aid))
            cb=catbindings[p['MainCategoryId']]
            if aid not in cb:cb[aid]={'attribute_id':aid,'axis':False,'card':False,'filter':False,'position':100+len(cb)}
            cb[aid]['filter'] |= any(r['AllowFiltering'] for r in rows)
        for _,aid in sorted(ranked)[:3]:catbindings[p['MainCategoryId']][aid]['card']=True
        pattrs[p['Id']]=attrs[:100]
    # Stable shared definitions with disambiguated names when the old catalog has duplicate specification labels.
    names=collections.Counter(d['name'].casefold() for d in list(defs.values())+list(axes.values()))
    for d in defs.values():
        if names[d['name'].casefold()]>1:d['name']=(d['name'][:62]+' — '+d['id'].split('-')[-1])[:80]
    catout=[]
    for c in sorted(rawcats,key=lambda x:x['Id']):
        bindings=[{'attribute_id':axes[k]['id'],'axis':True,'card':False,'filter':True,'position':j} for j,k in enumerate(cataxes.get(c['Id'],[]))]+list(catbindings[c['Id']].values())
        if len(bindings)>100:raise ValueError('category binding limit '+str(c['Id']))
        catout.append({'id':identity('cat',c['Id']),'name':bounded(c['Name'],160,f'category.{c["Id"]}.name',True),'english_name':bounded(c['NameEn'],160,f'category.{c["Id"]}.english_name',True),'description':bounded(c['Description'],4000,f'category.{c["Id"]}.description',True),'published':c['Published'],'parent_id':identity('cat',c['ParentCategoryId']) if c['ParentCategoryId'] in cats else '', 'bindings':bindings,'archived':False,'icon_id':image(c['PictureId']),'version':1,'seo':seo(c,f'category.{c["Id"]}')})
    brandout=[{'id':identity('brand',b['Id']),'name':clean(b['Name']),'active':b['Published'],'icon_id':image(b['PictureId']),'version':1} for b in sorted(rawbrands,key=lambda b:b['Id'])]
    productout=[];variantout=[];offers={};variantmap={};sourcepacmap={};usedsuppliers=set()
    now=dt.datetime.now(dt.timezone.utc)
    def regular(x,price):
        for key,testdate in [('OldPrice',True)]:
            old=money(x.get(key))
            if old<=price:continue
            active=True
            for field,sgn in [('OldPriceStartDateUtc',1),('OldPriceEndDateUtc',-1)]:
                if x.get(field):
                    t=dt.datetime.fromisoformat(x[field]).replace(tzinfo=dt.timezone.utc)
                    active=active and (t<=now if sgn==1 else t>=now)
            if active:return old
        return 0
    for p in chosen:
        pid=identity('product',p['Id']);photoids=list(dict.fromkeys(i for i in photos[p['Id']] if i in pictures))
        if not photoids:raise ValueError('Missing picture metadata '+str(p['Id']))
        def points(field):
            lines=[s.strip(' •-\t') for s in text(p[field]).splitlines() if s.strip(' •-\t')]
            if len(lines)>20:issues.append({'field':f'product.{p["Id"]}.{field}','action':'first_20_points','original_count':len(lines)})
            return [bounded(s,300,f'product.{p["Id"]}.{field}') for s in lines[:20]]
        full=clean(p['FullDescription'])
        if len(full)>20000:
            # Avoid cutting an HTML tag: preserve all original HTML in source_records, use a safe bounded text projection.
            full='<p>'+html.escape(text(full)[:19000]).replace('\n','</p><p>')+'</p>'
            if len(full)>20000:full='<p>'+html.escape(text(p['FullDescription'])[:10000])+'</p>'
            issues.append({'field':f'product.{p["Id"]}.full_description','action':'bounded_plain_paragraph_projection','original_length':len(p['FullDescription'])})
        shipping={'weight_unit':'g','dimension_unit':'mm'}
        for src,dst,m in [('Weight','weight',1),('Length','length',10),('Width','width',10),('Height','height',10)]:
            if p[src] and float(p[src])>0:shipping[dst]=round(float(p[src])*m,3)
        content={'short_description':bounded(p['ShortDescription'],2000,f'product.{p["Id"]}.short_description',True),'full_description':full,'advantages':points('Advantage'),'disadvantages':points('Disadvantage'),'reviews_enabled':p['AllowCustomerReviews'],'qa_enabled':False,'shipping':shipping,'seo':seo(p,f'product.{p["Id"]}'),'noindex':p['NoIndex'],'nofollow':p['NoFollow']}
        productout.append({'id':pid,'name':bounded(p['Name'],160,f'product.{p["Id"]}.name',True),'english_name':bounded(p['NameEn'],160,f'product.{p["Id"]}.english_name',True),'description':bounded(p['ShortDescription'],4000,f'product.{p["Id"]}.description',True),'category_id':identity('cat',p['MainCategoryId']),'brand_id':identity('brand',p['MainManufacturerId']),'model':'shared','owner_store_id':'','status':'published','version':1,'variant_version':1,'created_at':utc(p['CreatedOnUtc']),'attributes':pattrs[p['Id']],'image_id':image(photoids[0]),'gallery_ids':[image(i) for i in photoids[1:9]],'content':content})
        for pac in sorted(pacproducts[p['Id']],key=lambda x:x['Id']):
            av=[{'attribute_id':axes[k]['id'],'value':option(pac[axisfields[k]])} for k in cataxes[p['MainCategoryId']]]
            selection=json.dumps(av,sort_keys=True,ensure_ascii=False)
            vid=pid+'-'+hashlib.sha256(selection.encode()).hexdigest()[:12]
            if vid not in variantmap:
                sku=pac['Sku'] or ''
                if not re.fullmatch(r'[A-Za-z0-9][A-Za-z0-9._-]{0,63}',sku) or sku.casefold() in {v['sku'].casefold() for v in variantout}:sku=f'GSP-PAC-{pac["Id"]}'
                v={'id':identity('variant',vid.removeprefix('gsp-product-')),'product_id':pid,'sku':sku,'values':av,'image_id':image(pac['PictureId']),'version':1};variantmap[vid]=v;variantout.append(v)
            v=variantmap[vid];sourcepacmap[str(pac['Id'])]=v['id']
            for s in bypac[pac['Id']]:
                usedsuppliers.add(s['SupplierId']);sid=identity('store',s['SupplierId'])
                isbuybox=s['Id']==pac['SupplierProductAttributeCombinationId']
                price=money(pac['FinalPrice'] if isbuybox else s['OverriddenPrice'])
                reg=regular(pac if isbuybox else s,price)
                if not reg and isbuybox and money(pac['OverriddenPrice'])>price:reg=money(pac['OverriddenPrice'])
                key=(v['id'],sid)
                o={'id':identity('offer',s['Id']),'product_id':pid,'variant_id':v['id'],'store_id':sid,'price':price,'regular_price':reg,'stock':s['StockQuantity'],'reserved':0,'version':1}
                if key not in offers or (price,o['id'])<(offers[key]['price'],offers[key]['id']):offers[key]=o
    pacidset=sourcepacmap_as_int(sourcepacmap);usedvalues={p.get(f) for p in cpacs for f in axisfields.values()}|{x['ProductInventoryAttributeValueId'] for x in inventory_maps if x['ProductId'] in selected};leads={x['SupplierId']:x for x in read('supply-leads')};suppliernames={s['SupplierId']:s['SupplierName'] for s in spacs};storeout=[]
    for i in sorted(usedsuppliers):
        l=leads.get(i,{});n=l.get('SupplyDelay') or 0;period=l.get('SupplyDelayPeriodId');days=math.ceil(n/24) if period==0 else math.ceil(n/1440) if period==3 else n
        deadlines=[s['StoreDeliveryDeadlineDays'] or 0 for s in spacs if s['SupplierId']==i and s['ProductAttributeCombinationId'] in pacidset]
        storeout.append({'id':identity('store',i),'name':clean(suppliernames[i]),'owner':'','city':'','active':True,'model':'shared','lead_days':max([days]+deadlines),'content_policy':'review_required','content_policy_version':1})
    media=[]
    for i in sorted(usedmedia):
        p=pictures[i];ext={'image/jpeg':'jpeg','image/png':'png','image/webp':'webp','image/gif':'gif','image/svg+xml':'svg'}.get(p['MimeType'],'jpeg')
        fn=f'{i:07d}'+('_'+p['SeoFilename'] if p['SeoFilename'] else '')
        # Exact full-size public source asset; the importer checks and decodes, never trusts the extension.
        url='https://cdn.gooshishop.ir/images/thumbs/n/Entity/'+quote(fn+'.'+ext,safe='')
        media.append({'id':identity('media',i),'source_picture_id':i,'url':url,'mime_type':p['MimeType'],'alt':p['AltAttribute'] or '', 'title':p['TitleAttribute'] or ''})
    def coverage(rows,field,population):
        a=collections.Counter(p[field] for p in eligible);b=collections.Counter(p[field] for p in chosen)
        return [{'source_id':r['Id'],'name':r['Name'],'available_products':a[r['Id']],'selected_products':b[r['Id']],'target':20,'status':'met' if b[r['Id']]>=20 else 'source_shortage' if a[r['Id']]<20 else 'selection_shortage'} for r in population if r.get('Published')]
    h={'version':1,'rails':[{'key':key,'category_id':identity('cat',cid),'limit':10} for key,cid in [('mobile',1),('earbuds',273),('power-bank',163),('smartwatch',755)]]}
    payload={'schema_version':1,'dataset_id':'gooshishop-public-catalog-2026-11-22','exported_at':now.isoformat(),'currency':'IRT','namespace':'gsp-','purpose':'test-catalog-snapshot','selection':{'minimum_per_primary_brand':20,'minimum_per_primary_category':20,'ranking':'UpdatedOnUtc DESC, Id DESC','membership':'primary category and primary brand; secondary links retained in source_records','eligible_products':len(eligible)},'catalog':{'categories':catout,'brands':brandout,'attributes':list(axes.values())+list(defs.values()),'products':productout,'variants':variantout,'offers':sorted(offers.values(),key=lambda x:x['id']),'stores':storeout,'home_settings':h},'media':media,'coverage':{'brands':coverage(chosen,'MainManufacturerId',rawbrands),'categories':coverage(chosen,'MainCategoryId',rawcats)},'mapping_notes':{'prices':'IRT integer; source PAC.FinalPrice for the selected supplier; other suppliers use cash OverriddenPrice; no recomputation of campaigns or installment rules','dimensions':'Source SyncProductDimensionsTask writes grams and centimeters; centimeters multiplied by 10 for target mm; zero measurements omitted','qa_enabled':'false: source product has no corresponding flag','gold':'static price snapshot only; source gold inputs retained, no gold formula engine is invented','lead_days':'max of supplier profile converted to days and selected per-offer deadlines; source detail retained','media':'URLs in this JSON; importer downloads, decodes and stores PNG bytes in target database; CDN required during first import','limits':'Current model limits apply to projections; source records preserve complete original text and extra fields','banners':'Legacy storefront banners retained as source records only; not automatically activated as Cherriva campaigns'},'projection_adjustments':issues,'source_records':{'products':chosen,'categories':rawcats,'brands':rawbrands,'specifications':specs,'variants':cpacs,'offers':[s for s in spacs if s['ProductAttributeCombinationId'] in pacidset],'inventory_values':[v for i,v in values.items() if i in usedvalues],'product_category_mappings':[r for r in read('product-category-maps') if r['ProductId'] in selected],'product_brand_mappings':[r for r in read('product-brand-maps') if r['ProductId'] in selected],'product_picture_mappings':[r for r in read('product-pictures') if r['ProductId'] in selected],'variant_id_mapping':sourcepacmap,'homepage':read('active-home'),'supplier_leads':read('supply-leads'),'product_inventory_mappings':[x for x in inventory_maps if x['ProductId'] in selected],'gold_reference_prices':read('gold-reference'),'supplier_gold_prices':read('supplier-gold'),'gold_profiles':read('gold-profiles')}}
    if media_checks:
        checks={x['id']:x for x in json.loads(media_checks.read_text())}
        original_media=list(payload['media'])
        missing={m['id'] for m in original_media if 'sha256' not in checks.get(m['id'],{})}
        for p in payload['catalog']['products']:
            if p['image_id'] in missing:raise ValueError('Primary image not verified: '+p['id'])
            p['gallery_ids']=[i for i in p['gallery_ids'] if i not in missing]
        for v in payload['catalog']['variants']:
            if v['image_id'] in missing:v['image_id']=''
        for c in payload['catalog']['categories']+payload['catalog']['brands']:
            if c['icon_id'] in missing:c['icon_id']=''
        payload['source_records']['media_references']=original_media
        payload['media']=[dict(m,sha256=checks[m['id']]['sha256'],source_bytes=checks[m['id']]['bytes']) for m in original_media if m['id'] not in missing]
        payload['media_verification']={'checked':len(original_media),'available':len(payload['media']),'unavailable_optional':[checks.get(i,{'id':i,'error':'not_verified'}) for i in sorted(missing)],'primary_images_missing':0,'verified_at':now.isoformat()}
    output.parent.mkdir(parents=True,exist_ok=True)
    output.write_text(json.dumps(payload,ensure_ascii=False,separators=(',',':'))+'\n',encoding='utf-8')
    print(json.dumps({'file':str(output),'bytes':output.stat().st_size,'sha256':hashlib.sha256(output.read_bytes()).hexdigest(),'counts':{k:len(v) for k,v in payload['catalog'].items() if isinstance(v,list)},'media':len(payload['media']),'projection_adjustments':len(issues)},ensure_ascii=False))

def sourcepacmap_as_int(m):return {int(k) for k in m}
if __name__=='__main__':
    ap=argparse.ArgumentParser();ap.add_argument('--raw',type=Path,required=True);ap.add_argument('--output',type=Path,required=True);ap.add_argument('--media-checks',type=Path);args=ap.parse_args();build(args.raw,args.output,args.media_checks)
