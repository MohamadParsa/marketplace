import { useSearchParams } from "react-router-dom";
import { Alert } from "@mui/material";
import { PageHeader } from "../../shared/components";
import { ChoiceField } from "../../shared/option-select";
import { useWorkspace } from "../../shared/workspace";
import { ProductQA } from "./product-qa";

export function ProductQuestionsPage() {
  const { state } = useWorkspace();
  const [params, setParams] = useSearchParams();
  const id = params.get("product") ?? "";
  const product = state.products?.find((p) => p.id === id);
  return (
    <>
      <PageHeader
        title="پرسش و پاسخ محصولات"
        description="انتخاب محصول، پاسخ‌گویی و بررسی انتشار پرسش‌ها و پاسخ‌ها"
      />
      <section className="panel">
        <ChoiceField
          label="جست‌وجو و انتخاب محصول"
          value={product?.id ?? ""}
          clearLabel="انتخاب محصول"
          options={(state.products ?? []).map((p) => ({
            id: p.id,
            name: p.name + (p.public_code ? " · " + p.public_code : ""),
          }))}
          onChange={(value) => setParams(value ? { product: value } : {})}
        />
        {product ? (
          <ProductQA key={product.id} productId={product.id} />
        ) : (
          <Alert severity={id ? "warning" : "info"}>
            {id
              ? "محصول در دسترس نیست؛ محصول دیگری انتخاب کنید."
              : "برای بررسی پرسش‌ها، محصول موردنظر را انتخاب کنید."}
          </Alert>
        )}
      </section>
    </>
  );
}
