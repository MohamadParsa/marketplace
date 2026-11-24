import { useRef, useState, type FormEvent } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Alert,
  Button,
  Chip,
  CircularProgress,
  Stack,
  TextField,
} from "@mui/material";
import { ApiError } from "../../shared/api";
import type {
  ProductQAMutation,
  ProductQAQuestion,
  ProductQAStatus,
} from "../../shared/types";
import { useWorkspace } from "../../shared/workspace";

const labels: Record<ProductQAStatus, string> = {
  pending: "در انتظار تأیید",
  approved: "تأییدشده",
  rejected: "ردشده",
};

export function ProductQA({ productId }: { productId: string }) {
  const { loadProductQA } = useWorkspace();
  const [cursors, setCursors] = useState([0]);
  const after = cursors[cursors.length - 1];
  const query = useQuery({
    queryKey: ["product-qa", productId, after],
    queryFn: () => loadProductQA(productId, after),
  });

  return (
    <section
      className="product-workspace-section"
      aria-label="مدیریت پرسش و پاسخ"
    >
      <Stack spacing={2}>
        {query.isPending && (
          <CircularProgress size={24} aria-label="در حال دریافت پرسش‌ها" />
        )}
        {query.error && (
          <Alert
            severity="error"
            action={
              <Button onClick={() => void query.refetch()}>تلاش مجدد</Button>
            }
          >
            {query.error.message}
          </Alert>
        )}
        {query.data && query.data.questions.length === 0 && (
          <p>پرسشی در این صفحه وجود ندارد.</p>
        )}
        {query.data?.questions.map((question) => (
          <QuestionEditor
            key={question.id}
            productId={productId}
            question={question}
            canReply={query.data.can_reply}
            refresh={query.refetch}
          />
        ))}
        {(cursors.length > 1 || query.data?.next_cursor) && (
          <nav aria-label="صفحه‌های پرسش و پاسخ">
            <Button
              disabled={cursors.length < 2}
              onClick={() => setCursors((value) => value.slice(0, -1))}
            >
              صفحه قبل
            </Button>
            <Button
              disabled={!query.data?.next_cursor}
              onClick={() => {
                const cursor = query.data?.next_cursor;
                if (cursor) setCursors((value) => [...value, cursor]);
              }}
            >
              صفحه بعد
            </Button>
          </nav>
        )}
      </Stack>
    </section>
  );
}

type QAAction =
  | Omit<Extract<ProductQAMutation, { action: "reply" }>, "version">
  | Omit<Extract<ProductQAMutation, { action: "moderate" }>, "version">;

function QuestionEditor({
  productId,
  question,
  canReply,
  refresh,
}: {
  productId: string;
  question: ProductQAQuestion;
  canReply: boolean;
  refresh: () => Promise<unknown>;
}) {
  const { changeProductQA } = useWorkspace();
  const [text, setText] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const pending = useRef<{ signature: string; key: string } | null>(null);
  const running = useRef(false);

  async function change(action: QAAction) {
    if (running.current) return;
    const mutation: ProductQAMutation = {
      ...action,
      version: question.version,
    };
    const signature = JSON.stringify(mutation);
    if (pending.current?.signature !== signature)
      pending.current = { signature, key: crypto.randomUUID() };
    running.current = true;
    setBusy(true);
    setError("");
    try {
      await changeProductQA(
        productId,
        question.id,
        mutation,
        pending.current.key,
      );
      pending.current = null;
      if (action.action === "reply") setText("");
      await refresh();
    } catch (cause) {
      setError(
        cause instanceof Error ? cause.message : "ارتباط با سرور برقرار نشد.",
      );
      if (cause instanceof ApiError && cause.status === 409) {
        pending.current = null;
        await refresh();
      }
    } finally {
      running.current = false;
      setBusy(false);
    }
  }

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (text.trim()) void change({ action: "reply", text: text.trim() });
  }

  function moderation(
    status: ProductQAStatus,
    label: string,
    answerId?: string,
  ) {
    return (
      <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
        <Chip size="small" label={labels[status]} />
        <Button
          size="small"
          aria-label={`انتشار ${label}`}
          disabled={busy || status === "approved"}
          onClick={() =>
            void change({
              action: "moderate",
              status: "approved",
              answer_id: answerId,
            })
          }
        >
          انتشار
        </Button>
        <Button
          size="small"
          aria-label={`رد ${label}`}
          disabled={busy || status === "rejected"}
          onClick={() =>
            void change({
              action: "moderate",
              status: "rejected",
              answer_id: answerId,
            })
          }
        >
          رد
        </Button>
      </Stack>
    );
  }

  return (
    <Stack
      component="article"
      spacing={1}
      sx={{ border: 1, borderColor: "divider", borderRadius: 2, p: 2 }}
    >
      <strong>پرسش مشتری</strong>
      <p
        dir="auto"
        style={{ whiteSpace: "pre-wrap", overflowWrap: "anywhere" }}
      >
        {question.text}
      </p>
      {moderation(question.status, "پرسش")}
      {question.answers.map((answer, index) => (
        <Stack
          key={answer.id}
          spacing={1}
          sx={{ borderBlockStart: 1, borderColor: "divider", pt: 1 }}
        >
          <strong>پاسخ رسمی</strong>
          <p
            dir="auto"
            style={{ whiteSpace: "pre-wrap", overflowWrap: "anywhere" }}
          >
            {answer.text}
          </p>
          {moderation(answer.status, `پاسخ ${index + 1}`, answer.id)}
        </Stack>
      ))}
      {canReply && (
        <Stack
          component="form"
          onSubmit={submit}
          spacing={1}
          alignItems="start"
        >
          <TextField
            label="پاسخ رسمی"
            value={text}
            onChange={(event) => setText(event.target.value)}
            slotProps={{ htmlInput: { maxLength: 2000, dir: "auto" } }}
            disabled={busy}
            required
            fullWidth
            multiline
            minRows={2}
          />
          <Button
            type="submit"
            variant="contained"
            disabled={busy || !text.trim()}
          >
            ثبت پاسخ برای بررسی
          </Button>
        </Stack>
      )}
      {error && <Alert severity="error">{error}</Alert>}
    </Stack>
  );
}
