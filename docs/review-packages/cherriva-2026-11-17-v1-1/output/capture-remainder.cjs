async(page)=>{
 const root='http://127.0.0.1:4190/';const captures=[],errors=[];page.on('pageerror',e=>errors.push(e.message));
 const go=async url=>{if(await page.locator('#drawer[open]').count())await page.keyboard.press('Escape');await page.goto(root+url);await page.evaluate(()=>document.fonts.ready);};
 const shot=async(path)=>{await page.screenshot({path,fullPage:true});captures.push(path);};
 const bo=async(route,theme='a',role='admin')=>{await go('02-backoffice/index.html#'+route+'?role='+role+'&theme='+theme+'&scene=normal');await page.waitForFunction(r=>CVDesign.state.route===r,route);};
 await page.setViewportSize({width:1440,height:1000});await bo('dashboard');const pages=await page.evaluate(()=>CV.pages.map(x=>({id:x.id,code:x.code})));
 for(const route of ['login','my-profile','categories','products','pages','orders','dashboard']){await page.setViewportSize({width:390,height:844});await bo(route);await shot('02-backoffice/output/playwright/mobile-'+route+'.png');}
 await page.setViewportSize({width:1440,height:1000});
 const drawers=[['categories','[data-action=category][data-id="موبایل"]','category-attributes'],['brands','[data-next=brand]','brand-icon'],['products','[data-next=variants][data-id=P2]','product-variants'],['products','[data-next=product-offers][data-id=P2]','product-pricing-stock'],['stores','[data-action=detail][data-id=S1]','store-supply-lead'],['pages','[data-action=detail][data-id=SC1]','banner-desktop-mobile'],['pages','[data-action=detail][data-id=SC8]','manual-product-carousel'],['pages','[data-action=detail][data-id=SC7]','sticky-banner']];
 for(const [route,selector,name] of drawers){await bo(route);await page.locator(selector).first().click();await shot('02-backoffice/output/playwright/revision-'+name+'.png');}
 await bo('dashboard','a','seller');await shot('02-backoffice/output/playwright/seller-dashboard.png');
 for(const direction of ['balanced','sales','editorial']){for(const width of [1440,390]){await page.setViewportSize({width,height:width===1440?1000:844});for(const route of ['home','plp','pdp','cart','login','address','shipping','payment','result','profile']){await go('01-storefront/mockup.html?page='+route+'&direction='+direction);await shot('01-storefront/images/revision-v0-4/'+route+'-'+direction+'-'+width+'.png');}await go('01-storefront/mockup.html?page=pdp&product=parsian&scope=gold&direction='+direction);await shot('01-storefront/images/revision-v0-4/pdp-gold-'+direction+'-'+width+'.png');}}
 await page.setViewportSize({width:1440,height:1000});await go('01-storefront/mockup.html?page=home&direction=balanced&state=sticky-banner');await shot('01-storefront/images/revision-v0-4/home-sticky.png');await go('01-storefront/mockup.html?page=login&direction=balanced&state=password');await shot('01-storefront/images/revision-v0-4/login-password.png');
 await bo('attributes');await page.locator('[data-binding="موبایل:camera:card"]').uncheck();const previewDynamic=!(await page.locator('.category-preview').innerText()).includes('۴۸ مگاپیکسل');
 await go('01-storefront/mockup.html?page=pdp&direction=balanced');await page.locator('[data-memory-choice="24"]').click();const warrantySynchronized=(await page.locator('[data-seller-variant]').first().innerText()).includes('۲۴');
 return {count:captures.length,captures,errors,extraChecks:{previewDynamic,warrantySynchronized}};
}

