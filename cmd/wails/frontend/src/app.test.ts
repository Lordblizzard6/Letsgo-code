import { describe, it, expect, vi, beforeEach } from "vitest";
import * as ChatService from "../bindings/github.com/user/go-claude-code/cmd/wails/services/chatservice";

vi.mock("@wailsio/runtime", () => ({
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

const runtime = await import("@wailsio/runtime");

describe("chat bindings", () => {
  beforeEach(() => {
    vi.mocked(runtime.Call.ByID).mockReset();
  });

  it("ChatService.Send routes text through the bound Go method", async () => {
    vi.mocked(runtime.Call.ByID).mockResolvedValue(undefined);
    await ChatService.Send("hola mundo");
    expect(runtime.Call.ByID).toHaveBeenCalledWith(
      2907149197,
      "hola mundo",
    );
  });

  it("ChatService.Cancel stops the in-flight stream", async () => {
    vi.mocked(runtime.Call.ByID).mockResolvedValue(undefined);
    await ChatService.Cancel();
    expect(runtime.Call.ByID).toHaveBeenCalledWith(658564621);
  });
});