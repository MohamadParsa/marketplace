import { CommerceProvider } from "./commerce";
import {
  LoginPage,
  CartPage,
  CheckoutPage,
  OrderPage,
  WalletPage,
  AccountPage,
  BusinessPage,
} from "./commerce-pages";
import "./commerce.css";
import { StrictMode, useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import { HomePage } from "./home-page";
import { ProductPage } from "./product-page";
import "./product.css";
import { ListingPage } from "./listing-page";
import { installNavigation } from "./navigation";
import "./home.css";
import "./listing.css";
function App() {
  const [url, setURL] = useState(location.pathname + location.search);
  useEffect(() => {
    const sync = () => setURL(location.pathname + location.search);
    addEventListener("popstate", sync);
    addEventListener("cherriva:navigate", sync);
    const uninstall = installNavigation();
    return () => {
      uninstall();
      removeEventListener("popstate", sync);
      removeEventListener("cherriva:navigate", sync);
    };
  }, []);
  const path = url.split("?")[0];
  if (path === "/login") return <LoginPage />;
  if (path === "/cart") return <CartPage />;
  if (path === "/checkout") return <CheckoutPage />;
  if (path === "/wallet") return <WalletPage />;
  if (path === "/account/orders" || path.startsWith("/account/orders/"))
    return <OrderPage key={path} id={path.split("/")[3]} />;
  if (path === "/account" || path === "/account/addresses")
    return <AccountPage key={path} addresses={path.endsWith("/addresses")} />;
  if (path.startsWith("/info/"))
    return <BusinessPage slug={path.split("/")[2]} />;
  if (url.startsWith("/product/"))
    return <ProductPage key={url.split("?")[0]} url={url} />;
  return url.startsWith("/category/") ||
    /^\/(search|offers)(\?|$)/.test(url) ? (
    <ListingPage url={url} />
  ) : (
    <HomePage />
  );
}
createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <CommerceProvider>
      <App />
    </CommerceProvider>
  </StrictMode>,
);
