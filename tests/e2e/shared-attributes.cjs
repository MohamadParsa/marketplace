/* Run after a fresh test API on 8095: playwright-cli run-code --filename tests/e2e/shared-attributes.cjs */
async (page) => {
  const checks=[], errors=[], consoleErrors=[];
  page.on('pageerror',e=>errors.push(e.message));
  page.on('console',m=>{if(m.type()==='error')consoleErrors.push(m.text());});
  const check=(id,pass)=>{checks.push({id,pass:!!pass});if(!pass)throw new Error(id);};
  const capture=async name=>{const heading=await page.getByRole('dialog').isVisible() ? page.getByRole('dialog').getByRole('heading').first() : page.locator('.content h1');await heading.click();await page.getByText('تغییر با موفقیت ثبت شد.',{exact:true}).waitFor({state:'hidden'});return page.screenshot({path:`output/playwright/issue-5/${name}.png`,fullPage:true,animations:'disabled'});};
  const card=()=>page.getByTestId('card-mobile-sample');
  const dialog=()=>page.getByRole('dialog');
  async function login(viewer=false){
    if(viewer)await page.getByRole('button',{name:'مشاهده‌گر',exact:true}).click();
    await page.getByRole('button',{name:'دریافت کد ورود',exact:true}).click();
    const text=await page.getByRole('alert').filter({hasText:'کد آزمون'}).innerText();
    await page.getByLabel('کد یک‌بارمصرف').fill(text.match(/\d{6}/)[0]);
    await page.getByRole('button',{name:'ورود به پنل',exact:true}).click();
    await page.locator('.content h1').waitFor();
  }
  async function navigate(href){const link=page.locator(`nav a[href="${href}"]`);const opener=page.getByRole('button',{name:'باز کردن منو',exact:true});if(await opener.isVisible() && !(await page.locator('.sidebar').getAttribute('class')).includes('is-open'))await opener.click();await link.click();}
  async function library(){await navigate('/attributes');await page.getByRole('heading',{name:'کتابخانه خصوصیات',exact:true}).waitFor();}
  async function bindings(name){
    await navigate('/categories');
    await page.locator('tbody tr').filter({hasText:name}).getByRole('button',{name:'تنظیم خصوصیات',exact:true}).click();
    await dialog().waitFor();
  }
  async function saveBindings(){await dialog().getByRole('button',{name:'ثبت انتساب‌ها',exact:true}).click();await dialog().waitFor({state:'hidden'});}
  await page.setViewportSize({width:1440,height:1000});
  await page.goto('http://127.0.0.1:8095/');await login();await library();
  check('initial-cap-three',(await card().locator('li').count())===3 && (await card().locator('li').first().innerText()).includes('حافظه'));
  await page.getByLabel('حداکثر ویژگی کارت').fill('4');await page.getByRole('button',{name:'ثبت سقف کارت',exact:true}).click();
  await card().locator('li').nth(3).waitFor();check('cap-four-consumed',(await card().locator('li').count())===4);
  await capture('library-desktop');
  await bindings('موبایل آزمایشی');
  check('axes-color-warranty',(await dialog().getByLabel('محور رنگ',{exact:true}).isChecked()) && (await dialog().getByLabel('محور گارانتی',{exact:true}).isChecked()));
  check('used-axes-locked',await dialog().getByLabel('محور رنگ',{exact:true}).isDisabled());
  await dialog().getByLabel('فیلتر رنگ',{exact:true}).uncheck();
  await dialog().getByLabel('ترتیب نمایشگر',{exact:true}).fill('0');
  await capture('category-desktop');await saveBindings();await library();
  check('card-order-from-category',(await card().locator('li').first().innerText()).includes('نمایشگر'));
  await bindings('خانه و زندگی');
  check('shared-color-flags-independent',await dialog().getByLabel('فیلتر رنگ',{exact:true}).isChecked());
  await dialog().getByRole('button',{name:'انصراف',exact:true}).click();
  await bindings('شارژر آزمایشی');
  check('no-parent-memory-inheritance',!(await dialog().getByLabel('انتساب حافظه',{exact:true}).isChecked()));
  check('same-color-on-second-category',await dialog().getByLabel('انتساب رنگ',{exact:true}).isChecked());
  await dialog().getByRole('button',{name:'انصراف',exact:true}).click();
  await library();
  await page.getByRole('button',{name:'خصوصیت جدید',exact:true}).click();
  await dialog().getByRole('textbox',{name:'نام خصوصیت',exact:true}).fill('جنس آزمایشی');
  await dialog().getByRole('textbox',{name:'کد یکتا',exact:true}).fill('material-e2e');
  await dialog().getByLabel('گزینه‌ها، هر خط یک مقدار',{exact:true}).fill('فلز\nفلز');
  await dialog().getByRole('button',{name:'ثبت خصوصیت',exact:true}).click();
  check('duplicate-options-validation',await dialog().getByText('گزینه‌ها نباید خالی یا تکراری باشند.').isVisible());
  await dialog().getByLabel('گزینه‌ها، هر خط یک مقدار',{exact:true}).fill('فلز\nپلاستیک');
  await dialog().getByRole('button',{name:'ثبت خصوصیت',exact:true}).click();await dialog().waitFor({state:'hidden'});
  check('create-library-definition',(await page.locator('tbody tr').filter({hasText:'material-e2e'}).count())===1);
  await bindings('دسته خالی آزمایشی');
  await dialog().getByLabel('انتساب جنس آزمایشی',{exact:true}).check();
  await dialog().getByLabel('محور جنس آزمایشی',{exact:true}).check();await saveBindings();
  await library();await page.locator('tbody tr').filter({hasText:'material-e2e'}).getByRole('button',{name:'مشاهده و ویرایش',exact:true}).click();
  check('used-definition-delete-locked',await dialog().getByRole('button',{name:'حذف خصوصیت بدون استفاده',exact:true}).isDisabled());
  check('used-definition-schema-locked',await dialog().getByLabel('گزینه‌ها، هر خط یک مقدار',{exact:true}).isDisabled());
  check('impact-preview',await dialog().getByText(/دسته خالی آزمایشی/).isVisible());
  await dialog().getByRole('textbox',{name:'نام خصوصیت',exact:true}).fill('جنس مشترک آزمایشی');
  await dialog().getByRole('button',{name:'ثبت خصوصیت',exact:true}).click();await dialog().waitFor({state:'hidden'});
  check('shared-definition-renamed',await page.locator('tbody tr').filter({hasText:'material-e2e'}).getByText('جنس مشترک آزمایشی',{exact:true}).isVisible());
  await page.setViewportSize({width:390,height:844});await library();
  await page.getByLabel('حداکثر ویژگی کارت').fill('2');await page.getByRole('button',{name:'ثبت سقف کارت',exact:true}).click();
  await page.waitForFunction(()=>document.querySelector('[data-testid="saved-card-limit"]')?.textContent==='2');
  check('cap-two-consumed',(await card().locator('li').count())===2);
  check('mobile-no-page-overflow',await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
  await capture('library-mobile');await bindings('موبایل آزمایشی');
  check('mobile-dialog-no-overflow',await dialog().evaluate(e=>e.scrollWidth<=e.clientWidth));
  await capture('category-mobile');await dialog().getByRole('button',{name:'انصراف',exact:true}).click();
  await page.setViewportSize({width:1440,height:1000});
  await page.getByRole('button',{name:'خروج از حساب',exact:true}).click();await login(true);await library();
  check('viewer-create-disabled',await page.getByRole('button',{name:'خصوصیت جدید',exact:true}).isDisabled());
  check('viewer-setting-disabled',await page.getByRole('button',{name:'ثبت سقف کارت',exact:true}).isDisabled());
  await bindings('شارژر آزمایشی');
  check('viewer-binding-disabled',await dialog().getByRole('button',{name:'ثبت انتساب‌ها',exact:true}).isDisabled());
  check('no-runtime-errors',errors.length===0);
  check('no-console-errors',consoleErrors.length===0);
  return {checks,passed:checks.filter(x=>x.pass).length,total:checks.length,runtimeErrors:errors,consoleErrors};
}
