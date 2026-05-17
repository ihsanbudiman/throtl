import { useState } from "react";
import { useAuth } from "@/lib/auth";
import ThrotlLogo from "@/assets/throtl-logo.svg";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { LogIn } from "lucide-react";

export default function LoginPage() {
  const { login } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [shakeKey, setShakeKey] = useState(0);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSubmitting(true);
    try { await login(email, password); }
    catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
      setShakeKey((k) => k + 1);
    } finally { setSubmitting(false); }
  };

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-background p-4">
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,_var(--color-primary)/0.06,_transparent_50%)] dark:bg-[radial-gradient(ellipse_at_top,_var(--color-primary)/0.03,_transparent_50%)]" />
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_bottom_left,_var(--color-chart-2)/0.04,_transparent_50%)] dark:bg-[radial-gradient(ellipse_at_bottom_left,_var(--color-chart-2)/0.02,_transparent_50%)]" />
      <Card className="relative w-full max-w-sm overflow-hidden shadow-2xl shadow-black/10 dark:shadow-black/40 animate-[scale-in_0.35s_ease-out]">
        <div className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-primary/60 to-transparent bg-[length:200%_100%] animate-[gradient-sweep_3s_ease-in-out_infinite]" />
        <CardHeader className="text-center pb-4 pt-8">
          <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-[8px] bg-gradient-to-br from-primary/20 to-primary/5 shadow-sm">
            <img src={ThrotlLogo} alt="Throtl" className="h-8 w-8 object-contain" />
          </div>
          <h1 className="text-lg font-[400] tracking-tight">Welcome back</h1>
          <p className="text-sm text-muted-foreground mt-1">Sign in to your Throtl gateway</p>
        </CardHeader>
        <CardContent className="pb-8">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="email">Email</Label>
              <Input id="email" type="email" placeholder="admin@example.com" value={email} onChange={(e) => setEmail(e.target.value)} required autoFocus />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="password">Password</Label>
              <Input id="password" type="password" placeholder="Enter your password" value={password} onChange={(e) => setPassword(e.target.value)} required />
            </div>
            {error && (
              <div key={shakeKey} className="flex items-center gap-2 rounded-lg border border-destructive/30 bg-destructive/5 px-3 py-2.5 text-sm text-destructive error-shake">
                <LogIn className="h-3.5 w-3.5 shrink-0" />
                {error}
              </div>
            )}
            <Button type="submit" className="w-full h-9" disabled={submitting}>
              {submitting ? (
                <span className="flex items-center gap-2">
                  <span className="h-3.5 w-3.5 rounded-full border-2 border-current border-t-transparent animate-spin" />
                  Signing in...
                </span>
              ) : "Sign in"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
