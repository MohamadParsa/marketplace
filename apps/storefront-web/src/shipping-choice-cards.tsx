import type {
  ShippingAvailability,
  DeliverySelection,
  Quote,
} from "../../../packages/commerce-client/types";
import {
  deliveryPromise,
  shippingCharge,
  shippingClock,
  shippingDate,
} from "../../../packages/commerce-client/shipping-presentation";
import { CatalogImage } from "./catalog-image";
import { apiBase } from "./commerce";
import "./shipping-choice-cards.css";

export function ShippingChoiceCards({
  row,
  methodID,
  selection,
  items,
  onMethod,
  onSelection,
  disabled,
}: {
  row: ShippingAvailability;
  methodID?: string;
  selection?: DeliverySelection;
  items: Quote["lines"];
  onMethod: (id: string) => void;
  onSelection: (value: DeliverySelection) => void;
  disabled: boolean;
}) {
  return (
    <fieldset className="buy-shipping-group" disabled={disabled}>
      <legend>{row.store_name}</legend>
      <div className="buy-shipment-thumbnails" aria-label="اقلام این مرسوله">
        {items.map((item) => (
          <span key={item.offer_id}>
            <CatalogImage
              src={item.image_url ? apiBase + item.image_url : undefined}
              alt=""
              loading="lazy"
            />
            <small>
              {item.name} × {item.quantity.toLocaleString("fa-IR")}
            </small>
          </span>
        ))}
      </div>
      {!row.methods.length && (
        <p className="buy-alert" role="status">
          {row.reason} <a href="/cart">بازبینی اقلام</a>
        </p>
      )}
      {row.methods.map((available) => {
        const m = available.method,
          selected = methodID === m.id;
        const day = available.days.find((d) => d.date === selection?.date);
        const slot = day?.slots.find(
          (s) => s.plan.slot?.window_id === selection?.window_id,
        );
        return (
          <div
            className={"buy-method-card" + (selected ? " selected" : "")}
            key={m.id}
          >
            <label className="buy-method-choice">
              <input
                type="radio"
                name={"shipping-" + row.store_id}
                checked={selected}
                onChange={() => onMethod(m.id)}
              />
              <span className="buy-method-mark" aria-hidden="true">
                {selected ? "✓" : "◇"}
              </span>
              <span>
                <b>{m.name}</b>
                <small>{shippingCharge(m.collection_mode, m.fee_toman)}</small>
                {available.range && (
                  <small>
                    {deliveryPromise(
                      available.range,
                      available.range.delivery_from,
                      available.range.delivery_to,
                    )}
                  </small>
                )}
                {!available.range && <small>تحویل با انتخاب روز و ساعت</small>}
              </span>
            </label>
            {selected && !available.range && (
              <div className="buy-delivery-selection">
                <p id={"shipping-days-" + row.store_id}>
                  روز تحویل را انتخاب کنید
                </p>
                <div
                  className="buy-delivery-days"
                  role="group"
                  aria-labelledby={"shipping-days-" + row.store_id}
                >
                  {available.days.map((d) => (
                    <button
                      type="button"
                      key={d.date}
                      disabled={!d.open}
                      aria-pressed={d.date === selection?.date}
                      onClick={() =>
                        onSelection({ date: d.date, window_id: "" })
                      }
                    >
                      <span>{shippingDate(d.date)}</span>
                      {!d.open && <small>بسته یا پر</small>}
                    </button>
                  ))}
                </div>
                {day?.open && (
                  <div
                    className="buy-delivery-hours"
                    role="group"
                    aria-label={"ساعت تحویل " + row.store_name}
                  >
                    {day.slots.map((s) => (
                      <button
                        type="button"
                        key={s.plan.slot!.window_id}
                        aria-pressed={
                          selection?.window_id === s.plan.slot!.window_id
                        }
                        onClick={() =>
                          onSelection({
                            date: day.date,
                            window_id: s.plan.slot!.window_id,
                          })
                        }
                      >
                        {shippingClock(s.plan.slot!.start_at)} تا{" "}
                        {shippingClock(s.plan.slot!.end_at)}
                      </button>
                    ))}
                  </div>
                )}
                <p className="buy-selected-delivery" aria-live="polite">
                  {slot
                    ? deliveryPromise(
                        slot.plan,
                        slot.plan.delivery_from,
                        slot.plan.delivery_to,
                      )
                    : "روز و سپس ساعت تحویل را انتخاب کنید."}
                </p>
              </div>
            )}
          </div>
        );
      })}
    </fieldset>
  );
}
