import { Alert, Button, LinearProgress } from "@mui/material";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { useState } from "react";
import { useWorkspace } from "../../shared/workspace";
import { ApiError } from "../../shared/api";
import { DataTable } from "../../shared/components";

type Row = {
  order_id: string;
  purchase_number: string;
  store_id: string;
  store_name: string;
  paid_at: string;
  deadline: string;
  overdue_minutes: number;
  lead_days: number;
  day_basis: string;
  calendar_version: number;
};
type View = { rows: Row[]; checked_at: string; missing_deadline_count: number };
const instant = (value: string) =>
  new Date(value).toLocaleString("fa-IR", {
    timeZone: "Asia/Tehran",
    dateStyle: "medium",
    timeStyle: "short",
  });
export function SupplyOverduePanel() {
  const { authToken, surface, can, role } = useWorkspace();
  const [search, setSearch] = useState("");
  const enabled =
    surface === "seller" ||
    can("commerce.orders.read") ||
    ["operations", "support", "finance"].includes(role);
  const query = useQuery({
    queryKey: ["supply-overdue", surface, authToken],
    enabled,
    refetchInterval: 60_000,
    refetchOnWindowFocus: false,
    queryFn: async () => {
      const response = await fetch(
        (import.meta.env.VITE_API_BASE_URL ?? "") +
          `/${surface}/orders/supply-overdue`,
        {
          cache: "no-store",
          headers: { Authorization: `Bearer ${authToken}` },
        },
      );
      if (!response.ok) throw new ApiError(response.status);
      return response.json() as Promise<View>;
    },
  });
  if (!enabled) return null;
  return (
    <section className="panel supply-overdue-panel" aria-label="صف تأخیر تأمین">
      <div className="page-heading">
        <div>
          <h2>صف تأخیر تأمین</h2>
          <p>
            سفارش پرداخت‌شده و منتظر تأمین، از زمان پایان مهلت ثبت‌شدهٔ
            آماده‌سازی؛ زمان تهران.
          </p>
        </div>
        <Button
          onClick={() => void query.refetch()}
          disabled={query.isFetching}
        >
          بررسی دوبارهٔ مهلت‌ها
        </Button>
      </div>
      {query.isFetching && <LinearProgress aria-label="بررسی مهلت تأمین" />}
      {query.error && (
        <Alert severity="error">دریافت صف انجام نشد؛ دوباره تلاش کنید.</Alert>
      )}
      {!query.data && !query.error ? (
        <p role="status">دریافت صف تأخیر…</p>
      ) : (
        query.data &&
        !(
          query.error instanceof ApiError &&
          [401, 403].includes(query.error.status)
        ) && (
          <>
            <p className="muted">
              آخرین بررسی: {instant(query.data.checked_at)} ·{" "}
              {query.data.rows.length.toLocaleString("fa-IR")} مرسوله
            </p>
            {!!query.data.missing_deadline_count && (
              <Alert severity="warning">
                {query.data.missing_deadline_count.toLocaleString("fa-IR")}{" "}
                سفارش منتظر تأمین، مهلت تاریخی معتبر ندارد؛ در جزئیات سفارش
                بررسی شود.
              </Alert>
            )}
            <div className="filter-bar">
              <input
                className="search"
                aria-label="جست‌وجوی صف تأخیر تأمین"
                placeholder="شماره سفارش یا فروشگاه…"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            {!query.data.rows.length && (
              <p role="status">
                در محدودهٔ مجاز شما سفارشی خارج از مهلت تأمین نیست.
              </p>
            )}
            {!!query.data.rows.length && (
              <DataTable
                searchable={false}
                rows={query.data.rows
                  .filter((r) =>
                    (
                      r.order_id +
                      " " +
                      r.purchase_number +
                      " " +
                      r.store_name
                    ).includes(
                      search.replace(/[۰-۹]/g, (c) =>
                        String(c.charCodeAt(0) - 1776),
                      ),
                    ),
                  )
                  .map((r) => ({ ...r, id: r.order_id }))}
                columns={[
                  {
                    title: "سفارش / فروشگاه",
                    render: (r) => (
                      <>
                        <b>#{r.purchase_number || r.order_id}</b>
                        <small>{r.store_name}</small>
                      </>
                    ),
                  },
                  {
                    title: "مهلت ثبت‌شده",
                    render: (r) => (
                      <>
                        <b>{instant(r.deadline)}</b>
                        <small>
                          {r.lead_days.toLocaleString("fa-IR")} روز{" "}
                          {r.day_basis === "calendar" ? "تقویمی" : "کاری"} ·
                          تقویم نسخهٔ{" "}
                          {r.calendar_version.toLocaleString("fa-IR")}
                        </small>
                      </>
                    ),
                  },
                  {
                    title: "تأخیر",
                    render: (r) =>
                      `${r.overdue_minutes.toLocaleString("fa-IR")} دقیقه`,
                  },
                  {
                    title: "مبدأ پرداخت قطعی",
                    render: (r) => instant(r.paid_at),
                  },
                  {
                    title: "پیگیری",
                    render: (r) =>
                      surface === "backoffice" ? (
                        <div className="action-group">
                          <Button
                            component={Link}
                            to={"/orders/" + encodeURIComponent(r.order_id)}
                          >
                            جزئیات سفارش
                          </Button>
                          <Button
                            component={Link}
                            to={
                              "/support?order=" + encodeURIComponent(r.order_id)
                            }
                          >
                            پرونده‌های مرتبط
                          </Button>
                        </div>
                      ) : (
                        <span>جزئیات در فهرست سفارش‌های فروشگاه</span>
                      ),
                  },
                ]}
              />
            )}
          </>
        )
      )}
      <p className="muted">
        بررسی این صف، وضعیت سفارش یا موجودی کیف پول را تغییر نمی‌دهد؛ رسیدگی از
        مسیر جزئیات سفارش و پشتیبانی انجام می‌شود.
      </p>
    </section>
  );
}
