import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/svelte";
import Chat from "../components/chat.svelte";
import { useChat } from "../lib/store.svelte";

vi.mock("@wailsio/runtime", () => ({
  Events: {
    On: vi.fn(),
  },
  Call: {
    ByID: vi.fn().mockResolvedValue(undefined),
  },
  Create: {
    Map: (k: unknown, v: unknown) => (x: unknown) => x,
    Any: Symbol("any"),
    Array: (v: unknown) => (x: unknown) => x,
    Named: (v: unknown) => (x: unknown) => x,
  },
  CancellablePromise: class {},
}));

describe("Errores de red/API (US4 / FR-013)", () => {
  beforeEach(() => {
    useChat.reset();
  });

  it("muestra el banner accionable ante stream:error sin perder el historial", () => {
    useChat.startStream();
    useChat.appendDelta("Hola, esta es una respuesta previa.");
    useChat.endStream();
    useChat.setError("connection refused on api.anthropic.com");

    render(Chat);

    const banner = screen.getByText("Error:");
    expect(banner.parentElement!.textContent).toContain(
      "connection refused on api.anthropic.com",
    );
    expect(
      screen.getByText("Check your API key, connection and model, then retry."),
    ).toBeTruthy();
    expect(screen.getByRole("button", { name: "Retry" })).toBeTruthy();

    // El historial sigue intacto junto al banner.
    expect(screen.getByText("Hola, esta es una respuesta previa.")).toBeTruthy();
    expect(useChat.messages().length).toBe(1);
  });
});