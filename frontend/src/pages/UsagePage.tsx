import { useEffect, useMemo, useRef, useState } from "react";
import { api, type APIKey, type UsageLog } from "@/lib/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Activity, Clock, ArrowDownToLine, ArrowUpFromLine, BarChart3, Filter, ChevronDown } from "lucide-react";
import { Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis } from "recharts";

function formatTime(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

function formatLocalDateTime(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleString();
}

function statusVariant(status: number): "default" | "secondary" | "destructive" {
  if (status >= 200 && status < 300) return "default";
  if (status >= 400 && status < 500) return "secondary";
  return "destructive";
}

function latencyColor(ms: number): string {
  if (ms < 500) return "text-chart-2";
  if (ms < 2000) return "text-chart-4";
  return "text-destructive";
}

type DateRange = "today" | "7d" | "30d";

function getDateCutoff(range: DateRange): Date {
  const now = new Date();
  const cutoff = new Date(now);
  if (range === "today") {
    cutoff.setHours(0, 0, 0, 0);
  } else if (range === "7d") {
    cutoff.setDate(now.getDate() - 7);
  } else {
    cutoff.setDate(now.getDate() - 30);
  }
  return cutoff;
}

function aggregateByDay(logs: UsageLog[]): Array<{ date: string; requests: number; tokensIn: number; tokensOut: number }> {
  const map = new Map<string, { requests: number; tokensIn: number; tokensOut: number }>();
  for (const log of logs) {
    const d = new Date(log.created_at);
    const day = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
    const entry = map.get(day) || { requests: 0, tokensIn: 0, tokensOut: 0 };
    entry.requests++;
    entry.tokensIn += log.tokens_in;
    entry.tokensOut += log.tokens_out;
    map.set(day, entry);
  }
  return Array.from(map.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([date, data]) => ({ date, ...data }));
}

export default function UsagePage() {
  const [logs, setLogs] = useState<UsageLog[]>([]);
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [dateRange, setDateRange] = useState<DateRange>("7d");
  const [selectedKey, setSelectedKey] = useState<string>("all");
  const [keyMenuOpen, setKeyMenuOpen] = useState(false);
  const keyMenuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    Promise.all([api.getUsageLogs(), api.listKeys()]).then(([l, k]) => {
      setLogs(l || []);
      setKeys(k || []);
      setLoading(false);
    });
  }, []);

  useEffect(() => {
    if (!keyMenuOpen) return;
    const handler = (e: MouseEvent) => {
      if (keyMenuRef.current && !keyMenuRef.current.contains(e.target as Node)) {
        setKeyMenuOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [keyMenuOpen]);

  const filteredLogs = useMemo(() => {
    const cutoff = getDateCutoff(dateRange);
    return logs.filter((log) => {
      if (new Date(log.created_at) < cutoff) return false;
      if (selectedKey !== "all" && log.api_key_id !== selectedKey) return false;
      return true;
    });
  }, [logs, dateRange, selectedKey]);

  const chartData = useMemo(() => aggregateByDay(filteredLogs), [filteredLogs]);

  const totalTokensIn = useMemo(() => filteredLogs.reduce((s, l) => s + l.tokens_in, 0), [filteredLogs]);
  const totalTokensOut = useMemo(() => filteredLogs.reduce((s, l) => s + l.tokens_out, 0), [filteredLogs]);

  if (loading) {
    return (
      <div className="space-y-6 fade-in-up">
        <div className="space-y-2"><div className="h-8 w-40 shimmer rounded-lg" /><div className="h-4 w-64 shimmer rounded-lg" /></div>
        <Card><CardContent className="p-6"><div className="space-y-3">{[1, 2, 3, 4, 5].map((i) => <div key={i} className="h-12 shimmer rounded-lg" />)}</div></CardContent></Card>
      </div>
    );
  }

  return (
    <div className="space-y-6 fade-in-up">
      <div>
        <h2 className="text-2xl font-[400] tracking-tight">Usage Logs</h2>
        <p className="text-muted-foreground text-sm mt-1">Recent proxied requests through your gateway</p>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-3">
        <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
          <Filter className="h-4 w-4" />
          <span>Filter:</span>
        </div>
        <div className="flex rounded-[8px] border border-border/60 bg-card overflow-hidden">
          {([["today", "Today"], ["7d", "7 Days"], ["30d", "30 Days"]] as [DateRange, string][]).map(([value, label], i) => (
            <button
              key={value}
              onClick={() => setDateRange(value)}
              className={`px-3 py-1.5 text-xs font-medium transition-colors ${
                dateRange === value
                  ? "bg-primary/15 text-primary"
                  : "text-muted-foreground hover:text-foreground hover:bg-accent/30"
              } ${i !== 0 ? "border-l border-border/60" : ""}`}
            >
              {label}
            </button>
          ))}
        </div>
        {keys.length > 0 && (
          <div className="relative" ref={keyMenuRef}>
            <button
              onClick={() => setKeyMenuOpen(!keyMenuOpen)}
              className="rounded-[8px] border border-border/60 bg-card px-3 py-1.5 pr-7 text-xs text-foreground cursor-pointer hover:bg-accent/30 transition-colors flex items-center gap-1.5"
            >
              {selectedKey === "all" ? "All Keys" : keys.find((k) => k.id === selectedKey)?.name}
              <ChevronDown className={`h-3.5 w-3.5 text-muted-foreground transition-transform ${keyMenuOpen ? "rotate-180" : ""}`} />
            </button>
            {keyMenuOpen && (
              <div className="absolute top-full left-0 mt-1 min-w-[140px] rounded-[8px] border border-border/60 bg-card shadow-lg z-50 overflow-hidden">
                <button
                  onClick={() => { setSelectedKey("all"); setKeyMenuOpen(false); }}
                  className={`w-full text-left px-3 py-2 text-xs transition-colors ${
                    selectedKey === "all"
                      ? "bg-primary/15 text-primary font-medium"
                      : "text-foreground hover:bg-accent/30"
                  }`}
                >
                  All Keys
                </button>
                {keys.map((k) => (
                  <button
                    key={k.id}
                    onClick={() => { setSelectedKey(k.id); setKeyMenuOpen(false); }}
                    className={`w-full text-left px-3 py-2 text-xs transition-colors ${
                      selectedKey === k.id
                        ? "bg-primary/15 text-primary font-medium"
                        : "text-foreground hover:bg-accent/30"
                    }`}
                  >
                    {k.name}
                  </button>
                ))}
              </div>
            )}
          </div>
        )}
        <span className="text-xs text-muted-foreground ml-auto">
          {filteredLogs.length} request{filteredLogs.length !== 1 ? "s" : ""}
        </span>
      </div>

      {/* Chart */}
      {chartData.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <BarChart3 className="h-4 w-4 text-primary/70" />
              Usage Overview
            </CardTitle>
            <CardDescription>
              {totalTokensIn.toLocaleString()} tokens in, {totalTokensOut.toLocaleString()} tokens out across {filteredLogs.length} requests
            </CardDescription>
          </CardHeader>
          <CardContent className="pt-0">
            <div className="w-full">
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart
                    data={chartData}
                    margin={{ top: 16, right: 16, bottom: 8, left: 16 }}
                  >
                    <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" vertical={false} />
                    <XAxis
                      dataKey="date"
                      stroke="var(--color-muted-foreground)"
                      fontSize={12}
                      tickFormatter={(v) => v.slice(5)}
                      axisLine={false}
                      tickLine={false}
                      dy={8}
                    />
                    <Tooltip
                      cursor={{ fill: "var(--color-muted)" }}
                      content={({ active, payload }) => {
                        if (!active || !payload || payload.length === 0) return null;
                        const tokensOut = payload.find((p) => p.name === "Tokens Out");
                        const tokensIn = payload.find((p) => p.name === "Tokens In");
                        return (
                          <div style={{
                            backgroundColor: "var(--color-card)",
                            border: "1px solid var(--color-border)",
                            borderRadius: "8px",
                            padding: "8px 12px",
                            fontSize: "12px",
                            boxShadow: "0 4px 12px rgba(0,0,0,0.15)",
                          }}>
                            {tokensOut && (
                              <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: 16 }}>
                                <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
                                  <div style={{ width: 8, height: 8, borderRadius: "50%", backgroundColor: "#93c5fd" }} />
                                  <span style={{ color: "var(--color-muted-foreground)" }}>Tokens Out</span>
                                </div>
                                <span style={{ fontWeight: 600, color: "var(--color-foreground)" }}>
                                  {(tokensOut.value as number).toLocaleString()}
                                </span>
                              </div>
                            )}
                            {tokensIn && (
                              <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: 16, marginTop: 4 }}>
                                <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
                                  <div style={{ width: 8, height: 8, borderRadius: "50%", backgroundColor: "#3b82f6" }} />
                                  <span style={{ color: "var(--color-muted-foreground)" }}>Tokens In</span>
                                </div>
                                <span style={{ fontWeight: 600, color: "var(--color-foreground)" }}>
                                  {(tokensIn.value as number).toLocaleString()}
                                </span>
                              </div>
                            )}
                          </div>
                        );
                      }}
                    />
                    <Bar
                      dataKey="tokensOut"
                      name="Tokens Out"
                      stackId="tokens"
                      fill="#93c5fd"
                      radius={[0, 0, 0, 0]}
                      maxBarSize={48}
                      isAnimationActive
                    />
                    <Bar
                      dataKey="tokensIn"
                      name="Tokens In"
                      stackId="tokens"
                      fill="#3b82f6"
                      radius={[6, 6, 0, 0]}
                      maxBarSize={48}
                      isAnimationActive
                    />
                  </BarChart>
                </ResponsiveContainer>
              </div>
              <div className="flex items-center justify-center gap-6 pt-2 pb-1">
                <div className="flex items-center gap-2">
                  <div className="h-2 w-2 rounded-full" style={{ backgroundColor: "#3b82f6" }} />
                  <span className="text-xs text-muted-foreground">Tokens In</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="h-2 w-2 rounded-full" style={{ backgroundColor: "#93c5fd" }} />
                  <span className="text-xs text-muted-foreground">Tokens Out</span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Table */}
      <Card className="overflow-hidden">
        <CardHeader>
          <CardTitle className="text-base">Request History</CardTitle>
          <CardDescription>{filteredLogs.length} request{filteredLogs.length !== 1 ? "s" : ""} in selected period</CardDescription>
        </CardHeader>
        <CardContent className="p-0 sm:p-4 sm:pt-0">
          {filteredLogs.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <div className="flex h-16 w-16 items-center justify-center rounded-[8px] bg-gradient-to-br from-primary/10 to-primary/5 mb-4">
                <Activity className="h-8 w-8 text-primary/40 float-icon" />
              </div>
              <p className="text-muted-foreground text-sm">No requests found for the selected filters.</p>
            </div>
          ) : (
            <>
              {/* Desktop table */}
              <div className="hidden sm:block overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Time</TableHead>
                      <TableHead>Provider</TableHead>
                      <TableHead>Model</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Tokens In</TableHead>
                      <TableHead>Tokens Out</TableHead>
                      <TableHead>Latency</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredLogs.map((log) => (
                      <TableRow key={log.id} className="transition-colors duration-150 hover:bg-accent/30">
                        <TableCell className="text-sm text-muted-foreground">
                          <div className="flex items-center gap-1.5">
                            <Clock className="h-3.5 w-3.5 text-muted-foreground/60" />
                            {formatLocalDateTime(log.created_at)}
                          </div>
                        </TableCell>
                        <TableCell><Badge variant="secondary" className="text-xs border-border/50">{log.provider}</Badge></TableCell>
                        <TableCell className="font-mono text-xs text-foreground/80">{log.model || "\u2014"}</TableCell>
                        <TableCell>
                          <Badge variant={statusVariant(log.status)} className={`text-xs ${statusVariant(log.status) === "default" ? "bg-chart-2/15 text-chart-2 border-chart-2/20" : statusVariant(log.status) === "secondary" ? "bg-chart-4/15 text-chart-4 border-chart-4/20" : ""}`}>
                            {log.status}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {log.tokens_in > 0 ? (
                            <div className="flex items-center gap-1 text-sm">
                              <ArrowDownToLine className="h-3 w-3 text-chart-2/70" />
                              <span>{log.tokens_in.toLocaleString()}</span>
                            </div>
                          ) : <span className="text-muted-foreground">\u2014</span>}
                        </TableCell>
                        <TableCell>
                          {log.tokens_out > 0 ? (
                            <div className="flex items-center gap-1 text-sm">
                              <ArrowUpFromLine className="h-3 w-3 text-chart-4/70" />
                              <span>{log.tokens_out.toLocaleString()}</span>
                            </div>
                          ) : <span className="text-muted-foreground">\u2014</span>}
                        </TableCell>
                        <TableCell className={`text-sm font-mono ${latencyColor(log.latency_ms)}`}>{formatTime(log.latency_ms)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
              {/* Mobile cards */}
              <div className="sm:hidden space-y-3 p-4">
                {filteredLogs.map((log) => (
                  <div key={log.id} className="rounded-[8px] border border-border/60 bg-card p-4 space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
                        <Clock className="h-3.5 w-3.5" />
                        {formatLocalDateTime(log.created_at)}
                      </div>
                      <Badge variant={statusVariant(log.status)} className={`text-xs ${statusVariant(log.status) === "default" ? "bg-chart-2/15 text-chart-2" : statusVariant(log.status) === "secondary" ? "bg-chart-4/15 text-chart-4" : ""}`}>{log.status}</Badge>
                    </div>
                    <div className="flex items-center gap-2 text-sm">
                      <Badge variant="secondary" className="text-xs border-border/50">{log.provider}</Badge>
                      <span className="font-mono text-xs text-foreground/60">{log.model || "\u2014"}</span>
                    </div>
                    <div className="flex items-center justify-between text-xs text-muted-foreground">
                      <span>In: {log.tokens_in > 0 ? log.tokens_in.toLocaleString() : "\u2014"}</span>
                      <span>Out: {log.tokens_out > 0 ? log.tokens_out.toLocaleString() : "\u2014"}</span>
                      <span className={latencyColor(log.latency_ms)}>{formatTime(log.latency_ms)}</span>
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
