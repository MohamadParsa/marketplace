async(cliPage)=>{
 const ctx=await cliPage.context().browser().newContext({viewport:{width:1440,height:1000}}),page=await ctx.newPage(),checks=[],errors=[];const origin='http://localhost:5196';
 const check=(id,pass)=>{checks.push({id,pass:!!pass});if(!pass)throw Error(id)};page.on('pageerror',e=>errors.push(e.message));
 const req=async(path,body)=>{const r=await ctx.request.fetch(origin+path,{method:body?'POST':'GET',headers:{Authorization:'Bearer '+token,'Idempotency-Key':crypto.randomUUID()},...(body?{data:body}:{})});if(!r.ok())throw Error(path+' '+r.status());return r.json()};let token='';
 const until=async fn=>{for(let i=0;i<100;i++){if(await fn())return;await page.waitForTimeout(50)}throw Error('condition timed out')};
 try{
  const otp=await req('/storefront/auth/request-otp',{phone:'09000000003'});token=(await req('/storefront/auth/verify-otp',{request_id:otp.request_id,code:otp.test_code})).token;
  const cart=await req('/storefront/cart');await req('/storefront/cart',{action:'add',offer_id:'offer-b',quantity:1,version:cart.version});
  await page.goto(origin+'/checkout');await page.evaluate(t=>sessionStorage.setItem('cherriva:customer-token',t),token);await page.reload();await page.getByRole('radio',{name:'انتخاب این آدرس',exact:true}).first().check();await page.locator('input[name="shipping-s1"]').check();await page.getByRole('button',{name:'تأیید آدرس و ارسال؛ ادامه به پرداخت',exact:true}).click();await page.getByRole('heading',{name:'روش پرداخت و تأیید سفارش',exact:true}).waitFor();
  let previews=0;await page.route('**/storefront/checkout/preview',async route=>{previews++;const response=await route.fetch();await page.waitForTimeout(700);await route.fulfill({response})});
  await page.getByRole('textbox',{name:'کد تخفیف',exact:true}).fill('BROWSER10');await page.getByRole('button',{name:'اعمال کد تخفیف',exact:true}).evaluate(e=>{e.click();e.click()});await page.getByRole('progressbar').waitFor();check('adjustment-request-blocks-input-and-uses-overlay',await page.getByRole('textbox',{name:'کد تخفیف',exact:true}).isDisabled());await until(async()=>await page.locator('.buy-coupon-box [role="status"]').count()===1);check('double-click-produces-one-preview-request',previews===1);
  await page.getByRole('button',{name:'حذف کد تخفیف',exact:true}).click();await until(async()=>await page.locator('.buy-coupon-box [role="status"]').count()===0);check('remove-coupon-restores-server-total',await page.getByRole('textbox',{name:'کد تخفیف',exact:true}).inputValue()===''&&await page.locator('.buy-order-summary').first().innerText().then(t=>t.replace(/[٬,]/g,'').includes('100000')));
  await page.getByLabel('مبلغ مصرف از کیف پول',{exact:true}).fill('-1');await page.getByRole('button',{name:'اعمال اعتبار',exact:true}).click();await page.getByRole('alert').waitFor();check('invalid-wallet-preserves-stage-and-input',await page.getByLabel('مبلغ مصرف از کیف پول',{exact:true}).inputValue()==='-1'&&await page.getByRole('heading',{name:'روش پرداخت و تأیید سفارش',exact:true}).isVisible());
  check('edge-cases-no-runtime-errors',!errors.length);return{checks,errors,isolated_memory_fixture:true};
 }finally{await ctx.close()}
}
