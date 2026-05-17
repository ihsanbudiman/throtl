import { useEffect, useState } from "react";
import { api, type DashboardStats } from "@/lib/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { KeyRound, Server, Activity, Zap, ArrowDownToLine, ArrowUpFromLine, TrendingUp, TrendingDown } from "lucide-react";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from "recharts";

const CHART_COLORS = ["var(--color-chart-1)", "var(--color-chart-2)", "var(--color-chart-4)", "var(--color-chart-5)", "var(--color-chart-3)"];

function formatNumber(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + "M";
  if (n >= 1_000) return (n / 1_000).toFixed(1) + "K";
  return n.toString();
}

function AnimatedNumber({ value }: { value: number }) {
  const [display, setDisplay] = useState(0);
  useEffect(() => {
    let start = 0;
    const duration = 600;
    const steps = 20;
    const increment = value / steps;
    const timer = setInterval(() => {
      start += increment;
      if (start >= value) { setDisplay(value); clearInterval(timer); }
      else { setDisplay(Math.floor(start)); }
    }, duration / steps);
    return () => clearInterval(timer);
  }, [value]);
  return <>{formatNumber(display)}</>;
}

const statCards = [
  { title: "Total API Keys", key: "total_keys" as const, subKey: "active_keys" as const, sub: (s: DashboardStats) => `${s.active_keys} active`, icon: KeyRound, color: "text-chart-1", gradient: "from-chart-1/20 to-chart-1/5" },
  { title: "Providers", key: "total_providers" as const, subKey: null, sub: () => "Connected", icon: Server, color: "text-chart-2", gradient: "from-chart-2/20 to-chart-2/5" },
  { title: "Total Requests", key: "total_requests" as const, subKey: "requests_today" as const, sub: (s: DashboardStats) => `${s.requests_today} today`, icon: Activity, color: "text-chart-4", gradient: "from-chart-4/20 to-chart-4/5" },
  { title: "Tokens Processed", key: null, subKey: null, value: (s: DashboardStats) => s.total_tokens_in + s.total_tokens_out, sub: (s: DashboardStats) => `${formatNumber(s.total_tokens_in)} in / ${formatNumber(s.total_tokens_out)} out`, icon: Zap, color: "text-chart-1", gradient: "from-chart-1/20 to-chart-1/5" },
];

export default function OverviewPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => { api.getStats().then(setStats).finally(() => setLoading(false)); }, []);

  if (loading) {
    return (
      <div className="space-y-6 fade-in-up">
        <div className="space-y-2">
          <div className="h-8 w-40 shimmer rounded-lg" />
          <div className="h-4 w-64 shimmer rounded-lg" />
        </div>
        <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <Card key={i}><CardContent className="p-6"><div className="h-20 shimmer rounded-lg" /></CardContent></Card>
          ))}
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          {[1, 2].map((i) => (
            <Card key={i}><CardContent className="p-6"><div className="h-52 shimmer rounded-lg" /></CardContent></Card>
          ))}
        </div>
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
        <div className="flex h-16 w-16 items-center justify-center rounded-[8px] bg-destructive/10">
          <Activity className="h-8 w-8 text-destructive/60" />
        </div>
        <p className="text-sm text-muted-foreground">Failed to load stats</p>
      </div>
    );
  }

  const keyUsage = stats.key_usage ?? [];
  const modelBreakdown = stats.model_breakdown ?? [];

  return (
    <div className="space-y-6 animate-[fade-in-up_0.4s_ease-out]">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-[400] tracking-tight">Overview</h2>
          <p className="text-muted-foreground text-sm mt-1">Monitor your Throtl gateway at a glance</p>
        </div>
        <div className="hidden sm:flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 text-primary">
          <Activity className="h-4 w-4" />
        </div>
      </div>

      <div className="grid gap-4 grid-cols-2 lg:grid-cols-4 fade-in-stagger">
        {statCards.map((s) => {
          const Icon = s.icon;
          const val = s.value ? s.value(stats) : stats[s.key!];
          return (
            <Card key={s.title} className="relative overflow-hidden transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg hover:shadow-primary/5">
              <div className={`absolute top-0 right-0 w-24 h-24 bg-gradient-to-br ${s.gradient} rounded-bl-[3rem] opacity-60`} />
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">{s.title}</CardTitle>
                <div className={`flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br ${s.gradient} transition-transform duration-300`}>
                  <Icon className={`h-4 w-4 ${s.color}`} />
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-xl sm:text-2xl font-[400] tracking-tight">
                  <AnimatedNumber value={val as number} />
                </div>
                <p className="text-xs text-muted-foreground mt-1">{s.sub(stats)}</p>
              </CardContent>
            </Card>
          );
        })}
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card className="overflow-hidden">
          <div className="absolute top-0 inset-x-0 h-px bg-gradient-to-r from-transparent via-chart-1/30 to-transparent" />
          <CardHeader>
            <div className="flex items-center gap-2">
              <TrendingUp className="h-4 w-4 text-chart-1" />
              <CardTitle className="text-base">Key Usage</CardTitle>
            </div>
            <CardDescription>Requests per API key</CardDescription>
          </CardHeader>
          <CardContent>
            {keyUsage.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-48 text-sm text-muted-foreground gap-2">
                <Activity className="h-8 w-8 text-muted-foreground/20" />
                <span>No usage data yet</span>
              </div>
            ) : (
              <ResponsiveContainer width="100%" height={220}>
                <BarChart data={keyUsage}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" strokeOpacity={0.4} vertical={false} />
                  <XAxis dataKey="key_name" tick={{ fontSize: 12, fill: "var(--color-muted-foreground)" }} axisLine={{ stroke: "var(--color-border)", strokeOpacity: 0.4 }} tickLine={false} />
                  <YAxis tick={{ fontSize: 12, fill: "var(--color-muted-foreground)" }} axisLine={false} tickLine={false} />
                  <Tooltip contentStyle={{ background: "var(--color-popover)", border: "1px solid var(--color-border)", borderRadius: "var(--radius)", fontSize: 12, boxShadow: "0 8px 24px rgba(0,0,0,0.12)" }} />
                  <Bar dataKey="requests" fill="var(--color-chart-1)" radius={[4, 4, 0, 0]} maxBarSize={40} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        <Card className="overflow-hidden">
          <div className="absolute top-0 inset-x-0 h-px bg-gradient-to-r from-transparent via-chart-4/30 to-transparent" />
          <CardHeader>
            <div className="flex items-center gap-2">
              <TrendingDown className="h-4 w-4 text-chart-4" />
              <CardTitle className="text-base">Model Breakdown</CardTitle>
            </div>
            <CardDescription>Requests by model</CardDescription>
          </CardHeader>
          <CardContent>
            {modelBreakdown.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-48 text-sm text-muted-foreground gap-2">
                <Activity className="h-8 w-8 text-muted-foreground/20" />
                <span>No model data yet</span>
              </div>
            ) : (
              <ResponsiveContainer width="100%" height={220}>
                <PieChart>
                  <Pie data={modelBreakdown} dataKey="requests" nameKey="model" cx="50%" cy="50%" innerRadius={50} outerRadius={80} paddingAngle={2}>
                    {modelBreakdown.map((_, i) => <Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />)}
                  </Pie>
                  <Tooltip contentStyle={{ background: "var(--color-popover)", border: "1px solid var(--color-border)", borderRadius: "var(--radius)", fontSize: 12, boxShadow: "0 8px 24px rgba(0,0,0,0.12)" }} />
                </PieChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>

      <Card className="relative overflow-hidden">
        <div className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-primary/40 to-transparent bg-[length:200%_100%] animate-[gradient-sweep_3s_ease-in-out_infinite]" />
        <CardHeader>
          <div className="flex items-center gap-2">
            <Zap className="h-4 w-4 text-chart-1" />
            <CardTitle className="text-base">Token Flow</CardTitle>
          </div>
          <CardDescription>Input vs Output tokens</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="relative flex items-center gap-4 rounded-xl border border-border/60 bg-gradient-to-br from-chart-2/[0.03] to-transparent p-4 transition-all duration-200 hover:border-chart-2/40 hover:shadow-sm hover:shadow-chart-2/5">
              <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-chart-2/20 to-chart-2/5">
                <ArrowDownToLine className="h-5 w-5 text-chart-2" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Input Tokens</p>
                <p className="text-xl font-[400] tracking-tight">{formatNumber(stats.total_tokens_in)}</p>
              </div>
            </div>
            <div className="relative flex items-center gap-4 rounded-xl border border-border/60 bg-gradient-to-br from-chart-4/[0.03] to-transparent p-4 transition-all duration-200 hover:border-chart-4/40 hover:shadow-sm hover:shadow-chart-4/5">
              <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-chart-4/20 to-chart-4/5">
                <ArrowUpFromLine className="h-5 w-5 text-chart-4" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Output Tokens</p>
                <p className="text-xl font-[400] tracking-tight">{formatNumber(stats.total_tokens_out)}</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
