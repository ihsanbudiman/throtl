import { useEffect, useState } from "react";
import { api, type UsageLog } from "@/lib/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Activity, Clock, ArrowDownToLine, ArrowUpFromLine } from "lucide-react";

function formatTime(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
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

export default function UsagePage() {
  const [logs, setLogs] = useState<UsageLog[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => { api.getUsageLogs().then((l) => { setLogs(l || []); setLoading(false); }); }, []);

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

      <Card className="overflow-hidden">
        <CardHeader>
          <CardTitle className="text-base">Request History</CardTitle>
          <CardDescription>{logs.length} recent requests</CardDescription>
        </CardHeader>
        <CardContent className="p-0 sm:p-4 sm:pt-0">
          {logs.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <div className="flex h-16 w-16 items-center justify-center rounded-[8px] bg-gradient-to-br from-primary/10 to-primary/5 mb-4">
                <Activity className="h-8 w-8 text-primary/40 float-icon" />
              </div>
              <p className="text-muted-foreground text-sm">No requests logged yet. Share your API keys to start seeing traffic.</p>
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
                    {logs.map((log) => (
                      <TableRow key={log.id} className="transition-colors duration-150 hover:bg-accent/30">
                        <TableCell className="text-sm text-muted-foreground">
                          <div className="flex items-center gap-1.5">
                            <Clock className="h-3.5 w-3.5 text-muted-foreground/60" />
                            {new Date(log.created_at).toLocaleTimeString()}
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
                {logs.map((log) => (
                  <div key={log.id} className="rounded-[8px] border border-border/60 bg-card p-4 space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
                        <Clock className="h-3.5 w-3.5" />
                        {new Date(log.created_at).toLocaleTimeString()}
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
