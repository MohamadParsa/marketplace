import { Fragment, useEffect, useRef, type ReactNode } from "react";

export function FeedbackCarousel<T extends { id: string }>({
  items,
  label,
  moreLabel,
  hasMore,
  loading,
  error,
  onMore,
  renderItem,
}: {
  items: T[];
  label: string;
  moreLabel: string;
  hasMore: boolean;
  loading: boolean;
  error: boolean;
  onMore: () => void;
  renderItem: (item: T) => ReactNode;
}) {
  const track = useRef<HTMLDivElement>(null);
  const requestedStart = useRef<number>();
  useEffect(() => {
    const start = requestedStart.current;
    if (start == null || items.length <= start) return;
    const group =
      track.current?.querySelectorAll<HTMLElement>(".pdp-review-group")[
        Math.floor(start / 5)
      ];
    group?.focus({ preventScroll: true });
    group?.scrollIntoView({ block: "nearest", inline: "start" });
    requestedStart.current = undefined;
  }, [items.length]);
  return (
    <div
      className="pdp-review-track"
      ref={track}
      tabIndex={0}
      role="region"
      aria-label={label}
      aria-busy={loading}
    >
      {Array.from({ length: Math.ceil(items.length / 5) }, (_, i) =>
        items.slice(i * 5, i * 5 + 5),
      ).map((group) => (
        <div className="pdp-review-group" key={group[0].id} tabIndex={-1}>
          {group.map((item) => (
            <Fragment key={item.id}>{renderItem(item)}</Fragment>
          ))}
        </div>
      ))}
      {hasMore && (
        <div className="pdp-review-more">
          {error && (
            <p role="alert">دریافت موارد بیشتر ممکن نشد؛ دوباره تلاش کنید.</p>
          )}
          <button
            className="pdp-expandable-toggle"
            type="button"
            disabled={loading}
            aria-label={moreLabel}
            onClick={() => {
              requestedStart.current = items.length;
              onMore();
            }}
          >
            {loading ? "در حال دریافت…" : error ? "تلاش مجدد" : "مشاهده بیشتر"}
          </button>
        </div>
      )}
    </div>
  );
}
