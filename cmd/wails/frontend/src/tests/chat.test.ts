import { describe, it, expect, vi } from "vitest";
import { render } from "@testing-library/svelte";
import Chat from "../components/chat.svelte";
import { useChat } from "../lib/store.svelte";

vi.mock("@wailsio/runtime", () => ({
  Events: {
    On: vi.fn(),
  },
  Call: {
    ByID: vi.fn(),
  },
  Create: {
    Map: (k: unknown, v: unknown) => (x: unknown) => x,
    Any: Symbol("any"),
    Array: (v: unknown) => (x: unknown) => x,
    Named: (v: unknown) => (x: unknown) => x,
  },
  CancellablePromise: class {},
}));

describe("Chat streaming (US3)", () => {
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
