async (page) => {
 const admin=await page.context().newPage(),checks=[],errors=[];
 const check=(id,pass)=>{checks.push({id,pass:!!pass});if(!pass)throw Error(id)};
 admin.on('pageerror',e=>errors.push(e.message));await admin.setViewportSize({width:1440,height:1000});
 await admin.goto('http://127.0.0.1:8098/commerce');await admin.getByLabel(/شماره موبایل یا ایمیل/).fill('review-admin@example.invalid');await admin.getByRole('button',{name:'ادامه',exact:true}).click();await admin.getByLabel(/^رمز عبور/).fill('Review-only-27-pass');
 const logged=admin.waitForResponse(r=>r.url().endsWith('/backoffice/auth/password-login')&&r.ok());await admin.getByRole('button',{name:'ورود با رمز عبور',exact:true}).click();const token=(await(await logged).json()).token;
 await admin.getByRole('heading',{name:'سفارش‌ها و پرداخت‌های سایت',exact:true}).waitFor();
 const data=await(await admin.request.get('http://127.0.0.1:8098/backoffice/commerce',{headers:{Authorization:'Bearer '+token}})).json();
 const p=data.purchases.find(p=>p.deposits.some(d=>d.status==='approved'));if(!p)throw Error('Run checkout.cjs first for a synthetic credited order');
 await admin.locator('.commerce-list article').filter({hasText:'#'+p.number}).getByRole('button',{name:'جزئیات سفارش',exact:true}).click();const dialog=admin.getByRole('dialog');
 check('admin-order-five-distinct-boxes',await dialog.locator('.commerce-order-box').count()===5);
 check('payment-rows-read-as-list',await dialog.locator('.commerce-payment-box dl').evaluate(e=>getComputedStyle(e).display==='block'));
 await admin.screenshot({path:'output/playwright/purchase-refinements/admin-order-desktop.png',animations:'disabled'});
 await dialog.getByRole('heading',{name:/اقلام سفارش/}).scrollIntoViewIfNeeded();await admin.screenshot({path:'output/playwright/purchase-refinements/admin-order-items.png',animations:'disabled'});
 check('shipment-and-line-details-visible',await dialog.locator('.commerce-order-item').count()===p.quote.lines.length&&await dialog.getByRole('heading',{name:'مرسوله‌ها و ارسال',exact:true}).count()===1);
 await admin.setViewportSize({width:390,height:844});await dialog.locator('.MuiDialogContent-root').evaluate(e=>e.scrollTop=0);
 check('admin-order-mobile-no-horizontal-overflow',await dialog.evaluate(e=>e.scrollWidth<=e.clientWidth)&&await admin.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
 await admin.screenshot({path:'output/playwright/purchase-refinements/admin-order-mobile.png',animations:'disabled'});
 check('admin-order-content-scrolls',await dialog.locator('.MuiDialogContent-root').evaluate(e=>e.scrollHeight>e.clientHeight));
 await dialog.getByRole('button',{name:'بستن',exact:true}).click();await admin.setViewportSize({width:1440,height:1000});await admin.locator('a[href="/orders"]').first().click();await admin.getByRole('heading',{name:'سفارش‌ها',exact:true}).waitFor();
 check('operations-panel-shows-public-order-number',await admin.getByText('سفارش '+p.number,{exact:true}).count()>0);
 await admin.locator('a[href="/commerce-settings"]').first().click();await admin.getByRole('heading',{name:'استان‌ها و شهرها',exact:true}).waitFor();
 const editor=admin.getByRole('region',{name:'مدیریت استان و شهر',exact:true}),provinceName='استان ساختگی آزمون پنل';
 let province=editor.locator('.geography-province').filter({hasText:provinceName});
 if(await province.count()===0){await editor.getByRole('button',{name:'افزودن استان',exact:true}).click();province=editor.locator('.geography-province').last();await province.locator('summary').click();await province.getByRole('textbox',{name:'نام استان',exact:true}).fill(provinceName);await province.getByRole('button',{name:'افزودن شهر به این استان',exact:true}).click();await province.getByRole('textbox',{name:'نام شهر',exact:true}).fill('شهر ساختگی آزمون');}
 else {await province.locator('summary').click();await province.getByRole('checkbox',{name:'استان فعال',exact:true}).check();}
 const saved=admin.waitForResponse(r=>r.url().endsWith('/backoffice/commerce/settings')&&r.request().method()==='PATCH'&&r.status()===200);await editor.getByRole('button',{name:'ذخیره استان‌ها و شهرها',exact:true}).click();await saved;await admin.getByRole('alert').filter({hasText:'عملیات ثبت شد'}).waitFor();
 let pub=await(await admin.request.get('http://127.0.0.1:5197/storefront/commerce/config')).json();check('admin-created-province-city-reaches-site',pub.geography.some(p=>p.name===provinceName&&p.cities.some(c=>c.name==='شهر ساختگی آزمون')));
 await admin.screenshot({path:'output/playwright/purchase-refinements/admin-geography.png',animations:'disabled'});
 province=editor.locator('.geography-province').filter({hasText:provinceName});await province.locator('summary').click();await province.getByRole('checkbox',{name:'استان فعال',exact:true}).uncheck();
 const disabled=admin.waitForResponse(r=>r.url().endsWith('/backoffice/commerce/settings')&&r.request().method()==='PATCH'&&r.status()===200);await editor.getByRole('button',{name:'ذخیره استان‌ها و شهرها',exact:true}).click();await disabled;await admin.getByRole('alert').filter({hasText:'عملیات ثبت شد'}).waitFor();
 pub=await(await admin.request.get('http://127.0.0.1:5197/storefront/commerce/config')).json();check('deactivated-province-excluded-from-public-directory',!pub.geography.some(p=>p.name===provinceName));
 check('no-admin-runtime-pageerror',errors.length===0);return {checks,passed:checks.length,errors};
}
