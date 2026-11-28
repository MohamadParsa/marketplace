import { useState } from "react";
import { Alert, Button, Tab, Tabs } from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import {
  ActionDialog,
  DataTable,
  PageHeader,
  Status,
} from "../../shared/components";
import { labels, money, number, recordNumber } from "../../shared/format";
import type { Transfer } from "../../shared/types";
export function FinancePage() {
  const { state, can, execute } = useWorkspace();
  const [tab, setTab] = useState("all");
  const [selected, setSelected] = useState<Transfer | null>(null);
  const transfers = state.transfers ?? [];
  return (
    <>
      <PageHeader
        title="مرکز عملیات مالی"
        description="پرداخت، بازپرداخت، برداشت و تسویه؛ قابل ردیابی تا قلم و منشأ وجه"
      />
      <Alert severity="warning" sx={{ mb: 3 }}>
        آداپتر آزمون مالی فعال است. نتیجه‌های موفق، ناموفق و نامشخص شبیه‌سازی
        می‌شوند؛ هیچ انتقال بانکی واقعی انجام نمی‌شود.
      </Alert>
      <div className="metric-grid three">
        <div className="metric">
          <span>در انتظار تعیین نتیجه</span>
          <strong>
            {number(transfers.filter((t) => t.status === "pending").length)}
          </strong>
          <small>رزرو تا نتیجه قطعی حفظ می‌شود</small>
        </div>
        <div className="metric">
          <span>نیازمند بررسی</span>
          <strong>
            {number(transfers.filter((t) => t.status === "unknown").length)}
          </strong>
          <small>وضعیت نامشخص، موفقیت محسوب نمی‌شود</small>
        </div>
        <div className="metric">
          <span>بدهی فروشگاه‌ها</span>
          <strong className="amount">
            {money(
              transfers
                .filter((t) => t.kind === "debt")
                .reduce((sum, t) => sum + t.amount, 0),
            )}
          </strong>
          <small>ناشی از بازپرداخت پس از تسویه</small>
        </div>
      </div>
      <section className="panel">
        <Tabs
          value={tab}
          onChange={(_, value) => setTab(value)}
          variant="scrollable"
          scrollButtons="auto"
        >
          <Tab value="all" label="همه تراکنش‌ها" />
          {["payment", "refund", "withdrawal", "settlement", "debt"].map(
            (value) => (
              <Tab key={value} value={value} label={labels[value]} />
            ),
          )}
        </Tabs>
        <DataTable
          rows={transfers.filter((t) => tab === "all" || t.kind === tab)}
          columns={[
            {
              title: "مرجع",
              render: (t) => (
                <>
                  <b>{recordNumber(state, "transfer", t.id)}</b>
                  <small>{t.order_id || "کیف پول"}</small>
                </>
              ),
            },
            { title: "نوع", render: (t) => labels[t.kind] },
            { title: "مبلغ", render: (t) => money(t.amount) },
            {
              title: "منشأ / قلم",
              render: (t) => (
                <>
                  <span className="code">{t.source}</span>
                  <small>{t.item_id || "—"}</small>
                </>
              ),
            },
            {
              title: "وضعیت",
              render: (t) => (
                <>
                  <Status value={t.status} />
                  <small>{t.reserved ? "وجه در رزرو" : "بدون رزرو باز"}</small>
                </>
              ),
            },
            {
              title: "عملیات",
              render: (t) => (
                <Button
                  disabled={
                    !can("transfer.resolve", "finance") ||
                    !["pending", "unknown"].includes(t.status)
                  }
                  onClick={() => setSelected(t)}
                >
                  تعیین نتیجه آزمون
                </Button>
              ),
            },
          ]}
        />
      </section>
      <section className="panel">
        <div className="section-heading">
          <div>
            <h2>دفتر ثبت نتایج قطعی</h2>
            <p>هر نتیجه موفق یک بار ثبت می‌شود؛ retry اثر مالی مضاعف ندارد.</p>
          </div>
        </div>
        <DataTable
          rows={state.ledger ?? []}
          columns={[
            { title: "شناسه انتقال", render: (e) => recordNumber(state, "transfer", e.transfer_id) },
            { title: "نوع", render: (e) => labels[e.kind] },
            { title: "مبلغ", render: (e) => money(e.amount) },
            { title: "منشأ", render: (e) => e.source },
          ]}
        />
      </section>
      {selected && (
        <ActionDialog
          open
          title="ثبت نتیجه آداپتر مالی آزمون"
          fields={[
            {
              name: "value",
              label: "نتیجه",
              options: [
                { value: "success", label: "موفق" },
                { value: "failure", label: "شکست قطعی" },
                { value: "unknown", label: "نامشخص؛ ورود به صف بررسی" },
              ],
            },
            { name: "reason", label: "دلیل / مرجع بررسی" },
          ]}
          onClose={() => setSelected(null)}
          onSubmit={(values) =>
            execute({
              action: "transfer.resolve",
              target: selected.id,
              version: selected.version,
              ...values,
            })
          }
        />
      )}
    </>
  );
}
