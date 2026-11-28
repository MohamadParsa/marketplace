import { useCallback } from "react";
import "../../../../../packages/commerce-client/order-requests.css";
import { OrderRequestsPanel } from "../../../../../packages/commerce-client/order-requests-panel";
import { OrderHistoryPanel } from "../../../../../packages/commerce-client/order-history-panel";
import { useWorkspace } from "../../shared/workspace";
import { apiRequest } from "../../shared/api";

export function PurchaseRequests({id,onChanged}:{id:string;onChanged?:()=>void|Promise<void>}) {
 const {authToken,surface}=useWorkspace();
 const request=useCallback(<T,>(path:string,body?:unknown,key?:string)=>apiRequest<T>(path,authToken,body,key),[authToken]);
 return <OrderRequestsPanel id={id} prefix={`/${surface}/purchases`} request={request} onChanged={onChanged}/>;
}

export function PurchaseHistory({id,version}:{id:string;version:number}) {
 const {authToken,surface}=useWorkspace();
 const request=useCallback(<T,>(path:string)=>apiRequest<T>(path,authToken),[authToken]);
 return <OrderHistoryPanel path={`/${surface}/purchases/${encodeURIComponent(id)}/history`} version={version} request={request}/>;
}
