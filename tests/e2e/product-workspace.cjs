/* Start a fresh API fixture on 8096, then use playwright-cli run-code --filename tests/e2e/product-workspace.cjs */
async (page) => {
  const checks=[],runtimeErrors=[];
  page.on('pageerror',e=>runtimeErrors.push(e.message));
  const check=(id,pass)=>{checks.push({id,pass:!!pass});if(!pass)throw new Error(id);};
  const dialog=()=>page.getByRole('dialog').last();
  const panel=()=>page.getByTestId('product-workspace-mobile-sample');
  const row=(id)=>panel().locator('tbody tr').filter({has:page.getByText(id,{exact:true})});
  async function login(account){
    if(account)await page.getByRole('button',{name:account,exact:true}).click();
    await page.getByRole('button',{name:'دریافت کد ورود',exact:true}).click();
    const text=await page.getByRole('alert').filter({hasText:'کد آزمون'}).innerText();
    await page.getByLabel('کد یک‌بارمصرف').fill(text.match(/\d{6}/)[0]);
    await page.getByRole('button',{name:'ورود به پنل',exact:true}).click();
    await page.locator('.content h1').waitFor();
  }
  async function navigateProducts(){
    const opener=page.getByRole('button',{name:'باز کردن منو',exact:true});
    if(await opener.isVisible()&&!(await page.locator('.sidebar').getAttribute('class')).includes('is-open'))await opener.click();
    await page.locator('nav a[href="/products"]').click();
  }
  const productRow=()=>page.locator('tbody tr').filter({has:page.getByText('گوشی نمونه برای بررسی کارت',{exact:true})});
  async function capture(name){await page.locator('.content h1').click();await page.getByText('تغییر با موفقیت ثبت شد.',{exact:true}).waitFor({state:'hidden'});await page.screenshot({path:`output/playwright/issue-6/${name}.png`,fullPage:true,animations:'disabled'});}
  await page.setViewportSize({width:1440,height:1000});
  await page.goto('http://127.0.0.1:8096/');await login();await navigateProducts();
  await productRow().getByRole('button',{name:'قیمت و موجودی',exact:true}).click();
  await row('mob-a-black').waitFor();
  check('grid-shared-offer-workspace',(await panel().locator('tbody tr').count())===4);
  await row('mob-a-black').getByRole('button',{name:'ویرایش قیمت',exact:true}).click();
  await dialog().getByRole('spinbutton',{name:'قیمت پایه نقدی (تومان)',exact:true}).fill('12345000');
  await dialog().getByRole('button',{name:'ثبت تغییر عرضه',exact:true}).click();
  await page.getByRole('heading',{name:'ویرایش قیمت عرضه',exact:true}).waitFor({state:'hidden'});
  check('grid-price-saved-to-offer',(await row('mob-a-black').innerText()).includes('۱۲٬۳۴۵٬۰۰۰'));
  await page.getByRole('button',{name:'بستن',exact:true}).click();
  await productRow().getByRole('link',{name:'جزئیات محصول',exact:true}).click();await panel().waitFor();
  await panel().getByRole('tab',{name:'قیمت و موجودی',exact:true}).click();
  check('details-same-offer-and-price',(await row('mob-a-black').innerText()).includes('۱۲٬۳۴۵٬۰۰۰'));
  await row('mob-a-black').getByRole('button',{name:'ویرایش موجودی',exact:true}).click();
  check('stock-reservation-visible',await dialog().getByText('رزروشده: 2',{exact:true}).isVisible());
  const stockInput=dialog().getByRole('spinbutton',{name:'موجودی کل',exact:true});
  await stockInput.fill('1');
  check('stock-below-reservation-client-blocked',!(await stockInput.evaluate(e=>e.checkValidity())));
  await stockInput.fill('7');await dialog().getByRole('button',{name:'ثبت تغییر عرضه',exact:true}).click();
  await page.getByRole('heading',{name:'ویرایش موجودی عرضه',exact:true}).waitFor({state:'hidden'});
  check('details-stock-saved',(await row('mob-a-black').innerText()).includes('7 / 2 / 5'));
  await capture('offers-desktop');
  let raced=false;
  await page.route('**/backoffice/commands',async route=>{
    const c=route.request().postDataJSON();
    if(!raced&&c.action==='offer.price.update'&&c.target==='mob-a-black'){
      raced=true;
      const response=await page.request.post(route.request().url(),{headers:route.request().headers(),data:{...c,idempotency_key:'browser-racing-price',offer:{price:123400000}}});
      if(response.status()!==204)throw new Error('concurrent change failed');
    }
    await route.continue();
  });
  await row('mob-a-black').getByRole('button',{name:'ویرایش قیمت',exact:true}).click();
  await dialog().getByRole('spinbutton',{name:'قیمت پایه نقدی (تومان)',exact:true}).fill('12350000');
  await dialog().getByRole('button',{name:'ثبت تغییر عرضه',exact:true}).click();
  await dialog().getByRole('button',{name:'بازخوانی و بازبینی',exact:true}).waitFor();
  check('stale-version-preserves-draft',(await dialog().getByRole('spinbutton',{name:'قیمت پایه نقدی (تومان)',exact:true}).inputValue())==='12350000');
  check('stale-version-requires-review',await dialog().getByRole('button',{name:'ثبت تغییر عرضه',exact:true}).isDisabled());
  await dialog().getByRole('button',{name:'بازخوانی و بازبینی',exact:true}).click();
  await dialog().getByText(/نسخه تازه دریافت شد/).waitFor();
  await dialog().getByRole('button',{name:'ثبت تغییر عرضه',exact:true}).click();
  await page.getByRole('heading',{name:'ویرایش قیمت عرضه',exact:true}).waitFor({state:'hidden'});
  await page.unroute('**/backoffice/commands');
  check('conflict-reviewed-and-retried',(await row('mob-a-black').innerText()).includes('۱۲٬۳۵۰٬۰۰۰'));
  await panel().getByRole('tab',{name:'تنوع‌ها',exact:true}).click();
  await panel().getByRole('button',{name:'افزودن تنوع',exact:true}).click();
  await dialog().getByRole('textbox',{name:'SKU تنوع',exact:true}).fill('E2E-MOBILE-BLUE18');
  await dialog().getByRole('combobox',{name:'رنگ',exact:true}).click();await page.getByRole('option',{name:'آبی',exact:true}).click();
  await dialog().getByRole('combobox',{name:'گارانتی',exact:true}).click();await page.getByRole('option',{name:'۱۸ ماهه',exact:true}).click();
  await dialog().getByRole('button',{name:'ثبت تنوع',exact:true}).click();await page.getByRole('heading',{name:'افزودن تنوع محصول',exact:true}).waitFor({state:'hidden'});
  check('create-variant-from-dynamic-axes',await panel().getByText('E2E-MOBILE-BLUE18',{exact:true}).isVisible());
  await panel().getByRole('tab',{name:'قیمت و موجودی',exact:true}).click();await panel().getByRole('button',{name:'افزودن عرضه',exact:true}).click();
  await dialog().getByRole('combobox',{name:/^تنوع عرضه/}).click();await page.getByRole('option',{name:/E2E-MOBILE-BLUE18/}).click();
  await dialog().getByRole('spinbutton',{name:'قیمت پایه نقدی (تومان)',exact:true}).fill('12000000');
  await dialog().getByRole('spinbutton',{name:'موجودی کل',exact:true}).fill('4');
  check('seller-lead-read-only',(await dialog().getByRole('spinbutton').count())===2 && (await dialog().innerText()).includes('فقط خواندنی'));
  await dialog().getByRole('button',{name:'ثبت عرضه',exact:true}).click();await page.getByRole('heading',{name:'افزودن عرضه فروشنده',exact:true}).waitFor({state:'hidden'});
  check('new-offer-on-existing-variant',(await panel().locator('tbody tr').count())===5);
  await panel().getByRole('tab',{name:'پیش‌نمایش انتخاب',exact:true}).click();
  await page.getByTestId('resolved-variant').getByText('mobile-black-18',{exact:true}).waitFor();
  check('black-selection-exact-variant',(await page.getByTestId('resolved-variant').locator('tbody tr').count())===2);
  await panel().getByRole('button',{name:'رنگ: سفید',exact:true}).click();await panel().getByRole('button',{name:'گارانتی: ۲۴ ماهه',exact:true}).click();
  await page.getByTestId('resolved-variant').getByText('mobile-white-24',{exact:true}).waitFor();
  check('white-selection-excludes-fully-reserved',(await page.getByTestId('resolved-variant').locator('tbody tr').count())===1);
  await capture('variant-preview-desktop');
  await page.setViewportSize({width:390,height:844});await panel().getByRole('tab',{name:'قیمت و موجودی',exact:true}).click();
  check('mobile-page-no-overflow',await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));await capture('offers-mobile');
  await row('mob-a-black').getByRole('button',{name:'ویرایش موجودی',exact:true}).click();
  check('mobile-form-no-overflow',await dialog().evaluate(e=>e.scrollWidth<=e.clientWidth));
  await dialog().getByRole('heading').click();await page.screenshot({path:'output/playwright/issue-6/stock-mobile.png',fullPage:true,animations:'disabled'});
  await dialog().getByRole('button',{name:'انصراف',exact:true}).click();
  await page.setViewportSize({width:1440,height:1000});await page.getByRole('button',{name:'خروج از حساب',exact:true}).click();
  await login('فروشنده الف');await navigateProducts();
  await productRow().getByRole('button',{name:'قیمت و موجودی',exact:true}).click();await row('mob-a-black').waitFor();
  check('seller-only-own-offers',(await panel().locator('tbody tr').count())===3 && !(await panel().getByText('mob-b-black',{exact:true}).count()));
  await row('mob-a-black').getByRole('button',{name:'ویرایش موجودی',exact:true}).click();await dialog().getByRole('spinbutton',{name:'موجودی کل',exact:true}).fill('8');
  await dialog().getByRole('button',{name:'ثبت تغییر عرضه',exact:true}).click();await page.getByRole('heading',{name:'ویرایش موجودی عرضه',exact:true}).waitFor({state:'hidden'});
  check('seller-own-stock-write',(await row('mob-a-black').innerText()).includes('8 / 2 / 6'));
  await panel().getByRole('tab',{name:'تنوع‌ها',exact:true}).click();check('seller-cannot-edit-global-variants',await panel().getByRole('button',{name:'افزودن تنوع',exact:true}).isDisabled());
  await page.getByRole('button',{name:'بستن',exact:true}).click();await page.getByRole('button',{name:'خروج از حساب',exact:true}).click();
  await login('مشاهده‌گر');await navigateProducts();await productRow().getByRole('button',{name:'قیمت و موجودی',exact:true}).click();await row('mob-a-black').waitFor();
  check('viewer-price-and-stock-disabled',await row('mob-a-black').getByRole('button',{name:'ویرایش قیمت',exact:true}).isDisabled()&&await row('mob-a-black').getByRole('button',{name:'ویرایش موجودی',exact:true}).isDisabled());
  check('no-runtime-errors',runtimeErrors.length===0);
  return {checks,passed:checks.filter(x=>x.pass).length,total:checks.length,runtimeErrors,realConcurrentConflict:raced};
}
