from pathlib import Path
p=Path(__file__).parent
f=p/'design.js';s=f.read_text(encoding='utf-8').replace('e.outerHTML=purchaseControl(quantity)','e.outerHTML=purchaseControl(quantity,\'pdp\',sellers[selectedSeller].stock)');f.write_text(s,encoding='utf-8')
f=p/'checkout-design.js';s=f.read_text(encoding='utf-8').replace("let method='cash';let loginPhase=checkoutState==='otp'?'otp':'phone';", "let method=checkoutState==='credit'?'credit':'cash';let loginPhase=['otp','invalid'].includes(checkoutState)?'otp':'phone';")
s=s.replace('purchaseControl(i.qty,i.id)','purchaseControl(i.qty,i.id,i.limit)')
s=s.replace("location.href=stateUrl('result','pending')", "location.href=stateUrl('result','pending')+(method==='credit'?'&method=credit':'')")
s=s.replace('>پرداخت از کیف پول</dd>','>${params.get(\'method\')===\'credit\'?\'پرداخت اعتباری\':\'پرداخت از کیف پول\'}</dd>')
# Public GooshiShop source inspired: small item strip in summary; no policy or provider copied.
s=s.replace('<dl><div><dt>قیمت کالاها</dt>', '<div class="summary-item-strip">${items.map(i=>`<div><img src="assets/${i.image}" alt="${i.product.name}"><span>${fa(i.qty)} عدد</span></div>`).join(\'\')}</div><dl><div><dt>قیمت کالاها</dt>')
f.write_text(s,encoding='utf-8')
with (p/'checkout-design.css').open('a',encoding='utf-8') as f:f.write('''
.purchase-control.is-max .purchase-increase span{line-height:1.2;font-size:15px;display:flex;flex-direction:column;align-items:center;gap:1px}.purchase-control.is-max .purchase-increase small{font-size:8px;font-weight:400}.purchase-control.is-max .purchase-increase{cursor:default;opacity:1}
.summary-item-strip{display:flex;align-items:center;gap:18px;border-block:1px solid #f2e9ed;padding:12px 0;margin:15px 0}.summary-item-strip>div{display:flex;align-items:center;gap:6px}.summary-item-strip img{height:44px;width:38px;object-fit:contain}.summary-item-strip span{font-size:10px;color:var(--muted)}
''')
print('Refined purchase controls and checkout summaries')
