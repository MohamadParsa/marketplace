import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type FormEvent,
} from "react";
import { LoadingPlaceholder } from "./busy-region";
import "./product-reviews.css";

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
type ReviewPage = {
  reviews: Review[];
  next_cursor?: string;
  submissions_enabled: boolean;
};

export function ProductReviews({ productId }: { productId: string }) {
  const endpoint = `${api}/storefront/products/${encodeURIComponent(productId)}/reviews`;
  const [pages, setPages] = useState<Review[][]>([]);
  const [cursor, setCursor] = useState<string>();
  const [enabled, setEnabled] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const request = useRef<AbortController | null>(null);
  const track = useRef<HTMLDivElement>(null);
  const nextPage = useRef<HTMLDivElement>(null);

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
        const response = await fetch(url, { signal: controller.signal });
        if (!response.ok) throw new Error("reviews_unavailable");
        const page: ReviewPage = await response.json();
        if (controller.signal.aborted) return;
        setPages((current) =>
          after ? [...current, page.reviews] : [page.reviews],
        );
        setCursor(page.next_cursor);
        setEnabled(page.submissions_enabled);
      } catch {
        if (!controller.signal.aborted) setError(true);
      } finally {
        if (request.current === controller) request.current = null;
        if (!controller.signal.aborted) setLoading(false);
      }
    },
    [endpoint],
  );

  useEffect(() => {
    void loadPage();
    return () => {
      request.current?.abort();
      request.current = null;
    };
  }, [loadPage]);

  useEffect(() => {
    if (!cursor || loading || error || !track.current || !nextPage.current)
      return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) void loadPage(cursor);
      },
      { root: track.current, threshold: 0.25 },
    );
    observer.observe(nextPage.current);
    return () => observer.disconnect();
  }, [cursor, loading, error, loadPage]);

  const reviewsExist = Boolean(pages[0]?.length);
  return (
    <section className="pdp-section" id="reviews">
      <h2>نظرات کاربران</h2>
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
      {pages.length > 0 && !reviewsExist && <p>هنوز نظری منتشر نشده است.</p>}
      {reviewsExist && (
        <div
          className="pdp-review-track"
          ref={track}
          tabIndex={0}
          role="region"
          aria-label="کارت‌های نظرات محصول"
        >
          {pages.map(
            (page) =>
              page.length > 0 && (
                <div className="pdp-review-group" key={page[0].id}>
                  {page.map((review) => (
                    <ReviewCard key={review.id} review={review} />
                  ))}
                </div>
              ),
          )}
          {cursor && (
            <div className="pdp-review-more" ref={nextPage}>
              {loading ? (
                <LoadingPlaceholder label="در حال دریافت نظرات بیشتر…" />
              ) : (
                <button type="button" onClick={() => void loadPage(cursor)}>
                  {error
                    ? "تلاش دوباره برای دریافت نظرات"
                    : "نمایش نظرات بیشتر"}
                </button>
              )}
            </div>
          )}
        </div>
      )}
      {reviewsExist && cursor && (
        <button
          className="pdp-review-scroll"
          type="button"
          onClick={() =>
            nextPage.current?.scrollIntoView({
              block: "nearest",
              inline: "start",
            })
          }
        >
          رفتن به نظرات بیشتر
        </button>
      )}
      {pages.length > 0 && enabled && (
        <ReviewForm endpoint={endpoint} onClosed={() => setEnabled(false)} />
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
        <span aria-label={`امتیاز ${number.format(review.rating)} از ۵`}>
          ★ {number.format(review.rating)} از ۵
        </span>
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
      onSubmit={(event) => void submit(event)}
      aria-busy={sending}
    >
      <h3>ثبت نظر</h3>
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
          role="group"
          aria-labelledby="pdp-rating-label"
          className="pdp-review-rating"
        >
          <span id="pdp-rating-label">امتیاز</span>
          {[1, 2, 3, 4, 5].map((value) => (
            <label key={value}>
              <input
                type="radio"
                name="product-rating"
                required
                value={value}
                checked={rating === value}
                onChange={() => setRating(value)}
              />
              {number.format(value)}
            </label>
          ))}
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
        <button type="submit" disabled={closed}>
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
