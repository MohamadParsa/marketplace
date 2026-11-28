export interface ProgressEvent {action:string; at:string; order_status?:string; to?:string}
export interface ProgressStep {id:string;title:string;state:"done"|"current"|"pending";at?:string}

function earliest(events:ProgressEvent[],matches:(event:ProgressEvent)=>boolean) {
  return events.filter(e=>matches(e)&&Number.isFinite(Date.parse(e.at))).map(e=>e.at).sort((a,b)=>Date.parse(a)-Date.parse(b))[0];
}
function steps(ids:string[],titles:string[],current:number,dates:(string|undefined)[],exception:boolean):ProgressStep[]{
 return ids.map((id,index)=>({id,title:titles[index],state:exception?(dates[index]?"done":"pending"):index<current?"done":index===current?"current":"pending",at:dates[index]}));
}
export function purchaseProgress(status:string,created:string,events:ProgressEvent[]):ProgressStep[]{
 const ids=["placed","processing","completed"];
 const dates=[Number.isFinite(Date.parse(created))?created:undefined,earliest(events,e=>e.order_status==="processing"),earliest(events,e=>e.order_status==="completed")];
 return steps(ids,["ثبت سفارش","در حال پردازش","تکمیل"],ids.indexOf(status),dates,!ids.includes(status));
}
export function shipmentProgress(status:string,events:ProgressEvent[]):ProgressStep[]{
 const ids=["awaiting-supply","supplied","packing","shipped","delivered"];
 const actions=["fulfillment.start","fulfillment.supply","fulfillment.pack","fulfillment.ship","fulfillment.deliver"];
 const dates=actions.map(action=>earliest(events,e=>e.action===action));
 // Exceptions use only actual milestones; cancellation/hold is not a successful final stage.
 return steps(ids,["تأمین نشده","تأمین شده","آماده‌سازی","ارسال","تحویل"],ids.indexOf(status),dates,!ids.includes(status));
}
