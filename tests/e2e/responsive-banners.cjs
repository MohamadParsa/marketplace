/* Use a fresh APP_ENV=test API fixture on port 8097. Run with playwright-cli run-code --filename tests/e2e/responsive-banners.cjs */
async (page)=>{
  const checks=[],runtimeErrors=[];let uploads=0;
  page.on('pageerror',e=>runtimeErrors.push(e.message));
  page.on('request',r=>{if(r.method()==='POST'&&/\/banners\/[^/]+\/media\//.test(r.url()))uploads++;});
  const check=(id,pass)=>{checks.push({id,pass:!!pass});if(!pass)throw new Error(id);};
  const dialog=()=>page.getByRole('dialog').last();
  const slot=(name)=>dialog().getByTestId(`banner-${name}-preview`);
  const desktop='tests/e2e/fixtures/banner-desktop.png',mobile='tests/e2e/fixtures/banner-mobile.png';
  async function login(viewer=false){if(viewer)await page.getByRole('button',{name:'مشاهده‌گر',exact:true}).click();await page.getByRole('button',{name:'دریافت کد ورود',exact:true}).click();const alert=await page.getByRole('alert').filter({hasText:'کد آزمون'}).innerText();await page.getByLabel('کد یک‌بارمصرف').fill(alert.match(/\d{6}/)[0]);await page.getByRole('button',{name:'ورود به پنل',exact:true}).click();await page.locator('.content h1').waitFor();}
  async function upload(name,file){const title=name==='desktop'?'دسکتاپ':'موبایل';await dialog().getByLabel(`فایل بنر ${title}`).setInputFiles(file);const save=dialog().getByRole('button',{name:`ثبت تصویر ${title}`,exact:true});await save.waitFor();const wait=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().includes(`/media/${name}`));await save.click();const response=await wait;if(response.status()!==200)throw new Error('upload failed');await save.waitFor({state:'hidden'});return (await response.json()).asset_id;}
  async function capture(name){await dialog().getByRole('heading').first().click();await page.getByText('تغییر با موفقیت ثبت شد.',{exact:true}).waitFor({state:'hidden'});await page.screenshot({path:`output/playwright/issue-7/${name}.png`,fullPage:true,animations:'disabled'});}
  async function publicRows(){return (await page.request.get('http://127.0.0.1:8097/content/banners')).json();}
  await page.setViewportSize({width:1440,height:1000});await page.goto('http://127.0.0.1:8097/');await login();await page.locator('nav a[href="/banners"]').click();
  await page.getByText('هنوز بنری تعریف نشده است.',{exact:true}).waitFor();check('empty-banner-state',await page.getByText('هنوز بنری تعریف نشده است.',{exact:true}).isVisible());
  await page.getByRole('button',{name:'بنر جدید',exact:true}).click();await dialog().getByRole('textbox',{name:'نام بنر',exact:true}).fill('بنر آزمون دو اندازه');await dialog().getByRole('button',{name:'ساخت پیش‌نویس',exact:true}).click();await page.getByRole('heading',{name:'بنر جدید',exact:true}).waitFor({state:'hidden'});
  await page.getByRole('button',{name:'ویرایش و پیش‌نمایش',exact:true}).click();
  check('incomplete-publish-disabled',await dialog().getByRole('button',{name:'انتشار نسخه',exact:true}).isDisabled());
  await dialog().getByLabel('فایل بنر دسکتاپ').setInputFiles(desktop);await dialog().getByRole('button',{name:'ثبت تصویر دسکتاپ',exact:true}).waitFor();
  check('local-selection-does-not-upload',uploads===0);await dialog().getByRole('button',{name:'لغو انتخاب دسکتاپ',exact:true}).click();
  check('cancel-keeps-public-empty',uploads===0&&(await publicRows()).length===0);
  const desktopID=await upload('desktop',desktop),mobileID=await upload('mobile',mobile);
  await slot('desktop').locator('img').waitFor();await slot('mobile').locator('img').waitFor();
  await page.waitForFunction(()=>{const d=document.querySelector('[data-testid="banner-desktop-preview"] img'),m=document.querySelector('[data-testid="banner-mobile-preview"] img');return d?.naturalWidth===1440&&m?.naturalWidth===600;});
  check('independent-preview-images',await slot('desktop').locator('img').evaluate(e=>e.naturalWidth===1440)&&await slot('mobile').locator('img').evaluate(e=>e.naturalHeight===800));
  check('draft-assets-not-public',(await page.request.get('http://127.0.0.1:8097/content/banner-media/'+desktopID)).status()===404&&(await page.request.get('http://127.0.0.1:8097/content/banner-media/'+mobileID)).status()===404);
  await dialog().getByRole('textbox',{name:'متن جایگزین تصویر',exact:true}).fill('چریوا، بنر آزمایشی');await dialog().getByRole('textbox',{name:'لینک داخلی (اختیاری)',exact:true}).fill('/products/mobile-sample');await dialog().getByRole('button',{name:'ذخیره پیش‌نویس',exact:true}).click();
  await dialog().getByText('پیش‌نویس ذخیره شد؛ نسخه عمومی تغییر نکرد.',{exact:true}).waitFor();
  check('saved-draft-not-public',(await publicRows()).length===0);
  await capture('editor-desktop');
  await dialog().getByRole('button',{name:'انتشار نسخه',exact:true}).click();await dialog().getByText('نسخه تازه منتشر شد.',{exact:true}).waitFor();
  const first=(await publicRows())[0];check('manual-publish-two-assets',first.revision===1&&first.config.desktop_id===desktopID&&first.config.mobile_id===mobileID);
  const publicPage=await page.context().newPage();await publicPage.setViewportSize({width:1440,height:900});await publicPage.goto(`http://127.0.0.1:8097/content/banners/${first.id}/view`);await publicPage.locator('img').evaluate(e=>e.decode());
  check('public-desktop-source',(await publicPage.locator('img').evaluate(e=>e.currentSrc)).endsWith(desktopID));
  check('public-view-noindex',await publicPage.locator('meta[name="robots"]').getAttribute('content')==='noindex,nofollow');
  await publicPage.screenshot({path:'output/playwright/issue-7/public-desktop.png',fullPage:true,animations:'disabled'});
  await publicPage.setViewportSize({width:390,height:844});await publicPage.locator('img').evaluate(e=>e.decode());await publicPage.waitForFunction(id=>document.querySelector('img')?.currentSrc.endsWith(id),mobileID);
  check('public-mobile-source',(await publicPage.locator('img').evaluate(e=>e.currentSrc)).endsWith(mobileID));await publicPage.screenshot({path:'output/playwright/issue-7/public-mobile.png',fullPage:true,animations:'disabled'});
  await page.bringToFront();await dialog().getByRole('textbox',{name:'لینک داخلی (اختیاری)',exact:true}).fill('javascript:alert(1)');await dialog().getByRole('button',{name:'ذخیره پیش‌نویس',exact:true}).click();await dialog().getByRole('button',{name:'بازخوانی و بازبینی',exact:true}).waitFor();
  check('unsafe-link-rejected-draft-preserved',(await dialog().getByRole('textbox',{name:'لینک داخلی (اختیاری)',exact:true}).inputValue())==='javascript:alert(1)'&&(await publicRows())[0].revision===1);
  await dialog().getByRole('textbox',{name:'لینک داخلی (اختیاری)',exact:true}).fill('/products/mobile-sample');
  await dialog().getByLabel('فایل بنر دسکتاپ').setInputFiles('tests/e2e/fixtures/invalid-icon.png');await slot('desktop').getByText(/تصویر خراب است/).waitFor();
  check('invalid-file-keeps-published-media',(await publicRows())[0].config.desktop_id===desktopID&&await slot('desktop').locator('img').evaluate(e=>e.naturalWidth===1440));
  const replacement=await upload('desktop',desktop);
  check('replacement-draft-preserves-old-public',(await publicRows())[0].config.desktop_id===desktopID&&(await page.request.get('http://127.0.0.1:8097/content/banner-media/'+replacement)).status()===404);
  await dialog().getByRole('button',{name:'انتشار نسخه',exact:true}).click();await page.waitForFunction(async id=>{const x=await (await fetch('/content/banners')).json();return x[0]?.config.desktop_id===id;},replacement);
  check('second-publication-atomic',(await publicRows())[0].revision===2&&(await publicRows())[0].config.desktop_id===replacement);
  await page.setViewportSize({width:390,height:844});await capture('editor-mobile');
  check('mobile-editor-no-overflow',await dialog().evaluate(e=>e.scrollWidth<=e.clientWidth));
  await dialog().getByRole('button',{name:'حذف تصویر موبایل',exact:true}).click();await dialog().getByText('تصویر از پیش‌نویس حذف شد؛ نسخه عمومی حفظ شد.',{exact:true}).waitFor();
  check('missing-media-blocks-publish-preserves-public',await dialog().getByRole('button',{name:'انتشار نسخه',exact:true}).isDisabled()&&(await publicRows())[0].revision===2);
  await dialog().getByRole('button',{name:'بستن',exact:true}).click();await page.setViewportSize({width:1440,height:1000});await page.getByRole('button',{name:'خروج از حساب',exact:true}).click();await login(true);await page.locator('nav a[href="/banners"]').click();
  check('viewer-create-disabled',await page.getByRole('button',{name:'بنر جدید',exact:true}).isDisabled());await page.getByRole('button',{name:'ویرایش و پیش‌نمایش',exact:true}).click();
  check('viewer-publish-upload-disabled',await dialog().getByRole('button',{name:'انتشار نسخه',exact:true}).isDisabled()&&await dialog().getByRole('button',{name:'انتخاب تصویر دسکتاپ',exact:true}).isDisabled());
  check('no-runtime-errors',runtimeErrors.length===0);await publicPage.close();
  return {checks,passed:checks.filter(x=>x.pass).length,total:checks.length,runtimeErrors,uploads};
}
