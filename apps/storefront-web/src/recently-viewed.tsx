import { useEffect, useRef } from "react";
import { apiBase, useCommerce } from "./commerce";

export function HistoryIcon() {
  return (
    <svg
      className="wishlist-heart-icon"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.6"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M3 10a9 9 0 1 1 1.8 7.4M3 4v6h6M12 7v5l3 2" />
    </svg>
  );
}
// A last-view bookmark is independent of raw analytics and anonymous traffic.
export function useRecordRecentlyViewed(productID: string) {
  const { token } = useCommerce();
  const sent = useRef("");
  useEffect(() => {
    if (!token || !productID) return;
    const key = token + ":" + productID;
    if (sent.current === key) return;
    sent.current = key;
    // Keep the completed visit if the user promptly moves to another page.
    void fetch(
      apiBase +
        "/storefront/account/recently-viewed/" +
        encodeURIComponent(productID),
      {
        method: "POST",
        headers: { Authorization: "Bearer " + token },
        cache: "no-store",
        keepalive: true,
      },
    )
      .then((response) => {
        if (response.status === 401)
          dispatchEvent(new Event("cherriva:logout"));
      })
      .catch(() => {});
  }, [token, productID]);
}
