async (page) => {
 const isolated=await page.context().browser().newContext();page=await isolated.newPage();
 const base='http://localhost:5197',checks=[],errors=[];
 const check=(id,ok)=>{checks.push({id,pass:!!ok});if(!ok)throw Error(id)};
 const shot=async(name)=>page.screenshot({path:'output/playwright/issue-42-urls/'+name+'.png',animations:'disabled'});
 page.on('pageerror',e=>errors.push(e.message));
 const home=await(await page.request.get(base+'/storefront/home')).json();
 check('home-root-and-rail-slugs',home.categories.every(c=>/^\/category\/[a-z0-9]+(-[a-z0-9]+)*$/.test(c.url)&&!c.url.includes('gsp-'))&&home.rails.every(r=>!r.url.includes('gsp-')));
 check('home-product-public-codes',[...home.special_offers,...home.rails.flatMap(r=>r.products)].every(p=>/^\/product\/chr-[0-9]+$/.test(p.url)));
 const banners=await(await page.request.get(base+'/content/banners')).json();
 check('banner-links-canonical',banners.every(b=>!b.config.link.includes('gsp-cat-')&&!b.config.link.includes('gsp-product-')));
 for(const id of ['18189','37354','39782']){
  const api=await page.request.get(base+'/storefront/products/chr-'+id);check('api-public-code-'+id,api.status()===200);
  const d=await api.json();check('internal-id-retained-'+id,d.id==='gsp-product-'+id&&d.public_code==='chr-'+id&&d.url==='/product/chr-'+id);
  const r=await page.request.get(base+'/product/gsp-product-'+id,{maxRedirects:0});check('one-permanent-redirect-'+id,r.status()===301&&r.headers().location==='/product/chr-'+id);
  const html=await page.request.get(base+d.url);check('canonical-html-'+id,html.status()===200&&(await html.text()).includes('rel="canonical" href="'+d.url+'"'));
 }
 for(const [id,slug] of [['1','mobile'],['755','smart-watch'],['163','power-bank']]){
  const r=await page.request.get(base+'/category/gsp-cat-'+id+'?sort=price-asc&category=gsp-cat-'+id,{maxRedirects:0});
  check('category-redirect-query-'+id,r.status()===301&&r.headers().location==='/category/'+slug+'?sort=price-asc');
  const api=await page.request.get(base+'/storefront/catalog?category='+slug),d=await api.json();check('category-internal-id-and-links-'+id,api.status()===200&&d.category.id==='gsp-cat-'+id&&d.category.url==='/category/'+slug&&d.products.every(p=>p.url.startsWith('/product/chr-')));
 }
 await page.setViewportSize({width:1440,height:1000});await page.goto(base);await page.locator('[data-rail="mobile"] .lh-product').first().waitFor();
 await page.locator('[data-rail="mobile"] .lh-product').first().click();await page.locator('.pdp-overview').waitFor();
 check('spa-home-to-pdp-canonical',new URL(page.url()).pathname.startsWith('/product/chr-')&&await page.locator('link[rel="canonical"]').getAttribute('href')===new URL(page.url()).pathname);
 await page.goto(base+'/product/gsp-product-39782');await page.locator('.pdp-overview').waitFor();check('old-39782-browser-url',new URL(page.url()).pathname==='/product/chr-39782');await page.waitForFunction(()=>[...document.querySelectorAll('button')].some(b=>b.textContent.includes('افزودن به سبد خرید')&&!b.disabled));await page.locator('.pdp-large-photo img').evaluate(i=>i.decode());await shot('pdp-watch-desktop');
 await page.goto(base+'/category/gsp-cat-755');await page.locator('.plp-grid .lh-product').first().waitFor();check('old-watch-browser-url',new URL(page.url()).pathname==='/category/smart-watch');
 check('watch-breadcrumb',await page.locator('.plp-breadcrumb li').count()>=2);await shot('plp-watch-desktop');
 await page.setViewportSize({width:390,height:844});check('plp-mobile-no-overflow',await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));await shot('plp-watch-mobile');
 const d=await(await page.request.get(base+'/storefront/products/chr-14687')).json();
 const selected=d.offers.find(o=>o.id===d.selected_offer_id);const query='?variant='+encodeURIComponent(d.selected_variant_id)+'&offer='+encodeURIComponent(selected.id);
 const old=await page.request.get(base+'/product/gsp-product-14687'+query,{maxRedirects:0});check('selection-query-preserved',old.status()===301&&new URL(old.headers().location,base).searchParams.get('variant')===d.selected_variant_id&&new URL(old.headers().location,base).searchParams.get('offer')===selected.id);
 await page.goto(base+'/product/gsp-product-14687'+query);await page.locator('.pdp-overview').waitFor();check('variant-axes-present',await page.locator('.pdp-axes fieldset').count()===2);check('public-code-visible',await page.getByText('chr-14687',{exact:true}).count()===1);
 const alternate=d.variants.find(v=>v.available&&v.id!==d.selected_variant_id);
 if(alternate){for(let i=0;i<d.axes.length;i++){const value=alternate.values.find(x=>x.attribute_id===d.axes[i].id).value;await page.locator('.pdp-axes fieldset').nth(i).getByRole('button',{name:value,exact:true}).click();await page.waitForFunction(()=>document.querySelector('.pdp-overview')?.getAttribute('aria-busy')==='false')};await page.waitForFunction(id=>document.querySelector('.pdp-purchase')?.getAttribute('data-variant-id')===id,alternate.id);check('variant-change-keeps-canonical',await page.locator('link[rel="canonical"]').getAttribute('href')==='/product/chr-14687');await page.goto(base+d.url+query);await page.locator('.pdp-overview').waitFor()}
 check('pdp-mobile-no-overflow',await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));await shot('pdp-earbuds-mobile');
 await page.getByRole('button',{name:'افزودن به سبد خرید',exact:true}).click();await page.getByRole('dialog').waitFor();await page.getByRole('link',{name:'مشاهده سبد خرید',exact:true}).click();await page.locator('.buy-cart-row').waitFor();
 check('cart-canonical-links',await page.locator('.buy-cart-row a[href^="/product/chr-14687"]').count()===2);await shot('cart-mobile');
 await page.getByRole('button',{name:/حذف از سبد/}).click();await page.waitForFunction(()=>!document.querySelector('.buy-cart-row'));check('isolated-cart-cleaned',true);
 await page.setViewportSize({width:1440,height:1000});await page.goto('http://localhost:8098');
 await page.getByRole('textbox',{name:'شماره موبایل یا ایمیل'}).fill('09000000001');await page.getByRole('button',{name:'ادامه',exact:true}).click();await page.getByRole('button',{name:'دریافت کد یک‌بارمصرف'}).click();
 const code=await page.getByRole('alert').filter({hasText:'کد آزمون:'}).innerText();await page.getByRole('textbox',{name:'کد یک‌بارمصرف',exact:true}).fill(code.match(/\d{6}/)[0]);await page.locator('form').getByRole('button',{name:/ورود|تأیید/}).click();await page.locator('a[href="/products"]').first().waitFor();
 await page.locator('a[href="/categories"]').first().click();await page.getByRole('textbox',{name:'جست‌وجو',exact:true}).fill('ساعت هوشمند');await page.locator('a[href="/home-catalog?category=gsp-cat-755"]').click();await page.getByRole('dialog').waitFor();
 const slug=page.getByRole('textbox',{name:'آدرس انگلیسی دسته (slug)',exact:false});check('backoffice-slug-field',await slug.inputValue()==='smart-watch');check('source-english-name-preserved',await page.getByRole('textbox',{name:'نام انگلیسی',exact:true}).inputValue()==='gadget');
 check('category-tree-with-slug',await page.getByRole('dialog').locator('.category-tree').count()===1);await shot('backoffice-category-slug');
 await slug.fill('bad/slug');check('slug-client-validation',await slug.evaluate(el=>!el.checkValidity()));await page.getByRole('dialog').getByRole('button',{name:/انصراف|بستن/}).last().click();
 await page.locator('a[href="/products"]').first().click();await page.getByRole('textbox',{name:'جست‌وجو',exact:true}).fill('gsp-product-39782');await page.locator('a[href="/home-catalog?product=gsp-product-39782"]').click();await page.getByRole('dialog').waitFor();const codeInput=page.getByRole('textbox',{name:'کد عمومی محصول (فقط‌خواندنی)',exact:false});
 check('backoffice-readonly-product-code',await codeInput.inputValue()==='chr-39782'&&await codeInput.getAttribute('readonly')!==null);await shot('backoffice-product-code');
 await page.getByRole('dialog').getByRole('button',{name:/انصراف|بستن/}).last().click();
 check('no-runtime-page-errors',errors.length===0);await isolated.close();return {checks,errors,no_order_or_payment_created:true,no_catalog_edit_saved:true};
}
