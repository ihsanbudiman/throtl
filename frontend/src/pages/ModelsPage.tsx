import { useEffect, useState } from "react";
import { api, type Model, type Provider } from "@/lib/api";
import {
  Card,
  CardContent,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Cpu, RefreshCw, Search } from "lucide-react";
import { Input } from "@/components/ui/input";

export function ModelsPage() {
  const [models, setModels] = useState<Model[]>([]);
  const [providers, setProviders] = useState<Provider[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [search, setSearch] = useState("");

  const loadData = async () => {
    const [modelsResp, provs] = await Promise.all([
      api.listModels(),
      api.listProviders(),
    ]);
    setModels(modelsResp?.data || []);
    setProviders(provs || []);
    setLoading(false);
    setRefreshing(false);
  };

  const refresh = async () => {
    setRefreshing(true);
    await loadData();
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleToggle = async (modelId: string, currentActive: boolean) => {
    const newActive = !currentActive;
    setModels((prev) =>
      prev.map((m) => (m.id === modelId ? { ...m, active: newActive } : m))
    );
    try {
      await api.toggleModel(modelId, newActive);
    } catch {
      setModels((prev) =>
        prev.map((m) => (m.id === modelId ? { ...m, active: currentActive } : m))
      );
    }
  };

  const providerMap = Object.fromEntries(providers.map((p) => [p.id, p]));

  const filtered = models.filter((m) =>
    m.id.toLowerCase().includes(search.toLowerCase())
  );

  const groupedByProvider = filtered.reduce<Record<string, Model[]>>(
    (acc, m) => {
      const pid = m.provider_id;
      if (!acc[pid]) acc[pid] = [];
      acc[pid].push(m);
      return acc;
    },
    {}
  );

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
        <Card>
          <CardContent className="p-6">
            <div className="space-y-3">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="h-10 w-full shimmer rounded-none" />
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
          <h2 className="text-2xl font-bold tracking-tight">Models</h2>
          <p className="text-muted-foreground text-sm mt-1">
            {models.length} models across {providers.length} provider{providers.length !== 1 ? "s" : ""}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
            <Input
              placeholder="Search models..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="h-8 w-56 pl-8"
            />
          </div>
          <Button variant="outline" size="sm" onClick={refresh} disabled={refreshing}>
            <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${refreshing ? "animate-spin" : ""}`} />
            Refresh
          </Button>
        </div>
      </div>

      {models.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16 text-center">
            <div className="flex h-14 w-14 items-center justify-center bg-gradient-to-br from-primary/10 to-primary/5 mb-4">
              <Cpu className="h-7 w-7 text-primary/50 float-icon" />
            </div>
            <p className="text-muted-foreground text-sm">
              No models found. Add a provider with a valid API key to see available models.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4 fade-in-stagger">
          {Object.entries(groupedByProvider).map(([providerId, providerModels]) => {
            const prov = providerMap[providerId];
            const activeCount = providerModels.filter((m) => m.active).length;
            return (
              <Card key={providerId}>
                <CardContent className="p-0">
                  <div className="flex items-center gap-3 px-5 py-3 border-b border-border bg-muted/30">
                    <div className="flex h-7 w-7 items-center justify-center bg-gradient-to-br from-primary/15 to-primary/5 text-primary">
                      <Cpu className="h-3.5 w-3.5" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <span className="text-sm font-semibold">{prov?.name || providerId}</span>
                      <span className="text-xs text-muted-foreground ml-2 font-mono">
                        {prov?.base_url || ""}
                      </span>
                    </div>
                    <Badge variant="secondary" className="text-xs">
                      {activeCount}/{providerModels.length} active
                    </Badge>
                  </div>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Model ID</TableHead>
                        <TableHead>Call As</TableHead>
                        <TableHead>Status</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {providerModels.map((m) => {
                        const bareId = m.id.includes("/") ? m.id.split("/").slice(1).join("/") : m.id;
                        return (
                          <TableRow
                            key={m.id}
                            className={`transition-colors duration-150 hover:bg-accent/50 ${!m.active ? "opacity-50" : ""}`}
                          >
                            <TableCell className="font-mono text-sm">
                              {bareId}
                            </TableCell>
                            <TableCell>
                              <code className="text-xs bg-muted px-2 py-0.5 border border-border/40">
                                {m.id}
                              </code>
                            </TableCell>
                            <TableCell>
                              <Switch
                                checked={m.active}
                                onCheckedChange={() => handleToggle(m.id, m.active)}
                              />
                            </TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
