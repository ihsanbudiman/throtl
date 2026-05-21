import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { BookOpen, ArrowRight, ChevronRight, ChevronDown, Terminal, KeyRound, Server, AlertTriangle, Info } from "lucide-react";

function SectionHeading({ num, title, description }: { num: string; title: string; description: string }) {
  return (
    <div className="mb-6">
      <div className="flex items-center gap-3 mb-2">
        <span className="font-mono text-xs font-semibold text-chart-1">{num}</span>
        <h3 className="text-xl font-[400] tracking-tight">{title}</h3>
      </div>
      <p className="text-sm text-muted-foreground max-w-2xl">{description}</p>
    </div>
  );
}

function Collapsible({ title, where, children, defaultOpen }: { title: string; where?: string; children: React.ReactNode; defaultOpen?: boolean }) {
  const [open, setOpen] = useState(defaultOpen ?? false);
  return (
    <div className="border border-hairline rounded-[8px] bg-card overflow-hidden mb-4">
      <button
        onClick={() => setOpen(!open)}
        className="w-full flex items-center gap-2 px-5 py-4 text-left hover:bg-canvas-soft/50 transition-colors"
      >
        {open ? <ChevronDown className="h-4 w-4 text-chart-1 shrink-0" /> : <ChevronRight className="h-4 w-4 text-chart-1 shrink-0" />}
        <span className="text-sm font-[500] flex-1">{title}</span>
        {where && <span className="font-mono text-[11px] text-muted-foreground">{where}</span>}
      </button>
      {open && <div className="px-5 pb-5 text-sm">{children}</div>}
    </div>
  );
}

function CodeBlock({ children }: { children: string }) {
  return (
    <pre className="bg-canvas/50 border border-hairline rounded-[8px] p-5 font-mono text-xs leading-relaxed overflow-x-auto my-4">
      {children}
    </pre>
  );
}

function Callout({ children, variant = "info" }: { children: React.ReactNode; variant?: "info" | "warn" }) {
  const Icon = variant === "warn" ? AlertTriangle : Info;
  return (
    <div className={`flex gap-3 p-5 rounded-[8px] border ${variant === "warn" ? "border-chart-1/30 bg-chart-1/5" : "border-hairline bg-canvas/30"} my-5`}>
      <Icon className={`h-4 w-4 shrink-0 mt-0.5 ${variant === "warn" ? "text-chart-1" : "text-chart-2"}`} />
      <div className="text-sm leading-relaxed">{children}</div>
    </div>
  );
}

function StepCard({ num, title, description, icon: Icon }: { num: number; title: string; description: string; icon: typeof Server }) {
  return (
    <Card className="group hover:border-chart-1/30 transition-all duration-200 hover:shadow-md">
      <CardHeader className="pb-3">
        <div className="flex items-center gap-2 mb-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-chart-1/10">
            <Icon className="h-4 w-4 text-chart-1" />
          </div>
          <span className="font-mono text-xs text-chart-1 font-semibold">Step {num}</span>
        </div>
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <CardDescription className="text-xs leading-relaxed">{description}</CardDescription>
      </CardContent>
    </Card>
  );
}

function FlowDiagram() {
  return (
    <div className="flex items-center gap-4 flex-wrap my-8 font-mono text-xs">
      <span className="bg-card border border-hairline rounded-lg px-4 py-2.5">Your AI Tool</span>
      <ArrowRight className="h-4 w-4 text-muted-foreground" />
      <span className="bg-chart-1/10 border border-chart-1/30 rounded-lg px-4 py-2.5 text-chart-1">Throtl Gateway</span>
      <ArrowRight className="h-4 w-4 text-muted-foreground" />
      <span className="bg-card border border-hairline rounded-lg px-4 py-2.5">Provider API</span>
    </div>
  );
}

const faqItems = [
  { q: "Can I connect multiple providers?", a: "Yes. Add as many providers as you need — OpenAI-compatible, Anthropic, or mixed. Models from different providers appear together and are referenced with their provider prefix: wafer/GLM-5.1, openai/gpt-4o." },
  { q: "Does Throtl support streaming?", a: "Yes. Both OpenAI-style Server-Sent Events and Anthropic-style streaming are supported. Throtl handles the format conversion transparently." },
  { q: "Can I revoke a share key instantly?", a: "Yes. Disable or delete any key from the API Keys page. The change takes effect immediately — no caching delay." },
  { q: "What happens to my real API key?", a: "It never leaves the Throtl server. Users only see and use sk-share-... keys. Your provider keys are stored securely in the database." },
  { q: "Can I limit which models a key can access?", a: "Yes. When creating a share key, you can specify allowed models. Requests to models not in the allowed list are blocked before reaching the provider." },
  { q: "How do I monitor usage?", a: "The Usage page shows real-time charts: requests per key, model breakdown, token flow (input vs output), and date range filters." },
  { q: "Can I use Throtl with other tools?", a: "Anything that speaks the OpenAI API format works — Cursor, Windsurf, VS Code extensions, custom scripts, the OpenAI SDK in any language." },
];

export default function DocumentationPage() {
  return (
    <div className="space-y-8 animate-[fade-in-up_0.4s_ease-out]">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-[400] tracking-tight">Documentation</h2>
          <p className="text-muted-foreground text-sm mt-1">How to use Throtl with your AI tools</p>
        </div>
        <div className="hidden sm:flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 text-primary">
          <BookOpen className="h-4 w-4" />
        </div>
      </div>

      {/* TL;DR */}
      <Card className="border-l-2 border-l-chart-1">
        <CardContent className="p-5">
          <div className="flex gap-3">
            <Info className="h-4 w-4 text-chart-1 shrink-0 mt-0.5" />
            <div className="text-sm leading-relaxed">
              <strong className="text-foreground">Quick start:</strong> Add your provider key in the dashboard. Create a share key for each user or project. Point your AI tool's API endpoint to your Throtl URL and use the share key as the API key. Throtl handles rate limiting, model access control, and usage logging automatically.
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Flow diagram */}
      <FlowDiagram />

      {/* ═══════════════════════════════════ */}
      <section>
        <SectionHeading
          num="01"
          title="Getting Started"
          description="Three steps to go from zero to a working share key. Each step takes less than a minute."
        />

        <div className="grid gap-6 sm:grid-cols-3 fade-in-stagger">
          <StepCard
            num={1}
            title="Add a Provider"
            icon={Server}
            description="Go to Providers in the dashboard. Enter your provider's base URL, API key, and available models. Throtl supports any OpenAI-compatible endpoint and Anthropic (with automatic format conversion)."
          />
          <StepCard
            num={2}
            title="Create a Share Key"
            icon={KeyRound}
            description="Go to API Keys. Set a name, rate limits, and allowed models. Throtl generates a sk-share-... key. Share this key — not your real provider key."
          />
          <StepCard
            num={3}
            title="Test the Connection"
            icon={Terminal}
            description="Verify everything works with a simple curl request before configuring your AI tools."
          />
        </div>

        <div className="mt-6">
        <Collapsible title="Test your share key" where="terminal">
          <p className="text-muted-foreground mb-2">Replace the URL and key with your actual values:</p>
          <CodeBlock>{`# Test the chat completions endpoint
curl https://your-throtl.example.com/v1/chat/completions \\
  -H "Authorization: Bearer sk-share-..." \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "provider-id/GLM-5.1",
    "messages": [{"role": "user", "content": "Say hello!"}]
  }'

# Discover available models
curl https://your-throtl.example.com/v1/models \\
  -H "Authorization: Bearer sk-share-..."`}</CodeBlock>
        </Collapsible>
        </div>
      </section>

      {/* ═══════════════════════════════════ */}
      <section>
        <SectionHeading
          num="02"
          title="Using with Claude Code"
          description="Claude Code speaks the Anthropic API format. Throtl automatically converts between OpenAI and Anthropic formats, so your underlying provider can be anything."
        />

        <Collapsible title="Quick start: environment variables" where="~/.bashrc or ~/.zshrc" defaultOpen>
          <p className="text-muted-foreground mb-2">The fastest way to route Claude Code through Throtl:</p>
          <CodeBlock>{`# Point Claude Code to your Throtl gateway
export ANTHROPIC_BASE_URL="https://your-throtl.example.com/v1"

# Use your Throtl share key
export ANTHROPIC_API_KEY="sk-share-..."

# Optional: override the default model
export ANTHROPIC_MODEL="provider-id/GLM-5.1"`}</CodeBlock>
          <p className="text-muted-foreground mt-2">Then launch Claude Code as usual. All requests flow through Throtl.</p>
        </Collapsible>

        <Collapsible title="Persistent config: settings.json" where="~/.claude/settings.json">
          <p className="text-muted-foreground mb-2">For a persistent configuration across all projects:</p>
          <CodeBlock>{`// ~/.claude/settings.json
{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "env": {
    "ANTHROPIC_BASE_URL": "https://your-throtl.example.com/v1",
    "ANTHROPIC_API_KEY": "sk-share-..."
  }
}`}</CodeBlock>
          <div className="mt-3 text-xs text-muted-foreground space-y-1">
            <p><code className="bg-canvas/50 px-1.5 py-0.5 rounded">~/.claude/settings.json</code> — You, every project</p>
            <p><code className="bg-canvas/50 px-1.5 py-0.5 rounded">.claude/settings.json</code> — Everyone in the project</p>
            <p><code className="bg-canvas/50 px-1.5 py-0.5 rounded">.claude/settings.local.json</code> — You, this project only</p>
          </div>
        </Collapsible>

        <Collapsible title="Model discovery from Throtl" where="env var">
          <p className="text-muted-foreground mb-2">Claude Code can discover available models directly from your Throtl gateway:</p>
          <CodeBlock>{`export CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY="1"`}</CodeBlock>
          <p className="text-muted-foreground mt-2">This makes Claude Code query your Throtl /v1/models endpoint to find available models.</p>
        </Collapsible>

        <Callout variant="warn">
          <strong>Format note.</strong> Claude Code requires an Anthropic-compatible endpoint. Throtl handles the format conversion automatically — your underlying provider can be any OpenAI-compatible API.
        </Callout>
      </section>

      {/* ═══════════════════════════════════ */}
      <section>
        <SectionHeading
          num="03"
          title="Using with OpenCode"
          description="OpenCode uses the Vercel AI SDK and supports any OpenAI-compatible endpoint through @ai-sdk/openai-compatible."
        />

        <Collapsible title="Project-level config" where="./opencode.json" defaultOpen>
          <p className="text-muted-foreground mb-2">Create or edit opencode.json in your project root:</p>
          <CodeBlock>{`// ./opencode.json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "throtl": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Throtl Gateway",
      "options": {
        "baseURL": "https://your-throtl.example.com/v1",
        "apiKey": "{env:THROTL_API_KEY}"
      },
      "models": {
        "glm-5.1": {
          "name": "GLM 5.1",
          "model": "provider-id/GLM-5.1"
        }
      }
    }
  }
}`}</CodeBlock>
          <p className="text-muted-foreground mt-2">Then set the API key as an environment variable:</p>
          <CodeBlock>{`export THROTL_API_KEY="sk-share-..."`}</CodeBlock>
        </Collapsible>

        <Collapsible title="Global config" where="~/.config/opencode/opencode.json">
          <p className="text-muted-foreground mb-2">For a configuration that applies across all projects:</p>
          <CodeBlock>{`// ~/.config/opencode/opencode.json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "throtl": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Throtl Gateway",
      "options": {
        "baseURL": "https://your-throtl.example.com/v1",
        "apiKey": "{env:THROTL_API_KEY}"
      }
    }
  }
}`}</CodeBlock>
          <div className="mt-3 text-xs text-muted-foreground space-y-1">
            <p><code className="bg-canvas/50 px-1.5 py-0.5 rounded">~/.config/opencode/opencode.json</code> — Global user defaults</p>
            <p><code className="bg-canvas/50 px-1.5 py-0.5 rounded">./opencode.json</code> — Project-specific (overrides global)</p>
            <p><code className="bg-canvas/50 px-1.5 py-0.5 rounded">OPENCODE_CONFIG_CONTENT</code> — Inline JSON (runtime override)</p>
          </div>
        </Collapsible>

        <Callout>
          <strong>Tip.</strong> Use <code className="bg-canvas/50 px-1.5 py-0.5 rounded">{`{env:VAR_NAME}`}</code> syntax to reference environment variables in your config. Never hardcode API keys in config files.
        </Callout>
      </section>

      {/* ═══════════════════════════════════ */}
      <section>
        <SectionHeading
          num="04"
          title="Model Format"
          description="Throtl uses a composite model identifier that combines the provider ID with the model name."
        />

        <CodeBlock>{`// Model format: provider-id/model-name
"model": "wafer/GLM-5.1"
"model": "kimi/K2.6"
"model": "openai/gpt-4o"`}</CodeBlock>

        <Collapsible title="Discover available models" where="/v1/models">
          <p className="text-muted-foreground mb-2">Users can discover which models are available:</p>
          <CodeBlock>{`curl https://your-throtl.example.com/v1/models \\
  -H "Authorization: Bearer sk-share-..."`}</CodeBlock>
          <p className="text-muted-foreground mt-2">This returns all enabled models from connected providers. Disabled models won't appear.</p>
        </Collapsible>

        <Callout>
          <strong>Model access control.</strong> As an admin, disable any model from the Models page. Disabled models won't appear in the models list and requests are blocked at the gateway.
        </Callout>
      </section>

      {/* ═══════════════════════════════════ */}
      <section>
        <SectionHeading
          num="05"
          title="Rate Limits"
          description="Each share key has independent daily limits that reset at midnight UTC."
        />

        <Card>
          <CardContent className="p-0">
            <div className="divide-y divide-hairline">
              <div className="grid grid-cols-3 gap-4 px-4 py-3 bg-canvas/30">
                <span className="font-mono text-[11px] text-muted-foreground uppercase tracking-wider">Limit Type</span>
                <span className="font-mono text-[11px] text-muted-foreground uppercase tracking-wider">What It Counts</span>
                <span className="font-mono text-[11px] text-muted-foreground uppercase tracking-wider">Default</span>
              </div>
              <div className="grid grid-cols-3 gap-4 px-4 py-3 text-sm">
                <code className="text-xs bg-canvas/50 px-1.5 py-0.5 rounded self-center">limit_daily</code>
                <span className="text-muted-foreground">Total requests per day</span>
                <span className="text-muted-foreground">Unlimited</span>
              </div>
              <div className="grid grid-cols-3 gap-4 px-4 py-3 text-sm">
                <code className="text-xs bg-canvas/50 px-1.5 py-0.5 rounded self-center">limit_tokens_in_daily</code>
                <span className="text-muted-foreground">Input tokens per day</span>
                <span className="text-muted-foreground">Unlimited</span>
              </div>
              <div className="grid grid-cols-3 gap-4 px-4 py-3 text-sm">
                <code className="text-xs bg-canvas/50 px-1.5 py-0.5 rounded self-center">limit_tokens_out_daily</code>
                <span className="text-muted-foreground">Output tokens per day</span>
                <span className="text-muted-foreground">Unlimited</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Collapsible title="What happens when a limit is exceeded" where="HTTP 429">
          <p className="text-muted-foreground mb-2">When a share key hits its daily limit:</p>
          <CodeBlock>{`HTTP/1.1 429 Too Many Requests
Content-Type: application/json

{
  "error": "rate_limited",
  "message": "Daily request limit exceeded"
}`}</CodeBlock>
          <p className="text-muted-foreground mt-2">The limit resets at midnight UTC. You can also manually reset a key's limits from the API Keys page.</p>
        </Collapsible>

        <Callout>
          <strong>Per-key isolation.</strong> Each share key has its own independent limits. One key hitting its limit doesn't affect other keys.
        </Callout>
      </section>

      {/* ═══════════════════════════════════ */}
      <section>
        <SectionHeading
          num="06"
          title="FAQ"
          description="Common questions about using Throtl."
        />

        <div className="space-y-4">
          {faqItems.map((item) => (
            <Collapsible key={item.q} title={item.q}>
              <p className="text-muted-foreground">{item.a}</p>
            </Collapsible>
          ))}
        </div>
      </section>

      {/* Footer */}
      <div className="pt-6 border-t border-hairline flex items-center justify-between text-xs text-muted-foreground">
        <span className="flex items-center gap-2">
          <BookOpen className="h-3.5 w-3.5" />
          Throtl Usage Guide
        </span>
        <a
          href="https://github.com/ihsanbudiman/throtl"
          target="_blank"
          rel="noopener noreferrer"
          className="hover:text-foreground transition-colors"
        >
          github.com/ihsanbudiman/throtl
        </a>
      </div>
    </div>
  );
}
