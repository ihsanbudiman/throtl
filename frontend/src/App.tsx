import { lazy, Suspense } from "react";
import { AuthProvider, useAuth } from "@/lib/auth";
import { Sidebar, type Page } from "@/components/Sidebar";
import { ToastProvider } from "@/hooks/use-toast";
import { ToastContainer } from "@/components/ToastContainer";
import { TooltipProvider } from "@/components/ui/tooltip";
import {
  BrowserRouter,
  Routes,
  Route,
  Navigate,
  useNavigate,
  useLocation,
} from "react-router-dom";
import { cn } from "@/lib/utils";

const OverviewPage = lazy(() => import("@/pages/OverviewPage"));
const KeysPage = lazy(() => import("@/pages/KeysPage"));
const ProvidersPage = lazy(() => import("@/pages/ProvidersPage"));
const ModelsPage = lazy(() => import("@/pages/ModelsPage"));
const UsagePage = lazy(() => import("@/pages/UsagePage"));
const LoginPage = lazy(() => import("@/pages/LoginPage"));
const SetupPage = lazy(() => import("@/pages/SetupPage"));

function PageLoader() {
  return (
    <div className="flex min-h-[50vh] items-center justify-center">
      <div className="flex flex-col items-center gap-3">
        <div className="h-8 w-8 rounded-full border-2 border-primary/30 border-t-primary animate-spin" />
        <p className="text-xs text-muted-foreground">Loading...</p>
      </div>
    </div>
  );
}

const PATH_TO_PAGE: Record<string, Page> = {
  "/": "overview",
  "/keys": "keys",
  "/providers": "providers",
  "/models": "models",
  "/usage": "usage",
};

const PAGE_TO_PATH: Record<Page, string> = {
  overview: "/",
  keys: "/keys",
  providers: "/providers",
  models: "/models",
  usage: "/usage",
};

function AppContent() {
  const { user, loading, setupRequired } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  if (loading || setupRequired === null) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <div className="flex flex-col items-center gap-3">
          <div className="h-8 w-8 rounded-full border-2 border-primary/30 border-t-primary animate-spin" />
          <div className="animate-pulse text-muted-foreground text-sm">Loading...</div>
        </div>
      </div>
    );
  }

  if (setupRequired) {
    return (
      <Suspense fallback={<PageLoader />}>
        <SetupPage />
      </Suspense>
    );
  }

  if (!user) {
    return (
      <Suspense fallback={<PageLoader />}>
        <LoginPage />
      </Suspense>
    );
  }

  const currentPage = PATH_TO_PAGE[location.pathname] ?? "overview";

  const handleNavigate = (page: Page) => {
    navigate(PAGE_TO_PATH[page]);
  };

  return (
    <TooltipProvider>
      <div className="min-h-screen bg-background">
        <Sidebar current={currentPage} onNavigate={handleNavigate} />
        <main className={cn(
          "p-4 sm:p-6 lg:p-8 pt-20 lg:pt-8",
          "lg:ml-64"
        )}>
          <Suspense fallback={<PageLoader />}>
            <Routes>
              <Route path="/" element={<OverviewPage />} />
              <Route path="/keys" element={<KeysPage />} />
              <Route path="/providers" element={<ProvidersPage />} />
              <Route path="/models" element={<ModelsPage />} />
              <Route path="/usage" element={<UsagePage />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </Suspense>
        </main>
      </div>
    </TooltipProvider>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <ToastProvider>
        <BrowserRouter>
          <AppContent />
          <ToastContainer />
        </BrowserRouter>
      </ToastProvider>
    </AuthProvider>
  );
}
