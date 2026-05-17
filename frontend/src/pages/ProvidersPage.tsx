import { useEffect, useState } from "react";
import { api, type Provider } from "@/lib/api";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Plus, MoreHorizontal, Trash2, Server, Globe, KeyRound } from "lucide-react";

const PROVIDER_ACCENTS = [
  "border-t-primary/30 from-primary/[0.03] to-transparent",
  "border-t-chart-2/30 from-chart-2/[0.03] to-transparent",
  "border-t-chart-4/30 from-chart-4/[0.03] to-transparent",
  "border-t-chart-1/30 from-chart-1/[0.03] to-transparent",
  "border-t-chart-3/30 from-chart-3/[0.03] to-transparent",
];

export function ProvidersPage() {
  const [providers, setProviders] = useState<Provider[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleting, setDeleting] = useState<string | null>(null);

  // Form state
  const [formID, setFormID] = useState("");
  const [formName, setFormName] = useState("");
  const [formType, setFormType] = useState("openai");
  const [formBaseURL, setFormBaseURL] = useState("");
  const [formAPIKey, setFormAPIKey] = useState("");

  const loadData = async () => {
    const p = await api.listProviders();
    setProviders(p || []);
    setLoading(false);
  };

  useEffect(() => {
    loadData();
  }, []);

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
    loadData();
  };

  const handleDelete = async (id: string) => {
    setDeleting(id);
    setTimeout(async () => {
      await api.deleteProvider(id);
      setDeleting(null);
      loadData();
    }, 200);
  };

  if (loading) {
    return (
      <div className="space-y-6 fade-in-up">
        <div className="flex items-center justify-between">
          <div className="space-y-2">
            <div className="h-8 w-36 shimmer rounded-none" />
            <div className="h-4 w-64 shimmer rounded-none" />
          </div>
          <div className="h-9 w-32 shimmer rounded-none" />
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Card key={i}>
              <CardContent className="p-6">
                <div className="space-y-3">
                  <div className="h-6 w-32 shimmer rounded-none" />
                  <div className="h-4 w-full shimmer rounded-none" />
                  <div className="h-4 w-48 shimmer rounded-none" />
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
          <h2 className="text-2xl font-bold tracking-tight">Providers</h2>
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
                <select
                  id="ptype"
                  className="flex h-9 w-full rounded-none border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  value={formType}
                  onChange={(e) => setFormType(e.target.value)}
                >
                  <option value="openai">OpenAI Compatible</option>
                  <option value="anthropic">Anthropic</option>
                </select>
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

      {/* Provider cards */}
      {providers.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16 text-center">
            <div className="flex h-14 w-14 items-center justify-center bg-gradient-to-br from-primary/10 to-primary/5 mb-4">
              <Server className="h-7 w-7 text-primary/50 float-icon" />
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
              className={`relative group border-t-2 ${PROVIDER_ACCENTS[idx % PROVIDER_ACCENTS.length]} transition-all duration-300 hover:-translate-y-1 hover:shadow-lg hover:shadow-primary/5 bg-gradient-to-b ${
                deleting === p.id ? "opacity-0 scale-95" : "opacity-100 scale-100"
              }`}
            >
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base flex items-center gap-2">
                    <div className="flex h-8 w-8 items-center justify-center bg-gradient-to-br from-primary/15 to-primary/5 text-primary">
                      <Server className="h-4 w-4" />
                    </div>
                    {p.name}
                    <Badge variant="outline" className="text-xs ml-2">
                      {p.type === "anthropic" ? "Anthropic" : "OpenAI"}
                    </Badge>
                  </CardTitle>
                  <DropdownMenu>
                    <DropdownMenuTrigger>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-all duration-200"
                      >
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem
                        className="text-destructive"
                        onClick={() => handleDelete(p.id)}
                      >
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="flex items-center gap-2 text-sm">
                  <Globe className="h-3.5 w-3.5 text-muted-foreground/60" />
                  <span className="text-muted-foreground font-mono text-xs truncate">
                    {p.base_url}
                  </span>
                </div>
                <div className="flex items-center gap-2 text-sm">
                  <KeyRound className="h-3.5 w-3.5 text-muted-foreground/60" />
                  <code className="text-xs bg-muted px-2 py-0.5 border border-border/40">
                    {p.api_key}
                  </code>
                </div>
                <Badge variant="secondary" className="text-xs flex items-center gap-1.5 w-fit">
                  <span className="inline-block h-1.5 w-1.5 rounded-none bg-chart-2 pulse-dot" />
                  Added {new Date(p.created_at).toLocaleDateString()}
                </Badge>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
