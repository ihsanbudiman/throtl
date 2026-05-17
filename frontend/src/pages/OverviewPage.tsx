import { useEffect, useState } from "react";
import { api, type DashboardStats } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { KeyRound, Server, Activity, Zap, ArrowDownToLine, ArrowUpFromLine } from "lucide-react";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from "recharts";

const CHART_COLORS = [
  "var(--color-chart-1)",
  "var(--color-chart-2)",
  "var(--color-chart-4)",
  "var(--color-chart-5)",
  "var(--color-chart-3)",
];

const ACCENT_COLORS = [
  "bg-primary/5 border-l-primary",
  "bg-chart-2/[0.06] border-l-chart-2",
  "bg-chart-4/[0.06] border-l-chart-4",
  "bg-chart-1/[0.06] border-l-chart-1",
];

const ICON_GRADIENTS = [
  "from-primary/20 to-primary/5",
  "from-chart-2/20 to-chart-2/5",
  "from-chart-4/20 to-chart-4/5",
  "from-chart-1/20 to-chart-1/5",
];

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
      if (start >= value) {
        setDisplay(value);
        clearInterval(timer);
      } else {
        setDisplay(Math.floor(start));
      }
    }, duration / steps);
    return () => clearInterval(timer);
  }, [value]);

  return <>{formatNumber(display)}</>;
}

export function OverviewPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getStats().then(setStats).finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="space-y-6 fade-in-up">
        <div className="space-y-2">
          <div className="h-8 w-40 shimmer rounded-none" />
          <div className="h-4 w-64 shimmer rounded-none" />
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <Card key={i}>
              <CardContent className="p-6">
                <div className="h-16 shimmer rounded-none" />
              </CardContent>
            </Card>
          ))}
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          {[1, 2].map((i) => (
            <Card key={i}>
              <CardContent className="p-6">
                <div className="h-48 shimmer rounded-none" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="flex flex-col items-center justify-center h-64 gap-2">
        <Activity className="h-8 w-8 text-destructive/50" />
        <p className="text-sm text-muted-foreground">Failed to load stats</p>
      </div>
    );
  }

  const statCards = [
    {
      title: "Total API Keys",
      value: stats.total_keys,
      sub: `${stats.active_keys} active`,
      icon: KeyRound,
      color: "text-primary",
    },
    {
      title: "Providers",
      value: stats.total_providers,
      sub: "Connected",
      icon: Server,
      color: "text-chart-2",
    },
    {
      title: "Total Requests",
      value: stats.total_requests,
      sub: `${stats.requests_today} today`,
      icon: Activity,
      color: "text-chart-4",
    },
    {
      title: "Tokens Processed",
      value: stats.total_tokens_in + stats.total_tokens_out,
      sub: `${formatNumber(stats.total_tokens_in)} in / ${formatNumber(stats.total_tokens_out)} out`,
      icon: Zap,
      color: "text-chart-1",
    },
  ];

  const keyUsage = stats.key_usage ?? [];
  const modelBreakdown = stats.model_breakdown ?? [];

  return (
    <div className="space-y-6 animate-[fade-in-up_0.4s_ease-out]">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Overview</h2>
        <p className="text-muted-foreground text-sm mt-1">
          Monitor your Throtl gateway at a glance
        </p>
      </div>

      {/* Stat cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 fade-in-stagger">
        {statCards.map((s, idx) => {
          const Icon = s.icon;
          return (
            <Card
              key={s.title}
              className={`relative border-l-2 ${ACCENT_COLORS[idx]} transition-all duration-300 hover:-translate-y-1 hover:shadow-lg hover:shadow-primary/5`}
            >
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {s.title}
                </CardTitle>
                <div className={`flex h-8 w-8 items-center justify-center bg-gradient-to-br ${ICON_GRADIENTS[idx]} transition-transform duration-300 group-hover/card:scale-110`}>
                  <Icon className={`h-4 w-4 ${s.color}`} />
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold tracking-tight">
                  <AnimatedNumber value={s.value} />
                </div>
                <p className="text-xs text-muted-foreground mt-1">{s.sub}</p>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Charts row */}
      <div className="grid gap-4 md:grid-cols-2">
        {/* Key usage chart */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Key Usage</CardTitle>
            <CardDescription>Requests per API key</CardDescription>
          </CardHeader>
          <CardContent>
            {keyUsage.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-48 text-sm text-muted-foreground gap-2">
                <Activity className="h-8 w-8 text-muted-foreground/30" />
                <span>No usage data yet</span>
              </div>
            ) : (
              <ResponsiveContainer width="100%" height={200}>
                <BarChart data={keyUsage}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" strokeOpacity={0.5} />
                  <XAxis
                    dataKey="key_name"
                    tick={{ fontSize: 12, fill: "var(--color-muted-foreground)" }}
                    axisLine={{ stroke: "var(--color-border)", strokeOpacity: 0.5 }}
                  />
                  <YAxis
                    tick={{ fontSize: 12, fill: "var(--color-muted-foreground)" }}
                    axisLine={{ stroke: "var(--color-border)", strokeOpacity: 0.5 }}
                  />
                  <Tooltip
                    contentStyle={{
                      background: "var(--color-popover)",
                      border: "1px solid var(--color-border)",
                      borderRadius: "0",
                      fontSize: 12,
                      boxShadow: "0 4px 12px rgba(0,0,0,0.08)",
                    }}
                  />
                  <Bar dataKey="requests" fill="var(--color-primary)" radius={[0, 0, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        {/* Model breakdown */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Model Breakdown</CardTitle>
            <CardDescription>Requests by model</CardDescription>
          </CardHeader>
          <CardContent>
            {modelBreakdown.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-48 text-sm text-muted-foreground gap-2">
                <Activity className="h-8 w-8 text-muted-foreground/30" />
                <span>No model data yet</span>
              </div>
            ) : (
              <ResponsiveContainer width="100%" height={200}>
                <PieChart>
                  <Pie
                    data={modelBreakdown}
                    dataKey="requests"
                    nameKey="model"
                    cx="50%"
                    cy="50%"
                    outerRadius={70}
                    label={({ name, percent }: any) =>
                      `${name ?? ''} ${((percent ?? 0) * 100).toFixed(0)}%`
                    }
                    labelLine={false}
                  >
                    {modelBreakdown.map((_, i) => (
                      <Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{
                      background: "var(--color-popover)",
                      border: "1px solid var(--color-border)",
                      borderRadius: "0",
                      fontSize: 12,
                      boxShadow: "0 4px 12px rgba(0,0,0,0.08)",
                    }}
                  />
                </PieChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Token flow */}
      <Card className="relative overflow-hidden">
        <div className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-primary/40 to-transparent bg-[length:200%_100%] animate-[gradient-sweep_3s_ease-in-out_infinite]" />
        <CardHeader>
          <CardTitle className="text-base">Token Flow</CardTitle>
          <CardDescription>Input vs Output tokens</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <div className="flex items-center gap-4 rounded-none border border-border p-4 transition-all duration-200 hover:border-chart-2/50 hover:bg-chart-2/[0.03] hover:shadow-sm">
              <div className="flex h-10 w-10 items-center justify-center bg-gradient-to-br from-chart-2/20 to-chart-2/5">
                <ArrowDownToLine className="h-5 w-5 text-chart-2" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Input Tokens</p>
                <p className="text-xl font-bold tracking-tight">{formatNumber(stats.total_tokens_in)}</p>
              </div>
            </div>
            <div className="flex items-center gap-4 rounded-none border border-border p-4 transition-all duration-200 hover:border-chart-4/50 hover:bg-chart-4/[0.03] hover:shadow-sm">
              <div className="flex h-10 w-10 items-center justify-center bg-gradient-to-br from-chart-4/20 to-chart-4/5">
                <ArrowUpFromLine className="h-5 w-5 text-chart-4" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Output Tokens</p>
                <p className="text-xl font-bold tracking-tight">{formatNumber(stats.total_tokens_out)}</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
