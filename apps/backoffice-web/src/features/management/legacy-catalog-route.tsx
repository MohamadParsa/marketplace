import { Navigate, useLocation } from "react-router-dom";

// Only routing changes here; authorization belongs to the destination and API.
export function LegacyCatalogRoute() {
  const { search } = useLocation();
  const params = new URLSearchParams(search);
  const kind =
    params.has("category") || params.get("create") === "category"
      ? "categories"
      : params.has("brand") || params.get("create") === "brand"
        ? "brands"
        : params.has("product") || params.get("create") === "product"
          ? "products"
          : "home-settings";
  return <Navigate to={`/${kind}${search}`} replace />;
}
