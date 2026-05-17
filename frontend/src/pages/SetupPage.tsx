import { useState } from "react";
import { useAuth } from "@/lib/auth";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ShieldCheck } from "lucide-react";

export default function SetupPage() {
  const { setup } = useAuth();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [shakeKey, setShakeKey] = useState(0);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    if (password !== confirm) { setError("Passwords do not match"); setShakeKey((k) => k + 1); return; }
    if (password.length < 8) { setError("Password must be at least 8 characters"); setShakeKey((k) => k + 1); return; }
    setSubmitting(true);
    try { await setup(email, password, name); }
    catch (err) {
      setError(err instanceof Error ? err.message : "Setup failed");
      setShakeKey((k) => k + 1);
    } finally { setSubmitting(false); }
  };

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-background p-4">
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,_var(--color-primary)/0.06,_transparent_50%)] dark:bg-[radial-gradient(ellipse_at_top,_var(--color-primary)/0.03,_transparent_50%)]" />
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_bottom_right,_var(--color-chart-2)/0.04,_transparent_50%)] dark:bg-[radial-gradient(ellipse_at_bottom_right,_var(--color-chart-2)/0.02,_transparent_50%)]" />
      <Card className="relative w-full max-w-sm overflow-hidden shadow-2xl shadow-black/10 dark:shadow-black/40 animate-[scale-in_0.35s_ease-out]">
        <div className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-primary/60 to-transparent bg-[length:200%_100%] animate-[gradient-sweep_3s_ease-in-out_infinite]" />
        <CardHeader className="text-center pt-8 pb-4">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-[8px] bg-gradient-to-br from-primary/20 to-primary/5 shadow-sm">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm shadow-primary/20">
              <ShieldCheck className="h-5 w-5" />
            </div>
          </div>
          <CardTitle className="text-xl font-[400] tracking-tight">Create Admin Account</CardTitle>
          <CardDescription className="text-sm mt-1">First-time setup for your Throtl gateway</CardDescription>
        </CardHeader>
        <CardContent className="pb-8">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="name">Full Name</Label>
              <Input id="name" placeholder="Admin" value={name} onChange={(e) => setName(e.target.value)} required autoFocus />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="email">Email</Label>
              <Input id="email" type="email" placeholder="admin@example.com" value={email} onChange={(e) => setEmail(e.target.value)} required />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="password">Password</Label>
              <Input id="password" type="password" placeholder="Min. 8 characters" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="confirm">Confirm Password</Label>
              <Input id="confirm" type="password" placeholder="Repeat your password" value={confirm} onChange={(e) => setConfirm(e.target.value)} required />
            </div>
            {error && (
              <div key={shakeKey} className="flex items-center gap-2 rounded-lg border border-destructive/30 bg-destructive/5 px-3 py-2.5 text-sm text-destructive error-shake">
                <ShieldCheck className="h-3.5 w-3.5 shrink-0" />
                {error}
              </div>
            )}
            <Button type="submit" className="w-full h-9" disabled={submitting}>
              {submitting ? (
                <span className="flex items-center gap-2">
                  <span className="h-3.5 w-3.5 rounded-full border-2 border-current border-t-transparent animate-spin" />
                  Creating account...
                </span>
              ) : "Create Admin Account"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
