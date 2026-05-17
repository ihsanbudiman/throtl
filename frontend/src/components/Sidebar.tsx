import { type LucideIcon, LayoutDashboard, KeyRound, Server, Activity, LogOut, Cpu, Menu, X } from "lucide-react";
import ThrotlIcon from "@/assets/throtl-icon.svg";
import { useState } from "react";
import { useAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";

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

function NavItem({ item, current, onNavigate, onClick }: { item: NavItem; current: Page; onNavigate: (page: Page) => void; onClick?: () => void }) {
  const Icon = item.icon;
  const isActive = current === item.id;

  return (
    <button
      onClick={() => { onNavigate(item.id); onClick?.(); }}
      className={cn(
        "group relative flex w-full items-center gap-3 px-3 py-2 text-sm font-[400] transition-all duration-150",
        isActive
          ? "text-sidebar-foreground"
          : "text-body-mid hover:text-sidebar-foreground"
      )}
    >
      {isActive && (
        <span className="absolute left-0 top-1/2 -translate-y-1/2 h-5 w-0.5 bg-sidebar-primary rounded-full" />
      )}
      <Icon className={cn(
        "h-4 w-4 transition-all duration-150 shrink-0",
        isActive ? "text-sidebar-primary" : ""
      )} />
      <span>{item.label}</span>
    </button>
  );
}

export function Sidebar({ current, onNavigate }: SidebarProps) {
  const { user, logout } = useAuth();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [logoutOpen, setLogoutOpen] = useState(false);

  const handleNavigate = (page: Page) => {
    onNavigate(page);
    setMobileOpen(false);
  };

  const sidebarContent = (
    <>
      <div className="flex items-center gap-3 px-5 py-5 border-b border-sidebar-border">
        <div className="flex h-7 w-7 items-center justify-center overflow-hidden shrink-0">
          <img src={ThrotlIcon} alt="Throtl" className="h-full w-full object-contain invert" />
        </div>
        <div>
          <h1 className="text-sm font-[400] tracking-tight text-sidebar-foreground">Throtl</h1>
          <p className="text-[11px] text-body-mid tracking-widest uppercase font-mono">API Gateway</p>
        </div>
      </div>
      <nav className="flex-1 px-3 py-3 space-y-0.5 overflow-y-auto">
        {nav.map((item) => (
          <NavItem key={item.id} item={item} current={current} onNavigate={handleNavigate} />
        ))}
      </nav>
      <div className="border-t border-sidebar-border px-3 py-3 space-y-0.5">
        <button
          onClick={() => setLogoutOpen(true)}
          className="flex w-full items-center gap-3 px-3 py-2 text-sm font-[400] text-body-mid hover:text-destructive transition-all duration-150"
        >
          <LogOut className="h-4 w-4 shrink-0" />
          Sign out
        </button>
        {user && (
          <div className="px-3 pt-2 text-xs text-body-mid truncate border-t border-sidebar-border/50 mt-2">
            {user.email}
          </div>
        )}
      </div>
    </>
  );

  return (
    <>
      <div className="fixed top-0 left-0 right-0 z-40 flex items-center gap-3 border-b border-border bg-background/80 backdrop-blur-md px-4 h-14 lg:hidden">
        <button
          onClick={() => setMobileOpen(true)}
          className="flex items-center justify-center h-8 w-8 rounded-[8px] text-body-mid hover:text-foreground hover:bg-canvas-soft transition-colors"
          aria-label="Open menu"
        >
          <Menu className="h-5 w-5" />
        </button>
        <div className="flex items-center gap-2">
          <div className="flex h-5 w-5 items-center justify-center overflow-hidden">
            <img src={ThrotlIcon} alt="Throtl" className="h-full w-full object-contain invert" />
          </div>
          <span className="text-sm font-[400]">Throtl</span>
        </div>
      </div>
      {mobileOpen && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <div className="fixed inset-0 bg-black/60 animate-[fade-in_0.15s_ease-out]" onClick={() => setMobileOpen(false)} />
          <aside className="fixed left-0 top-0 z-50 h-full w-72 bg-sidebar border-r border-sidebar-border rounded-r-[8px] flex flex-col animate-[slide-in-right_0.2s_ease-out]">
            <div className="flex items-center justify-between px-4 pt-4 pb-0">
              <span className="text-xs font-mono text-body-mid tracking-widest uppercase">Navigation</span>
              <button
                onClick={() => setMobileOpen(false)}
                className="flex items-center justify-center h-7 w-7 rounded-[8px] text-body-mid hover:text-foreground hover:bg-sidebar-accent transition-colors"
                aria-label="Close menu"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            </div>
            {sidebarContent}
          </aside>
        </div>
      )}
      <aside className="hidden lg:flex fixed left-0 top-0 z-30 h-screen w-60 border-r border-sidebar-border bg-sidebar flex-col">
        {sidebarContent}
      </aside>

      <Dialog open={logoutOpen} onOpenChange={setLogoutOpen}>
        <DialogContent showCloseButton={false} className="max-w-xs">
          <DialogHeader>
            <DialogTitle>Sign out?</DialogTitle>
            <DialogDescription>You'll need to sign in again to access the dashboard.</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" size="sm" onClick={() => setLogoutOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" size="sm" onClick={() => { logout(); setLogoutOpen(false); }}>
              Sign out
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
