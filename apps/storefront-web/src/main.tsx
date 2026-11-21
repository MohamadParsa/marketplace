import { StrictMode, useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import { HomePage } from "./home-page";
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
  return url.startsWith("/category/") ||
    /^\/(search|offers)(\?|$)/.test(url) ? (
    <ListingPage url={url} />
  ) : (
    <HomePage />
  );
}
createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
