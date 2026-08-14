import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/svelte";
import App from "../App.svelte";
import type { Config } from "../../bindings/github.com/user/go-claude-code/internal/config/models.js";

const noKeyConfig = {
  APIKey: "",
  AnthropicAPIKey: "",
  OpenAIAPIKey: "",
  GroqAPIKey: "",
  OpenRouterAPIKey: "",
  Model: "",
  ThemeVariant: "dark",
  AutoApprove: {},
} as Config;

const withKeyConfig = {
  ...noKeyConfig,
  APIKey: "sk-test",
  AnthropicAPIKey: "sk-test",
  Model: "claude-sonnet-4-5",
} as Config;

// Config is a class: the bindings pass through Create.Map/Any. Provide the
// field map via the runtime mock so Config.createFrom returns plain data.
let values = { cfg: noKeyConfig };

vi.mock("@wailsio/runtime", () => ({
  Events: { On: vi.fn() },
  Call: {
    ByID: vi.fn((method: number) => {
      switch (method) {
        case 3812644894: // SettingsService.GetConfig
          return Promise.resolve(values.cfg);
        case 794155347: // SettingsService.SaveConfig
          values.cfg = withKeyConfig;
          return Promise.resolve(withKeyConfig);
        case 2757779676: // ThemeService.Get
          return Promise.resolve("dark");
        case 3767476829: // AccountService.GetStatus
          return Promise.resolve({ online: false, busy: false, provider: "", model: "" });
        default:
          return Promise.resolve(undefined);
      }
    }),
  },
  Create: {
    Map: (k: unknown, v: unknown) => (x: unknown) => x,
    Any: Symbol("any"),
    Array: (v: unknown) => (x: unknown) => x,
    Named: (v: unknown) => (x: unknown) => x,
  },
  CancellablePromise: class {},
}));

describe("onboarding (US4)", () => {
  it("con GetConfig sin key muestra solo el flujo de proveedor/key y el chat se habilita al guardar", async () => {
    const { container } = render(App);

    // Gate: la única superficie visible es el onboarding.
    expect(await screen.findByText("Welcome to LetsGO")).toBeTruthy();
    expect(container.querySelector("nav.rail")).toBeNull();
    expect(screen.queryByPlaceholderText("Type your message...")).toBeNull();

    // Flujo key → modelo → chat (FR-011, gui-contract §6).
    await fireEvent.input(screen.getByLabelText("API Key"), {
      target: { value: "sk-test" },
    });
    await fireEvent.submit(screen.getByRole("button", { name: /Save & start chatting/ }));

    // Tras SaveConfig con key válida se habilita el chat.
    await waitFor(() => {
      expect(container.querySelector("nav.rail")).toBeTruthy();
    });
    expect(await screen.findByPlaceholderText("Type your message...")).toBeTruthy();
  });
});