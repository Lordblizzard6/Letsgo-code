import { test, expect } from "@playwright/test";
import type { Page, Route } from "@playwright/test";

// The Wails dev server serves the SPA on 127.0.0.1:9245 without a live Go
// backend, so every runtime call goes to `POST /wails/runtime`. We fulfil
// those calls from here with the same contract the bindings expect, so the
// E2E is deterministic and needs no API key or external network (FR-011).
const METHODS: Record<number, (args?: unknown[], state?: State) => unknown> = {
  // SettingsService.GetConfig
  3812644894: () => ({
    APIKey: "",
    AnthropicAPIKey: "",
    OpenAIAPIKey: "",
    GroqAPIKey: "",
    OpenRouterAPIKey: "",
    Model: "",
    BaseURL: "",
    Shell: "",
    Verbose: false,
    Temperature: 0,
    MaxTokens: 0,
    Stream: false,
    AutoApprove: {},
    Statusline: { Show: false, Fields: [] },
    Rail: { Collapsed: false },
    ThemeVariant: "dark",
  }),
  // SettingsService.SaveConfig
  794155347: (args: unknown[] | undefined, state?: State) => ({
    ...(state?.cfg ?? {}),
    APIKey: "sk-test",
    AnthropicAPIKey: "sk-test",
    Model: "claude-sonnet-4-5",
  }),
  // ThemeService.Get
  2757779676: () => "dark",
  // AccountService.GetStatus
  3767476829: () => ({ online: false, busy: false, provider: "", model: "" }),
};

type State = { cfg: Record<string, unknown> };

function mockRuntime(page: Page) {
  const state: State = { cfg: METHODS[3812644894]() as Record<string, unknown> };
  page.route("**/wails/runtime", async (route: Route) => {
    if (route.request().method() !== "POST") {
      await route.fulfill({ status: 404 });
      return;
    }
    let body: { args?: { methodID?: number; args?: unknown[] } } = {};
    try {
      body = JSON.parse(route.request().postData() ?? "{}");
    } catch {
      body = {};
    }
    const method = body.args?.methodID ?? -1;
    const handler = METHODS[method];
    let payload: unknown;
    if (method === 794155347) {
      state.cfg = {
        ...state.cfg,
        APIKey: "sk-test",
        AnthropicAPIKey: "sk-test",
        Model: "claude-sonnet-4-5",
      };
      payload = handler(body.args?.args, state);
    } else if (handler) {
      payload = handler();
    } else {
      await route.fulfill({ status: 500, body: `unknown method ${method}` });
      return;
    }
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(payload),
    });
  });
  return state;
}

test("onboarding: sin key solo se ve el flujo de proveedor, con key se habilita el chat", async ({
  page,
}) => {
  mockRuntime(page);
  await page.goto("/");

  // Gate (T078): config cargada sin credenciales => solo onboarding.
  await expect(
    page.getByRole("heading", { name: "Welcome to LetsGO" }),
  ).toBeVisible();
  await expect(page.locator("nav.rail")).toHaveCount(0);

  // FR-011: clave + modelo => guardar inicia el chat sin reiniciar (FR-016).
  await page.getByLabel("API Key").fill("sk-test");
  await page.getByRole("button", { name: "Save & start chatting" }).click();

  await expect(page.locator("nav.rail")).toBeVisible();
  await expect(
    page.getByPlaceholder("Type your message..."),
  ).toBeVisible();
});