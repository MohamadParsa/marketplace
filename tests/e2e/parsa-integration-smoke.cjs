async(cliPage)=>{
 const ctx=await cliPage.context().browser().newContext({viewport:{width:1440,height:1000}}),page=await ctx.newPage(),origin='http://localhost:5197',checks=[],errors=[],responses=[];
 const check=(id,pass)=>{checks.push({id,pass:!!pass});if(!pass)throw Error(id)};page.on('pageerror',e=>errors.push(e.message));page.on('response',r=>{if(/\/storefront\/products\/[^/]+\/(reviews|questions)(\?|$)/.test(r.url()))responses.push({kind:r.url().includes('/reviews')?'reviews':'questions',status:r.status()})});
 try{
  const home=await(await ctx.request.get(origin+'/storefront/home')).json(),product=home.rails[0].products[0];
  await page.goto(origin+product.url);await page.getByRole('heading',{name:product.name,exact:true}).waitFor();await page.getByRole('heading',{name:'نظرات کاربران',exact:true}).waitFor();await page.getByRole('heading',{name:'پرسش و پاسخ',exact:true}).waitFor();
  check('integrated-both-interaction-sections',await page.locator('#reviews').count()===1&&await page.locator('#questions').count()===1);
  for(const kind of ['reviews','questions']){const r=await ctx.request.get(origin+'/storefront/products/'+encodeURIComponent(product.id)+'/'+kind);check('integrated-'+kind+'-endpoint',r.status()===200&&r.headers()['cache-control']==='no-store')}
  const d=await(await ctx.request.get(origin+'/storefront/products/'+encodeURIComponent(product.id))).json();check('integrated-independent-public-flags',d.interactions.reviews===d.content.reviews_enabled&&d.interactions.qa===d.content.qa_enabled);
  check('integrated-wishlist-and-product-tabs-preserved',await page.locator('.pdp-wishlist button').count()===1&&await page.locator('.pdp-tabs a[href="#pros-cons"]').count()===1);
  await page.setViewportSize({width:390,height:844});await page.locator('#questions').scrollIntoViewIfNeeded();check('integrated-mobile-no-overflow',await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
  check('integrated-no-runtime-errors',errors.length===0);return{checks,errors,responses,product_url:product.url,no_order_or_payment_created:true};
 }finally{await ctx.close()}
}
