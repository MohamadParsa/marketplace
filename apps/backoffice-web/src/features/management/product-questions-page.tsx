import { useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Alert, Button, CircularProgress, Stack } from "@mui/material";
import { PageHeader } from "../../shared/components";
import { ChoiceField } from "../../shared/option-select";
import { FilterPanel } from "../../shared/filter-panel";
import { useWorkspace } from "../../shared/workspace";
import { api } from "../../shared/api";
import { QuestionEditor } from "./product-qa";

export function ProductQuestionsPage() {
  const { state, authToken } = useWorkspace();
  const [params, setParams] = useSearchParams();
  const product = params.get("product") ?? "";
  const status = params.get("status") ?? "pending";
  const [draftProduct, setDraftProduct] = useState(product);
  const [draftStatus, setDraftStatus] = useState(status);
  const [cursors, setCursors] = useState([0]);
  const after = cursors[cursors.length - 1];
  const query = useQuery({
    queryKey: ["product-questions", authToken, product, status, after],
    queryFn: () => api.productQuestions(authToken, product, status, after),
  });
  return <>
    <PageHeader title="پرسش و پاسخ محصولات" description="بررسی پرسش‌های تمام محصولات و پاسخ‌گویی؛ محصول فیلتر اختیاری است." />
    <section className="panel">
      <FilterPanel onSearch={() => {
        setCursors([0]);
        setParams({ ...(draftProduct ? { product: draftProduct } : {}), status: draftStatus });
        if (product === draftProduct && status === draftStatus) void query.refetch();
      }}>
        <ChoiceField label="وضعیت بررسی پرسش‌ها" value={draftStatus} clearLabel="همه وضعیت‌ها" options={[
          { id: "pending", name: "در انتظار بررسی" }, { id: "approved", name: "تأییدشده" }, { id: "rejected", name: "ردشده" },
        ]} onChange={setDraftStatus} />
        <ChoiceField label="محصول" value={draftProduct} clearLabel="همه محصولات" options={(state.products ?? []).map(p => ({ id: p.id, name: p.name }))} onChange={setDraftProduct} />
      </FilterPanel>
      <Stack spacing={2}>
        {query.isPending && <CircularProgress size={24} aria-label="در حال دریافت پرسش‌ها" />}
        {query.error && <Alert severity="error" action={<Button onClick={() => void query.refetch()}>تلاش مجدد</Button>}>{query.error.message}</Alert>}
        {query.data?.questions.length === 0 && <p>پرسشی با این فیلتر وجود ندارد.</p>}
        {query.data?.questions.map(q => <Stack key={q.id} spacing={1}>
          <Link to={"/products/" + encodeURIComponent(q.product_id!)}>{q.product_name || "محصول در دسترس نیست"}</Link>
          <QuestionEditor productId={q.product_id!} question={q} canReply={!!q.can_reply} refresh={query.refetch} />
        </Stack>)}
        {(cursors.length > 1 || query.data?.next_cursor) && <nav aria-label="صفحه‌های پرسش و پاسخ">
          <Button disabled={cursors.length < 2} onClick={() => setCursors(v => v.slice(0, -1))}>صفحه قبل</Button>
          <Button disabled={!query.data?.next_cursor} onClick={() => { const next = query.data?.next_cursor; if (next) setCursors(v => [...v, next]); }}>صفحه بعد</Button>
        </nav>}
      </Stack>
    </section>
  </>;
}
