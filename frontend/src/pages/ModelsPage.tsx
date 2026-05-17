import { useEffect, useState, useMemo, useCallback } from "react";
import { api, type Model, type Provider } from "@/lib/api";
import { Card, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { useDebounce } from "@/lib/useDebounce";
import { useToast } from "@/hooks/use-toast";
import { Cpu, RefreshCw, Search, GripHorizontal } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export default function ModelsPage() {
  const { toast } = useToast();
  const [models, setModels] = useState<Model[]>([]);
  const [providers, setProviders] = useState<Provider[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [search, setSearch] = useState("");
  const [selectedProviderIds, setSelectedProviderIds] = useState<string[]>([]);
  const [editingMultiplier, setEditingMultiplier] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    const [modelsResp, provs] = await Promise.all([api.listModels(), api.listProviders()]);
    setModels(modelsResp?.data || []);
    setProviders(provs || []);
    setLoading(false);
    setRefreshing(false);
  }, []);

  const refresh = async () => { setRefreshing(true); await loadData(); };

  useEffect(() => {
    const fetch = async () => {
      const [modelsResp, provs] = await Promise.all([api.listModels(), api.listProviders()]);
      setModels(modelsResp?.data || []);
      setProviders(provs || []);
      setLoading(false);
      setRefreshing(false);
    };
    fetch();
  }, []);

  const handleToggle = async (modelId: string, currentActive: boolean) => {
    const newActive = !currentActive;
    setModels((prev) => prev.map((m) => (m.id === modelId ? { ...m, active: newActive } : m)));
    try { await api.toggleModel(modelId, newActive); toast({ title: newActive ? "Model enabled" : "Model disabled", variant: "success" }); }
    catch { setModels((prev) => prev.map((m) => (m.id === modelId ? { ...m, active: currentActive } : m))); toast({ title: "Failed to update model", variant: "destructive" }); }
  };

  const handleMultiplierChange = async (modelId: string, value: string) => {
    const mult = parseInt(value, 10);
    if (isNaN(mult) || mult < 1) {
      toast({ title: "Multiplier must be a positive number", variant: "destructive" });
      return;
    }
    const prevMult = models.find((m) => m.id === modelId)?.request_multiplier ?? 1;
    setModels((prev) => prev.map((m) => (m.id === modelId ? { ...m, request_multiplier: mult } : m)));
    try {
      await api.updateModel(modelId, { request_multiplier: mult });
      toast({ title: `Multiplier set to ${mult}x`, variant: "success" });
    } catch {
      setModels((prev) => prev.map((m) => (m.id === modelId ? { ...m, request_multiplier: prevMult } : m)));
      toast({ title: "Failed to update multiplier", variant: "destructive" });
    }
    setEditingMultiplier(null);
  };

  const providerMap = useMemo(() => Object.fromEntries(providers.map((p) => [p.id, p])), [providers]);
  const debouncedSearch = useDebounce(search, 200);

  const filtered = useMemo(() => {
    let result = models;
    if (debouncedSearch) {
      result = result.filter((m) => m.id.toLowerCase().includes(debouncedSearch.toLowerCase()));
    }
    if (selectedProviderIds.length > 0) {
      result = result.filter((m) => selectedProviderIds.includes(m.provider_id));
    }
    return result;
  }, [models, debouncedSearch, selectedProviderIds]);

  const toggleProvider = (providerId: string) => {
    setSelectedProviderIds((prev) =>
      prev.includes(providerId)
        ? prev.filter((id) => id !== providerId)
        : [...prev, providerId]
    );
  };

  const allProvidersSelected = selectedProviderIds.length === 0 || selectedProviderIds.length === providers.length;
  const filteredProviderCount = useMemo(() => new Set(filtered.map((m) => m.provider_id)).size, [filtered]);

  if (loading) {
    return (
      <div className="space-y-6 fade-in-up">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="space-y-2"><div className="h-8 w-36 shimmer rounded-lg" /><div className="h-4 w-64 shimmer rounded-lg" /></div>
          <div className="h-9 w-32 shimmer rounded-lg shrink-0" />
        </div>
        <Card><CardContent className="p-6"><div className="space-y-3">{[1, 2, 3, 4, 5].map((i) => <div key={i} className="h-10 w-full shimmer rounded-lg" />)}</div></CardContent></Card>
      </div>
    );
  }

  return (
    <div className="space-y-6 fade-in-up">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-2xl font-[400] tracking-tight">Models</h2>
          <p className="text-muted-foreground text-sm mt-1">
            {filtered.length} model{filtered.length !== 1 ? "s" : ""} across {filteredProviderCount} provider{filteredProviderCount !== 1 ? "s" : ""}
            {filtered.length !== models.length && (
              <span className="text-muted-foreground/60"> (filtered from {models.length})</span>
            )}
          </p>
        </div>
        <div className="flex items-center gap-2 flex-wrap">
          <div className="relative flex-1 sm:flex-none min-w-0">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
            <Input placeholder="Search models..." value={search} onChange={(e) => setSearch(e.target.value)} className="h-8 w-full sm:w-56 pl-8" />
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger className="group inline-flex h-7 shrink-0 cursor-default items-center justify-center gap-1 rounded-full border border-hairline bg-transparent px-3 text-xs font-[400] text-ink whitespace-nowrap outline-none select-none transition-all duration-150 hover:bg-canvas-soft hover:text-ink focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/30">
              <Cpu className="h-3.5 w-3.5 shrink-0" />
              {allProvidersSelected
                ? "All Providers"
                : `${selectedProviderIds.length} provider${selectedProviderIds.length !== 1 ? "s" : ""}`}
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48">
              <DropdownMenuGroup>
              <DropdownMenuLabel>Filter by provider</DropdownMenuLabel>
              <DropdownMenuSeparator />
              {providers.map((p) => {
                const modelCount = models.filter((m) => m.provider_id === p.id).length;
                const checked = selectedProviderIds.includes(p.id);
                return (
                  <DropdownMenuCheckboxItem
                    key={p.id}
                    checked={checked}
                    onCheckedChange={() => toggleProvider(p.id)}
                  >
                    <span className="flex w-full items-center justify-between gap-2">
                      <span>{p.name}</span>
                      <span className="text-xs text-muted-foreground">{modelCount}</span>
                    </span>
                  </DropdownMenuCheckboxItem>
                );
              })}
              </DropdownMenuGroup>
            </DropdownMenuContent>
          </DropdownMenu>
          <Button variant="outline" size="sm" onClick={refresh} disabled={refreshing}>
            <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${refreshing ? "animate-spin" : ""}`} />Refresh
          </Button>
        </div>
      </div>

      {models.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16 text-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-[8px] bg-gradient-to-br from-primary/10 to-primary/5 mb-4">
              <Cpu className="h-8 w-8 text-primary/40 float-icon" />
            </div>
            <p className="text-muted-foreground text-sm">No models found. Add a provider with a valid API key to see available models.</p>
          </CardContent>
        </Card>
      ) : (
        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                  <TableRow>
                    <TableHead>Provider</TableHead>
                    <TableHead>Model ID</TableHead>
                    <TableHead>Call As</TableHead>
                    <TableHead>Req. Weight</TableHead>
                    <TableHead>Status</TableHead>
                  </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.length === 0 ? (
                  <TableRow>
                    <TableCell                     colSpan={5} className="text-center text-muted-foreground text-sm py-8">
                      No models match the current filters.
                    </TableCell>
                  </TableRow>
                ) : (
                  filtered.map((m) => {
                    const bareId = m.id.includes("/") ? m.id.split("/").slice(1).join("/") : m.id;
                    const prov = providerMap[m.provider_id];
                    return (
                      <TableRow key={m.id} className={`transition-colors duration-150 hover:bg-accent/30 ${!m.active ? "opacity-50" : ""}`}>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <div className="flex h-6 w-6 items-center justify-center rounded-md bg-gradient-to-br from-primary/15 to-primary/5 text-primary shrink-0">
                              <Cpu className="h-3 w-3" />
                            </div>
                            <span className="text-sm">{prov?.name || m.provider_id}</span>
                          </div>
                        </TableCell>
                        <TableCell className="font-mono text-sm">{bareId}</TableCell>
                        <TableCell><code className="text-xs bg-muted px-2 py-0.5 rounded border border-border/40">{m.id}</code></TableCell>
                        <TableCell>
                          {editingMultiplier === m.id ? (
                            <Input
                              type="number"
                              min={1}
                              defaultValue={m.request_multiplier ?? 1}
                              className="h-7 w-16 text-xs text-center"
                              autoFocus
                              onBlur={(e) => handleMultiplierChange(m.id, e.target.value)}
                              onKeyDown={(e) => { if (e.key === "Enter") handleMultiplierChange(m.id, (e.target as HTMLInputElement).value); if (e.key === "Escape") setEditingMultiplier(null); }}
                            />
                          ) : (
                            <button
                              className="inline-flex items-center gap-1 text-xs font-mono bg-muted px-2 py-1 rounded border border-border/40 hover:bg-accent transition-colors cursor-pointer min-w-[3rem] justify-center"
                              onClick={() => setEditingMultiplier(m.id)}
                              title="Click to edit request weight"
                            >
                              <GripHorizontal className="h-3 w-3 text-muted-foreground" />
                              {m.request_multiplier ?? 1}x
                            </button>
                          )}
                        </TableCell>
                        <TableCell><Switch checked={m.active} onCheckedChange={() => handleToggle(m.id, m.active)} /></TableCell>
                      </TableRow>
                    );
                  })
                )}
              </TableBody>
            </Table>
          </div>
        </Card>
      )}
    </div>
  );
}
