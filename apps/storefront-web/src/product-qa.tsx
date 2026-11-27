import { useEffect, useRef, useState, type FormEvent } from "react";
import { LoadingPlaceholder } from "./busy-region";
import { call, CommerceError, operationKey, useCommerce } from "./commerce";
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
  const [cursors, setCursors] = useState([0]);
  const after = cursors[cursors.length - 1];
  const [reload, setReload] = useState(0);
  const [page, setPage] = useState<{ after: number; data: QuestionPage }>();
  const [loadError, setLoadError] = useState("");
  const [text, setText] = useState("");
  const [sending, setSending] = useState(false);
  const [sendError, setSendError] = useState("");
  const [notice, setNotice] = useState("");
  const pending = useRef<{ text: string; key: string } | null>(null);
  const sendingRef = useRef(false);

  useEffect(() => {
    let active = true;
    void call<QuestionPage>(`${path}${after ? `?after=${after}` : ""}`)
      .then((data) => {
        if (active) setPage({ after, data });
      })
      .catch((error: unknown) => {
        if (active) setLoadError(message(error));
      });
    return () => {
      active = false;
    };
  }, [path, after, reload]);

  function refresh() {
    setLoadError("");
    setReload((value) => value + 1);
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmed = text.trim();
    if (!trimmed || !token || sendingRef.current) return;
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

  const visible = page?.after === after ? page.data : undefined;
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
      {loadError && (
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
            <div
              className="pdp-review-track"
              role="region"
              aria-label="کارت‌های پرسش و پاسخ"
              tabIndex={0}
            >
              {Array.from(
                { length: Math.ceil(visible.questions.length / 5) },
                (_, index) => visible.questions.slice(index * 5, index * 5 + 5),
              ).map((group) => (
                <div className="pdp-review-group" key={group[0].id}>
                  {group.map((question) => (
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
                  ))}
                </div>
              ))}
            </div>
          )}
        </>
      )}
      {(cursors.length > 1 || visible?.next_cursor) && (
        <nav aria-label="صفحه‌های پرسش و پاسخ" className="product-qa-pages">
          <button
            type="button"
            disabled={cursors.length < 2}
            onClick={() => {
              setLoadError("");
              setCursors((value) => value.slice(0, -1));
            }}
          >
            صفحه قبل
          </button>
          <button
            type="button"
            disabled={!visible?.next_cursor}
            onClick={() => {
              const cursor = visible?.next_cursor;
              if (cursor) {
                setLoadError("");
                setCursors((value) => [...value, cursor]);
              }
            }}
          >
            صفحه بعد
          </button>
        </nav>
      )}
      {page && !page.data.can_ask && (
        <p>ثبت پرسش جدید برای این محصول غیرفعال است.</p>
      )}
      {page?.data.can_ask && !token && (
        <a href={`/login?return=${encodeURIComponent(returnTo)}`}>
          برای ثبت پرسش وارد شوید
        </a>
      )}
      {token && (page?.data.can_ask || text) && (
        <form
          className="product-qa-form"
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
            disabled={!page?.data.can_ask || sending || !text.trim()}
          >
            {sending ? "در حال ارسال…" : "ثبت پرسش"}
          </button>
        </form>
      )}
      {sendError && <p role="alert">{sendError}</p>}
      {notice && <p role="status">{notice}</p>}
    </section>
  );
}
