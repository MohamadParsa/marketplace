import { StoresPage } from "../features/management/stores-page";
import { StorePreparationPage } from "../features/management/store-preparation-page";
import { UsersPage } from "../features/management/users-page";
import { RolesPage } from "../features/management/access-panel";
import { TemporaryPassword } from "../../../../packages/identity-client/temporary-password";
import { CouponsPage } from "../features/commerce/coupons-page";
import { useCallback, useEffect, useState } from "react";
import {
  BrowserRouter,
  Link,
  NavLink,
  Navigate,
  Route,
  Routes,
  useLocation,
  useNavigate,
} from "react-router-dom";
import {
  QueryClient,
  QueryClientProvider,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  Alert,
  Button,
  CircularProgress,
  Snackbar,
  ThemeProvider,
  createTheme,
} from "@mui/material";
import {
  DashboardOutlined,
  PeopleOutline,
  StorefrontOutlined,
  CategoryOutlined,
  LabelOutlined,
  Inventory2Outlined,
  PercentOutlined,
  LocalShippingOutlined,
  ReceiptLongOutlined,
  SupportAgentOutlined,
  SwapHorizOutlined,
  AccountBalanceWalletOutlined,
  NotificationsNoneOutlined,
  AssessmentOutlined,
  LogoutOutlined,
  MenuOutlined,
  RateReviewOutlined,
} from "@mui/icons-material";
import {
  CommercePage,
  SellerCheckoutPage,
} from "../features/commerce/commerce-page";
import { SharedAttributesPage } from "../features/management/shared-attributes";
import { ProductDetailsPage } from "../features/management/product-workspace";
import { ProfilePage } from "../features/management/profile-page";
import { SalesMethodsPage } from "../features/commerce/sales-methods-page";
import { CatalogPage } from "../features/management/catalog-pages";
import { LegacyCatalogRoute } from "../features/management/legacy-catalog-route";
import { ProductReviewsPage } from "../features/management/product-reviews-page";
import { BannersPage } from "../features/management/banners-page";
import { api, ApiError } from "../shared/api";
import type { Command, Role, Session } from "../shared/types";
import { can, WorkspaceContext } from "../shared/workspace";
import { DashboardPage } from "../features/dashboard/dashboard-page";
import { ManagementPage } from "../features/management/management-page";
import { OperationsPage } from "../features/operations/operations-page";
import { FinancePage } from "../features/finance/finance-page";
import { LoginPage } from "./login-page";
const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false, refetchOnWindowFocus: false } },
});
const theme = createTheme({
  direction: "rtl",
  typography: {
    fontFamily: "Vazirmatn, Tahoma, sans-serif",
    button: { textTransform: "none", fontWeight: 600 },
  },
  palette: { primary: { main: "#B51F46" }, background: { default: "#F7F7F9" } },
  shape: { borderRadius: 10 },
  components: {
    MuiButton: { defaultProps: { disableElevation: true } },
    MuiDialog: { defaultProps: { dir: "rtl" } },
  },
});
const menu = [
  {
    heading: "نمای کلی",
    items: [{ path: "/", label: "داشبورد", icon: <DashboardOutlined /> }],
  },
  {
    heading: "مدیریت بازارگاه",
    items: [
      {
        path: "/sales-methods",
        label: "روش‌های فروش",
        icon: <Inventory2Outlined />,
      },
      {
        path: "/users",
        label: "کاربران",
        icon: <PeopleOutline />,
        roles: ["super-admin"],
      },
      {
        path: "/roles",
        label: "نقش‌ها و مجوزها",
        icon: <DashboardOutlined />,
        roles: ["super-admin"],
      },
      { path: "/stores", label: "فروشگاه‌ها", icon: <StorefrontOutlined /> },
      { path: "/store-preparation", label: "وقفه و تقویم فروشگاه", icon: <LocalShippingOutlined /> },
      {
        path: "/product-reviews",
        label: "نظرات محصولات",
        icon: <RateReviewOutlined />,
        roles: ["super-admin"],
      },
      {
        path: "/pricing",
        label: "قیمت و کمیسیون",
        icon: <PercentOutlined />,
        roles: ["super-admin", "finance"],
      },
      {
        path: "/shipping",
        label: "ارسال و تقویم",
        icon: <LocalShippingOutlined />,
      },
    ],
  },
  {
    heading: "کاتالوگ",
    items: [
      {
        path: "/categories",
        label: "دسته‌بندی‌ها",
        icon: <CategoryOutlined />,
      },
      {
        path: "/attributes",
        label: "کتابخانه خصوصیات",
        icon: <CategoryOutlined />,
      },
      { path: "/brands", label: "برندها", icon: <LabelOutlined /> },
      {
        path: "/products",
        label: "محصولات و محتوا",
        icon: <Inventory2Outlined />,
      },
    ],
  },
  {
    heading: "محتوا",
    items: [
      {
        path: "/home-settings",
        label: "مدیریت صفحه اصلی",
        icon: <DashboardOutlined />,
      },
      {
        path: "/banners",
        label: "بنرهای دو اندازه",
        icon: <Inventory2Outlined />,
      },
    ],
  },
  {
    heading: "عملیات و پشتیبانی",
    items: [
      {
        path: "/commerce",
        label: "سفارش و پرداخت سایت",
        icon: <ReceiptLongOutlined />,
      },
      {
        path: "/coupons",
        label: "کدهای تخفیف",
        icon: <PercentOutlined />,
      },
      {
        path: "/commerce-settings",
        label: "تنظیمات خرید و حمل",
        icon: <LocalShippingOutlined />,
      },
      { path: "/orders", label: "سفارش‌ها", icon: <ReceiptLongOutlined /> },
      {
        path: "/support",
        label: "پشتیبانی و پرونده‌ها",
        icon: <SupportAgentOutlined />,
      },
      {
        path: "/assignments",
        label: "تأمین جایگزین",
        icon: <SwapHorizOutlined />,
      },
      {
        path: "/finance",
        label: "عملیات مالی",
        icon: <AccountBalanceWalletOutlined />,
        roles: ["super-admin", "finance"],
      },
    ],
  },
  {
    heading: "گزارش و اطلاع‌رسانی",
    items: [
      {
        path: "/notifications",
        label: "اعلان‌ها",
        icon: <NotificationsNoneOutlined />,
      },
      {
        path: "/reports",
        label: "گزارش فعالیت‌ها",
        icon: <AssessmentOutlined />,
      },
    ],
  },
];
function Workspace({
  session,
  logout,
}: {
  session: Session;
  logout: () => void;
}) {
  const cache = useQueryClient();
  const location = useLocation();
  const [menuOpen, setMenuOpen] = useState(false);
  const [notice, setNotice] = useState("");
  const loadCatalogMedia = useCallback(
    (id: string) => api.catalogMedia(session.token, id),
    [session.token],
  );
  const loadProfile = useCallback(
    () => api.profile(session.token, session.user.Surface),
    [session.token],
  );
  const query = useQuery({
    queryKey: ["workspace", session.token],
    queryFn: () => api.workspace(session.token, session.user.Surface),
  });
  const mutation = useMutation({
    mutationFn: (command: Command) =>
      api.execute(session.token, command, session.user.Surface),
    onSuccess: async (_result, command) => {
      await cache.invalidateQueries({ queryKey: ["workspace"] });
      await cache.invalidateQueries({ queryKey: ["product"] });
      await cache.invalidateQueries({ queryKey: ["banners"] });
      await cache.invalidateQueries({ queryKey: ["variant-resolution"] });
      setNotice(
        command.action.startsWith("banner.") ? "" : "تغییر با موفقیت ثبت شد.",
      );
    },
    onError: (error) => {
      if (error instanceof ApiError && error.status === 401) logout();
    },
  });
  useEffect(() => {
    if (query.error instanceof ApiError && query.error.status === 401) logout();
  }, [query.error, logout]);
  const catalogActive = [
    "/categories",
    "/attributes",
    "/brands",
    "/products",
  ].some(
    (path) =>
      location.pathname === path || location.pathname.startsWith(path + "/"),
  );
  const [catalogOpen, setCatalogOpen] = useState(catalogActive);
  useEffect(() => {
    if (catalogActive) setCatalogOpen(true);
  }, [catalogActive, location.pathname]);
  const current = menu
    .flatMap((g) => g.items)
    .find((i) => i.path === location.pathname);
  const role = session.user.Role;
  const permissions = session.user.Permissions ?? [];
  const canCreditWallet =
    role === "super-admin" ||
    role === "finance" ||
    permissions.includes("commerce.wallet.credit");
  const financeVisible =
    role === "super-admin" ||
    role === "finance" ||
    permissions.includes("finance.read");
  const visible = (roles?: string[]) =>
    !roles ||
    roles.includes(role) ||
    (roles.includes("finance") && financeVisible);
  const visibleMenu = menu
    .map((group) => ({
      ...group,
      items: group.items.filter((i) =>
        session.user.Surface === "seller"
          ? ["/products", "/commerce", "/store-preparation"].includes(i.path)
          : i.path === "/users"
            ? canCreditWallet || permissions.includes("users.read")
            : ["/stores", "/store-preparation"].includes(i.path)
              ? role === "super-admin" ||
                role === "operations" ||
                permissions.includes("stores.read")
              : visible("roles" in i ? i.roles : undefined),
      ),
    }))
    .filter((group) => group.items.length > 0);
  return (
    <div className="app-layout">
      <aside className={"sidebar " + (menuOpen ? "is-open" : "")}>
        <Link className="brand" to="/">
          <img
            src="/brand/cherriva-wordmark.svg"
            alt="Cherriva"
            width="160"
            height="44"
          />
          <small>
            فضای مدیریت
            <br />
            چریوا
          </small>
        </Link>
        <div className="workspace-badge">
          <span className="workspace-symbol">◈</span>
          <div>
            فضای مدیریت<small>محیط آزمون و بررسی</small>
          </div>
          <span className="live-dot" />
        </div>
        <nav>
          {visibleMenu.map((group) => (
            <div className="nav-group" key={group.heading}>
              {group.heading === "کاتالوگ" ? (
                <button
                  className={
                    "catalog-nav-parent " + (catalogActive ? "is-current" : "")
                  }
                  aria-expanded={catalogOpen}
                  aria-controls="catalog-nav-children"
                  onClick={() => setCatalogOpen((v) => !v)}
                >
                  <CategoryOutlined />
                  <span>کاتالوگ</span>
                  <span aria-hidden="true">{catalogOpen ? "⌃" : "⌄"}</span>
                </button>
              ) : (
                <h3>{group.heading}</h3>
              )}
              <div
                id={
                  group.heading === "کاتالوگ"
                    ? "catalog-nav-children"
                    : undefined
                }
                className={
                  group.heading === "کاتالوگ"
                    ? "catalog-nav-children"
                    : undefined
                }
                hidden={group.heading === "کاتالوگ" && !catalogOpen}
              >
                {group.items.map((item) => (
                  <NavLink
                    key={item.path}
                    to={item.path}
                    end={item.path === "/"}
                    onClick={() => setMenuOpen(false)}
                  >
                    {item.icon}
                    <span>{item.label}</span>
                    {item.path === "/support" && (
                      <b>
                        {query.data?.cases?.filter((c) => c.status !== "closed")
                          .length ?? 0}
                      </b>
                    )}
                  </NavLink>
                ))}
              </div>
            </div>
          ))}
        </nav>
        <div className="sidebar-bottom">
          <div className="avatar">{role === "super-admin" ? "م" : "ک"}</div>
          <div>
            <b>کاربر آزمایشی</b>
            <small>
              {role === "super-admin" ? "ادمین" : role || "نقش سفارشی"}
            </small>
          </div>
          <button aria-label="خروج از حساب" onClick={logout}>
            <LogoutOutlined />
          </button>
        </div>
      </aside>
      {menuOpen && (
        <button
          className="sidebar-overlay"
          aria-label="بستن منو"
          onClick={() => setMenuOpen(false)}
        />
      )}
      <div className="workspace-main">
        <header className="topbar">
          <div className="breadcrumb">
            <button
              className="mobile-menu"
              aria-label="باز کردن منو"
              onClick={() => setMenuOpen(true)}
            >
              <MenuOutlined />
            </button>
            <span>
              {session.user.Surface === "seller" ? "پنل فروشنده" : "بک‌آفیس"}
            </span>
            <i>/</i>
            <strong>
              {current?.label ??
                (location.pathname.startsWith("/products/")
                  ? "جزئیات محصول"
                  : "جزئیات سفارش")}
            </strong>
          </div>
          <div className="topbar-actions">
            <span className="environment">
              <i />
              محیط آزمون
            </span>
            <Link aria-label="اعلان‌ها" to="/notifications">
              <NotificationsNoneOutlined />
            </Link>
            <Link
              to="/profile"
              className="avatar small-avatar"
              aria-label="پروفایل و تنظیم رمز"
            >
              ک
            </Link>
          </div>
        </header>
        <main className="content">
          {query.isPending ? (
            <div className="loading">
              <CircularProgress />
              <p>در حال دریافت اطلاعات…</p>
            </div>
          ) : query.isError ? (
            <Alert
              severity="error"
              action={
                <Button onClick={() => query.refetch()}>تلاش مجدد</Button>
              }
            >
              {query.error.message}
            </Alert>
          ) : (
            <WorkspaceContext.Provider
              value={{
                authToken: session.token,
                actorID: session.user.ID,
                readCustomerWallet: (user, key) =>
                  api.readCustomerWallet(session.token, user, key),
                creditCustomerWallet: async (input, key) => {
                  try {
                    return await api.creditCustomerWallet(
                      session.token,
                      input,
                      key,
                    );
                  } catch (error) {
                    if (error instanceof ApiError && error.status === 401)
                      logout();
                    throw error;
                  }
                },
                state: query.data,
                role,
                surface: session.user.Surface,
                loadProfile,
                updateProfile: (input) =>
                  api.updateProfile(session.token, input, session.user.Surface),
                loadCatalogMedia,
                loadBanners: () => api.banners(session.token),
                loadBannerAsset: (id) => api.bannerAsset(session.token, id),
                uploadBanner: async (input) => {
                  const result = await api.uploadBanner(session.token, input);
                  await cache.invalidateQueries({ queryKey: ["banners"] });
                  return result;
                },
                loadProductQA: (id, after) =>
                  api.productQA(session.token, id, after),
                changeProductQA: (id, question, mutation, key) =>
                  api.changeProductQA(
                    session.token,
                    id,
                    question,
                    mutation,
                    key,
                  ),
                loadProduct: (id) =>
                  api.productWorkspace(session.token, id, session.user.Surface),
                resolveVariant: (id, values) =>
                  api.resolveVariant(
                    session.token,
                    id,
                    values,
                    session.user.Surface,
                  ),
                uploadIcon: async (input) => {
                  try {
                    await api.uploadIcon(session.token, input);
                    await cache.invalidateQueries({ queryKey: ["workspace"] });
                    setNotice("آیکن با موفقیت ثبت شد.");
                  } catch (error) {
                    if (error instanceof ApiError && error.status === 401)
                      logout();
                    throw error;
                  }
                },
                can: (action, area) =>
                  role === "super-admin" ||
                  (!!area && can(role, area)) ||
                  permissions.includes(action),
                execute: (command) => mutation.mutateAsync(command),
                requestAdminCode: () => api.requestAdminCode(session.token),
                busy: mutation.isPending,
              }}
            >
              <Routes>
                <Route path="/sales-methods" element={<SalesMethodsPage />} />
                <Route
                  path="/coupons"
                  element={<CouponsPage token={session.token} />}
                />
                <Route
                  path="/commerce"
                  element={
                    session.user.Surface === "seller" ? (
                      <SellerCheckoutPage token={session.token} />
                    ) : (
                      <CommercePage token={session.token} />
                    )
                  }
                />
                <Route
                  path="/commerce-settings"
                  element={
                    session.user.Surface === "seller" ? (
                      <Navigate to="/products" />
                    ) : (
                      <CommercePage token={session.token} settingsOnly />
                    )
                  }
                />
                <Route path="/profile" element={<ProfilePage />} />
                <Route path="/store-preparation" element={<StorePreparationPage />} />
                <Route
                  path="/stores"
                  element={
                    session.user.Surface === "backoffice" &&
                    (role === "super-admin" ||
                      role === "operations" ||
                      permissions.includes("stores.read")) ? (
                      <StoresPage />
                    ) : (
                      <Navigate to="/" replace />
                    )
                  }
                />
                <Route
                  path="/home-catalog"
                  element={
                    session.user.Surface === "seller" ? (
                      <Navigate to="/" replace />
                    ) : (
                      <LegacyCatalogRoute />
                    )
                  }
                />
                <Route
                  path="/"
                  element={
                    session.user.Surface === "seller" ? (
                      <Navigate to="/products" replace />
                    ) : (
                      <DashboardPage />
                    )
                  }
                />
                {(["categories", "brands", "products", "home"] as const).map(
                  (section) => (
                    <Route
                      key={section}
                      path={
                        section === "home" ? "/home-settings" : "/" + section
                      }
                      element={
                        session.user.Surface === "seller" &&
                        section !== "products" ? (
                          <Navigate to="/products" replace />
                        ) : (
                          <CatalogPage key={section} section={section} />
                        )
                      }
                    />
                  ),
                )}
                <Route path="/products/:id" element={<ProductDetailsPage />} />
                <Route
                  path="/banners"
                  element={
                    session.user.Surface === "seller" ? (
                      <Navigate to="/products" replace />
                    ) : (
                      <BannersPage />
                    )
                  }
                />
                <Route
                  path="/product-reviews"
                  element={
                    session.user.Surface === "backoffice" &&
                    role === "super-admin" ? (
                      <ProductReviewsPage token={session.token} />
                    ) : (
                      <Navigate to="/" replace />
                    )
                  }
                />
                <Route
                  path="/attributes"
                  element={
                    session.user.Surface === "seller" ? (
                      <Navigate to="/products" replace />
                    ) : (
                      <SharedAttributesPage />
                    )
                  }
                />
                <Route
                  path="/users"
                  element={
                    session.user.Surface === "backoffice" &&
                    (canCreditWallet || permissions.includes("users.read")) ? (
                      <UsersPage />
                    ) : (
                      <Navigate to="/" replace />
                    )
                  }
                />
                <Route
                  path="/roles"
                  element={
                    session.user.Surface === "backoffice" &&
                    role === "super-admin" ? (
                      <RolesPage />
                    ) : (
                      <Navigate to="/" replace />
                    )
                  }
                />
                {(["pricing", "shipping"] as const).map((section) => (
                  <Route
                    key={section}
                    path={"/" + section}
                    element={
                      session.user.Surface === "seller" ||
                      (section === "pricing" && !financeVisible) ? (
                        <Navigate to="/" replace />
                      ) : (
                        <ManagementPage key={section} section={section} />
                      )
                    }
                  />
                ))}
                {(
                  [
                    "orders",
                    "support",
                    "assignments",
                    "notifications",
                    "reports",
                  ] as const
                ).map((section) => (
                  <Route
                    key={section}
                    path={"/" + section}
                    element={<OperationsPage key={section} section={section} />}
                  />
                ))}
                <Route
                  path="/orders/:id"
                  element={<OperationsPage section="orders" />}
                />
                <Route
                  path="/finance"
                  element={
                    financeVisible ? (
                      <FinancePage />
                    ) : (
                      <Navigate to="/" replace />
                    )
                  }
                />
                <Route
                  path="*"
                  element={
                    <section className="panel detail">
                      <h1>صفحه پیدا نشد</h1>
                      <Link to="/">بازگشت به داشبورد</Link>
                    </section>
                  }
                />
              </Routes>
            </WorkspaceContext.Provider>
          )}
          <footer className="page-footer">
            <span>مارکت‌پلیس · بک‌آفیس</span>
            <span>داده آزمون · بدون اتصال واقعی</span>
          </footer>
        </main>
      </div>
      <Snackbar
        open={!!notice}
        autoHideDuration={3500}
        onClose={() => setNotice("")}
        message={notice}
      />
    </div>
  );
}
function SessionApp() {
  const location = useLocation(),
    navigate = useNavigate();
  const loginSurface =
    location.pathname === "/seller-login" ? "seller" : "backoffice";
  const [session, setSession] = useState<Session | null>(null);
  const logout = () => {
    if (session)
      void api
        .logout(session.token, session.user.Surface)
        .catch(() => undefined);
    setSession(null);
    queryClient.clear();
  };
  return session?.user.PasswordChangeRequired ? (
    <TemporaryPassword<Session>
      token={session.token}
      surface={session.user.Surface}
      base={import.meta.env.VITE_API_BASE_URL ?? ""}
      onComplete={setSession}
      onCancel={logout}
    />
  ) : session ? (
    <Workspace session={session} logout={logout} />
  ) : (
    <LoginPage
      key={loginSurface}
      surface={loginSurface}
      onLogin={(s) => {
        setSession(s);
        if (s.user.Surface === "seller")
          navigate("/products", { replace: true });
      }}
    />
  );
}
export function App() {
  return (
    <ThemeProvider theme={theme}>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <SessionApp />
        </BrowserRouter>
      </QueryClientProvider>
    </ThemeProvider>
  );
}
