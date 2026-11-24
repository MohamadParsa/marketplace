async(cliPage) => {
 const ctx=await cliPage.context().browser().newContext({viewport:{width:1440,height:900}}),page=await ctx.newPage();
 const origin='http://localhost:5197',home=await (await ctx.request.get(origin+'/storefront/home')).json();
 const banners=await(await ctx.request.get(origin+'/content/banners')).json();
 const rail=home.rails.find(r=>r.products.length),product=rail.products[0],result=[];
 try {
 for(const viewport of [{width:1440,height:900},{width:390,height:844}]){
  await page.setViewportSize(viewport);
  for(const [name,url,ready] of [['home','/','.lh-product'],['plp',rail.url,'.plp-grid .lh-product'],['pdp',product.url,'.pdp-purchase']]){
   await page.goto(origin+url);await page.locator(ready).first().waitFor();await page.evaluate(()=>document.fonts.ready);if(name==='home'&&banners.some(b=>b.config.placement!=='promo'))await page.locator('.lh-hero.with-media [data-image-state=ready]').waitFor();if(name==='pdp'){await page.locator('.pdp-layout[aria-busy="false"]').waitFor();await page.locator('.pdp-purchase .purchase-control[aria-busy="false"]').waitFor();}
   await page.waitForFunction(()=>[...document.images].filter(i=>{const r=i.getBoundingClientRect();return r.top<innerHeight&&r.bottom>0&&r.left<innerWidth&&r.right>0}).every(i=>i.complete),null,{timeout:10000});
   await page.screenshot({path:`output/playwright/issues-102-103-104/after-${name}-${viewport.width}.png`,animations:'disabled'});
   result.push({name,url,width:viewport.width,...await page.evaluate((sel)=>{const r=document.querySelector(sel).getBoundingClientRect();return{content_top:r.top,document_height:document.documentElement.scrollHeight,overflow:document.documentElement.scrollWidth>innerWidth}},ready)});
  }
 }
 return result;
 }finally{await ctx.close()}
}
