import { useEffect, useState, useCallback } from "react";
import { api, type Provider } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useToast } from "@/hooks/use-toast";
import { Trash2, Server, Globe, KeyRound, Plus } from "lucide-react";

const PROVIDER_ACCENTS = [
  "from-chart-1/[0.04] to-transparent",
  "from-chart-2/[0.04] to-transparent",
  "from-chart-4/[0.04] to-transparent",
  "from-chart-5/[0.04] to-transparent",
  "from-chart-3/[0.04] to-transparent",
];

const PROVIDER_BORDERS = [
  "border-t-chart-1/30",
  "border-t-chart-2/30",
  "border-t-chart-4/30",
  "border-t-chart-5/30",
  "border-t-chart-3/30",
];

export default function ProvidersPage() {
  const { toast } = useToast();
  const [providers, setProviders] = useState<Provider[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleteConfirm, setDeleteConfirm] = useState<Provider | null>(null);

  // Form state
  const [dialogOpen, setDialogOpen] = useState(false);
  const [formID, setFormID] = useState("");
  const [formName, setFormName] = useState("");
  const [formType, setFormType] = useState("openai");
  const [formBaseURL, setFormBaseURL] = useState("");
  const [formAPIKey, setFormAPIKey] = useState("");

  const loadProviders = useCallback(async () => {
    const p = await api.listProviders();
    setProviders(p || []);
    setLoading(false);
  }, []);

  useEffect(() => {
    loadProviders();
  }, [loadProviders]);

  const handleCreate = async () => {
    await api.createProvider({
      id: formID,
      name: formName,
      type: formType,
      base_url: formBaseURL,
      api_key: formAPIKey,
    });
    setDialogOpen(false);
    setFormID("");
    setFormName("");
    setFormType("openai");
    setFormBaseURL("");
    setFormAPIKey("");
    loadProviders();
  };

  const handleDelete = async (id: string) => {
    try {
      await api.deleteProvider(id);
      toast({ title: "Provider deleted", variant: "destructive" });
      loadProviders();
    } catch (err) {
      toast({ title: "Delete failed", description: err instanceof Error ? err.message : "Unknown error", variant: "destructive" });
    }
  };

  if (loading) {
    return (
      <div className="space-y-6 fade-in-up">
        <div className="flex items-center justify-between">
          <div className="space-y-2">
            <div className="h-8 w-36 shimmer rounded-lg" />
            <div className="h-4 w-64 shimmer rounded-lg" />
          </div>
          <div className="h-9 w-32 shimmer rounded-lg" />
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Card key={i}>
              <CardContent className="p-6">
                <div className="space-y-3">
                  <div className="h-6 w-32 shimmer rounded-lg" />
                  <div className="h-4 w-full shimmer rounded-lg" />
                  <div className="h-4 w-48 shimmer rounded-lg" />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6 fade-in-up">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-[400] tracking-tight">Providers</h2>
          <p className="text-muted-foreground text-sm mt-1">
            Manage upstream AI providers and their API keys
          </p>
        </div>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger>
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Add Provider
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-lg">
            <DialogHeader>
              <DialogTitle>Add Provider</DialogTitle>
              <DialogDescription>
                Connect an upstream AI provider. The API key is stored locally and used to proxy requests.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="pid">Provider ID</Label>
                <Input
                  id="pid"
                  placeholder="e.g. wafer, openai, anthropic"
                  value={formID}
                  onChange={(e) => setFormID(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ""))}
                />
                <p className="text-xs text-muted-foreground">
                  Used in model calls: <code className="bg-muted px-1">{formID || "id"}/model-name</code>
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="pname">Provider Name</Label>
                <Input
                  id="pname"
                  placeholder="e.g. OpenAI, Anthropic"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="ptype">Provider Type</Label>
                <Select value={formType} onValueChange={(v) => { if (v) setFormType(v); }}>
                  <SelectTrigger id="ptype">
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="openai">OpenAI Compatible</SelectItem>
                    <SelectItem value="anthropic">Anthropic</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  {formType === "openai"
                    ? "Any API that follows the /v1/chat/completions format"
                    : "Anthropic Messages API (/v1/messages format)"}
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="purl">Base URL</Label>
                <Input
                  id="purl"
                  placeholder="e.g. https://api.openai.com"
                  value={formBaseURL}
                  onChange={(e) => setFormBaseURL(e.target.value)}
                />
                <p className="text-xs text-muted-foreground">
                  {formType === "openai"
                    ? "OpenAI-compatible endpoint. No trailing slash."
                    : "Anthropic API endpoint. e.g. https://api.anthropic.com"}
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="pkey">API Key</Label>
                <Input
                  id="pkey"
                  type="password"
                  placeholder="sk-..."
                  value={formAPIKey}
                  onChange={(e) => setFormAPIKey(e.target.value)}
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDialogOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleCreate}
                disabled={!formID || !formName || !formBaseURL || !formAPIKey}
              >
                Add Provider
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {providers.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16 text-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-[8px] bg-gradient-to-br from-primary/10 to-primary/5 mb-4">
              <Server className="h-8 w-8 text-primary/40 float-icon" />
            </div>
            <p className="text-muted-foreground text-sm">
              No providers configured. Add one to start sharing AI access.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 fade-in-stagger">
          {providers.map((p, idx) => (
            <Card
              key={p.id}
              className={`relative group border-t-2 ${PROVIDER_BORDERS[idx % PROVIDER_BORDERS.length]} overflow-hidden transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg hover:shadow-primary/5 bg-gradient-to-b ${PROVIDER_ACCENTS[idx % PROVIDER_ACCENTS.length]}`}
            >
              <div className="absolute top-0 right-0 w-32 h-32 bg-gradient-to-bl from-chart-1/[0.03] to-transparent rounded-bl-[4rem]" />
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base font-[400] flex items-center gap-2">
                    <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-gradient-to-br from-primary/15 to-primary/5 text-primary">
                      <Server className="h-4 w-4" />
                    </div>
                    {p.name}
                    <Badge variant="outline" className="text-xs ml-2">
                      {p.type === "anthropic" ? "Anthropic" : "OpenAI"}
                    </Badge>
                  </CardTitle>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 opacity-50 hover:opacity-100 transition-all duration-200 text-body-mid hover:text-destructive"
                    onClick={() => setDeleteConfirm(p)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="flex items-center gap-2 text-sm">
                  <Globe className="h-3.5 w-3.5 text-muted-foreground/60 shrink-0" />
                  <span className="text-muted-foreground font-mono text-xs truncate">
                    {p.base_url}
                  </span>
                </div>
                <div className="flex items-center gap-2 text-sm">
                  <KeyRound className="h-3.5 w-3.5 text-muted-foreground/60 shrink-0" />
                  <code className="text-xs bg-muted px-2 py-0.5 rounded border border-border/40 truncate">
                    {p.api_key}
                  </code>
                </div>
                <Badge
                  variant="secondary"
                  className="text-xs flex items-center gap-1.5 w-fit"
                >
                  <span className="inline-block h-1.5 w-1.5 rounded-full bg-chart-2 pulse-dot" />
                  Added {new Date(p.created_at).toLocaleDateString()}
                </Badge>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog
        open={!!deleteConfirm}
        onOpenChange={(open) => !open && setDeleteConfirm(null)}
      >
        <DialogContent showCloseButton={false} className="max-w-xs">
          <DialogHeader>
            <DialogTitle>Delete provider?</DialogTitle>
            <DialogDescription>
              This will permanently delete{" "}
              <code className="text-xs bg-muted px-1.5 py-0.5 rounded">
                {deleteConfirm?.name}
              </code>{" "}
              and all its models. This cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setDeleteConfirm(null)}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => {
                if (deleteConfirm) {
                  const id = deleteConfirm.id;
                  setDeleteConfirm(null);
                  handleDelete(id);
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
