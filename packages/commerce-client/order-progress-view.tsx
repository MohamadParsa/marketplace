import type { ProgressStep } from "./order-progress";
export function OrderProgress({steps,label}:{steps:ProgressStep[];label:string}) {
 return <ol className="order-progress" aria-label={label}>{steps.map((step,index)=><li key={step.id} data-state={step.state} aria-current={step.state==="current"?"step":undefined}>
  <span className="order-progress-marker" aria-hidden="true">{step.state==="done"?"✓":(index+1).toLocaleString("fa-IR")}</span>
  <div><strong>{step.title}</strong><small>{step.state==="done"?"انجام شده":step.state==="current"?"مرحله فعلی":"در انتظار"}</small>
  {step.at&&<time dateTime={step.at}>{new Date(step.at).toLocaleString("fa-IR",{timeZone:"Asia/Tehran"})}</time>}</div>
 </li>)}</ol>;
}
