import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/svelte";
import { tick } from "svelte";
import Composer from "../components/composer.svelte";
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

describe("Composer Enviar/Detener (US3)", () => {
  it("alterna el botón Send↔Stop y mantiene el input activo durante el stream", async () => {
    useChat.reset();

    const { container } = render(Composer);
    const textarea = container.querySelector("textarea") as HTMLTextAreaElement;
    expect(textarea.disabled).toBe(false);

    expect(screen.getByRole("button", { name: "Send" })).toBeTruthy();

    useChat.setStreaming(true);
    await tick();

    expect(screen.getByRole("button", { name: "Stop" })).toBeTruthy();
    expect(textarea.disabled).toBe(false);

    useChat.setStreaming(false);
    await tick();

    expect(screen.getByRole("button", { name: "Send" })).toBeTruthy();
  });
});
