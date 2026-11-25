/* Run against opt-in TestAddressBrowserHarness on5198; synthetic data only.
   Official pinned SDK JS/CSS must be downloaded to the private runtime first.
   Tile images and reverse replies are fake: no real-service activation claim. */
async (cliPage) => {
 const context=await cliPage.context().browser().newContext({viewport:{width:1440,height:1000}});
 const page=await context.newPage(), checks=[],errors=[],origin="http://localhost:5198";
 page.on("pageerror",e=>errors.push(e.message));
 const check=(id,pass)=>{checks.push({id,pass:!!pass});if(!pass)throw Error(id);};
 const request=async(path,body,token="")=>{const r=await context.request.fetch(origin+path,{method:body?"POST":"GET",headers:{Authorization:`Bearer ${token}`,"Idempotency-Key":crypto.randomUUID()},...(body?{data:body}:{})});if(!r.ok())throw Error(path+" "+r.status());return r.json();};
 const screenshot=async(name)=>{await dialog().evaluate(e=>e.scrollTop=0);return page.screenshot({path:`output/playwright/issues-100-101/${name}.png`,fullPage:true,animations:"disabled"});};
 const dialog=()=>page.getByRole("dialog");
 const fillAddress=async()=>{
  await dialog().getByLabel("عنوان آدرس").fill("خانه ساختگی");await dialog().getByLabel("استان",{exact:false}).selectOption("تهران");await dialog().getByLabel("شهر",{exact:false}).selectOption("تهران");
  await dialog().getByLabel("آدرس (خیابان و کوچه)").fill("خیابان آزمون، کوچه ساختگی");await dialog().getByLabel("پلاک",{exact:false}).fill("12");await dialog().getByLabel("کد پستی").fill("1234567890");
 };
 try{
  const challenge=await request("/storefront/auth/request-otp",{phone:"09000000003"});const login=await request("/storefront/auth/verify-otp",{request_id:challenge.request_id,code:challenge.test_code});
  await context.addInitScript(token=>sessionStorage.setItem("cherriva:customer-token",token),login.token);
  // Use the real pinned Leaflet SDK; local image tiles avoid provider calls/billing.
  await context.route(/https:\/\/.*neshan\.org\/.*/,async route=>{
   const u=route.request().url();if(u.endsWith("/index.js"))return route.fulfill({path:"C:/Users/Elias/.codex/local-runtimes/cherriva/neshan-sdk-review.js",contentType:"application/javascript"});
   if(u.endsWith("/index.css"))return route.fulfill({path:"C:/Users/Elias/.codex/local-runtimes/cherriva/neshan-sdk-review.css",contentType:"text/css"});
   return route.fulfill({path:"tests/e2e/fixtures/catalog-icon.png",contentType:"image/png"});
  });
  await page.goto(origin+"/account/addresses");await page.getByText("این آدرس قدیمی است",{exact:false}).waitFor();
  await page.getByRole("button",{name:"ویرایش",exact:true}).first().click();
  check("legacy-edit-preserves-recipient",!(await dialog().getByRole("switch").isChecked())&&await dialog().getByLabel("نام و نام خانوادگی گیرنده").inputValue()==="گیرنده قدیمی");
  await page.keyboard.press("Escape");check("modal-escape",await dialog().count()===0);
  await page.getByRole("button",{name:"افزودن آدرس",exact:true}).click();check("new-self-default-and-profile-required",await dialog().getByRole("switch").isChecked()&&await dialog().getByRole("button",{name:"ذخیره آدرس"}).isDisabled());
  await fillAddress();await dialog().getByLabel("نام و نام خانوادگی",{exact:false}).fill("مشتری ساختگی مرورگر");await dialog().getByLabel("کد ملی",{exact:false}).fill("1111111111");
  await dialog().getByRole("button",{name:"ثبت اطلاعات در پروفایل"}).click();await dialog().getByRole("alert").waitFor();check("invalid-national-id-keeps-draft",await dialog().getByLabel("پلاک").inputValue()==="12");
  await dialog().getByLabel("کد ملی",{exact:false}).fill("۰۰۱۳۵۴۷۸۹۵");await dialog().getByRole("button",{name:"ثبت اطلاعات در پروفایل"}).click();await dialog().getByText("کد ملی از پروفایل ثبت می‌شود.").waitFor();
  await screenshot("self-profile-desktop");await dialog().getByRole("button",{name:"ذخیره آدرس"}).click();await dialog().waitFor({state:"hidden"});
  let addresses=await request("/storefront/account/addresses",null,login.token);let self=addresses.find(a=>a.recipient_self===true);
  check("self-profile-snapshot-persisted",self?.recipient==="مشتری ساختگی مرورگر"&&self?.recipient_national_id==="0013547895"&&self?.phone==="09000000003"&&self?.house_number==="12");
  await page.getByRole("button",{name:"افزودن آدرس",exact:true}).click();await fillAddress();await dialog().getByRole("switch").uncheck();
  await dialog().getByLabel("نام و نام خانوادگی گیرنده").fill("گیرنده دیگر آزمون");await dialog().getByLabel("کد ملی گیرنده").fill("0013547895");
  check("other-mobile-required",await dialog().getByLabel("شماره موبایل گیرنده").getAttribute("required")!==null);
  await dialog().getByLabel("شماره موبایل گیرنده").fill("09000000111");
  await dialog().getByLabel("استان",{exact:false}).selectOption("قم");check("province-change-clears-city",await dialog().getByLabel("شهر",{exact:false}).inputValue()==="");await dialog().getByLabel("شهر",{exact:false}).selectOption("قم");
  await dialog().getByRole("switch").check();await dialog().getByRole("switch").uncheck();check("recipient-toggle-preserves-other-draft",await dialog().getByLabel("نام و نام خانوادگی گیرنده").inputValue()==="گیرنده دیگر آزمون");
  await page.setViewportSize({width:390,height:844});await screenshot("other-recipient-mobile");
  check("mobile-no-horizontal-overflow",await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)&&await dialog().evaluate(e=>e.scrollWidth<=e.clientWidth));
  await dialog().getByRole("button",{name:"ذخیره آدرس"}).click();await dialog().waitFor({state:"hidden"});addresses=await request("/storefront/account/addresses",null,login.token);check("other-recipient-persisted",addresses.some(a=>a.recipient_self===false&&a.recipient==="گیرنده دیگر آزمون"&&a.recipient_national_id==="0013547895"));
  await page.setViewportSize({width:1440,height:1000});await page.getByRole("button",{name:"افزودن آدرس",exact:true}).click();await fillAddress();
  await dialog().getByRole("button",{name:/انتخاب موقعیت روی نقشه/}).click();await dialog().locator(".leaflet-container").waitFor();
  check("real-pinned-sdk-map-initialized",await dialog().locator(".leaflet-container").count()===1);
  const canvas=dialog().locator(".address-map-canvas");await canvas.click({position:{x:180,y:100}});
  const picked=await dialog().getByLabel("عرض جغرافیایی").inputValue();check("real-sdk-click-selects-pin",picked!==""&&await dialog().locator(".address-map-pin").count()===1);
  const pin=await dialog().locator(".address-map-pin").boundingBox();if(!pin)throw Error("pin missing");await page.mouse.move(pin.x+14,pin.y+14);await page.mouse.down();await page.mouse.move(pin.x+42,pin.y+32,{steps:8});await page.mouse.up();
  check("real-sdk-drag-moves-pin",await dialog().getByLabel("عرض جغرافیایی").inputValue()!==picked);
  // Ignore AbortSignal only in this test to force genuinely out-of-order replies.
  await page.evaluate(()=>{const original=window.fetch;window.fetch=(input,init)=>original(input,String(input).includes("/address-map/reverse")?{...init,signal:undefined}:init);});
  const lat=dialog().getByLabel("عرض جغرافیایی"),lng=dialog().getByLabel("طول جغرافیایی");
  await lat.fill("35.70");await lng.fill("51.40");await dialog().getByRole("button",{name:"انتخاب این مختصات"}).click();
  await lat.fill("35.72");await dialog().getByRole("button",{name:"انتخاب این مختصات"}).click();
  await dialog().getByLabel("نشانی کامل قابل ویرایش").fill("نشانی دستی تازه مشتری");
  await dialog().locator(".address-map-suggestion").waitFor();await page.waitForTimeout(550);
  check("late-reverse-does-not-overwrite-manual",await dialog().getByLabel("نشانی کامل قابل ویرایش").inputValue()==="نشانی دستی تازه مشتری"&&await lat.inputValue()==="35.72"&&(await dialog().locator(".address-map-suggestion").innerText()).includes("35.72")&&!(await dialog().locator(".address-map-suggestion").innerText()).includes("35.70"));
  await dialog().getByRole("button",{name:"استفاده از نشانی پیشنهادی"}).click();check("explicit-suggestion-applies-active-geography",await dialog().getByLabel("نشانی کامل قابل ویرایش").inputValue()==="نشانی پیشنهادی ساختگی 35.72"&&await dialog().getByLabel("شهر",{exact:false}).inputValue()==="تهران");
  await lat.fill("36.2");await dialog().getByRole("button",{name:"انتخاب این مختصات"}).click();await dialog().getByText("استان یا شهر پیشنهادشده",{exact:false}).waitFor();
  await dialog().getByRole("button",{name:"استفاده از نشانی پیشنهادی"}).click();check("unknown-geography-keeps-valid-city",await dialog().getByLabel("شهر",{exact:false}).inputValue()==="تهران");
  await page.evaluate(()=>Object.defineProperty(navigator,"geolocation",{configurable:true,value:{getCurrentPosition:(_ok,fail)=>fail({code:1})}}));
  await dialog().getByRole("button",{name:"موقعیت فعلی من"}).click();await dialog().getByText("دسترسی به موقعیت ممکن نشد",{exact:false}).waitFor();check("geolocation-denied-manual-form-usable",await dialog().getByRole("button",{name:"ذخیره آدرس"}).isEnabled());
  await page.route("**/storefront/account/address-map/reverse",r=>r.fulfill({status:503,contentType:"application/json",body:'{"code":"map_provider_unavailable"}'}),{times:1});
  await lat.fill("35.73");await dialog().getByRole("button",{name:"انتخاب این مختصات"}).click();await dialog().getByText("دریافت نشانی انجام نشد",{exact:false}).waitFor();check("provider-failure-manual-fallback",await dialog().getByRole("button",{name:"ذخیره آدرس"}).isEnabled());
  await screenshot("map-manual-fallback-desktop");await dialog().getByRole("button",{name:"ذخیره آدرس"}).click();await dialog().waitFor({state:"hidden"});
  addresses=await request("/storefront/account/addresses",null,login.token);const mapped=addresses.find(a=>a.latitude===35.73);check("coordinates-and-final-text-persist",mapped?.longitude===51.4&&mapped?.full_address==="نشانی پیشنهادی ساختگی 36.20");
  await page.locator(".buy-address").filter({hasText:"نشانی پیشنهادی ساختگی 36.20"}).getByRole("button",{name:"ویرایش",exact:true}).click();await dialog().getByRole("button",{name:/انتخاب موقعیت روی نقشه/}).click();await dialog().locator(".leaflet-container").waitFor();
  check("edit-restores-saved-position-and-text",await dialog().getByLabel("عرض جغرافیایی").inputValue()==="35.73"&&await dialog().getByLabel("نشانی کامل قابل ویرایش").inputValue()==="نشانی پیشنهادی ساختگی 36.20");
  await dialog().getByLabel("عرض جغرافیایی").fill("35.74");await dialog().getByRole("button",{name:"انتخاب این مختصات"}).click();await dialog().getByLabel("نشانی کامل قابل ویرایش").fill("نشانی اصلاح‌شده در ویرایش");await dialog().getByRole("button",{name:"ذخیره آدرس"}).click();await dialog().waitFor({state:"hidden"});
  addresses=await request("/storefront/account/addresses",null,login.token);const edited=addresses.find(a=>a.id===mapped.id);check("edit-persists-same-address-and-updated-position",edited?.latitude===35.74&&edited?.full_address==="نشانی اصلاح‌شده در ویرایش"&&edited?.version===mapped.version+1);
  // The editor must still work when SDK load fails or no keys are configured.
  const fallback=await context.newPage();fallback.on("pageerror",e=>errors.push(e.message));
  await fallback.route("**/neshan-sdk/v1.0.8/index.js",r=>r.abort("failed"));await fallback.goto(origin+"/account/addresses");await fallback.getByRole("button",{name:"افزودن آدرس",exact:true}).click();
  await fallback.getByRole("button",{name:/انتخاب موقعیت روی نقشه/}).click();await fallback.getByText("نقشه بارگذاری نشد",{exact:false}).waitFor();
  check("sdk-load-error-preserves-manual-fields",await fallback.getByLabel("آدرس (خیابان و کوچه)").isEnabled());
  await fallback.keyboard.press("Escape");await fallback.route("**/storefront/account/address-map",r=>r.fulfill({status:200,contentType:"application/json",body:'{"enabled":false,"web_key":""}'}));
  await fallback.getByRole("button",{name:"افزودن آدرس",exact:true}).click();await fallback.getByRole("button",{name:/انتخاب موقعیت روی نقشه/}).click();await fallback.getByText("نقشه هنوز فعال نشده است",{exact:false}).waitFor();
  check("unconfigured-provider-manual-fallback",await fallback.getByLabel("آدرس (خیابان و کوچه)").isEnabled());
  await fallback.close();check("no-runtime-pageerror",errors.length===0);return {checks,errors,passed:checks.filter(x=>x.pass).length,real_pinned_sdk:true,fake_tiles_and_geocoder:true,live_provider_not_activated:true};
 }catch(e){throw Error(e.message+"; runtime="+JSON.stringify(errors)+"; passed="+checks.filter(c=>c.pass).map(c=>c.id).join(","));}finally{await context.close();}
}
