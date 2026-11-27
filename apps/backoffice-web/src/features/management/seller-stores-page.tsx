import { useState } from "react";
import { useSearchParams, Navigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  LinearProgress,
} from "@mui/material";
import { DataTable, PageHeader } from "../../shared/components";
import { EditorTabs } from "../../shared/editor-tabs";
import { useWorkspace } from "../../shared/workspace";
import { ApiError } from "../../shared/api";
import {
  StoreShippingPanel,
  useStoreShippingDirectory,
} from "../commerce/store-shipping-page";
import {
  StorePreparationPage,
  type StorePreparation,
} from "./store-preparation-page";

export function LegacyStoreShippingRoute() {
  const { surface } = useWorkspace();
  const [params] = useSearchParams();
  const next = new URLSearchParams();
  if (params.get("store")) {
    next.set("store", params.get("store")!);
    next.set("tab", "shipping");
  }
  return (
    <Navigate
      replace
      to={
        (surface === "seller" ? "/my-stores" : "/stores") +
        (next.size ? "?" + next.toString() : "")
      }
    />
  );
}

export function SellerStoresPage() {
  const { authToken, surface } = useWorkspace();
  const [params, setParams] = useSearchParams();
  const [shippingBusy, setShippingBusy] = useState(false),
    [preparationBusy, setPreparationBusy] = useState(false);
  const busy = shippingBusy || preparationBusy;
  const { query: shipping } = useStoreShippingDirectory();
  const preparation = useQuery<StorePreparation[]>({
    queryKey: ["store-preparation", surface, authToken],
    queryFn: async ({ signal }) => {
      const response = await fetch(
        (import.meta.env.VITE_API_BASE_URL ?? "") +
          "/seller/stores/preparation",
        {
          signal,
          cache: "no-store",
          headers: { Authorization: "Bearer " + authToken },
        },
      );
      if (!response.ok) throw new ApiError(response.status);
      return response.json();
    },
  });
  const stores = new Map<
    string,
    { id: string; name: string; shipping: boolean; preparation: boolean }
  >();
  for (const store of shipping.data?.stores ?? [])
    stores.set(store.id, {
      id: store.id,
      name: store.name,
      shipping: true,
      preparation: false,
    });
  for (const store of preparation.data ?? [])
    stores.set(store.id, {
      id: store.id,
      name: store.name,
      shipping: stores.get(store.id)?.shipping ?? false,
      preparation: true,
    });
  const rows = [...stores.values()].sort((a, b) =>
    a.name.localeCompare(b.name, "fa"),
  );
  const selected = rows.find((store) => store.id === params.get("store"));
  const requestedTab = params.get("tab");
  const tab =
    selected?.shipping && requestedTab !== "preparation"
      ? "shipping"
      : selected?.preparation
        ? "preparation"
        : "shipping";
  const changeTab = (tab: string) => {
    const next = new URLSearchParams(params);
    next.set("tab", tab);
    setParams(next, { replace: true });
  };
  const close = () => {
    const next = new URLSearchParams(params);
    next.delete("store");
    next.delete("tab");
    setParams(next, { replace: true });
  };
  const panels = selected
    ? [
        ...(selected.shipping
          ? [
              {
                id: "shipping",
                label: "روش‌های حمل",
                content: (
                  <StoreShippingPanel
                    key={selected.id}
                    storeID={selected.id}
                    enabled={tab === "shipping"}
                    onBusyChange={setShippingBusy}
                  />
                ),
              },
            ]
          : []),
        ...(selected.preparation
          ? [
              {
                id: "preparation",
                label: "وقفه و تقویم",
                content: (
                  <StorePreparationPage
                    key={selected.id}
                    storeID={selected.id}
                    embedded
                    enabled={tab === "preparation"}
                    onBusyChange={setPreparationBusy}
                  />
                ),
              },
            ]
          : []),
      ]
    : [];
  return (
    <>
      <PageHeader
        title="فروشگاه‌های من"
        description="تنظیم روش‌های حمل، وقفه و تقویم فقط برای فروشگاه‌های مجاز شما"
      />
      {(shipping.isPending || preparation.isPending) && (
        <LinearProgress aria-label="دریافت فروشگاه‌های من" />
      )}
      {shipping.error && (
        <Alert severity="error">
          دریافت دسترسی حمل انجام نشد.{" "}
          <Button onClick={() => void shipping.refetch()}>تلاش دوباره</Button>
        </Alert>
      )}
      {preparation.error && (
        <Alert severity="error">
          دریافت دسترسی تقویم انجام نشد.{" "}
          <Button onClick={() => void preparation.refetch()}>
            تلاش دوباره
          </Button>
        </Alert>
      )}
      {!shipping.isPending &&
        !preparation.isPending &&
        !rows.length &&
        !shipping.error &&
        !preparation.error && (
          <Alert severity="info">
            هنوز فروشگاهی با مجوز تنظیمات حمل یا تقویم در دسترس نیست.
          </Alert>
        )}
      <DataTable
        rows={rows}
        searchable={false}
        columns={[
          { title: "فروشگاه", render: (store) => <b>{store.name}</b> },
          {
            title: "تنظیمات مجاز",
            render: (store) =>
              [
                store.shipping ? "روش‌های حمل" : "",
                store.preparation ? "وقفه و تقویم" : "",
              ]
                .filter(Boolean)
                .join(" · "),
          },
          {
            title: "عملیات",
            render: (store) => (
              <Button
                onClick={() => {
                  setParams({
                    store: store.id,
                    tab: store.shipping ? "shipping" : "preparation",
                  });
                }}
              >
                تنظیمات فروشگاه
              </Button>
            ),
          },
        ]}
      />
      {params.has("store") &&
        !selected &&
        !shipping.isPending &&
        !preparation.isPending && (
          <Alert severity="warning">
            فروشگاه در دسترس نیست یا مجوز مشاهدهٔ آن را ندارید.
            <Button onClick={close}>بازگشت به فروشگاه‌های من</Button>
          </Alert>
        )}
      {selected && (
        <Dialog open fullWidth maxWidth="lg" onClose={() => !busy && close()}>
          <DialogTitle>تنظیمات فروشگاه · {selected.name}</DialogTitle>
          <DialogContent>
            <EditorTabs
              label="بخش‌های فروشگاه من"
              value={tab}
              onChange={changeTab}
              disabled={busy}
              panels={panels}
            />
          </DialogContent>
          <DialogActions>
            <Button disabled={busy} onClick={close}>
              بستن
            </Button>
          </DialogActions>
        </Dialog>
      )}
    </>
  );
}
