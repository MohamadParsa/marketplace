import { useState } from "react";
import "./catalog-loading.css";

type Props = {
  src?: string;
  mobileSrc?: string;
  onError?: () => void;
  fetchPriority?: "high" | "low" | "auto";
  alt: string;
  loading?: "eager" | "lazy";
  className?: string;
};
/** A keyed frame prevents a late image event from revealing the previous selection. */
export function CatalogImage(props: Props) {
  return <ImageFrame key={props.src ?? "missing"} {...props} />;
}
function ImageFrame({
  src,
  mobileSrc,
  onError,
  fetchPriority,
  alt,
  loading = "eager",
  className = "",
}: Props) {
  const [state, setState] = useState<"loading" | "ready" | "failed">(
    src ? "loading" : "failed",
  );
  return (
    <span
      className={`sf-catalog-image ${className}`}
      data-image-state={state}
      aria-busy={state === "loading"}
    >
      {state === "loading" && (
        <span className="sf-skeleton sf-image-skeleton" aria-hidden="true" />
      )}
      {state === "failed" ? (
        <span className="sf-image-fallback">تصویر در دسترس نیست</span>
      ) : (
        <picture>
          {mobileSrc && (
            <source media="(max-width: 600px)" srcSet={mobileSrc} />
          )}
          <img
            src={src}
            alt={alt}
            loading={loading}
            decoding="async"
            fetchPriority={fetchPriority}
            ref={(img) => {
              if (img?.complete)
                setState(img.naturalWidth > 0 ? "ready" : "failed");
            }}
            onLoad={() => setState("ready")}
            onError={() => {
              setState("failed");
              onError?.();
            }}
          />
        </picture>
      )}
    </span>
  );
}
