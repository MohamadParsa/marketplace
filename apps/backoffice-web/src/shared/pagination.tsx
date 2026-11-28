import { ChoiceField } from "./option-select";
import { useEffect, useState } from "react";
import { Button } from "@mui/material";

export function usePagedRows<T extends { id: string }>(
  rows: T[],
  initialSize = 20,
) {
  const [page, setPage] = useState(0);
  const [size, setSize] = useState(initialSize);
  const signature = rows.map((r) => r.id).join("\0");
  useEffect(() => setPage(0), [signature, size]);
  const pages = Math.max(1, Math.ceil(rows.length / size));
  const current = Math.min(page, pages - 1);
  return {
    rows: rows.slice(current * size, (current + 1) * size),
    controls: {
      page: current,
      pages,
      size,
      total: rows.length,
      setPage,
      setSize,
    },
  };
}
export function PaginationControls({
  page,
  pages,
  size,
  total,
  setPage,
  setSize,
}: {
  page: number;
  pages: number;
  size: number;
  total: number;
  setPage: (page: number) => void;
  setSize: (size: number) => void;
}) {
  const fa = (n: number) => new Intl.NumberFormat("fa-IR").format(n);
  return (
    <nav className="data-pagination" aria-label="صفحه‌بندی">
      <span aria-live="polite">
        {total
          ? `${fa(page * size + 1)} تا ${fa(Math.min((page + 1) * size, total))} از ${fa(total)}`
          : "موردی وجود ندارد"}
      </span>
      <div className="pagination-sizes">
        <ChoiceField
          label="تعداد در صفحه"
          value={String(size)}
          options={[20, 50, 100].map((n) => ({ id: String(n), name: fa(n) }))}
          onChange={(value) => setSize(Number(value))}
        />
      </div>{" "}
      <div className="pagination-pages">
        <Button disabled={page === 0} onClick={() => setPage(page - 1)}>
          قبلی
        </Button>
        <span>
          صفحه {fa(page + 1)} از {fa(pages)}
        </span>
        <Button disabled={page + 1 >= pages} onClick={() => setPage(page + 1)}>
          بعدی
        </Button>
      </div>
    </nav>
  );
}
