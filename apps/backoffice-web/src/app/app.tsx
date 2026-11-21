import { useEffect, useState } from "react";
import {
  BrowserRouter,
  Link,
  NavLink,
  Navigate,
  Route,
  Routes,
  useLocation,
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
} from "@mui/icons-material";
import { SharedAttributesPage } from "../features/management/shared-attributes";
import { ProductDetailsPage } from "../features/management/product-workspace";
import { BannersPage } from "../features/management/banners-page";
import { api, ApiError } from "../shared/api";
import type { Command, Role, Session } from "../shared/types";
import { WorkspaceContext } from "../shared/workspace";
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
  palette: { primary: { main: "#207863" }, background: { default: "#f5f7f8" } },
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
        path: "/users",
        label: "کاربران و نقش‌ها",
        icon: <PeopleOutline />,
        roles: ["super-admin"],
      },
      { path: "/stores", label: "فروشگاه‌ها", icon: <StorefrontOutlined /> },
      {
        path: "/categories",
        label: "دسته‌ها و خصوصیات",
        icon: <CategoryOutlined />,
      },
      { path: "/brands", label: "برندها", icon: <LabelOutlined /> },
      {
        path: "/banners",
        label: "بنرهای دو اندازه",
        icon: <Inventory2Outlined />,
      },
      {
        path: "/attributes",
        label: "کتابخانه خصوصیات",
        icon: <CategoryOutlined />,
      },
      {
        path: "/products",
        label: "محصولات و محتوا",
        icon: <Inventory2Outlined />,
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
    heading: "عملیات و پشتیبانی",
    items: [
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
  });
  useEffect(() => {
    if (query.error instanceof ApiError && query.error.status === 401) logout();
  }, [query.error, logout]);
  const current = menu
    .flatMap((g) => g.items)
    .find((i) => i.path === location.pathname);
  const role = session.user.Role;
  const visible = (roles?: string[]) => !roles || roles.includes(role);
  return (
    <div className="app-layout">
      <aside className={"sidebar " + (menuOpen ? "is-open" : "")}>
        <Link className="brand" to="/">
          <span className="brand-mark">م</span>
          <span>
            مارکت‌پلیس<small>مرکز مدیریت بازارگاه</small>
          </span>
        </Link>
        <div className="workspace-badge">
          <span className="workspace-symbol">◈</span>
          <div>
            فضای مدیریت<small>محیط آزمون و بررسی</small>
          </div>
          <span className="live-dot" />
        </div>
        <nav>
          {menu.map((group) => (
            <div className="nav-group" key={group.heading}>
              <h3>{group.heading}</h3>
              {group.items
                .filter((i) =>
                  session.user.Surface === "seller"
                    ? i.path === "/products"
                    : visible("roles" in i ? i.roles : undefined),
                )
                .map((item) => (
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
          ))}
        </nav>
        <div className="sidebar-bottom">
          <div className="avatar">{role === "super-admin" ? "م" : "ک"}</div>
          <div>
            <b>کاربر آزمایشی</b>
            <small>{role}</small>
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
            <div className="avatar small-avatar">م</div>
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
                state: query.data,
                role,
                surface: session.user.Surface,
                loadBanners: () => api.banners(session.token),
                loadBannerAsset: (id) => api.bannerAsset(session.token, id),
                uploadBanner: async (input) => {
                  const result = await api.uploadBanner(session.token, input);
                  await cache.invalidateQueries({ queryKey: ["banners"] });
                  return result;
                },
                loadProduct: (id) =>
                  api.productWorkspace(session.token, id, session.user.Surface),
                resolveVariant: (id, values) =>
                  api.resolveVariant(
                    session.token,
                    id,
                    values,
                    session.user.Surface,
                  ),
                execute: (command) => mutation.mutateAsync(command),
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
                busy: mutation.isPending,
              }}
            >
              <Routes>
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
                <Route path="/attributes" element={<SharedAttributesPage />} />
                {(
                  [
                    "users",
                    "stores",
                    "categories",
                    "brands",
                    "products",
                    "pricing",
                    "shipping",
                  ] as const
                ).map((section) => (
                  <Route
                    key={section}
                    path={"/" + section}
                    element={
                      (session.user.Surface === "seller" &&
                        section !== "products") ||
                      (section === "users" && role !== "super-admin") ||
                      (section === "pricing" &&
                        !["super-admin", "finance"].includes(role)) ? (
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
                    ["super-admin", "finance"].includes(role) ? (
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
  const [session, setSession] = useState<Session | null>(null);
  const logout = () => {
    if (session)
      void api
        .logout(session.token, session.user.Surface)
        .catch(() => undefined);
    setSession(null);
    queryClient.clear();
  };
  return session ? (
    <Workspace session={session} logout={logout} />
  ) : (
    <LoginPage onLogin={setSession} />
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
