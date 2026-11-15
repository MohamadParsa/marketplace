"""تشخیص خطای runtime، پیمایش صفحات و ثبت عملیات روی داده ساختگی."""
import json
import re
from pathlib import Path
from playwright.sync_api import sync_playwright, expect

BASE='http://127.0.0.1:8087'
ARTIFACTS=Path('/tmp/marketplace-backoffice-review')
ARTIFACTS.mkdir(exist_ok=True)
with sync_playwright() as p:
    browser=p.chromium.launch(headless=True,channel="chrome")
    page=browser.new_page(viewport={'width':1440,'height':1100},device_scale_factor=1)
    failures=[]
    page.on('pageerror', lambda error: failures.append(str(error)))
    page.goto(BASE)
    page.wait_for_load_state('networkidle')
    page.screenshot(path=str(ARTIFACTS/'login.png'),full_page=True)
    page.get_by_role('button',name='دریافت کد ورود').click()
    code=page.get_by_role('alert').filter(has_text='کد آزمون').inner_text()
    code=re.search(r'\d{6}',code).group(0)
    page.get_by_label('کد یک‌بارمصرف').fill(code)
    page.get_by_role('button',name='ورود به پنل',exact=True).click()
    expect(page.get_by_role('heading',name='نمای کلی بازارگاه')).to_be_visible()
    page.screenshot(path=str(ARTIFACTS/'dashboard.png'),full_page=True)
    routes={'/users':'کاربران و نقش‌ها','/stores':'فروشگاه‌ها','/categories':'دسته‌ها و خصوصیات','/products':'محصولات و بررسی محتوا','/pricing':'قوانین کمیسیون','/shipping':'ارسال و تقویم کاری','/orders':'سفارش‌ها','/support':'پشتیبانی و پرونده‌ها','/assignments':'تأمین جایگزین','/finance':'مرکز عملیات مالی','/notifications':'اعلان‌ها','/reports':'گزارش‌ها و تاریخچه'}
    for path,title in routes.items():
        page.locator('nav a[href="'+path+'"]').click()
        expect(page.get_by_role('heading',name=title,exact=True)).to_be_visible()
        page.screenshot(path=str(ARTIFACTS/(path[1:]+'.png')),full_page=True)
    page.locator('nav a[href="/stores"]').click()
    page.get_by_role('button',name='افزودن فروشگاه').click()
    page.get_by_role('textbox',name='نام فروشگاه').fill('فروشگاه آزمون مرورگر')
    page.get_by_role('textbox',name='شهر',exact=True).fill('تهران')
    page.get_by_role('button',name='ثبت و تأیید').click()
    expect(page.get_by_role('dialog')).to_have_count(0)
    expect(page.get_by_text('فروشگاه آزمون مرورگر',exact=True)).to_be_visible()
    page.locator('nav a[href="/orders"]').click()
    page.get_by_role('link',name='ORD-1044',exact=True).click()
    expect(page.get_by_role('heading',name='جزئیات ORD-1044')).to_be_visible()
    page.screenshot(path=str(ARTIFACTS/'order-details.png'),full_page=True)
    page.locator('nav a[href="/finance"]').click()
    row=page.locator('tr').filter(has_text='TX-302')
    row.get_by_role('button',name='تعیین نتیجه آزمون').click()
    page.get_by_role('textbox',name='دلیل / مرجع بررسی').fill('نتیجه موفق آزمون مرورگر')
    page.get_by_role('button',name='ثبت و تأیید').click()
    expect(page.get_by_role('dialog')).to_have_count(0)
    expect(row.get_by_text('موفق',exact=True)).to_be_visible()
    page.locator('nav a[href="/reports"]').click()
    expect(page.get_by_text('store.create',exact=True)).to_be_visible()
    expect(page.get_by_text('transfer.resolve',exact=True)).to_be_visible()
    page.set_viewport_size({'width':390,'height':844})
    page.screenshot(path=str(ARTIFACTS/'mobile.png'),full_page=True)
    page.get_by_role('button',name='باز کردن منو').click()
    page.get_by_role('button',name='خروج از حساب').click()
    expect(page.get_by_role('heading',name='ورود به بک‌آفیس')).to_be_visible()
    assert not failures, failures
    print(json.dumps({'pages_checked':len(routes)+3,'operations':['create store','resolve financial result','audit','logout'],'runtime_errors':failures,'screenshots':str(ARTIFACTS)},ensure_ascii=False))
    browser.close()
