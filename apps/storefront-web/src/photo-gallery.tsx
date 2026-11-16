import { useEffect, useRef, useState, type PointerEvent } from "react";
import { money, type Product } from "./catalog";
import { Icon } from "./components";
import { productPhotos } from "./product-photos";

export function PhotoGallery({
  product,
  index,
  onChange,
  expanded = false,
}: {
  product: Product;
  index: number;
  onChange: (index: number) => void;
  expanded?: boolean;
}) {
  const photos = productPhotos(product);
  const photo = photos[index] ?? photos[0];
  const gallery = useRef<HTMLElement>(null);
  const rail = useRef<HTMLDivElement>(null);
  const zoomButton = useRef<HTMLButtonElement>(null);
  const gesture = useRef<{
    id: number;
    x: number;
    y: number;
    origin: [number, number];
  } | null>(null);
  const dragged = useRef(false);
  const [zoomed, setZoomed] = useState(false);
  const [origin, setOrigin] = useState<[number, number]>([50, 50]);
  const [railEnd, setRailEnd] = useState(false);
  const scale = photo.crop ? 1.45 : 2;
  const overflow = !expanded && photos.length > 4;

  useEffect(() => {
    setZoomed(false);
  }, [index]);

  useEffect(() => {
    const dialog = gallery.current?.closest("dialog");
    if (!dialog || !zoomed) return;
    function closeZoom(event: KeyboardEvent) {
      if (event.key !== "Escape") return;
      event.preventDefault();
      setZoomed(false);
    }
    dialog.addEventListener("keydown", closeZoom, true);
    return () => dialog.removeEventListener("keydown", closeZoom, true);
  }, [zoomed]);

  useEffect(() => {
    const strip = rail.current;
    const thumbnail = strip?.children[index] as HTMLElement | undefined;
    if (!strip || !thumbnail || expanded) return;
    const bounds = strip.getBoundingClientRect();
    const selected = thumbnail.getBoundingClientRect();
    if (selected.left < bounds.left)
      strip.scrollBy({
        left: selected.left - bounds.left - strip.clientWidth / 5,
      });
    else if (selected.right > bounds.right)
      strip.scrollBy({
        left: selected.right - bounds.right + strip.clientWidth / 5,
      });
  }, [index, expanded]);

  function show(next: number) {
    setZoomed(false);
    onChange((next + photos.length) % photos.length);
  }

  function position(x: number, y: number) {
    setOrigin([Math.max(0, Math.min(100, x)), Math.max(0, Math.min(100, y))]);
  }

  function positionFromPointer(
    surface: HTMLElement,
    clientX: number,
    clientY: number,
  ) {
    const bounds = surface.getBoundingClientRect();
    position(
      (100 * (clientX - bounds.left)) / bounds.width,
      (100 * (clientY - bounds.top)) / bounds.height,
    );
  }

  function startPointer(event: PointerEvent<HTMLDivElement>) {
    if (!event.isPrimary || event.button !== 0) return;
    gesture.current = {
      id: event.pointerId,
      x: event.clientX,
      y: event.clientY,
      origin,
    };
    dragged.current = false;
    event.currentTarget.setPointerCapture(event.pointerId);
  }

  function movePointer(event: PointerEvent<HTMLDivElement>) {
    const start = gesture.current;
    if (start?.id === event.pointerId) {
      const dx = event.clientX - start.x;
      const dy = event.clientY - start.y;
      dragged.current ||= Math.hypot(dx, dy) > 6;
      if (zoomed && dragged.current) {
        position(
          start.origin[0] -
            (100 * dx) / (event.currentTarget.clientWidth * (scale - 1)),
          start.origin[1] -
            (100 * dy) / (event.currentTarget.clientHeight * (scale - 1)),
        );
      }
    } else if (zoomed && event.pointerType === "mouse" && !event.buttons) {
      positionFromPointer(event.currentTarget, event.clientX, event.clientY);
    }
  }

  function endPointer(event: PointerEvent<HTMLDivElement>) {
    const start = gesture.current;
    gesture.current = null;
    if (!start || start.id !== event.pointerId || zoomed) return;
    const dx = event.clientX - start.x;
    const dy = event.clientY - start.y;
    if (Math.abs(dx) > 35 && Math.abs(dx) > Math.abs(dy) * 1.3) {
      dragged.current = true;
      show(index + (dx > 0 ? 1 : -1));
    }
  }

  return (
    <section
      ref={gallery}
      className={expanded ? "photo-gallery expanded" : "photo-gallery"}
      aria-label={`تصاویر ${product.name}`}
      onKeyDown={(event) => {
        if (event.key === "Escape" && zoomed) {
          event.preventDefault();
          setZoomed(false);
          return;
        }
        const horizontal =
          event.key === "ArrowRight" ? 1 : event.key === "ArrowLeft" ? -1 : 0;
        const vertical =
          event.key === "ArrowDown" ? 1 : event.key === "ArrowUp" ? -1 : 0;
        if (zoomed && (horizontal || vertical)) {
          event.preventDefault();
          position(origin[0] + 12 * horizontal, origin[1] + 12 * vertical);
        } else if (horizontal) {
          event.preventDefault();
          show(index - horizontal);
        }
      }}
    >
      <div className="photo-stage">
        <div
          className="photo-surface"
          data-zoomed={zoomed}
          onPointerDown={startPointer}
          onPointerMove={movePointer}
          onPointerUp={endPointer}
          onPointerCancel={() => {
            gesture.current = null;
          }}
          onClick={(event) => {
            if (dragged.current) return;
            positionFromPointer(
              event.currentTarget,
              event.clientX,
              event.clientY,
            );
            setZoomed(!zoomed);
            zoomButton.current?.focus({ preventScroll: true });
          }}
        >
          <div
            className="photo-view"
            style={
              zoomed
                ? {
                    transform: `scale(${scale})`,
                    transformOrigin: `${origin[0]}% ${origin[1]}%`,
                  }
                : undefined
            }
          >
            <img
              src={photo.src}
              alt={`${product.name}، ${photo.label}`}
              draggable={false}
              data-crop={photo.crop}
            />
          </div>
        </div>
        <button
          ref={zoomButton}
          type="button"
          className="photo-zoom"
          aria-label="بزرگ‌نمایی تصویر"
          aria-pressed={zoomed}
          onClick={() => {
            setOrigin([50, 50]);
            setZoomed(!zoomed);
          }}
        >
          <Icon name="search" />
        </button>
      </div>
      {expanded && (
        <div className="photo-caption">
          <span>{photo.label}</span>
          <div className="photo-nav" role="group" aria-label="پیمایش تصاویر">
            <button
              type="button"
              dir="ltr"
              aria-label="تصویر قبلی"
              onClick={() => show(index - 1)}
            >
              ›
            </button>
            <span dir="ltr" aria-live="polite">
              {money(index + 1)} / {money(photos.length)}
            </span>
            <button
              type="button"
              dir="ltr"
              aria-label="تصویر بعدی"
              onClick={() => show(index + 1)}
            >
              ‹
            </button>
          </div>
        </div>
      )}
      <div className="photo-strip" data-overflow={overflow}>
        <div
          ref={rail}
          className="photo-rail"
          role="group"
          aria-label="انتخاب تصویر"
          onScroll={(event) => {
            const strip = event.currentTarget;
            setRailEnd(
              -strip.scrollLeft >= strip.scrollWidth - strip.clientWidth - 1,
            );
          }}
        >
          {photos.map((frame, i) => (
            <button
              key={i}
              type="button"
              aria-label={`نمایش تصویر، ${frame.label}`}
              aria-pressed={i === index}
              onClick={() => show(i)}
            >
              <span>
                <img
                  src={frame.src}
                  alt=""
                  draggable={false}
                  data-crop={frame.crop}
                />
              </span>
            </button>
          ))}
        </div>
        {overflow && (
          <button
            type="button"
            className="photo-edge"
            data-back={railEnd}
            aria-label={
              railEnd ? "بازگشت به تصاویر قبلی" : "نمایش تصاویر بیشتر"
            }
            onClick={() => {
              const strip = rail.current;
              if (!strip) return;
              if (railEnd) strip.scrollTo({ left: 0 });
              else strip.scrollBy({ left: -strip.clientWidth });
            }}
          >
            {railEnd ? "›" : "‹"}
          </button>
        )}
      </div>
      {!expanded && (
        <span className="photo-count" dir="ltr">
          {money(index + 1)} / {money(photos.length)}
        </span>
      )}
    </section>
  );
}
