import { AuthProvider, useAuth } from "@/lib/auth";
import { Sidebar, type Page } from "@/components/Sidebar";
import { OverviewPage } from "@/pages/OverviewPage";
import { KeysPage } from "@/pages/KeysPage";
import { ProvidersPage } from "@/pages/ProvidersPage";
import { ModelsPage } from "@/pages/ModelsPage";
import { UsagePage } from "@/pages/UsagePage";
import { LoginPage } from "@/pages/LoginPage";
import { SetupPage } from "@/pages/SetupPage";
import { TooltipProvider } from "@/components/ui/tooltip";
import {
  BrowserRouter,
  Routes,
  Route,
  Navigate,
  useNavigate,
  useLocation,
} from "react-router-dom";

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
        <div className="animate-pulse text-muted-foreground text-sm">Loading...</div>
      </div>
    );
  }

  if (setupRequired) {
    return <SetupPage />;
  }

  if (!user) {
    return <LoginPage />;
  }

  const currentPage = PATH_TO_PAGE[location.pathname] ?? "overview";

  const handleNavigate = (page: Page) => {
    navigate(PAGE_TO_PATH[page]);
  };

  return (
    <TooltipProvider>
      <div className="min-h-screen bg-background">
        <Sidebar current={currentPage} onNavigate={handleNavigate} />
        <main className="ml-64 p-8">
          <Routes>
            <Route path="/" element={<OverviewPage />} />
            <Route path="/keys" element={<KeysPage />} />
            <Route path="/providers" element={<ProvidersPage />} />
            <Route path="/models" element={<ModelsPage />} />
            <Route path="/usage" element={<UsagePage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </main>
      </div>
    </TooltipProvider>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <AppContent />
      </BrowserRouter>
    </AuthProvider>
  );
}
