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

  const handleCreate = async () => {
    const key = await api.createKey({
      name: formName,
      limit_window: formLimitWindow ? parseInt(formLimitWindow) : 0,
      limit_daily: formLimitDaily ? parseInt(formLimitDaily) : 0,
      limit_window_hrs: formWindowHrs ? parseInt(formWindowHrs) : 5,
      allowed_models: formModels,
    });
    setOpen(false);
    setFormName("");
    setFormLimitWindow("");
    setFormLimitDaily("");
    setFormWindowHrs("5");
    setFormModels("");
    toast({ title: "Key created", description: key.name, variant: "success" });
    onSuccess(key);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
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
              <Label htmlFor="windowHrs">Window (hrs)</Label>
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
