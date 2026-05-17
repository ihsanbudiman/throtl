import { useEffect, useState } from "react";
import { api, type APIKey } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Plus, Copy, Trash2, Check, KeyRound } from "lucide-react";

function formatResetTime(iso: string): string {
  const d = new Date(iso);
  const now = new Date();
  const diffMs = d.getTime() - now.getTime();
  if (diffMs <= 0) return "now";
  const diffMin = Math.floor(diffMs / 60000);
  const diffHr = Math.floor(diffMin / 60);
  const remMin = diffMin % 60;
  if (diffHr > 0) return `${diffHr}h ${remMin}m`;
  return `${diffMin}m`;
}

function formatDailyReset(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleTimeString(undefined, { hour: "2-digit", minute: "2-digit", hour12: false });
}

export function KeysPage() {
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [copied, setCopied] = useState<string | null>(null);
  const [newKey, setNewKey] = useState<APIKey | null>(null);
  const [deleteKey, setDeleteKey] = useState<APIKey | null>(null);

  // Form state
  const [formName, setFormName] = useState("");
const [formLimitWindow, setFormLimitWindow] = useState("");
const [formLimitDaily, setFormLimitDaily] = useState("");
  const [formWindowHrs, setFormWindowHrs] = useState("5");
  const [formModels, setFormModels] = useState("");

  const loadData = async () => {
    const k = await api.listKeys();
    setKeys(k || []);
    setLoading(false);
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleCreate = async () => {
    const key = await api.createKey({
      name: formName,
    limit_window: formLimitWindow ? parseInt(formLimitWindow) : 0,
    limit_daily: formLimitDaily ? parseInt(formLimitDaily) : 0,
      limit_window_hrs: formWindowHrs ? parseInt(formWindowHrs) : 5,
      allowed_models: formModels,
    });
    setNewKey(key);
    setDialogOpen(false);
    setFormName("");
    setFormLimitWindow("");
    setFormLimitDaily("");
    setFormWindowHrs("5");
    setFormModels("");
    loadData();
  };

  const handleToggle = async (id: string, active: boolean) => {
    await api.toggleKey(id, !active);
    loadData();
  };

  const copyKey = (key: string) => {
    navigator.clipboard.writeText(key);
    setCopied(key);
    setTimeout(() => setCopied(null), 2000);
  };

  if (loading) {
    return (
      <div className="space-y-6 fade-in-up">
        <div className="flex items-center justify-between">
          <div className="space-y-2">
            <div className="h-8 w-32 shimmer rounded-none" />
            <div className="h-4 w-56 shimmer rounded-none" />
          </div>
          <div className="h-9 w-36 shimmer rounded-none" />
        </div>
        <Card>
          <CardContent className="p-6">
            <div className="space-y-3">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="h-12 shimmer rounded-none" />
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6 fade-in-up">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">API Keys</h2>
          <p className="text-muted-foreground text-sm mt-1">
            Generate and manage share keys for consumers
          </p>
        </div>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger>
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Generate Key
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-lg">
            <DialogHeader>
              <DialogTitle>Generate New API Key</DialogTitle>
              <DialogDescription>
                Create a share key that consumers will use to access the AI provider through your gateway.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="name">Key Name</Label>
                <Input
                  id="name"
                  placeholder="e.g. Team Alpha"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                />
              </div>
              
              <div className="grid grid-cols-3 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="limitWindow">Window Limit</Label>
                  <Input
                    id="limitWindow"
                    type="number"
                    placeholder="0 = unlimited"
                    value={formLimitWindow}
                    onChange={(e) => setFormLimitWindow(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="windowHrs">Window Duration (hrs)</Label>
                  <Input
                    id="windowHrs"
                    type="number"
                    min="1"
                    placeholder="5"
                    value={formWindowHrs}
                    onChange={(e) => setFormWindowHrs(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="limitDaily">Daily Limit</Label>
                  <Input
                    id="limitDaily"
                    type="number"
                    placeholder="0 = unlimited"
                    value={formLimitDaily}
                    onChange={(e) => setFormLimitDaily(e.target.value)}
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="models">Allowed Models</Label>
                <Input
                  id="models"
                  placeholder="e.g. gpt-4o,gpt-4o-mini (empty = all)"
                  value={formModels}
                  onChange={(e) => setFormModels(e.target.value)}
                />
                <p className="text-xs text-muted-foreground">
                  Comma-separated model names. Leave empty to allow all models.
                </p>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDialogOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleCreate} disabled={!formName}>
                Generate
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* New key reveal */}
      {newKey && (
        <Card className="border-primary/40 bg-gradient-to-r from-primary/[0.04] to-transparent glow-pulse">
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <Check className="h-4 w-4 text-primary" />
              Key Created — Copy Now
            </CardTitle>
            <CardDescription>
              This is the only time the full key will be shown.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <code className="flex-1 rounded-none bg-muted px-3 py-2 text-sm font-mono border border-border/50">
                {newKey.key}
              </code>
              <Button
                size="sm"
                variant="outline"
                onClick={() => copyKey(newKey.key)}
              >
                {copied === newKey.key ? (
                  <Check className="h-4 w-4" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setNewKey(null)}>
                Dismiss
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Keys table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">All Keys</CardTitle>
          <CardDescription>{keys.length} keys generated</CardDescription>
        </CardHeader>
        <CardContent>
          {keys.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-14 text-center">
              <div className="flex h-14 w-14 items-center justify-center bg-gradient-to-br from-primary/10 to-primary/5 mb-4">
                <KeyRound className="h-7 w-7 text-primary/50 float-icon" />
              </div>
              <p className="text-muted-foreground text-sm">
                No API keys yet. Generate one to start sharing access.
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Key</TableHead>
                  <TableHead>Rate Limit</TableHead>
                  <TableHead>Daily Limit</TableHead>
                  <TableHead>Models</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Last Used</TableHead>
                  <TableHead className="w-10"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {keys.map((key) => (
                  <TableRow key={key.id} className="transition-colors duration-150 hover:bg-accent/50">
                    <TableCell className="font-medium">{key.name}</TableCell>
                    <TableCell>
                      <code className="text-xs bg-muted px-2 py-1 rounded-none">
                        {key.key}
                      </code>
                    </TableCell>
                    <TableCell>
                      {key.limit_window > 0 ? (
                        <div>
                          {key.rate_limit?.reset_at ? (
                            <>
                              <span className="font-mono">{key.rate_limit.count}/{key.rate_limit.limit}</span>
                              <div className="text-xs text-muted-foreground font-mono">
                                resets {formatResetTime(key.rate_limit.reset_at)}
                              </div>
                            </>
                          ) : (
                            <span className="text-xs text-muted-foreground shimmer inline-block px-2 py-0.5">No requests yet</span>
                          )}
                        </div>
                      ) : (
                        <span className="text-muted-foreground">∞</span>
                      )}
                    </TableCell>
                    <TableCell>
                      {key.limit_daily > 0 ? (
                        <div>
                          {key.rate_limit?.daily_reset ? (
                            <>
                              <span className="font-mono">{key.rate_limit.daily_count}/{key.rate_limit.daily_limit}</span>
                              <div className="text-xs text-muted-foreground font-mono">
                                resets at {formatDailyReset(key.rate_limit.daily_reset)}
                              </div>
                            </>
                          ) : (
                            <span className="text-xs text-muted-foreground shimmer inline-block px-2 py-0.5">No requests yet</span>
                          )}
                        </div>
                      ) : (
                        <span className="text-muted-foreground">∞</span>
                      )}
                    </TableCell>
                    <TableCell>
                      {key.allowed_models ? (
                        <Badge variant="secondary" className="text-xs">
                          {key.allowed_models.split(",").length} models
                        </Badge>
                      ) : (
                        <span className="text-muted-foreground">All</span>
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Switch
                          checked={key.active}
                          onCheckedChange={() =>
                            handleToggle(key.id, key.active)
                          }
                        />
                      </div>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {key.last_used_at
                        ? new Date(key.last_used_at).toLocaleDateString()
                        : "Never"}
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-destructive"
                        onClick={() => setDeleteKey(key)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Dialog open={!!deleteKey} onOpenChange={(open) => !open && setDeleteKey(null)}>
        <DialogContent showCloseButton={false} className="max-w-xs">
          <DialogHeader>
            <DialogTitle>Delete key?</DialogTitle>
            <DialogDescription>
              This will permanently delete <code className="text-xs bg-muted px-1.5 py-0.5">{deleteKey?.name}</code> and all its usage logs. This cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" size="sm" onClick={() => setDeleteKey(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => {
                if (deleteKey) {
                  api.deleteKey(deleteKey.id).then(() => {
                    setDeleteKey(null);
                    loadData();
                  });
                }
              }}
            >
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
