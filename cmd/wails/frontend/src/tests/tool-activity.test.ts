import { describe, it, expect, vi, beforeEach } from "vitest";
import { render } from "@testing-library/svelte";
import ToolActivity from "../components/tool-activity.svelte";
import { useTools } from "../lib/store.svelte";

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

describe("Tool dedupe 1 row/call (US5 U5-A I-6)", () => {
  beforeEach(() => {
    useTools.clearActivity();
  });

  it("tool:start + tool:end con el mismo callId -> 1 sola fila actualizada", () => {
    useTools.upsertActivity({ callId: "call-1", name: "bash", success: true });
    useTools.upsertActivity({
      callId: "call-1",
      name: "bash",
      success: false,
      error: "exit 1",
    });

    const activity = useTools.toolActivity();
    expect(activity.length).toBe(1);
    expect(activity[0].name).toBe("bash");
    expect(activity[0].success).toBe(false);

    const { container } = render(ToolActivity);
    expect(container.querySelectorAll(".tool-row").length).toBe(1);
    expect(container.textContent).toContain("exit 1");
  });

  it("callIds distintos generan filas separadas", () => {
    useTools.upsertActivity({ callId: "a", name: "bash", success: true });
    useTools.upsertActivity({ callId: "b", name: "grep", success: true });

    expect(useTools.toolActivity().length).toBe(2);

    const { container } = render(ToolActivity);
    expect(container.querySelectorAll(".tool-row").length).toBe(2);
  });
});