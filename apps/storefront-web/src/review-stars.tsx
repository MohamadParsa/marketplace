import { useId } from "react";
const path =
  "M12 3 14.78 8.63 21 9.54 16.5 13.93 17.56 20.13 12 17.2 6.44 20.13 7.5 13.93 3 9.54 9.22 8.63 12 3Z";
export function ReviewStar({ fill = 0 }: { fill?: number }) {
  const id = "review-star-" + useId().replace(/:/g, "");
  const percent = Math.min(1, Math.max(0, fill)) * 100;
  return (
    <svg
      className="review-star"
      viewBox="0 0 24 24"
      aria-hidden="true"
      focusable="false"
    >
      <defs>
        <linearGradient id={id} x1="0" x2="1" y1="0" y2="0">
          <stop offset={`${percent}%`} stopColor="var(--review-gold,#dfa524)" />
          <stop
            offset={`${percent}%`}
            stopColor="var(--review-star-empty,#e8e4dc)"
          />
        </linearGradient>
      </defs>
      <path
        d={path}
        fill={`url(#${id})`}
        stroke={`url(#${id})`}
        strokeWidth="1.1"
        strokeLinejoin="round"
      />
    </svg>
  );
}
export function ReviewStars({
  value,
  label,
}: {
  value: number;
  label: string;
}) {
  return (
    <span className="review-stars" role="img" aria-label={label}>
      {[0, 1, 2, 3, 4].map((i) => (
        <ReviewStar key={i} fill={value - i} />
      ))}
    </span>
  );
}
