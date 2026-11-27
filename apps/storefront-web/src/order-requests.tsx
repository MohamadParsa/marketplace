import { useCallback } from "react";
import "../../../packages/commerce-client/order-requests.css";
import { OrderRequestsPanel } from "../../../packages/commerce-client/order-requests-panel";
import { call } from "./commerce";

export function CustomerRequests({id,token,onChanged}:{id:string;token:string;onChanged?:()=>void|Promise<void>}) {
 const request=useCallback(<T,>(path:string,body?:unknown,key?:string)=>call<T>(path,token,body,undefined,key),[token]);
 return <OrderRequestsPanel id={id} prefix="/storefront/account/orders" request={request} onChanged={onChanged}/>;
}
