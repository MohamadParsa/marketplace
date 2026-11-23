async (page) => {
 const site="http://127.0.0.1:5197",cases=[{"key": "mobile", "id": "gsp-product-18189", "category_id": "gsp-cat-1"}, {"key": "earbuds", "id": "gsp-product-14687", "category_id": "gsp-cat-273"}, {"key": "power-bank", "id": "gsp-product-28032", "category_id": "gsp-cat-163"}, {"key": "smartwatch", "id": "gsp-product-22763", "category_id": "gsp-cat-755"}, {"key": "coin", "id": "gsp-product-37354", "category_id": "gsp-cat-828"}, {"key": "bar", "id": "gsp-product-24572", "category_id": "gsp-cat-831"}, {"key": "melted", "id": "gsp-product-19687", "category_id": "gsp-cat-878"}],cartOffer={"id": "gsp-offer-134287", "product_id": "gsp-product-14687", "variant_id": "gsp-variant-14687-8d8476f60499", "store_id": "gsp-store-172", "price": 3959000, "regular_price": 4300000, "stock": 23, "reserved": 0, "version": 1},checks=[],errors=[];
 const check=(id,pass)=>{checks.push({id,pass:!!pass});if(!pass)throw Error(id)};
 const digits=s=>s.replace(/[۰-۹]/g,c=>String(c.charCodeAt(0)-1776)).replace(/[^0-9.]/g,"");
 const shot=async(name,fullPage=false)=>page.screenshot({path:"output/playwright/catalog-seed-local-2026-11-23/"+name+".png",fullPage,animations:"disabled"});
 page.on("pageerror",e=>errors.push(e.message));
 await page.setViewportSize({width:1440,height:1000});await page.goto(site);await page.locator('[data-rail="mobile"] .lh-product').first().waitFor();
 for(const key of ["mobile","earbuds","power-bank","smartwatch"]){check("home-ten-cards-"+key,(await page.locator('[data-rail="'+key+'"] .lh-product').count())===10)}
 check("special-offers-visible",(await page.locator('[data-rail="special"] .lh-product').count())===10);
 const home=(await (await page.request.get(site+'/storefront/home')).json());
 const catalog=(await (await page.request.get(site+'/storefront/catalog?category=gsp-cat-1')).json());
 const roots=new Set(catalog.navigation.filter(c=>!c.parent_id).map(c=>c.id));
 check("stocked-root-category-shortcuts",home.categories.length>0&&home.categories.every(c=>roots.has(c.id)&&c.count>0)&&(await page.locator('.lh-categories a').count())===home.categories.length);
 await shot("home-desktop");await page.locator('[data-rail="mobile"]').scrollIntoViewIfNeeded();await shot("home-mobile-carousel");
 const examples=[];
 for(const c of cases){
  await page.goto(site+"/category/"+c.category_id);await page.locator('.plp-grid .lh-product').first().waitFor();
  check("plp-breadcrumb-"+c.key,(await page.locator('.plp-breadcrumb li').count())>=2);
  check("plp-cards-and-features-"+c.key,(await page.locator('.plp-grid .lh-product').count())>0);
  if(c.key==='mobile')await shot("plp-mobile-category");
  const r=await page.request.get(site+'/storefront/products/'+c.id);if(r.status()!==200)throw Error('PDP API '+c.id);const d=await r.json();
  await page.goto(site+'/product/'+c.id);await page.locator('.pdp-overview').waitFor();await page.locator('.pdp-price strong').waitFor();
  const crumb=await page.locator('.plp-breadcrumb li').allTextContents();
  check('pdp-brand-breadcrumb-'+c.key,crumb.at(-2)?.trim()===d.brand.name&&crumb.at(-1)?.trim()===d.name);
  check('pdp-two-axes-'+c.key,(await page.locator('.pdp-axes fieldset').count())===2);
  for(let i=0;i<d.axes.length;i++){const used=new Set(d.variants.flatMap(v=>v.values.filter(x=>x.attribute_id===d.axes[i].id).map(x=>x.value)));check('pdp-relevant-options-'+c.key+'-'+i,(await page.locator('.pdp-axes fieldset').nth(i).getByRole('button').count())===used.size)}
  const selected=d.offers.find(o=>o.id===d.selected_offer_id);check('visible-toman-price-'+c.key,Number(digits(await page.locator('.pdp-price strong').innerText()))===selected.price/10);
  await page.locator('.pdp-large-photo img').evaluate(i=>i.decode());check('pdp-real-image-'+c.key,await page.locator('.pdp-large-photo img').evaluate(i=>i.naturalWidth>0));
  if(c.key==='mobile'||c.key==='coin')await shot('pdp-'+c.key);
  const alternate=d.variants.find(v=>v.available&&v.id!==d.selected_variant_id);
  if(alternate){for(let i=0;i<d.axes.length;i++){const value=alternate.values.find(x=>x.attribute_id===d.axes[i].id).value;await page.locator('.pdp-axes fieldset').nth(i).getByRole('button',{name:value,exact:true}).click();await page.waitForFunction(()=>document.querySelector('.pdp-overview')?.getAttribute('aria-busy')==='false')}
   await page.waitForFunction(id=>document.querySelector('.pdp-purchase')?.getAttribute('data-variant-id')===id,alternate.id);
   check('variant-selection-updates-'+c.key,true)}
  examples.push({key:c.key,url:site+'/product/'+c.id});
 }
 await page.goto(site+'/product/'+cartOffer.product_id+'?variant='+cartOffer.variant_id+'&offer='+cartOffer.id);
 await page.locator('.pdp-purchase button').first().waitFor();await page.getByRole('button',{name:'افزودن به سبد خرید',exact:true}).click();await page.getByRole('dialog').waitFor();
 check('add-to-cart-popup',await page.getByRole('link',{name:'مشاهده سبد خرید',exact:true}).isVisible());
 await page.getByRole('button',{name:'ادامه خرید',exact:true}).click();await page.goto(site+'/cart');await page.locator('.buy-cart-row').first().waitFor();
 check('cart-unit-price-in-toman',Number(digits(await page.locator('.buy-cart-row strong').first().innerText()))===cartOffer.price);
 await page.getByRole('button',{name:/افزایش تعداد/}).click();await page.waitForFunction(()=>document.querySelector('.buy-cart-row .purchase-control output')?.textContent==='۲');check('cart-quantity-increase',true);
 await page.getByRole('button',{name:/کاهش تعداد/}).click();await page.waitForFunction(()=>document.querySelector('.buy-cart-row .purchase-control output')?.textContent==='۱');check('cart-quantity-decrease',true);
 await shot('cart-desktop');await page.setViewportSize({width:390,height:844});await shot('cart-mobile');check('cart-mobile-no-overflow',await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
 await page.getByRole('button',{name:/حذف از سبد/}).click();await page.waitForFunction(()=>!document.querySelector('.buy-cart-row'));check('verification-cart-cleaned',true);
 await page.goto(site+'/category/gsp-cat-1');await page.locator('.plp-grid .lh-product').first().waitFor();check('plp-mobile-no-overflow',await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));await shot('plp-mobile');
 await page.goto(site+'/product/'+cases[4].id);await page.locator('.pdp-overview').waitFor();check('pdp-mobile-no-overflow',await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));await shot('pdp-coin-mobile');
 check('no-runtime-page-errors',errors.length===0);return {checks,errors,examples,guest_cart_only:true,no_order_placed:true};
}