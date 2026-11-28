async(cliPage) => {
 const ctx=await cliPage.context().browser().newContext({viewport:{width:1440,height:900}}),page=await ctx.newPage(),origin='http://localhost:5197';
 const checks=[],errors=[];const check=(id,pass)=>{checks.push({id,pass:!!pass});if(!pass)throw Error(id)};
 const until=async(fn)=>{for(let i=0;i<120;i++){if(await fn())return;await page.waitForTimeout(50)}throw Error('condition timeout')};
 page.on('pageerror',e=>errors.push(e.message));
 const home=await(await ctx.request.get(origin+'/storefront/home')).json(),rail=home.rails.find(r=>r.products.length),product=rail.products[0];
 const banners=await(await ctx.request.get(origin+'/content/banners')).json();
 const detail=await(await ctx.request.get(origin+'/storefront/products/'+product.url.split('/').pop())).json();
 const decode=(t)=>Buffer.from(t.slice(6),'base64url').toString('utf8');
 try {
  check('banner-and-rail-category-references-match',banners.filter(b=>b.config.catalog_category_id).every(b=>home.rails.some(r=>r.category_id===b.config.catalog_category_id)));
  check('public-home-ids-and-media-have-no-source-prefix',!JSON.stringify(home).includes('gsp-'));
  check('public-product-ids-and-media-have-no-source-prefix',!JSON.stringify(detail).includes('gsp-'));
  check('variant-offer-and-media-use-chr-tokens',detail.selected_variant_id.startsWith('chr-v-')&&detail.selected_offer_id.startsWith('chr-o-')&&detail.gallery.every(u=>u.includes('/chr-m-')));
  const oldV=decode(detail.selected_variant_id),oldO=decode(detail.selected_offer_id);
  const old=await(await ctx.request.get(origin+'/storefront/products/'+decode(detail.id)+'?variant='+oldV+'&offer='+oldO)).json();
  check('legacy-selection-same-offer-price-and-image',old.selected_offer_id===detail.selected_offer_id&&JSON.stringify(old.offers)===JSON.stringify(detail.offers)&&JSON.stringify(old.gallery)===JSON.stringify(detail.gallery));
  await page.goto(origin+product.url+'?variant='+oldV+'&offer='+oldO);await page.locator('.pdp-information h1').waitFor();await until(()=>!page.url().includes('gsp-'));check('old-link-canonicalized-with-same-selection',new URL(page.url()).searchParams.get('variant')===detail.selected_variant_id&&new URL(page.url()).searchParams.get('offer')===detail.selected_offer_id);
  check('canonical-excludes-selection-query',(await page.locator('link[rel=canonical]').getAttribute('href'))===product.url);
  const invalid=await ctx.request.get(origin+'/storefront/products/'+detail.id+'?variant=chr-v-!');check('malformed-token-is-rejected',invalid.status()===400);
  for(const [name,url,pattern,skeleton,ready] of [
   ['home','/','**/storefront/home','.sf-home-skeleton','.lh-rail[data-rail]'],
   ['plp',rail.url,'**/storefront/catalog?*','.sf-card-skeleton','.plp-grid .lh-product:not(.sf-card-skeleton)'],
   ['pdp',product.url,'**/storefront/products/'+product.url.split('/').pop(),'.sf-product-skeleton','.pdp-purchase']
  ]) {
   let release;const gate=new Promise(r=>release=r);
   await page.route(pattern,async route=>{const response=await route.fetch();await gate;await route.fulfill({response})});
   await page.goto(origin+url);await page.locator(skeleton).first().waitFor();await page.waitForTimeout(160);
   check(name+'-slow-data-shaped-skeleton',await page.locator(skeleton).first().isVisible());
   await page.screenshot({path:'output/playwright/issues-102-103-104/loading-'+name+'.png',animations:'disabled'});
   release();await page.locator(ready).first().waitFor();await until(async()=>await page.locator(skeleton).count()===0);check(name+'-skeleton-clears-after-data',await page.locator(skeleton).count()===0);await page.unroute(pattern);
  }
  // Image frame dimensions must survive delayed image completion and an error.
  let releaseImages;const imageGate=new Promise(r=>releaseImages=r);
  await page.route('**/catalog/icons/chr-m-*',async route=>{const response=await route.fetch();await imageGate;await route.fulfill({response})});
  await page.goto(origin+product.url);await page.locator('.pdp-large-photo .sf-image-skeleton').waitFor();
  const before=await page.locator('.pdp-large-photo').boundingBox();check('slow-image-shows-skeleton',await page.locator('.pdp-large-photo [aria-busy=true]').count()===1);releaseImages();
  await page.locator('.pdp-large-photo [data-image-state=ready]').waitFor();const after=await page.locator('.pdp-large-photo').boundingBox();check('image-frame-height-and-width-stable',before.height===after.height&&before.width===after.width);await page.unroute('**/catalog/icons/chr-m-*');
  await page.route('**/catalog/icons/chr-m-*',route=>route.fulfill({status:404,body:'missing'}));await page.reload();await page.locator('.pdp-large-photo .sf-image-fallback').waitFor();check('broken-image-fallback-without-permanent-skeleton',await page.locator('.pdp-large-photo .sf-image-skeleton').count()===0);await page.unroute('**/catalog/icons/chr-m-*');
  // Explicit network failure and retry; no immortal placeholder.
  await page.route('**/storefront/products/'+product.url.split('/').pop(),route=>route.fulfill({status:503,contentType:'application/json',body:'{}'}));await page.reload();await page.locator('.pdp-state[role=alert]').waitFor();check('api-failure-replaces-skeleton-with-retry',await page.locator('.sf-product-skeleton').count()===0&&await page.getByRole('button',{name:'تلاش مجدد',exact:true}).isVisible());await page.unroute('**/storefront/products/'+product.url.split('/').pop());await page.getByRole('button',{name:'تلاش مجدد',exact:true}).click();await page.locator('.pdp-purchase').waitFor();check('retry-restores-product',await page.locator('.pdp-purchase').count()===1);
  // Empty listing must be an empty state, not a skeleton.
  await page.goto(origin+rail.url+'?q=cherriva-no-match-4491637');await page.locator('.plp-empty').waitFor();check('empty-listing-is-not-a-skeleton',await page.locator('.sf-card-skeleton').count()===0);
  await page.emulateMedia({reducedMotion:'reduce'});
  let releaseReduced;const reducedGate=new Promise(r=>releaseReduced=r);const pattern='**/storefront/products/'+product.url.split('/').pop();await page.route(pattern,async route=>{const response=await route.fetch();await reducedGate;await route.fulfill({response})});await page.goto(origin+product.url);await page.locator('.sf-product-skeleton .sf-skeleton').first().waitFor();check('reduced-motion-disables-shimmer',await page.locator('.sf-product-skeleton .sf-skeleton').first().evaluate(e=>getComputedStyle(e,'::after').animationName)==='none');releaseReduced();await page.locator('.pdp-purchase').waitFor();await page.unroute(pattern);
  // Browser Back during a slow PDP fetch must never restore it over the new page.
  let releaseLate;const lateGate=new Promise(r=>releaseLate=r);await page.route(pattern,async route=>{const response=await route.fetch();await lateGate;try{await route.fulfill({response})}catch{}});await page.goto(origin+rail.url);await page.locator('.plp-grid .lh-product').first().waitFor();await page.locator('.plp-grid .lh-product').first().click();await page.locator('.sf-product-skeleton').waitFor();await page.goBack();releaseLate();await page.locator('.plp-grid .lh-product').first().waitFor();check('late-navigation-response-does-not-replace-new-page',new URL(page.url()).pathname===rail.url&&await page.locator('.sf-product-skeleton').count()===0);await page.unroute(pattern);
  await page.setViewportSize({width:390,height:844});await page.goto(origin+product.url);await page.locator('.pdp-purchase').waitFor();check('mobile-product-no-overflow',await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));check('one-mobile-design-purchase-control',await page.locator('.pdp-purchase .purchase-control').count()===1);check('breadcrumb-and-brand-remain',await page.locator('.plp-breadcrumb').innerText().then(t=>t.includes(detail.brand?.name??detail.category.name)));
  await page.keyboard.press('Tab');check('keyboard-focus-visible',await page.evaluate(()=>document.activeElement!==document.body));await page.goto(origin+rail.url);await page.locator('.plp-grid .lh-product').first().waitFor();
  await page.getByRole('button',{name:'فیلترها',exact:true}).click();await page.getByRole('dialog',{name:'فیلتر کالاها'}).waitFor();check('compact-mobile-filter-dialog-opens',await page.getByRole('dialog',{name:'فیلتر کالاها'}).isVisible());await page.keyboard.press('Escape');await page.getByRole('dialog',{name:'فیلتر کالاها'}).waitFor({state:'hidden'});check('filter-escape-restores-focus',await page.evaluate(()=>document.activeElement?.textContent?.trim()==='فیلترها'));
  await page.getByRole('radio',{name:'پرفروش‌ترین',exact:true}).focus();await page.keyboard.press('Space');await page.waitForURL('**sort=best-selling');await page.locator('.plp-main .sf-busy-region[aria-busy=false]').waitFor();check('horizontal-sort-keyboard-works',await page.getByRole('radio',{name:'پرفروش‌ترین',exact:true}).isChecked());check('compact-mobile-toolbar-no-page-overflow',await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
  check('no-runtime-errors',errors.length===0);
  return {checks,errors,product_url:product.url,no_order_or_payment_created:true};
 } finally {await ctx.close()}
}
