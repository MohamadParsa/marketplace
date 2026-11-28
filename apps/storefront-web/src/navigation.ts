export function navigate(
  href: string,
  options: { preserveScroll?: boolean } = {},
) {
  const next = new URL(href, location.href);
  if (next.origin !== location.origin) {
    location.assign(next.href);
    return;
  }
  history.replaceState(
    { ...history.state, cherrivaScroll: scrollY },
    "",
    location.href,
  );
  history.pushState(
    { cherrivaScroll: options.preserveScroll ? scrollY : 0 },
    "",
    next.href,
  );
  dispatchEvent(new Event("cherriva:navigate"));
}
export function installNavigation() {
  history.scrollRestoration = "manual";
  const click = (event: MouseEvent) => {
    if (
      event.defaultPrevented ||
      event.button !== 0 ||
      event.ctrlKey ||
      event.metaKey ||
      event.shiftKey ||
      event.altKey
    )
      return;
    const a = (event.target as Element)?.closest<HTMLAnchorElement>("a[href]");
    if (!a || a.target || a.hasAttribute("download")) return;
    const next = new URL(a.href, location.href);
    if (
      next.origin !== location.origin ||
      (next.pathname !== "/" &&
        !["/login", "/cart", "/checkout", "/wallet", "/account"].includes(
          next.pathname,
        ) &&
        !next.pathname.startsWith("/account/") &&
        !next.pathname.startsWith("/info/") &&
        next.pathname !== "/search" &&
        next.pathname !== "/offers" &&
        !next.pathname.startsWith("/category/") &&
        !next.pathname.startsWith("/product/"))
    )
      return;
    if (
      next.pathname === location.pathname &&
      next.search === location.search &&
      next.hash
    )
      return;
    event.preventDefault();
    navigate(next.href);
  };
  document.addEventListener("click", click);
  return () => document.removeEventListener("click", click);
}
