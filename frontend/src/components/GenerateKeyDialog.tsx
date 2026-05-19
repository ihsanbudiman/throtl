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
  const [formLimitDaily, setFormLimitDaily] = useState("");
  const [formTokensInDaily, setFormTokensInDaily] = useState("");
  const [formTokensOutDaily, setFormTokensOutDaily] = useState("");
  const [formModels, setFormModels] = useState("");

  const resetForm = () => {
    setFormName("");
    setFormLimitDaily("");
    setFormTokensInDaily("");
    setFormTokensOutDaily("");
    setFormModels("");
  };

  const handleCreate = async () => {
    try {
      const key = await api.createKey({
        name: formName,
        limit_daily: formLimitDaily ? parseInt(formLimitDaily) : 0,
        limit_tokens_in_daily: formTokensInDaily ? parseInt(formTokensInDaily) : 0,
        limit_tokens_out_daily: formTokensOutDaily ? parseInt(formTokensOutDaily) : 0,
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

  const dailyLimit = formLimitDaily ? parseInt(formLimitDaily) : 0;
  const tokensInDaily = formTokensInDaily ? parseInt(formTokensInDaily) : 0;
  const tokensOutDaily = formTokensOutDaily ? parseInt(formTokensOutDaily) : 0;

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
                <Label htmlFor="limitDaily" className="text-sm font-medium">Daily Request Limit</Label>
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

            <div>
              <div className="flex items-center justify-between mb-2">
                <Label htmlFor="tokensInDaily" className="text-sm font-medium">Daily Token-In Limit</Label>
                {tokensInDaily > 0 && (
                  <Badge variant="outline" className="text-xs">Active</Badge>
                )}
              </div>
              <Input
                id="tokensInDaily"
                type="number"
                min="0"
                placeholder="0"
                value={formTokensInDaily}
                autoComplete="off"
                onChange={(e) => setFormTokensInDaily(e.target.value)}
              />
              <p className="text-xs text-muted-foreground mt-1.5">
                {tokensInDaily > 0
                  ? `Limits to ${tokensInDaily} input tokens per day. Resets at midnight UTC.`
                  : "No input token limit."}
              </p>
            </div>

            <div>
              <div className="flex items-center justify-between mb-2">
                <Label htmlFor="tokensOutDaily" className="text-sm font-medium">Daily Token-Out Limit</Label>
                {tokensOutDaily > 0 && (
                  <Badge variant="outline" className="text-xs">Active</Badge>
                )}
              </div>
              <Input
                id="tokensOutDaily"
                type="number"
                min="0"
                placeholder="0"
                value={formTokensOutDaily}
                autoComplete="off"
                onChange={(e) => setFormTokensOutDaily(e.target.value)}
              />
              <p className="text-xs text-muted-foreground mt-1.5">
                {tokensOutDaily > 0
                  ? `Limits to ${tokensOutDaily} output tokens per day. Resets at midnight UTC.`
                  : "No output token limit."}
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
