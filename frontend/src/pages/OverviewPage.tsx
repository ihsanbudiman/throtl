import { useEffect, useState } from "react";
import { api, type DashboardStats } from "@/lib/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { KeyRound, Server, Activity, Zap, ArrowDownToLine, ArrowUpFromLine, TrendingUp, TrendingDown } from "lucide-react";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, PieChart, Pie, Cell } from "recharts";

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
  { title: "Total API Keys", key: "total_keys" as const, subKey: "active_keys" as const, sub: (s: DashboardStats) => `${s.active_keys} active`, icon: KeyRound, color: "text-chart-1", accent: "chart-1" },
  { title: "Providers", key: "total_providers" as const, subKey: null, sub: () => "Connected", icon: Server, color: "text-chart-2", accent: "chart-2" },
  { title: "Total Requests", key: "total_requests" as const, subKey: "requests_today" as const, sub: (s: DashboardStats) => `${s.requests_today} today`, icon: Activity, color: "text-chart-4", accent: "chart-4" },
  { title: "Tokens Processed", key: null, subKey: null, value: (s: DashboardStats) => s.total_tokens_in + s.total_tokens_out, sub: (s: DashboardStats) => `${formatNumber(s.total_tokens_in)} in / ${formatNumber(s.total_tokens_out)} out`, icon: Zap, color: "text-chart-1", accent: "chart-1" },
];

export default function OverviewPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getStats()
      .then(setStats)
      .catch(() => setStats(null))
      .finally(() => setLoading(false));
  }, []);

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
            <Card key={s.title} className={`transition-all duration-200 hover:shadow-md border-t-2 border-t-chart-${s.accent} bg-gradient-to-b from-chart-${s.accent}/[0.03] to-transparent`}>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">{s.title}</CardTitle>
                <div className={`flex h-8 w-8 items-center justify-center rounded-lg bg-chart-${s.accent}/10`}>
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
        <Card>
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
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" opacity={0.4} vertical={false} />
                  <XAxis dataKey="key_name" stroke="var(--color-muted-foreground)" fontSize={12} axisLine={false} tickLine={false} />
                  <YAxis stroke="var(--color-muted-foreground)" fontSize={12} axisLine={false} tickLine={false} />
                  <Legend wrapperStyle={{ fontSize: "12px", paddingTop: "8px" }} iconType="circle" />
                  <Tooltip
                    cursor={{ fill: "var(--color-muted)" }}
                    contentStyle={{
                      backgroundColor: "var(--color-card)",
                      border: "1px solid var(--color-border)",
                      borderRadius: "8px",
                      fontSize: "12px",
                      boxShadow: "0 4px 12px rgba(0,0,0,0.15)",
                    }}
                    formatter={(value) => [typeof value === "number" ? value.toLocaleString() : "0"]}
                  />
                  <Bar dataKey="requests" name="Requests" fill="var(--color-chart-1)" radius={[6, 6, 0, 0]} maxBarSize={48} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        <Card>
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
                  <Pie data={modelBreakdown} dataKey="requests" nameKey="model" cx="50%" cy="45%" innerRadius={45} outerRadius={70} paddingAngle={3}>
                    {modelBreakdown.map((_, i) => <Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />)}
                  </Pie>
                  <Legend
                    wrapperStyle={{ fontSize: "12px", paddingTop: "8px" }}
                    iconType="circle"
                    formatter={(value) => value}
                  />
                  <Tooltip
                    cursor={{ fill: "transparent" }}
                    contentStyle={{
                      backgroundColor: "var(--color-card)",
                      border: "1px solid var(--color-border)",
                      borderRadius: "8px",
                      fontSize: "12px",
                      boxShadow: "0 4px 12px rgba(0,0,0,0.15)",
                    }}
                    formatter={(value) => [typeof value === "number" ? value.toLocaleString() : "0"]}
                  />
                </PieChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Zap className="h-4 w-4 text-chart-1" />
            <CardTitle className="text-base">Token Flow</CardTitle>
          </div>
          <CardDescription>Input vs Output tokens</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="flex items-center gap-4 rounded-xl border border-chart-2/20 bg-chart-2/[0.03] p-4">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-chart-2/10">
                  <ArrowDownToLine className="h-5 w-5 text-chart-2" />
                </div>
                <div className="flex-1">
                  <p className="text-sm text-muted-foreground">Input Tokens</p>
                  <p className="text-xl font-[400] tracking-tight">{formatNumber(stats.total_tokens_in)}</p>
                </div>
              </div>
              <div className="flex items-center gap-4 rounded-xl border border-chart-4/20 bg-chart-4/[0.03] p-4">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-chart-4/10">
                  <ArrowUpFromLine className="h-5 w-5 text-chart-4" />
                </div>
                <div className="flex-1">
                  <p className="text-sm text-muted-foreground">Output Tokens</p>
                  <p className="text-xl font-[400] tracking-tight">{formatNumber(stats.total_tokens_out)}</p>
                </div>
              </div>
            </div>
            <div className="space-y-2">
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>Ratio</span>
                <span>{((stats.total_tokens_in / Math.max(stats.total_tokens_in + stats.total_tokens_out, 1)) * 100).toFixed(0)}% in / {((stats.total_tokens_out / Math.max(stats.total_tokens_in + stats.total_tokens_out, 1)) * 100).toFixed(0)}% out</span>
              </div>
              <div className="flex h-2 rounded-full overflow-hidden bg-muted">
                <div
                  className="bg-chart-2 rounded-l-full transition-all duration-500"
                  style={{ width: `${(stats.total_tokens_in / Math.max(stats.total_tokens_in + stats.total_tokens_out, 1)) * 100}%` }}
                />
                <div
                  className="bg-chart-4 rounded-r-full transition-all duration-500"
                  style={{ width: `${(stats.total_tokens_out / Math.max(stats.total_tokens_in + stats.total_tokens_out, 1)) * 100}%` }}
                />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
