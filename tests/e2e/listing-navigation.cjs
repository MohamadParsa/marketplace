/* Regression for cross-page footer anchors and metadata; owned local test storefront. */
async page=>{
 const shop=await page.context().newPage(),errors=[];shop.on('pageerror',e=>errors.push(e.message));
 const response=await shop.request.get('http://127.0.0.1:5197/storefront/catalog');if(!response.ok())throw Error('Catalog unavailable');
 const data=await response.json(),category=data.navigation.find(c=>c.name==='گوشی موبایل');if(!category)throw Error('Run seed-listing.cjs first');
 await shop.goto('http://127.0.0.1:5197'+category.url);await shop.locator('.plp-grid .lh-product').first().waitFor();
 await shop.locator('a[href="/#about"]').first().click();await shop.locator('#about').waitFor();
 await shop.waitForFunction(()=>scrollY>0&&document.getElementById('about').getBoundingClientRect().top<innerHeight);
 const hash=shop.url().endsWith('/#about');
 await shop.goto('http://127.0.0.1:5197/');await shop.locator('.lh-product').first().waitFor();
 const title=await shop.title(),description=await shop.locator('meta[name=description]').getAttribute('content');await shop.close();
 const checks=[{id:'plp-footer-home-hash-scroll',pass:hash},{id:'home-seo-reset',pass:title==='Cherriva | دنیایی برای انتخاب تو'&&description.includes('چریوا؛')},{id:'no-pageerror-final-navigation',pass:errors.length===0}];
 if(checks.some(c=>!c.pass))throw Error(JSON.stringify(checks));return {checks,runtimeErrors:errors,passed:checks.length,total:checks.length}
}
