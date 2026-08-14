import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/svelte";
import Chat from "../components/chat.svelte";
import { useChat } from "../lib/store.svelte";
import type { ChatMessage } from "../lib/store.svelte";

const { byID } = vi.hoisted(() => ({ byID: vi.fn() }));

vi.mock("@wailsio/runtime", () => ({
  Events: {
    On: vi.fn(),
  },
  Call: {
    ByID: byID,
  },
  Create: {
    Map: (k: unknown, v: unknown) => (x: unknown) => x,
    Any: Symbol("any"),
    Array: (v: unknown) => (x: unknown) => x,
    Named: (v: unknown) => (x: unknown) => x,
  },
  CancellablePromise: class {},
}));

beforeEach(() => {
  byID.mockReset();
  byID.mockResolvedValue(undefined);
});

describe("Chat streaming (US3)", () => {
  afterEach(() => {
    cleanup();
    useChat.reset();
  });

  it("acumula los deltas del stream y renderiza markdown sanitizado", () => {
    useChat.reset();
    useChat.startStream();
    useChat.appendDelta("# Hola\n\n");
    useChat.appendDelta("**mundo** ");
    useChat.appendDelta("<script>alert(1)</script>");
    useChat.endStream();

    render(Chat);

    const html = document.body.innerHTML;
    expect(html).toContain("<h1>Hola</h1>");
    expect(html).toContain("<strong>mundo</strong>");
    expect(html).not.toContain("<script");
  });
});

describe("Retry reproduce el último turno (US5 U5-A I-2)", () => {
  beforeEach(() => {
    useChat.reset();
    cleanup();
  });

  it("Reintentar reenvía Start + Send con el último mensaje del usuario", async () => {
    const prev: ChatMessage[] = [
      { id: "u1", role: "user", content: "Explica el parser" },
      { id: "a1", role: "assistant", content: "Respuesta previa" },
    ];
    useChat.setMessages(prev);
    useChat.setError("connection refused on api.anthropic.com");

    render(Chat);
    expect(screen.getByText("Respuesta previa")).toBeTruthy();

    await fireEvent.click(screen.getByRole("button", { name: "Retry" }));

    await waitFor(() => {
      const calls = byID.mock.calls.filter(
        (c) => c[0] === 4150834633 || c[0] === 2907149197,
      );
      expect(calls.some((c) => c[0] === 4150834633)).toBe(true);
      expect(
        calls.some((c) => c[0] === 2907149197 && c[1] === "Explica el parser"),
      ).toBe(true);
    });
    expect(byID).not.toHaveBeenCalledWith(2907149197, "Respuesta previa");
  });

  it("sin último turno de usuario no dispara el reintento", async () => {
    useChat.setMessages([
      { id: "a1", role: "assistant", content: "Solo asistente" },
    ]);
    useChat.setError("boom");

    render(Chat);
    await fireEvent.click(screen.getByRole("button", { name: "Retry" }));

    await waitFor(() => {
      const sends = byID.mock.calls.filter((c) => c[0] === 2907149197);
      expect(sends).toHaveLength(0);
    });
  });
});
