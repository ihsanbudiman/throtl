import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
} from "@/components/ui/dialog";
import { api, type APIKey } from "@/lib/api";
import { useToast } from "@/hooks/use-toast";
import { Plus } from "lucide-react";

interface GenerateKeyDialogProps {
  onSuccess: (key: APIKey) => void;
}

export default function GenerateKeyDialog({ onSuccess }: GenerateKeyDialogProps) {
  const { toast } = useToast();
  const [open, setOpen] = useState(false);
  const [formName, setFormName] = useState("");
  const [formLimitWindow, setFormLimitWindow] = useState("");
  const [formLimitDaily, setFormLimitDaily] = useState("");
  const [formWindowHrs, setFormWindowHrs] = useState("5");
  const [formModels, setFormModels] = useState("");

  const resetForm = () => {
    setFormName("");
    setFormLimitWindow("");
    setFormLimitDaily("");
    setFormWindowHrs("5");
    setFormModels("");
  };

  const handleCreate = async () => {
    try {
      const key = await api.createKey({
        name: formName,
        limit_window: formLimitWindow ? parseInt(formLimitWindow) : 0,
        limit_daily: formLimitDaily ? parseInt(formLimitDaily) : 0,
        limit_window_hrs: formWindowHrs ? parseInt(formWindowHrs) : 5,
        allowed_models: formModels,
      });
      setOpen(false);
      resetForm();
      toast({ title: "Key created", description: key.name, variant: "success" });
      onSuccess(key);
    } catch (err) {
      toast({ title: "Failed to create key", description: err instanceof Error ? err.message : "Unknown error", variant: "destructive" });
    }
  };

  return (
    <Dialog open={open} onOpenChange={(open) => { setOpen(open); if (!open) resetForm(); }}>
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
            Create a share key that consumers will use to access the AI provider
            through your gateway.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="name">Key Name</Label>
            <Input
              id="name"
              placeholder="e.g. Team Alpha"
              value={formName}
              autoComplete="off"
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
                autoComplete="off"
                onChange={(e) => setFormLimitWindow(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="windowHrs">Window (hrs)</Label>
              <Input
                id="windowHrs"
                type="number"
                min="1"
                placeholder="5"
                value={formWindowHrs}
                autoComplete="off"
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
                autoComplete="off"
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
              autoComplete="off"
              onChange={(e) => setFormModels(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              Comma-separated model names. Leave empty to allow all models.
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <Button onClick={handleCreate} disabled={!formName}>
            Generate
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
