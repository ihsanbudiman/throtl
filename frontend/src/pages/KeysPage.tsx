import { useEffect, useState, useCallback } from "react";
import { api, type APIKey } from "@/lib/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { useToast } from "@/hooks/use-toast";
import GenerateKeyDialog from "@/components/GenerateKeyDialog";
import { Copy, Trash2, Check, KeyRound, Infinity as InfinityIcon, RotateCcw } from "lucide-react";

function formatDailyReset(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleTimeString(undefined, { hour: "2-digit", minute: "2-digit", hour12: false });
}

export default function KeysPage() {
  const { toast } = useToast();
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [copied, setCopied] = useState<string | null>(null);
  const [newKey, setNewKey] = useState<APIKey | null>(null);
  const [deleteKey, setDeleteKey] = useState<APIKey | null>(null);

  const loadKeys = useCallback(async () => {
    const k = await api.listKeys();
    setKeys(k || []);
    setLoading(false);
  }, []);

  useEffect(() => {
    loadKeys();
  }, [loadKeys]);

  const handleCreateSuccess = (key: APIKey) => {
    setNewKey(key);
    loadKeys();
  };

  const handleToggle = async (id: string, active: boolean) => { await api.toggleKey(id, !active); loadKeys(); };

  const handleReset = async (id: string, name: string) => {
    await api.resetKeyLimit(id);
    toast({ title: "Limits reset", description: name });
    loadKeys();
  };

  const copyKey = (key: string) => {
    navigator.clipboard.writeText(key);
    setCopied(key);
    toast({ title: "Copied to clipboard", variant: "success" });
    setTimeout(() => setCopied(null), 2000);
  };

  if (loading) {
    return (
      <div className="space-y-6 fade-in-up">
        <div className="flex items-center justify-between">
          <div className="space-y-2"><div className="h-8 w-32 shimmer rounded-lg" /><div className="h-4 w-56 shimmer rounded-lg" /></div>
          <div className="h-9 w-36 shimmer rounded-lg" />
        </div>
        <Card><CardContent className="p-6"><div className="space-y-3">{[1, 2, 3, 4, 5].map((i) => <div key={i} className="h-12 shimmer rounded-lg" />)}</div></CardContent></Card>
      </div>
    );
  }

  return (
    <div className="space-y-6 fade-in-up">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-[400] tracking-tight">API Keys</h2>
          <p className="text-muted-foreground text-sm mt-1">Generate and manage share keys for consumers</p>
        </div>
        <GenerateKeyDialog onSuccess={handleCreateSuccess} />
      </div>

      {newKey && (
        <Card className="relative overflow-hidden border-primary/10 bg-gradient-to-r from-primary/[0.05] to-transparent animate-[fade-in_0.3s_ease-out]">
          <div className="absolute top-0 inset-x-0 h-px bg-gradient-to-r from-transparent via-primary/20 to-transparent" />
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <Check className="h-4 w-4 text-primary" />
              Key Created — Copy Now
            </CardTitle>
            <CardDescription>This is the only time the full key will be shown.</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <code className="flex-1 rounded-lg bg-muted px-3 py-2 text-sm font-mono border border-border/50 truncate">{newKey.key}</code>
              <Button size="sm" variant="outline" onClick={() => copyKey(newKey.key)}>{copied === newKey.key ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}</Button>
              <Button size="sm" variant="ghost" onClick={() => setNewKey(null)}>Dismiss</Button>
            </div>
          </CardContent>
        </Card>
      )}

      <Card className="overflow-hidden">
        <CardHeader>
          <CardTitle className="text-base">All Keys</CardTitle>
          <CardDescription>{keys.length} keys generated</CardDescription>
        </CardHeader>
        <CardContent className="p-0 sm:p-4 sm:pt-0">
          {keys.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <div className="flex h-16 w-16 items-center justify-center rounded-[8px] bg-gradient-to-br from-primary/10 to-primary/5 mb-4">
                <KeyRound className="h-8 w-8 text-primary/40 float-icon" />
              </div>
              <p className="text-muted-foreground text-sm">No API keys yet. Generate one to start sharing access.</p>
            </div>
          ) : (
            <>
              {/* Desktop table */}
              <div className="hidden sm:block overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                       <TableHead>Name</TableHead>
                       <TableHead>Key</TableHead>
                       <TableHead>Token Limits</TableHead>
                       <TableHead>Daily Requests</TableHead>
                       <TableHead>Models</TableHead>
                       <TableHead>Status</TableHead>
                       <TableHead>Last Used</TableHead>
                       <TableHead>Reset</TableHead>
                       <TableHead className="w-10" />
                     </TableRow>
                  </TableHeader>
                  <TableBody>
                    {keys.map((key) => (
                      <TableRow key={key.id} className="transition-colors duration-150 hover:bg-accent/40">
                        <TableCell className="font-medium">{key.name}</TableCell>
                        <TableCell><code className="text-xs bg-muted px-2 py-1 rounded">{key.key}</code></TableCell>
                        <TableCell>
                          {(key.limit_tokens_in_daily > 0 || key.limit_tokens_out_daily > 0) ? (
                            <div className="space-y-1">
                              {key.limit_tokens_in_daily > 0 && (
                                <div>
                                  {key.rate_limit?.daily_reset ? (
                                    <><span className="font-mono text-xs">In: {key.rate_limit.daily_tokens_in_count}/{key.rate_limit.daily_tokens_in_limit}</span><div className="text-xs text-muted-foreground font-mono">resets {formatDailyReset(key.rate_limit.daily_reset)}</div></>
                                  ) : <span className="text-xs text-muted-foreground shimmer inline-block px-2 py-0.5 rounded">In: No requests yet</span>}
                                </div>
                              )}
                              {key.limit_tokens_out_daily > 0 && (
                                <div>
                                  {key.rate_limit?.daily_reset ? (
                                    <><span className="font-mono text-xs">Out: {key.rate_limit.daily_tokens_out_count}/{key.rate_limit.daily_tokens_out_limit}</span><div className="text-xs text-muted-foreground font-mono">resets {formatDailyReset(key.rate_limit.daily_reset)}</div></>
                                  ) : <span className="text-xs text-muted-foreground shimmer inline-block px-2 py-0.5 rounded">Out: No requests yet</span>}
                                </div>
                              )}
                            </div>
                          ) : <InfinityIcon className="h-3.5 w-3.5 text-muted-foreground" />}
                        </TableCell>
                        <TableCell>
                          {key.limit_daily > 0 ? (
                            <span className="font-mono text-xs">{key.rate_limit?.daily_count ?? 0}/{key.rate_limit?.daily_limit ?? key.limit_daily}</span>
                          ) : <InfinityIcon className="h-3.5 w-3.5 text-muted-foreground" />}
                        </TableCell>
                        <TableCell>{key.allowed_models ? <Badge variant="secondary" className="text-xs">{key.allowed_models.split(",").length} models</Badge> : <span className="text-muted-foreground">All</span>}</TableCell>
                        <TableCell><Switch checked={key.active} onCheckedChange={() => handleToggle(key.id, key.active)} /></TableCell>
                        <TableCell className="text-sm text-muted-foreground">{key.last_used_at ? new Date(key.last_used_at).toLocaleDateString() : "Never"}</TableCell>
                        <TableCell>
                          <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-primary" onClick={() => handleReset(key.id, key.name)}>
                            <RotateCcw className="h-4 w-4" />
                          </Button>
                        </TableCell>
                        <TableCell>
                          <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-destructive" onClick={() => setDeleteKey(key)}>
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
              {/* Mobile cards */}
              <div className="sm:hidden space-y-3 p-4">
                {keys.map((key) => (
                  <div key={key.id} className="rounded-[8px] border border-border/60 bg-card p-4 space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-sm">{key.name}</span>
                      <Switch checked={key.active} onCheckedChange={() => handleToggle(key.id, key.active)} size="sm" />
                    </div>
                    <code className="block text-xs bg-muted px-2 py-1.5 rounded font-mono truncate">{key.key}</code>
                    <div className="flex items-center justify-between text-xs text-muted-foreground">
                      <span>In: {key.limit_tokens_in_daily > 0 ? `${key.rate_limit?.daily_tokens_in_count ?? 0}/${key.rate_limit?.daily_tokens_in_limit ?? key.limit_tokens_in_daily}` : "∞"}</span>
                      <span>Out: {key.limit_tokens_out_daily > 0 ? `${key.rate_limit?.daily_tokens_out_count ?? 0}/${key.rate_limit?.daily_tokens_out_limit ?? key.limit_tokens_out_daily}` : "∞"}</span>
                    </div>
                    <div className="flex items-center justify-between text-xs text-muted-foreground">
                      <span>Requests: {key.limit_daily > 0 ? `${key.rate_limit?.daily_count ?? 0}/${key.rate_limit?.daily_limit ?? key.limit_daily}` : "∞"}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-muted-foreground">{key.last_used_at ? `Last used ${new Date(key.last_used_at).toLocaleDateString()}` : "Never used"}</span>
                      <div className="flex items-center gap-1">
                        <Button variant="ghost" size="icon-sm" className="text-muted-foreground hover:text-primary" onClick={() => handleReset(key.id, key.name)}>
                          <RotateCcw className="h-3.5 w-3.5" />
                        </Button>
                        <Button variant="ghost" size="icon-sm" className="text-muted-foreground hover:text-destructive" onClick={() => setDeleteKey(key)}>
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}
        </CardContent>
      </Card>

      <Dialog open={!!deleteKey} onOpenChange={(open) => !open && setDeleteKey(null)}>
        <DialogContent showCloseButton={false} className="max-w-xs">
          <DialogHeader>
            <DialogTitle>Delete key?</DialogTitle>
            <DialogDescription>This will permanently delete <code className="text-xs bg-muted px-1.5 py-0.5 rounded">{deleteKey?.name}</code> and all its usage logs. This cannot be undone.</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" size="sm" onClick={() => setDeleteKey(null)}>Cancel</Button>
            <Button variant="destructive" size="sm" onClick={() => { if (deleteKey) { api.deleteKey(deleteKey.id).then(() => { setDeleteKey(null); toast({ title: "Key deleted", description: deleteKey.name, variant: "destructive" }); loadKeys(); }); } }}>Delete</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
