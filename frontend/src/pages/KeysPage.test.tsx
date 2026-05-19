/// <reference types="vitest/globals" />
import { render, screen } from "@testing-library/react";
import { ToastProvider } from "@/hooks/use-toast";
import KeysPage from "@/pages/KeysPage";

vi.mock("@/lib/api", () => {
  const keys = [
    {
      id: "1",
      name: "Test Key",
      key: "sk-share-test123",
      limit_daily: 100,
      limit_tokens_in_daily: 50000,
      limit_tokens_out_daily: 25000,
      allowed_models: "wafer/gpt-4o,wafer/gpt-4o-mini",
      active: true,
      created_at: "2026-01-01T00:00:00Z",
      last_used_at: "2026-05-01T00:00:00Z",
      rate_limit: {
        daily_count: 42,
        daily_limit: 100,
        daily_tokens_in_count: 12000,
        daily_tokens_in_limit: 50000,
        daily_tokens_out_count: 8000,
        daily_tokens_out_limit: 25000,
        daily_reset: "2026-05-20T00:00:00Z",
      },
    },
    {
      id: "2",
      name: "Unlimited Key",
      key: "sk-share-unlimited",
      limit_daily: 0,
      limit_tokens_in_daily: 0,
      limit_tokens_out_daily: 0,
      allowed_models: "",
      active: true,
      created_at: "2026-01-01T00:00:00Z",
      last_used_at: null,
      rate_limit: {
        daily_count: 0,
        daily_limit: 0,
        daily_tokens_in_count: 0,
        daily_tokens_in_limit: 0,
        daily_tokens_out_count: 0,
        daily_tokens_out_limit: 0,
      },
    },
  ];
  return {
    api: {
      listKeys: vi.fn().mockResolvedValue(keys),
      toggleKey: vi.fn().mockResolvedValue(undefined),
      deleteKey: vi.fn().mockResolvedValue(undefined),
    },
  };
});

vi.mock("@/hooks/use-toast", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/hooks/use-toast")>();
  return {
    ...actual,
    useToast: () => ({ toast: vi.fn(), toasts: [], dismiss: vi.fn() }),
  };
});

describe("KeysPage", () => {
  it("renders token limits and daily requests columns", async () => {
    render(
      <ToastProvider>
        <KeysPage />
      </ToastProvider>,
    );

    expect(await screen.findAllByText("Test Key")).toHaveLength(2);
    expect(screen.getAllByText("Unlimited Key")).toHaveLength(2);

    expect(screen.getByText("Token Limits")).toBeInTheDocument();
    expect(screen.getByText("Daily Requests")).toBeInTheDocument();

    expect(screen.getByText("42/100")).toBeInTheDocument();

    expect(screen.getAllByText("In: 12000/50000")).toHaveLength(2);
    expect(screen.getAllByText("Out: 8000/25000")).toHaveLength(2);

    expect(screen.queryByText(/Rolling Window/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/Per \(hours\)/i)).not.toBeInTheDocument();
  });
});
