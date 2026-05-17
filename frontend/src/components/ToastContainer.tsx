import { useEffect, useState } from "react";
import { useToast, type Toast } from "@/hooks/use-toast";
import { X, CheckCircle, AlertCircle, Info } from "lucide-react";
import { cn } from "@/lib/utils";

function ToastIcon({ variant }: { variant?: Toast["variant"] }) {
  switch (variant) {
    case "success": return <CheckCircle className="h-4 w-4 text-success" />;
    case "destructive": return <AlertCircle className="h-4 w-4 text-destructive" />;
    default: return <Info className="h-4 w-4 text-primary" />;
  }
}

function ToastItem({ toast }: { toast: Toast }) {
  const [exiting, setExiting] = useState(false);
  const { dismiss } = useToast();

  useEffect(() => {
    const timer = setTimeout(() => setExiting(true), 3700);
    const remove = setTimeout(() => dismiss(toast.id), 4000);
    return () => { clearTimeout(timer); clearTimeout(remove); };
  }, [toast.id, dismiss]);

  return (
    <div
      role="alert"
      className={cn(
        "flex items-start gap-3 rounded-[8px] border bg-card/95 backdrop-blur-md p-3.5 text-sm shadow-xl ring-1 ring-black/5 dark:ring-white/5",
        "transition-all duration-300",
        exiting ? "animate-[slide-out-right_0.3s_ease-in_forwards]" : "animate-[slide-in-right_0.3s_ease-out]"
      )}
    >
      <ToastIcon variant={toast.variant} />
      <div className="flex-1 min-w-0">
        <p className="font-medium text-foreground">{toast.title}</p>
        {toast.description && <p className="text-xs text-muted-foreground mt-0.5">{toast.description}</p>}
      </div>
      <button onClick={() => setExiting(true)} className="shrink-0 text-muted-foreground hover:text-foreground transition-colors">
        <X className="h-3.5 w-3.5" />
      </button>
    </div>
  );
}

export function ToastContainer() {
  const { toasts } = useToast();
  if (toasts.length === 0) return null;
  return (
    <div className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 w-80 max-w-[calc(100vw-2rem)]">
      {toasts.map((t) => <ToastItem key={t.id} toast={t} />)}
    </div>
  );
}
