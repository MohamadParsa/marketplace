async(cliPage)=>{
 const ctx=await cliPage.context().browser().newContext({viewport:{width:390,height:844}}),page=await ctx.newPage(),origin='http://localhost:5197',checks=[];
 const check=(id,pass)=>{checks.push({id,pass:!!pass});if(!pass)throw Error(id)};
 const home=await(await ctx.request.get(origin+'/storefront/home')).json(),p=home.rails[0].products[0];
 const pattern='**/storefront/products/'+p.url.split('/').pop();
 await page.route(pattern,async route=>{
  const response=await route.fetch(),d=await response.json();
  d.name='نام طولانی محصول طلای آزمونِ چیدمان و خوانایی '.repeat(7);d.gallery=[];
  const axis=d.axes[0],v=d.variants.find(v=>v.id===d.selected_variant_id);axis.name='وزن طلا';axis.unit='گرم';axis.groups=[];axis.options=Array.from({length:16},(_,i)=>String((i+1)/10));
  d.variants=axis.options.map((value,i)=>({...v,id:i===0?v.id:'chr-v-'+Buffer.from('layout-only-'+i).toString('base64url'),values:v.values.map(x=>x.attribute_id===axis.id?{...x,value}:x)}));
  await route.fulfill({response,json:d});
 });
 try{
  await page.goto(origin+p.url);await page.locator('.pdp-purchase').waitFor();
  check('long-title-and-missing-image-no-overflow',await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
  check('missing-image-fallback-visible',await page.locator('.pdp-large-photo .sf-image-fallback').isVisible());
  const axis=page.locator('.pdp-variant-axis').first();check('many-weight-choices-initially-bounded',await axis.locator('.pdp-axis-chip').count()===6);
  await axis.locator('.pdp-axis-toggle').click();check('all-weight-choices-can-expand',await axis.locator('.pdp-axis-chip').count()===16);
  check('expanded-weight-choices-no-overflow',await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
  await page.screenshot({path:'output/playwright/issues-102-103-104/long-title-gold-weights-mobile.png',animations:'disabled'});
  await axis.locator('.pdp-axis-toggle').click();check('weight-collapse-restores-bounded-layout',await axis.locator('.pdp-axis-chip').count()===6);
  // Desktop also retains separate purchase and specification regions.
  await page.setViewportSize({width:1440,height:900});check('long-title-desktop-no-overflow',await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
  await page.screenshot({path:'output/playwright/issues-102-103-104/long-title-gold-weights-desktop.png',animations:'disabled'});
  return{checks,fixture:'browser response only: long title, no image, 16 weights; no database writes'};
 }finally{await ctx.close()}
}
