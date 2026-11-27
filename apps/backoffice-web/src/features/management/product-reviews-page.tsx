import { ChoiceField } from "../../shared/option-select";
import { FilterPanel } from "../../shared/filter-panel";
import { useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Paper,
  Stack,
  Tab,
  Tabs,
  TextField,
  Typography,
} from "@mui/material";
import { api, ApiError } from "../../shared/api";
import type {
  AdminProductReview,
  ProductReviewStatus,
  ReviewAction,
} from "../../shared/types";

const statusTabs: { value: ProductReviewStatus; label: string }[] = [
  { value: "pending", label: "در انتظار بررسی" },
  { value: "approved", label: "تأییدشده" },
  { value: "rejected", label: "ردشده" },
];

const date = new Intl.DateTimeFormat("fa-IR", { dateStyle: "medium" });
const number = new Intl.NumberFormat("fa-IR");
const reviewTextStyle = { whiteSpace: "pre-wrap", overflowWrap: "anywhere" };

export function ProductReviewsPage({ token }: { token: string }) {
  const cache = useQueryClient();
  const [status, setStatus] = useState<ProductReviewStatus>("pending");
  const [draftStatus, setDraftStatus] =
    useState<ProductReviewStatus>("pending");
  const [cursors, setCursors] = useState<string[]>([""]);
  const cursor = cursors[cursors.length - 1];
  const query = useQuery({
    queryKey: ["product-reviews", token, status, cursor],
    queryFn: () => api.productReviews(token, status, cursor),
  });
  const page = query.data;

  async function refresh() {
    await cache.invalidateQueries({ queryKey: ["product-reviews", token] });
  }

  return (
    <section className="panel detail">
      <Typography component="h1" variant="h5">
        نظرات محصولات
      </Typography>
      <FilterPanel
        onSearch={() => {
          setStatus(draftStatus);
          setCursors([""]);
          if (status === draftStatus) void query.refetch();
        }}
      >
        <ChoiceField
          label="وضعیت بررسی نظرات"
          value={draftStatus}
          options={statusTabs.map((t) => ({ id: t.value, name: t.label }))}
          onChange={(value) => setDraftStatus(value as ProductReviewStatus)}
        />
      </FilterPanel>
      {query.isPending && (
        <CircularProgress size={24} aria-label="در حال دریافت نظرات" />
      )}
      {query.error && (
        <Alert
          severity="error"
          action={
            <Button onClick={() => void query.refetch()}>تلاش دوباره</Button>
          }
        >
          {query.error.message}
        </Alert>
      )}
      {page && (
        <Stack spacing={2} sx={{ marginBlockStart: 2 }}>
          {!page.reviews.length && (
            <Typography>نظری در این بخش وجود ندارد.</Typography>
          )}
          {page.reviews.map((review) => (
            <ReviewDecisionCard
              key={review.id}
              review={review}
              token={token}
              refresh={refresh}
            />
          ))}
          {(cursors.length > 1 || page.next_cursor) && (
            <Stack
              component="nav"
              aria-label="صفحه‌بندی نظرات"
              direction="row"
              spacing={1}
            >
              <Button
                disabled={cursors.length === 1}
                onClick={() => setCursors((current) => current.slice(0, -1))}
              >
                قبلی
              </Button>
              <Button
                disabled={!page.next_cursor}
                onClick={() =>
                  setCursors((current) => [...current, page.next_cursor!])
                }
              >
                بعدی
              </Button>
            </Stack>
          )}
        </Stack>
      )}
    </section>
  );
}

function ReviewDecisionCard({
  review,
  token,
  refresh,
}: {
  review: AdminProductReview;
  token: string;
  refresh: () => Promise<void>;
}) {
  const [text, setText] = useState("");
  const [formAction, setFormAction] = useState<ReviewAction>();
  const [retry, setRetry] = useState<{
    action: ReviewAction;
    text: string;
    version: number;
    key: string;
  }>();
  const mutation = useMutation({
    mutationFn: (decision: {
      action: ReviewAction;
      text: string;
      version: number;
      idempotency_key: string;
    }) => api.decideProductReview(token, review.id, decision),
    onSuccess: async () => {
      setText("");
      setFormAction(undefined);
      setRetry(undefined);
      await refresh();
    },
  });

  function decide(action: ReviewAction) {
    if (mutation.isPending) return;
    const value = action === "approve" ? "" : text.trim();
    if (action !== "approve" && !value) return;
    const version = review.version;
    const key =
      retry?.action === action &&
      retry.text === value &&
      retry.version === version
        ? retry.key
        : crypto.randomUUID();
    setRetry({ action, text: value, version, key });
    mutation.mutate({ action, text: value, version, idempotency_key: key });
  }

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (formAction) decide(formAction);
  }

  return (
    <Paper component="article" variant="outlined" sx={{ p: 2 }}>
      <Typography component="h2" fontWeight={700}>
        {review.product_name || review.product_id}
      </Typography>
      <Typography color="text.secondary">
        {review.anonymous
          ? ["بدون نام در سایت", review.name].filter(Boolean).join(" · ")
          : review.name}{" "}
        · {number.format(review.rating)} از ۵ ·{" "}
        {date.format(new Date(review.created_at))}
      </Typography>
      <Typography sx={reviewTextStyle}>{review.text}</Typography>
      {review.rejection_reason && (
        <Typography sx={reviewTextStyle}>
          دلیل رد: {review.rejection_reason}
        </Typography>
      )}
      {review.reply && (
        <Typography sx={reviewTextStyle}>پاسخ چریوا: {review.reply}</Typography>
      )}
      {review.status === "pending" && !formAction && (
        <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
          <Button
            variant="contained"
            disabled={mutation.isPending}
            onClick={() => decide("approve")}
          >
            تأیید و انتشار
          </Button>
          <Button
            color="error"
            disabled={mutation.isPending}
            onClick={() => setFormAction("reject")}
          >
            رد نظر
          </Button>
        </Stack>
      )}
      {review.status === "approved" && !review.reply && !formAction && (
        <Button
          disabled={mutation.isPending}
          onClick={() => setFormAction("publish_reply")}
        >
          نوشتن پاسخ
        </Button>
      )}
      {formAction && (
        <Box component="form" onSubmit={submit} sx={{ mt: 2 }}>
          <TextField
            fullWidth
            required
            multiline
            rows={3}
            label={formAction === "reject" ? "دلیل رد" : "متن پاسخ تیم"}
            value={text}
            disabled={mutation.isPending}
            inputProps={{ maxLength: 2000 }}
            onChange={(event) => setText(event.target.value)}
          />
          <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
            <Button
              type="submit"
              variant="contained"
              disabled={mutation.isPending || !text.trim()}
            >
              {formAction === "reject" ? "ثبت رد نظر" : "انتشار پاسخ"}
            </Button>
            <Button
              disabled={mutation.isPending}
              onClick={() => {
                setFormAction(undefined);
                setText("");
                mutation.reset();
              }}
            >
              انصراف
            </Button>
          </Stack>
        </Box>
      )}
      {mutation.error && (
        <Alert
          severity="error"
          sx={{ mt: 1 }}
          action={
            mutation.error instanceof ApiError &&
            mutation.error.status === 409 ? (
              <Button onClick={() => void refresh()}>به‌روزرسانی فهرست</Button>
            ) : undefined
          }
        >
          {mutation.error.message}
        </Alert>
      )}
    </Paper>
  );
}
