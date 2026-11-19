from pathlib import Path
p=Path(__file__).parent
f=p/'design.js';s=f.read_text(encoding='utf-8')
s=s.replace("const page = ['home','plp','pdp'].includes(params.get('page'))", "const page = ['home','plp','pdp','cart','login','address','shipping','payment','result'].includes(params.get('page'))")
s=s.replace('<div class="quantity-cart"><div class="quantity-stepper"><button id="qty-plus" aria-label="افزایش تعداد">+</button><span id="quantity">۱</span><button id="qty-minus" aria-label="کاهش تعداد">−</button></div><button class="primary add-cart" data-add>افزودن به سبد خرید</button></div>', '<div class="quantity-cart">${purchaseControl()}</div><a class="purchase-cart-link" href="${urls(\'cart\')}">مشاهده سبد خرید ${icon(\'arrow\')}</a>')
s=s.replace('<button class="primary add-cart" data-add>افزودن به سبد خرید</button><div><span id="mobile-price">', '${purchaseControl()}<div class="mobile-price-copy"><span id="mobile-price">')
s=s.replace("page==='plp'?plp():pdp()", "page==='plp'?plp():page==='pdp'?pdp():'<div id=\"checkout-root\"></div>'+footer()")
start=s.index("document.getElementById('cart-button').addEventListener")
end=s.index("document.getElementById('newsletter-form')",start)
s=s[:start]+"document.getElementById('cart-button').addEventListener('click',()=>location.href=urls('cart'));\n"+s[end:]
s=s.replace("let quantity=1;let memory=256;let color=product.type==='gold'?'یک گرم':'تیتانیوم طبیعی';", """let quantity=0;let memory=256;let color=product.type==='gold'?'یک گرم':'تیتانیوم طبیعی';
 const basketQuantities=new Map();
 function activeOfferKey(){return [product.id,color,memory,selectedSeller].join('|')}
 function syncPurchase(){quantity=basketQuantities.get(activeOfferKey())||0;document.querySelectorAll('[data-purchase-key="pdp"]').forEach(e=>e.outerHTML=purchaseControl(quantity));cartCount=[...basketQuantities.values()].reduce((a,b)=>a+b,0);document.getElementById('cart-count').textContent=fa(cartCount)}
 document.addEventListener('click',e=>{const b=e.target.closest('[data-purchase-action]');if(!b||!b.closest('[data-purchase-key="pdp"]'))return;const key=activeOfferKey();const old=basketQuantities.get(key)||0;const next=b.dataset.purchaseAction==='decrease'?Math.max(0,old-1):Math.min(sellers[selectedSeller].stock,old+1);if(next===old&&next>0){notify('به سقف موجودی این عرضه رسیده‌ای.');return}if(next)basketQuantities.set(key,next);else basketQuantities.delete(key);syncPurchase();notify(next?'تعداد همین تنوع و فروشنده در سبد: '+fa(next):'این انتخاب از سبد حذف شد.');});
""")
s=s.replace("document.getElementById('quantity').textContent=fa(quantity)","syncPurchase()")
start=s.index(" document.getElementById('qty-plus')")
end=s.index(" document.querySelectorAll('[data-anchor]')",start)
s=s[:start]+s[end:]
needle="if(params.get('state')==='search')"
s=s.replace(needle,"if(page==='pdp'&&params.get('state')==='added'){document.querySelector('[data-purchase-action=add]').click()}\n"+needle)
f.write_text(s,encoding='utf-8')
f=p/'mockup.html';s=f.read_text(encoding='utf-8').replace('<script src="design.js"></script>','<script src="purchase-controls.js"></script><script src="design.js"></script><script src="checkout-design.js"></script>');s=s.replace('</head>','<link rel="stylesheet" href="checkout-design.css">\n</head>');f.write_text(s,encoding='utf-8')
print('Extended design artboards only')
