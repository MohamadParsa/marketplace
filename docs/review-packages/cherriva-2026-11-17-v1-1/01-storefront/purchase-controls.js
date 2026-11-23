/* Shared design component only. No order or remote state. */
function purchaseControl(qty=0,key='pdp',max=99,disabled=false){
 return `<div class="purchase-control ${qty?'is-added':''} ${qty>=max?'is-max':''}" data-purchase-key="${key}">${qty?`<button class="purchase-increase" data-purchase-action="increase" aria-label="افزایش تعداد، ${Number(qty).toLocaleString('fa-IR')} عدد در سبد${qty>=max?'، حداکثر موجودی':''}" ${qty>=max?'disabled':''}><span>${Number(qty).toLocaleString('fa-IR')}${qty>=max?'<small>حداکثر</small>':''}</span></button><button class="purchase-decrease" data-purchase-action="decrease" aria-label="${qty===1?'حذف از سبد':'کاهش تعداد'}">−</button>`:`<button class="purchase-initial" data-purchase-action="add" ${disabled?'disabled':''}>افزودن به سبد</button>`}</div>`;
}
