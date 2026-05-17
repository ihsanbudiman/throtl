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
import { api } from "@/lib/api";
import { useToast } from "@/hooks/use-toast";
import { Plus } from "lucide-react";

interface AddProviderDialogProps {
  onSuccess: () => void;
}

export default function AddProviderDialog({ onSuccess }: AddProviderDialogProps) {
  const { toast } = useToast();
  const [open, setOpen] = useState(false);
  const [formID, setFormID] = useState("");
  const [formName, setFormName] = useState("");
  const [formBaseURL, setFormBaseURL] = useState("");
  const [formAPIKey, setFormAPIKey] = useState("");

  const handleCreate = async () => {
    await api.createProvider({
      id: formID,
      name: formName,
      base_url: formBaseURL,
      api_key: formAPIKey,
    });
    setOpen(false);
    setFormID("");
    setFormName("");
    setFormBaseURL("");
    setFormAPIKey("");
    toast({ title: "Provider added", description: formName, variant: "success" });
    onSuccess();
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
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
            Connect an upstream AI provider. The API key is stored locally and
            used to proxy requests.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="pid">Provider ID</Label>
            <Input
              id="pid"
              placeholder="e.g. wafer, openai, anthropic"
              value={formID}
              onChange={(e) =>
                setFormID(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ""))
              }
            />
            <p className="text-xs text-muted-foreground">
              Used in model calls:{" "}
              <code className="bg-muted px-1 rounded">
                {formID || "id"}/model-name
              </code>
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
            <Label htmlFor="purl">Base URL</Label>
            <Input
              id="purl"
              placeholder="e.g. https://api.openai.com"
              value={formBaseURL}
              onChange={(e) => setFormBaseURL(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              OpenAI-compatible endpoint. No trailing slash.
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
          <Button variant="outline" onClick={() => setOpen(false)}>
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
  );
}
