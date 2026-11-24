async(cliPage)=>{
 const browser=cliPage.context().browser(),ctx=await browser.newContext({viewport:{width:1440,height:1000}}),page=await ctx.newPage(),origin='http://localhost:5197';
 const checks=[],errors=[];let mutations=0,guest='';const check=(id,pass)=>{checks.push({id,pass:!!pass});if(!pass)throw Error(id)};
 const until=async fn=>{for(let i=0;i<100;i++){if(await fn())return;await page.waitForTimeout(50)}throw Error('condition timed out')};
 page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{if(r.method()==='POST'&&r.url()===origin+'/storefront/cart')mutations++});
 const cart=async()=>{if(!guest)guest=await page.evaluate(()=>localStorage.getItem('cherriva:guest-cart'));const r=await ctx.request.get(origin+'/storefront/cart',{headers:{'X-Cart-Token':guest}});if(!r.ok())throw Error('cart read '+r.status());return r.json()};
 const home=await(await ctx.request.get(origin+'/storefront/home')).json();const product=home.rails.flatMap(r=>r.products).find(p=>p.price>0&&p.url);
 if(!product)throw Error('available product required');
 try{
  await page.goto(origin+product.url);await page.getByRole('heading',{name:product.name,exact:true}).waitFor();const add=()=>page.getByRole('button',{name:'افزودن به سبد خرید',exact:true});await until(async()=>await add().isEnabled());
  check('desktop-one-active-design-control',await add().count()===1&&await add().isEnabled());
  await page.route('**/storefront/cart',async route=>{if(route.request().method()!=='POST')return route.continue();const response=await route.fetch();await page.waitForTimeout(500);await route.fulfill({response})});
  await add().evaluate(e=>{e.click();e.click()});await page.locator('.purchase-control [role="progressbar"]').waitFor();check('add-shows-progress-overlay',await page.locator('.purchase-control').getAttribute('aria-busy')==='true');
  await page.getByRole('heading',{name:'به سبد خرید اضافه شد',exact:true}).waitFor();const first=await cart();check('desktop-double-click-one-add',mutations===1&&first.items.length===1&&first.items[0].quantity===1);check('popup-links-to-real-cart',await page.getByRole('link',{name:'مشاهده سبد خرید',exact:true}).getAttribute('href')==='/cart');
  await page.getByRole('button',{name:'ادامه خرید',exact:true}).click();check('same-control-becomes-counter',await page.locator('.pdp-purchase .purchase-control output').innerText()==='۱');await page.locator('.pdp-purchase').scrollIntoViewIfNeeded();await page.screenshot({path:'output/playwright/pdp-add-enabled/desktop.png',animations:'disabled'});
  await page.getByRole('button',{name:'حذف از سبد '+product.name,exact:true}).click();await until(async()=>(await cart()).items.length===0);await until(async()=>await add().isEnabled());check('desktop-delete-restores-add',await add().isEnabled());
  await page.setViewportSize({width:390,height:844});await page.reload();await until(async()=>await add().isEnabled());check('mobile-add-enabled',await add().isVisible());await add().click();await page.getByRole('heading',{name:'به سبد خرید اضافه شد',exact:true}).waitFor();check('mobile-add-saved-in-server-cart',(await cart()).items[0]?.quantity===1);await page.getByRole('link',{name:'مشاهده سبد خرید',exact:true}).click();await page.waitForURL('**/cart');await page.getByText(product.name,{exact:true}).first().waitFor();check('cart-page-shows-selected-product',await page.getByText(product.name,{exact:true}).count()>0);
  await page.goto(origin+product.url);await page.locator('.pdp-purchase .purchase-control output').waitFor();await page.screenshot({path:'output/playwright/pdp-add-enabled/mobile.png',animations:'disabled'});await page.getByRole('button',{name:'حذف از سبد '+product.name,exact:true}).click();await until(async()=>(await cart()).items.length===0);check('guest-cart-cleaned-up',(await cart()).items.length===0);
  await page.route('**/storefront/products/*',async route=>{if(route.request().method()!=='GET')return route.continue();const response=await route.fetch();const d=await response.json();d.offers=d.offers.map(o=>({...o,available:false,price:null}));d.variants=d.variants.map(v=>({...v,available:false}));return route.fulfill({response,json:d})});await page.reload();await page.getByRole('button',{name:'ناموجود',exact:true}).waitFor();check('unavailable-product-still-blocked',await page.getByRole('button',{name:'ناموجود',exact:true}).isDisabled());check('no-runtime-errors',errors.length===0);
  return{checks,errors,product_url:product.url,guest_cart_empty:true,no_order_or_payment_created:true};
 }finally{
  if(guest){const current=await cart();for(const item of current.items){const latest=await cart();await ctx.request.post(origin+'/storefront/cart',{headers:{'X-Cart-Token':guest,'Idempotency-Key':crypto.randomUUID()},data:{action:'remove',offer_id:item.offer_id,quantity:0,version:latest.version}})}}
  await ctx.close();
 }
}
