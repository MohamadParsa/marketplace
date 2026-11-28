import { ReviewStar, ReviewStars } from "./review-stars";
import { FeedbackDisclosure } from "./feedback-disclosure";
import { FeedbackCarousel } from "./feedback-carousel";
import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type FormEvent,
} from "react";
import { LoadingPlaceholder } from "./busy-region";
import "./product-reviews.css";
import "./product-feedback.css";

const api = import.meta.env.VITE_API_BASE_URL ?? "";
const number = new Intl.NumberFormat("fa-IR");
const date = new Intl.DateTimeFormat("fa-IR", { dateStyle: "medium" });

type Review = {
  id: string;
  name?: string;
  text: string;
  rating: number;
  created_at: string;
  reply?: string;
};
export type RatingSummary = { count: number; average: number | null };
type ReviewPage = {
  rating_summary: RatingSummary;
  reviews: Review[];
  next_cursor?: string;
  submissions_enabled: boolean;
};

export function ProductReviews({
  productId,
  onSummary,
}: {
  productId: string;
  onSummary?: (summary: RatingSummary | undefined) => void;
}) {
  const summaryCallback = useRef(onSummary);
  summaryCallback.current = onSummary;
  const endpoint = `${api}/storefront/products/${encodeURIComponent(productId)}/reviews`;
  const [pages, setPages] = useState<Review[][]>([]);
  const [summary, setSummary] = useState<RatingSummary>();
  const [cursor, setCursor] = useState<string>();
  const [enabled, setEnabled] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const request = useRef<AbortController | null>(null);

  const loadPage = useCallback(
    async (after?: string) => {
      if (request.current) return;
      const controller = new AbortController();
      request.current = controller;
      setLoading(true);
      setError(false);
      try {
        const url = after
          ? endpoint + `?cursor=${encodeURIComponent(after)}`
          : endpoint;
        const response = await fetch(url, {
          signal: controller.signal,
          cache: "no-store",
        });
        if (!response.ok) throw new Error("reviews_unavailable");
        const page: ReviewPage = await response.json();
        if (controller.signal.aborted) return;
        setPages((current) =>
          after ? [...current, page.reviews] : [page.reviews],
        );
        setSummary(page.rating_summary);
        summaryCallback.current?.(page.rating_summary);
        setCursor(page.next_cursor);
        setEnabled(page.submissions_enabled);
      } catch {
        if (!controller.signal.aborted) {
          setError(true);
          summaryCallback.current?.(undefined);
        }
      } finally {
        if (request.current === controller) request.current = null;
        if (!controller.signal.aborted) setLoading(false);
      }
    },
    [endpoint],
  );

  useEffect(() => {
    void loadPage();
    const refresh = () => void loadPage();
    addEventListener("focus", refresh);
    return () => {
      removeEventListener("focus", refresh);
      request.current?.abort();
      request.current = null;
    };
  }, [loadPage]);

  const reviewsExist = Boolean(pages[0]?.length);
  return (
    <section className="pdp-section" id="reviews">
      <h2>نظرات کاربران</h2>
      {summary && (
        <div
          className="pdp-rating-summary"
          aria-label="میانگین امتیاز محصول"
          aria-busy={loading}
          aria-live="polite"
        >
          {loading && (
            <span
              className="pdp-rating-progress"
              role="progressbar"
              aria-label="به‌روزرسانی امتیاز محصول"
            />
          )}
          <ReviewStars
            value={summary.average ?? 0}
            label={
              summary.average == null
                ? "هنوز امتیازی ثبت نشده"
                : `امتیاز ${new Intl.NumberFormat("fa-IR", { minimumFractionDigits: 1, maximumFractionDigits: 1 }).format(summary.average)} از ۵ بر اساس ${number.format(summary.count)} نظر تأییدشده`
            }
          />
          {summary.average == null ? (
            <p>هنوز امتیازی ثبت نشده</p>
          ) : (
            <>
              <strong>
                {new Intl.NumberFormat("fa-IR", {
                  minimumFractionDigits: 1,
                  maximumFractionDigits: 1,
                }).format(summary.average)}{" "}
                <small>از ۵</small>
              </strong>
              <span>بر اساس {number.format(summary.count)} نظر تأییدشده</span>
            </>
          )}
        </div>
      )}
      {!pages.length && loading && (
        <LoadingPlaceholder label="در حال دریافت نظرات…" />
      )}
      {!pages.length && error && (
        <p role="alert">
          دریافت نظرات ممکن نشد.{" "}
          <button type="button" onClick={() => void loadPage()}>
            تلاش دوباره
          </button>
        </p>
      )}
      {pages.length > 0 && error && !cursor && (
        <p role="alert">
          به‌روزرسانی نظرات ممکن نشد؛ آخرین اطلاعات دریافت‌شده نمایش داده
          می‌شود.{" "}
          <button type="button" onClick={() => void loadPage()}>
            تلاش دوباره
          </button>
        </p>
      )}
      {pages.length > 0 && !reviewsExist && <p>هنوز نظری منتشر نشده است.</p>}
      {reviewsExist && (
        <FeedbackCarousel
          items={pages.flat()}
          label="کارت‌های نظرات محصول"
          moreLabel="مشاهده بیشتر نظرات"
          hasMore={!!cursor}
          loading={loading}
          error={error}
          onMore={() => void loadPage(cursor)}
          renderItem={(review) => <ReviewCard review={review} />}
        />
      )}
      {pages.length > 0 && enabled && (
        <FeedbackDisclosure label="ثبت نظر">
          <ReviewForm endpoint={endpoint} onClosed={() => setEnabled(false)} />
        </FeedbackDisclosure>
      )}
      {pages.length > 0 && !enabled && (
        <p>ثبت نظر برای این محصول غیرفعال است.</p>
      )}
    </section>
  );
}

function ReviewCard({ review }: { review: Review }) {
  return (
    <article className="pdp-review-card">
      <header>
        <strong>{review.name || "بدون نام"}</strong>
        <ReviewStars
          value={review.rating}
          label={`امتیاز ${number.format(review.rating)} از ۵`}
        />
      </header>
      <time dateTime={review.created_at}>
        {date.format(new Date(review.created_at))}
      </time>
      <p>{review.text}</p>
      {review.reply && (
        <div className="pdp-review-reply">
          <strong>پاسخ چریوا</strong>
          <p>{review.reply}</p>
        </div>
      )}
    </article>
  );
}

function ReviewForm({
  endpoint,
  onClosed,
}: {
  endpoint: string;
  onClosed: () => void;
}) {
  const [name, setName] = useState("");
  const [text, setText] = useState("");
  const [rating, setRating] = useState(0);
  const [hoverRating, setHoverRating] = useState(0);
  const [anonymous, setAnonymous] = useState(false);
  const [sending, setSending] = useState(false);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");
  const [closed, setClosed] = useState(false);
  const retry = useRef<{ data: string; key: string }>();
  const active = useRef(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (active.current) return;
    if (!text.trim() || (!anonymous && !name.trim()) || !rating) {
      setError("متن نظر، امتیاز و در صورت نمایش نام، نام نمایشی را وارد کنید.");
      return;
    }
    const review = {
      name: name.trim(),
      text: text.trim(),
      rating,
      anonymous,
    };
    const data = JSON.stringify(review);
    if (retry.current?.data !== data)
      retry.current = { data, key: crypto.randomUUID() };
    active.current = true;
    setSending(true);
    setError("");
    setSent(false);
    try {
      const response = await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ...review, idempotency_key: retry.current.key }),
      });
      if (response.status === 403) {
        setClosed(true);
        setError("ثبت نظر برای این محصول غیرفعال شده است.");
        return;
      }
      if (response.status === 429) {
        setError("تعداد درخواست‌ها زیاد است. کمی بعد دوباره تلاش کنید.");
        return;
      }
      if (!response.ok) throw new Error("review_not_submitted");
      retry.current = undefined;
      setName("");
      setText("");
      setRating(0);
      setAnonymous(false);
      setSent(true);
    } catch {
      setError("ثبت نظر ممکن نشد. دوباره تلاش کنید.");
    } finally {
      active.current = false;
      setSending(false);
    }
  }

  return (
    <form
      className="pdp-review-form"
      aria-label="فرم ثبت نظر"
      onSubmit={(event) => void submit(event)}
      aria-busy={sending}
    >
      <fieldset disabled={sending}>
        <label className="pdp-review-checkbox">
          <input
            type="checkbox"
            checked={anonymous}
            onChange={(event) => setAnonymous(event.target.checked)}
          />
          انتشار بدون نام
        </label>
        <label>
          نام نمایشی
          <input
            required={!anonymous}
            maxLength={80}
            value={name}
            aria-describedby={anonymous ? "pdp-review-private-name" : undefined}
            onChange={(event) => setName(event.target.value)}
          />
          {anonymous && (
            <small id="pdp-review-private-name">
              اختیاری و قابل مشاهده فقط برای ادمین
            </small>
          )}
        </label>
        <div
          role="radiogroup"
          onMouseLeave={() => setHoverRating(0)}
          aria-labelledby="pdp-rating-label"
          className="pdp-review-rating"
        >
          <span id="pdp-rating-label">امتیاز</span>
          <div className="pdp-rating-input-stars">
            {" "}
            {[1, 2, 3, 4, 5].map((value) => (
              <label
                key={value}
                className="pdp-rating-choice"
                onMouseEnter={() => setHoverRating(value)}
              >
                <input
                  aria-label={`${number.format(value)} از ۵ ستاره`}
                  type="radio"
                  name="product-rating"
                  required
                  value={value}
                  checked={rating === value}
                  onChange={() => setRating(value)}
                />
                <ReviewStar fill={(hoverRating || rating) >= value ? 1 : 0} />
              </label>
            ))}
          </div>
          <span className="pdp-rating-feedback" aria-live="polite">
            {
              ["انتخاب کنید", "بسیار ضعیف", "ضعیف", "متوسط", "خوب", "عالی"][
                hoverRating || rating
              ]
            }
          </span>
        </div>
        <label>
          متن نظر{" "}
          <textarea
            required
            maxLength={2000}
            rows={4}
            value={text}
            onChange={(event) => setText(event.target.value)}
          />
        </label>
        <button type="submit" className="pdp-feedback-action" disabled={closed}>
          {sending ? "در حال ارسال…" : "ارسال برای بررسی"}
        </button>
      </fieldset>
      {error && <p role="alert">{error}</p>}
      {closed && (
        <button type="button" onClick={onClosed}>
          بستن فرم
        </button>
      )}
      {sent && (
        <p role="status">نظر شما ثبت شد و پس از تأیید نمایش داده می‌شود.</p>
      )}
    </form>
  );
}

export function ProductRatingLink({ summary }: { summary?: RatingSummary }) {
  const average = summary?.average;
  const formatted =
    average == null
      ? ""
      : new Intl.NumberFormat("fa-IR", {
          minimumFractionDigits: 1,
          maximumFractionDigits: 1,
        }).format(average);
  return (
    <a
      className="pdp-rating-link"
      href="#reviews"
      aria-label={
        summary && average != null
          ? `امتیاز ${formatted} از ۵؛ مشاهده ${number.format(summary.count)} دیدگاه`
          : "مشاهده نظرات کاربران"
      }
    >
      {summary && average != null ? (
        <>
          <ReviewStars value={average} label={`امتیاز ${formatted} از ۵`} />
          <b>{formatted}</b>
          <span>({number.format(summary.count)} دیدگاه)</span>
        </>
      ) : (
        <span>{summary ? "هنوز امتیازی ثبت نشده" : "نظرات کاربران"}</span>
      )}
    </a>
  );
}
