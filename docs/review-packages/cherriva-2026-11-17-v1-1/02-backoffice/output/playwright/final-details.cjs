async (page)=>{
  await page.setViewportSize({width:1440,height:1000});
  await page.goto('http://127.0.0.1:4179/#login?role=admin&theme=a&scene=normal');
  await page.reload();
  await page.getByRole('button',{name:'دریافت کد ورود نمونه'}).click();
  await page.getByRole('textbox',{name:'کد تأیید'}).waitFor();
  await page.screenshot({path:'output/playwright/login-otp-a.png'});
  await page.goto('http://127.0.0.1:4179/#orders?role=seller&theme=a&scene=normal');
  await page.locator('[data-action="detail"][data-id="ORD-1048-1"]').first().click();
  await page.screenshot({path:'output/playwright/seller-order-private.png'});
  await page.goto('http://127.0.0.1:4179/#finance?role=finance&theme=a&scene=normal');
  await page.locator('[data-action="detail"][data-id="PAY-2081"]').click();
  await page.screenshot({path:'output/playwright/payment-unknown.png'});
  await page.goto('http://127.0.0.1:4179/#dashboard?role=admin&theme=a&scene=normal');
  return {updated:3,viewport:'1440x1000'};
}
