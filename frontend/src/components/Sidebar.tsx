import { type LucideIcon, LayoutDashboard, KeyRound, Server, Activity, Moon, Sun, LogOut, Cpu } from "lucide-react";
import ThrotlIcon from "@/assets/throtl-icon.svg";
import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth";

export type Page = "overview" | "keys" | "providers" | "models" | "usage";

interface NavItem {
  id: Page;
  label: string;
  icon: LucideIcon;
}

const nav: NavItem[] = [
  { id: "overview", label: "Overview", icon: LayoutDashboard },
  { id: "keys", label: "API Keys", icon: KeyRound },
  { id: "providers", label: "Providers", icon: Server },
  { id: "models", label: "Models", icon: Cpu },
  { id: "usage", label: "Usage", icon: Activity },
];

interface SidebarProps {
  current: Page;
  onNavigate: (page: Page) => void;
}

export function Sidebar({ current, onNavigate }: SidebarProps) {
  const { user, logout } = useAuth();
  const [dark, setDark] = useState(() => {
    if (typeof window !== "undefined") {
      const saved = localStorage.getItem("throtl-theme");
      if (saved !== null) return saved === "dark";
      return document.documentElement.classList.contains("dark");
    }
    return false;
  });

  useEffect(() => {
    document.documentElement.classList.toggle("dark", dark);
    localStorage.setItem("throtl-theme", dark ? "dark" : "light");
  }, [dark]);

  return (
    <aside className="fixed left-0 top-0 z-40 h-screen w-64 border-r border-border bg-sidebar flex flex-col">
      {/* Logo */}
      <div className="flex items-center gap-3 px-6 py-5 border-b border-border">
        <div className="flex h-8 w-8 items-center justify-center overflow-hidden">
          <img src={ThrotlIcon} alt="Throtl" className="h-full w-full object-contain" />
        </div>
        <div>
          <h1 className="text-sm font-semibold tracking-tight">Throtl</h1>
          <p className="text-xs text-muted-foreground">API Gateway</p>
        </div>
      </div>

      {/* Nav */}
      <nav className="flex-1 px-3 py-4 space-y-0.5">
        {nav.map((item) => {
          const Icon = item.icon;
          const isActive = current === item.id;
          return (
            <button
              key={item.id}
              onClick={() => onNavigate(item.id)}
              className={`group relative flex w-full items-center gap-3 px-3 py-2.5 text-sm font-medium overflow-hidden ${
                isActive
                  ? "text-primary"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              {/* Active indicator bar */}
              <span
                className={`absolute left-0 top-1/2 -translate-y-1/2 h-5 w-0.5 bg-primary transition-all duration-200 ${
                  isActive ? "opacity-100 scale-y-100" : "opacity-0 scale-y-0 group-hover:opacity-50 group-hover:scale-y-75"
                }`}
              />
              {/* Active/ hover background */}
              <span
                className={`absolute inset-0 transition-all duration-200 ${
                  isActive
                    ? "bg-primary/[0.08]"
                    : "group-hover:bg-accent"
                }`}
              />
              {/* Active glow */}
              {isActive && (
                <span className="absolute inset-y-0 left-0 w-32 bg-gradient-to-r from-primary/[0.06] to-transparent pointer-events-none" />
              )}
              <Icon className={`h-4 w-4 relative z-10 transition-all duration-200 group-hover:scale-110 ${
                isActive ? "text-primary" : ""
              }`} />
              <span className="relative z-10 transition-all duration-200 group-hover:translate-x-0.5">{item.label}</span>
            </button>
          );
        })}
      </nav>

      {/* Footer */}
      <div className="border-t border-border px-3 py-3 space-y-0.5">
        <button
          onClick={() => setDark(!dark)}
          className="flex w-full items-center gap-3 px-3 py-2.5 text-sm font-medium text-muted-foreground hover:bg-accent hover:text-foreground transition-all duration-200"
        >
          {dark ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
          {dark ? "Light mode" : "Dark mode"}
        </button>
        <button
          onClick={logout}
          className="flex w-full items-center gap-3 px-3 py-2.5 text-sm font-medium text-muted-foreground hover:bg-destructive/10 hover:text-destructive transition-all duration-200"
        >
          <LogOut className="h-4 w-4" />
          Sign out
        </button>
        {user && (
          <div className="px-3 pt-2 text-xs text-muted-foreground truncate border-t border-border/50 mt-2">
            {user.email}
          </div>
        )}
      </div>
    </aside>
  );
}
