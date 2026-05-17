import { useState, memo } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { api, type APIKey } from "@/lib/api";
import { useToast } from "@/hooks/use-toast";
import { Plus } from "lucide-react";

interface GenerateKeyDialogProps {
  onSuccess: (key: APIKey) => void;
}

export default function GenerateKeyDialog({ onSuccess }: GenerateKeyDialogProps) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={(open) => setOpen(open)}>
      <DialogTrigger>
        <Button>
          <Plus className="h-4 w-4 mr-2" />
          Generate Key
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <GenerateKeyForm
          onSuccess={onSuccess}
          onClose={() => setOpen(false)}
        />
      </DialogContent>
    </Dialog>
  );
}

interface GenerateKeyFormProps {
  onSuccess: (key: APIKey) => void;
  onClose: () => void;
}

const GenerateKeyForm = memo(function GenerateKeyForm({
  onSuccess,
  onClose,
}: GenerateKeyFormProps) {
  const { toast } = useToast();
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
      resetForm();
      onClose();
      toast({ title: "Key created", description: key.name, variant: "success" });
      onSuccess(key);
    } catch (err) {
      toast({ title: "Failed to create key", description: err instanceof Error ? err.message : "Unknown error", variant: "destructive" });
    }
  };

  const handleCancel = () => {
    resetForm();
    onClose();
  };

  const windowLimit = formLimitWindow ? parseInt(formLimitWindow) : 0;
  const dailyLimit = formLimitDaily ? parseInt(formLimitDaily) : 0;

  return (
    <>
      <DialogHeader>
        <DialogTitle>Generate New API Key</DialogTitle>
        <p className="text-sm text-muted-foreground">
          Create a share key (<code className="bg-muted px-1 rounded text-xs">sk-share-...</code>) for each person or project.
          The generated key works with any OpenAI-compatible client — just swap the base
          URL and API key.
        </p>
      </DialogHeader>
      <div className="space-y-6 py-4">
        <div className="space-y-2">
          <Label htmlFor="name">Key Name</Label>
          <Input
            id="name"
            placeholder="e.g. Team Alpha, Project X"
            value={formName}
            autoComplete="off"
            onChange={(e) => setFormName(e.target.value)}
          />
          <p className="text-xs text-muted-foreground">
            A friendly label to identify who or what this key is for.
          </p>
        </div>

        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <div className="h-px flex-1 bg-border/50" />
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              Rate Limits
            </span>
            <div className="h-px flex-1 bg-border/50" />
          </div>
          <p className="text-xs text-muted-foreground">
            Set limits to control how much each key can be used. Leave both at 0 for
            unrestricted access.
          </p>

          <div className="rounded-[8px] border border-border/50 bg-muted/30 p-4 space-y-4">
            <div>
              <div className="flex items-center justify-between mb-2">
                <Label className="text-sm font-medium">Rolling Window Limit</Label>
                {windowLimit > 0 && (
                  <Badge variant="outline" className="text-xs">Active</Badge>
                )}
              </div>
              <div className="grid grid-cols-3 gap-3">
                <div className="col-span-2 space-y-1.5">
                  <Label htmlFor="limitWindow" className="text-xs text-muted-foreground font-normal">
                    Max requests
                  </Label>
                  <Input
                    id="limitWindow"
                    type="number"
                    min="0"
                    placeholder="0"
                    value={formLimitWindow}
                    autoComplete="off"
                    onChange={(e) => setFormLimitWindow(e.target.value)}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="windowHrs" className="text-xs text-muted-foreground font-normal">
                    Per (hours)
                  </Label>
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
              </div>
              <p className="text-xs text-muted-foreground mt-1.5">
                {windowLimit > 0
                  ? `Limits to ${windowLimit} requests per ${formWindowHrs || "5"} hours (sliding window).`
                  : "No window limit — requests are not throttled by time window."}
              </p>
            </div>

            <div>
              <div className="flex items-center justify-between mb-2">
                <Label htmlFor="limitDaily" className="text-sm font-medium">Daily Limit</Label>
                {dailyLimit > 0 && (
                  <Badge variant="outline" className="text-xs">Active</Badge>
                )}
              </div>
              <Input
                id="limitDaily"
                type="number"
                min="0"
                placeholder="0"
                value={formLimitDaily}
                autoComplete="off"
                onChange={(e) => setFormLimitDaily(e.target.value)}
              />
              <p className="text-xs text-muted-foreground mt-1.5">
                {dailyLimit > 0
                  ? `Caps at ${dailyLimit} requests per day. Resets at midnight UTC.`
                  : "No daily cap — requests are not limited per day."}
              </p>
            </div>
          </div>
        </div>

        <div className="space-y-2">
          <Label htmlFor="models">Allowed Models</Label>
          <Input
            id="models"
            placeholder="e.g. wafer/gpt-4o,wafer/gpt-4o-mini"
            value={formModels}
            autoComplete="off"
            onChange={(e) => setFormModels(e.target.value)}
          />
          <p className="text-xs text-muted-foreground">
            Comma-separated model names in{" "}
            <code className="bg-muted px-1 rounded text-xs">provider/model</code> format.
            Leave empty to allow all models.
          </p>
        </div>
      </div>
      <DialogFooter>
        <Button variant="outline" onClick={handleCancel}>
          Cancel
        </Button>
        <Button onClick={handleCreate} disabled={!formName}>
          Generate
        </Button>
      </DialogFooter>
    </>
  );
});
