import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type FormEvent,
} from "react";
import { FeedbackDisclosure } from "./feedback-disclosure";
import { FeedbackCarousel } from "./feedback-carousel";
import { LoadingPlaceholder } from "./busy-region";
import {
  apiBase,
  call,
  CommerceError,
  operationKey,
  useCommerce,
} from "./commerce";
import "./product-qa.css";
import "./product-reviews.css";
import "./product-feedback.css";

type PublicQuestion = {
  id: string;
  text: string;
  answers: { id: string; text: string }[];
};
type QuestionPage = {
  questions: PublicQuestion[];
  next_cursor?: number;
  can_ask: boolean;
};

function message(error: unknown) {
  return error instanceof Error ? error.message : "ارتباط با سرور برقرار نشد.";
}

export function ProductQA({ productId }: { productId: string }) {
  const { token } = useCommerce();
  const path = `/storefront/products/${encodeURIComponent(productId)}/questions`;
  const [page, setPage] = useState<QuestionPage & { shown: number }>();
  const [loading, setLoading] = useState(true);
  const request = useRef<AbortController | null>(null);
  const [loadError, setLoadError] = useState("");
  const [text, setText] = useState("");
  const [sending, setSending] = useState(false);
  const [sendError, setSendError] = useState("");
  const [notice, setNotice] = useState("");
  const pending = useRef<{ text: string; key: string } | null>(null);
  const sendingRef = useRef(false);

  const loadPage = useCallback(
    async (after = 0) => {
      if (request.current) return;
      const controller = new AbortController();
      request.current = controller;
      setLoading(true);
      setLoadError("");
      try {
        const response = await fetch(
          `${apiBase}${path}${after ? `?after=${after}` : ""}`,
          { signal: controller.signal, cache: "no-store" },
        );
        if (!response.ok) throw new Error("دریافت پرسش‌ها ممکن نشد.");
        const data: QuestionPage = await response.json();
        if (!controller.signal.aborted)
          setPage((current) => ({
            ...data,
            questions: after
              ? [...(current?.questions ?? []), ...data.questions]
              : data.questions,
            shown: after ? (current?.shown ?? 0) + 5 : 5,
          }));
      } catch (error) {
        if (!controller.signal.aborted) setLoadError(message(error));
      } finally {
        if (request.current === controller) request.current = null;
        if (!controller.signal.aborted) setLoading(false);
      }
    },
    [path],
  );
  useEffect(() => {
    void loadPage();
    return () => {
      request.current?.abort();
      request.current = null;
    };
  }, [loadPage]);

  function refresh() {
    void loadPage();
  }
  function showMore() {
    if (!page || loading) return;
    if (page.shown < page.questions.length) {
      setPage((current) => current && { ...current, shown: current.shown + 5 });
    } else if (page.next_cursor) {
      void loadPage(page.next_cursor);
    }
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmed = text.trim();
    if (!token || sendingRef.current || !page?.can_ask) return;
    if (!trimmed) {
      setSendError("متن پرسش را وارد کنید.");
      return;
    }
    if (pending.current?.text !== trimmed)
      pending.current = { text: trimmed, key: operationKey() };
    sendingRef.current = true;
    setSending(true);
    setSendError("");
    setNotice("");
    try {
      await call<void>(
        path,
        token,
        { text: trimmed },
        "POST",
        pending.current.key,
      );
      setText("");
      pending.current = null;
      setNotice("پرسش ثبت شد و پس از تأیید نمایش داده می‌شود.");
    } catch (error) {
      setSendError(
        error instanceof CommerceError && error.status === 403
          ? "در حال حاضر ثبت پرسش برای این محصول غیرفعال است."
          : message(error),
      );
      if (error instanceof CommerceError && error.status === 403) refresh();
      if (error instanceof CommerceError && error.status === 409)
        pending.current = null;
    } finally {
      sendingRef.current = false;
      setSending(false);
    }
  }

  const visible = page;
  const returnTo = location.pathname + location.search + "#questions";
  return (
    <section
      className="pdp-section product-qa"
      id="questions"
      aria-labelledby="product-qa-heading"
    >
      <h2 id="product-qa-heading">پرسش و پاسخ</h2>
      {!visible && !loadError && (
        <LoadingPlaceholder label="در حال دریافت پرسش‌ها" />
      )}
      {loadError && !page && (
        <div role="alert">
          <p>{loadError}</p>
          <button type="button" onClick={refresh}>
            تلاش مجدد
          </button>
        </div>
      )}
      {visible && (
        <>
          {visible.questions.length === 0 && (
            <p>هنوز پرسش تأییدشده‌ای ثبت نشده است.</p>
          )}
          {visible.questions.length > 0 && (
            <FeedbackCarousel
              items={visible.questions.slice(0, visible.shown)}
              label="کارت‌های پرسش و پاسخ"
              moreLabel="مشاهده بیشتر پرسش و پاسخ"
              hasMore={
                visible.shown < visible.questions.length ||
                !!visible.next_cursor
              }
              loading={loading}
              error={!!loadError}
              onMore={showMore}
              renderItem={(question) => (
                <article
                  key={question.id}
                  className="pdp-review-card product-qa-question"
                >
                  <header>
                    <strong>پرسش کاربر</strong>
                  </header>
                  <p dir="auto">{question.text}</p>
                  {question.answers.map((answer) => (
                    <div
                      key={answer.id}
                      className="pdp-review-reply product-qa-answer"
                    >
                      <strong>پاسخ رسمی</strong>
                      <p dir="auto">{answer.text}</p>
                    </div>
                  ))}
                </article>
              )}
            />
          )}
        </>
      )}
      {page && !page.can_ask && (
        <p>ثبت پرسش جدید برای این محصول غیرفعال است.</p>
      )}
      {(page?.can_ask || text) && (
        <FeedbackDisclosure label="ثبت پرسش">
          {!token && (
            <a href={`/login?return=${encodeURIComponent(returnTo)}`}>
              برای ثبت پرسش وارد شوید
            </a>
          )}
          {token && (
            <form
              className="product-qa-form"
              aria-label="فرم ثبت پرسش"
              onSubmit={(event) => void submit(event)}
            >
              <label htmlFor="product-qa-text">پرسش شما</label>
              <textarea
                id="product-qa-text"
                rows={3}
                dir="auto"
                maxLength={2000}
                required
                disabled={sending}
                value={text}
                onChange={(event) => {
                  const value = event.target.value;
                  setText(value);
                  setSendError("");
                  setNotice("");
                  if (pending.current?.text !== value.trim())
                    pending.current = null;
                }}
              />
              <button
                type="submit"
                className="pdp-feedback-action"
                disabled={!page?.can_ask || sending}
              >
                {sending ? "در حال ارسال…" : "ثبت پرسش"}
              </button>
            </form>
          )}
          {sendError && <p role="alert">{sendError}</p>}
          {notice && <p role="status">{notice}</p>}
        </FeedbackDisclosure>
      )}
    </section>
  );
}
