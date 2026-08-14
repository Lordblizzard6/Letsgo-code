import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/svelte";
import SessionsPanel from "../components/sessions-panel.svelte";
import App from "../App.svelte";
import { useChat, useSession } from "../lib/store.svelte";
import type { Message } from "../../bindings/github.com/user/go-claude-code/internal/db/models.js";
import type { Session } from "../../bindings/github.com/user/go-claude-code/internal/db/models.js";

const sessions: Session[] = [
  { id: "s1", name: "Proyecto", project_path: "/tmp", is_active: false, created_at: "", updated_at: "" },
] as Session[];

const messages: Message[] = [
  { id: 1, role: "user", content: "Hola LetsGO" },
  { id: 2, role: "assistant", content: "Aquí tienes la respuesta" },
] as Message[];

vi.mock("@wailsio/runtime", () => ({
  Events: {
    On: vi.fn(),
  },
  Call: {
    ByID: vi.fn((method: number) => {
      switch (method) {
        case 1532743028: // SessionsService.List
          return Promise.resolve(sessions);
        case 2574219276: // SessionsService.Open
          return Promise.resolve(messages);
        case 3812644894: // SettingsService.GetConfig (App gate)
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

const withKeyConfig = {
  APIKey: "sk-test",
  AnthropicAPIKey: "sk-test",
  OpenAIAPIKey: "",
  GroqAPIKey: "",
  OpenRouterAPIKey: "",
  Model: "claude-sonnet-4-5",
  ThemeVariant: "dark",
  AutoApprove: {},
};

describe("Abrir sesión entra al chat (US5 U5-A I-4)", () => {
  beforeEach(() => {
    useChat.reset();
    useSession.setActive(null);
    window.location.hash = "#sessions";
  });

  it("carga el historial y navega a chat (hash vacío) al abrir una sesión", async () => {
    render(SessionsPanel);

    await screen.findByText("Proyecto");
    await fireEvent.click(screen.getByText("Proyecto"));

    await waitFor(() => {
      expect(window.location.hash).toBe("");
    });
    expect(useSession.activeId()).toBe("s1");
    expect(useChat.messages()).toHaveLength(2);
    expect(useChat.messages()[0].content).toBe("Hola LetsGO");
  });
});

describe("Esc cierra Settings/Usage/Help (US5 U5-A I-5)", () => {
  beforeEach(() => {
    window.location.hash = "";
  });

  it("Esc desde settings vuelve al chat", async () => {
    render(App);
    await waitFor(() => {
      expect(document.querySelector("nav.rail")).toBeTruthy();
    });

    window.location.hash = "#settings";
    await screen.findByRole("heading", { name: "Settings" });

    await fireEvent.keyDown(window, { key: "Escape" });
    await waitFor(() => {
      expect(window.location.hash).toBe("");
    });
  });
});