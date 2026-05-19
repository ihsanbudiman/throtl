/// <reference types="vitest/globals" />
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ToastProvider } from "@/hooks/use-toast";
import GenerateKeyDialog from "@/components/GenerateKeyDialog";

// Mock the API module
vi.mock("@/lib/api", () => ({
  api: {
    createKey: vi.fn().mockResolvedValue({ id: "1", name: "Test Key", key: "sk-share-test" }),
  },
}));

// Mock the toast hook
vi.mock("@/hooks/use-toast", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/hooks/use-toast")>();
  return {
    ...actual,
    useToast: () => ({ toast: vi.fn(), toasts: [], dismiss: vi.fn() }),
  };
});

describe("GenerateKeyDialog", () => {
  it("renders token limit labels when opened", async () => {
    const user = userEvent.setup();
    const onSuccess = vi.fn();

    render(
      <ToastProvider>
        <GenerateKeyDialog onSuccess={onSuccess} />
      </ToastProvider>,
    );

    // Click the trigger button to open dialog
    await user.click(screen.getByText("Generate Key"));

    // Wait for dialog content to appear
    expect(screen.getByText("Generate New API Key")).toBeInTheDocument();

    // Token limit labels should be present
    expect(screen.getByText("Daily Token-In Limit")).toBeInTheDocument();
    expect(screen.getByText("Daily Token-Out Limit")).toBeInTheDocument();

    // Window-related labels should NOT be present
    expect(screen.queryByText(/Rolling Window/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/Per \(hours\)/i)).not.toBeInTheDocument();
  });
});
