async (page) => {
  const errors=[];page.on('pageerror',e=>errors.push(e.message));
  const go=async url=>{await page.goto(url);await page.waitForFunction(()=>window.CVDesign&&CVDesign.state.route===(location.hash.slice(1).split('?')[0]||'login'));};
  await page.setViewportSize({width:1440,height:1000});
  await page.goto('http://127.0.0.1:4179/#dashboard?role=admin&theme=a&scene=normal');
  await page.reload();
  const pages=await page.evaluate(()=>CV.pages.map(x=>({id:x.id,code:x.code,title:x.title})));
  const results=[];
  for(const theme of ['a','b','c']) {
    for(const p of pages) {
      await go(`http://127.0.0.1:4179/#${p.id}?role=admin&theme=${theme}&scene=normal`);
      await page.evaluate(()=>document.fonts.ready);
      await page.locator('main h1').waitFor();
      const layout=await page.evaluate(()=>({overflow:document.documentElement.scrollWidth>innerWidth+1,title:document.querySelector('main h1').textContent}));
      await page.screenshot({path:`output/playwright/${p.code.toLowerCase()}-${p.id}-${theme}.png`,fullPage:true});
      results.push({route:p.id,theme,...layout});
    }
    await go(`http://127.0.0.1:4179/#login?role=admin&theme=${theme}&scene=normal`);
    await page.screenshot({path:`output/playwright/login-${theme}.png`,fullPage:true});
  }
  for(const route of ['login','dashboard','orders','pages','sale-methods','roles','finance']) {
    await page.setViewportSize({width:390,height:844});
    await go(`http://127.0.0.1:4179/#${route}?role=admin&theme=a&scene=normal`);
    await page.screenshot({path:`output/playwright/mobile-${route}.png`,fullPage:true});
    results.push({route,viewport:'390x844',overflow:await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth+1)});
  }
  await page.setViewportSize({width:1440,height:1000});
  for(const scene of ['loading','empty','error','forbidden']){
    await page.goto(`http://127.0.0.1:4179/#orders?role=admin&theme=a&scene=${scene}`);
    await page.screenshot({path:`output/playwright/state-${scene}.png`,fullPage:true});
  }
  await page.goto('http://127.0.0.1:4179/#dashboard?role=seller&theme=a&scene=normal');
  await page.screenshot({path:'output/playwright/seller-dashboard.png',fullPage:true});
  await page.goto('http://127.0.0.1:4179/#orders?role=seller&theme=a&scene=normal');
  await page.locator('[data-action="detail"][data-id="ORD-1048-1"]').first().click();
  await page.screenshot({path:'output/playwright/seller-order-private.png',fullPage:false});
  await page.goto('http://127.0.0.1:4179/#finance?role=finance&theme=a&scene=normal');
  await page.locator('[data-action="detail"][data-id="PAY-2081"]').click();
  await page.screenshot({path:'output/playwright/payment-unknown.png',fullPage:false});
  await page.goto('http://127.0.0.1:4179/#dashboard?role=admin&theme=a&scene=normal');
  return {screens:results,errors};
}
