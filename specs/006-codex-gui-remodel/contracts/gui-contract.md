# Contrato de la GUI Wails — Codex Remodel

**Branch**: `006-codex-gui-remodel` | **Fecha**: 2026-08-14 | **Spec**: [spec.md](../spec.md)

**Propósito**: contrato del remodelado visual **solo frontend** del feature 006.
Las secciones §1..§7 (estructura de navegación, chat, sesiones, paneles, accesos
rápidos, onboarding, retirada de Fyne) y §8 (sistema de diseño y gates UX de la
enmienda US5) se heredan de `specs/005-bubbletea-tui-wails/contracts/gui-contract.md`
y se mantienen vigentes. Este documento añade **§9**, que es la fuente única de los
tokens finales, la dirección de diseño por superficie, el checklist del smoke
visual G10 y las reglas no negociables del remodelado 006.

## 9. Codex Remodel (feature 006)

### 9.1 Tokens finales (una sola fuente)

Viven **enteros** en `cmd/wails/frontend/src/app.css` bajo
`html[data-theme="dark"|"light"]`; los componentes consumen `var(--token)` y nunca
hex literal (G1). Verificación de contraste AA (G2): texto ≥4.5:1, no-texto ≥3:1,
texto sobre acento ≥4.5:1; focus ring ≥3:1.

```css
:root {
  /* Tipografía */
  --font-ui: "Inter", "Segoe UI", system-ui, sans-serif;
  --font-mono: "JetBrains Mono", "Cascadia Mono", Consolas, monospace;

  /* Espaciado (escala 4px — G4) */
  --space-1: 4px;   --space-2: 8px;   --space-3: 12px;
  --space-4: 16px;  --space-5: 20px;  --space-6: 24px;
  --space-8: 32px;  --space-10: 40px; --space-12: 48px;

  /* Radio */
  --radius-sm: 6px;  --radius-md: 8px; --radius-lg: 10px; --radius-xl: 12px;
  --radius-pill: 999px;

  /* Sombras (solo overlays/flotantes; la página no lleva sombra) */
  --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.30);
  --shadow-md: 0 4px 8px -2px rgb(0 0 0 / 0.35);
  --shadow-lg: 0 12px 28px -6px rgb(0 0 0 / 0.45);

  /* Foco y motion */
  --focus-ring: 2px solid var(--accent);
  --focus-offset: 2px;
  --dur-fast: 120ms; --dur-normal: 150ms;          /* ≤150ms salvo progreso stream (G9) */
  --ease: cubic-bezier(0.2, 0, 0, 1);

  /* Capas */
  --z-rail: 10; --z-status: 20; --z-flyout: 40; --z-overlay: 50; --z-palette: 60;
}

html[data-theme="dark"] {                       /* base #0e1116 (heredado D8) */
  --bg: #0e1116;                                /* canvas principal */
  --bg-panel: #14181e;                          /* rail, status line, paneles */
  --bg-elevated: #1a1f26;                       /* code block, inputs hover */
  --bg-input: #1f252d;                          /* textarea, inputs */
  --bg-active: #252c36;                         /* burbuja usuario, row activa */
  --bg-hover: #1c2129;                          /* hover rows */
  --bg-skeleton: #1a1f26;
  --border: #2e3540;                            /* hairline */
  --border-strong: #3d4654;                     /* inputs, composer shell */
  --text: #e8ebf0;                              /* 15.8:1 */
  --text-secondary: #9aa3b0;                    /* 7.4:1 */
  --text-tertiary: #6f7885;                     /* 3.8:1 (solo caps/labels ≥14px) */
  --accent: #8b94f8;                            /* indigo — 6.9:1 */
  --accent-on: #101218;                         /* texto sobre accent */
  --accent-soft: color-mix(in srgb, var(--accent) 14%, transparent);
  --ok: #4ade80;  --ok-on: #0e1a12;
  --warn: #fbbf24; --warn-on: #1a1404;
  --err: #f87171; --err-on: #1c0d0d;  --err-soft: color-mix(in srgb, var(--err) 14%, transparent);
  --overlay: rgba(0, 0, 0, 0.18);               /* drawers: chat visible (UX-05) */
  --overlay-strong: rgba(0, 0, 0, 0.42);        /* paleta Ctrl+K */
}

html[data-theme="light"] {                      /* base #f6f7f9 (heredado D8) */
  --bg: #f6f7f9;
  --bg-panel: #ffffff;
  --bg-elevated: #eef0f3;
  --bg-input: #e9ebef;
  --bg-active: #e4e7ec;
  --bg-hover: #eef0f3;
  --bg-skeleton: #e4e7ec;
  --border: #dfe3e8;
  --border-strong: #c4cbd4;
  --text: #1c2026;                              /* 16.4:1 */
  --text-secondary: #5c6570;                    /* 5.9:1 */
  --text-tertiary: #8b939e;                     /* 3.2:1 (solo caps/labels ≥14px) */
  --accent: #4f46e5;                            /* indigo — 6.3:1 */
  --accent-on: #ffffff;
  --accent-soft: color-mix(in srgb, var(--accent) 12%, transparent);
  --ok: #15803d;  --ok-on: #ffffff;
  --warn: #b45309; --warn-on: #ffffff;
  --err: #dc2626; --err-on: #ffffff; --err-soft: color-mix(in srgb, var(--err) 10%, transparent);
  --overlay: rgba(15, 18, 24, 0.10);            /* drawers */
  --overlay-strong: rgba(15, 18, 24, 0.30);     /* paleta */
}

/* Tipografía (base) */
body { font-family: var(--font-ui); font-size: 14px; line-height: 1.6; }
code, pre { font-family: var(--font-mono); font-size: 13px; }
.caps { font-size: 11px; font-weight: 600; letter-spacing: 0.08em; text-transform: uppercase; }
.tabular { font-variant-numeric: tabular-nums; }
```

Escala de texto: 11 (caps) / 13 (mono, meta) / 14 (base chat) / 15 (paleta input)
/ 16 (títulos de panel) / 18 (h2) / 20 (h1, kpi value) / 24 (brand). Pesos 400/500/600.

### 9.2 Dirección de diseño por superficie

- **Chrome**: casi monocromo. Rail + status line + paneles con `--bg-panel` y
  hairline `--border`; superficies de contenido por luminancia (`--bg-elevated`,
  `--bg-active`), no por borde de tarjeta.
- **Transcript (chat-canvas)**: columna ≤760px centered; **usuario = burbuja
  derecha** (`--bg-active`, radius-lg, ≤70% ancho); **asistente = bloque full-width
  sin borde** con header-per-message (mark + "Assistant" + timestamp + modelo,
  13px `--text-tertiary`) y cursor de escritura `▍` + línea de progreso 2px durante
  el stream. Sin fila "Streaming…"; sin `.msg` con borde.
- **CodeBlock**: objeto elevado (`--bg-elevated`, `--border`, radius-md) con header
  (language caps + botón Copy on hover/focus, estado "Copied" ≤1.5s), resaltado de
  sintaxis por tema, overflow horizontal dentro del bloque. Todo `<pre>` vía este
  componente (G8).
- **Rail**: 48px (icon-only) colapsable a 56px con tooltip `Label Alt+n`; SVG
  stroke 16px monocromo; marcador activo único (icono `--accent` + barra 3px +
  `--accent-soft`); secciones WORK/TOOLS/SYSTEM en caps; footer del rail: collapse
  + chip de cuenta/theme (absorbe las pills flotantes).
- **Status line** (28px, `--bg-panel`, borde superior hairline, caps 11px):
  proveedor · modelo · uso hoy · dot busy (CSS circle, `aria-live`). El estado
  "Ready/Streaming" vive aquí, no en el transcript.
- **Composer (command shell)**: docked, radius-lg, `--border-strong`, `--shadow-sm`;
  textarea siempre activa; Enter=send / Shift+Enter=nueva línea / Enter=stop en
  stream; izquierda: chip de modelo; derecha: Send (accent fill `--accent-on`) ⇄
  Stop (err-soft outline) en la misma ranura, min 44px; hint row 11px `Ctrl+K
  commands · / plan · @ files · Alt+1-9 surfaces`; affordance "Streaming — Enter
  stops" en stream. Placeholder "Type your message..." (string bloqueado).
- **Tool activity**: inline en el turno; cards colapsables (header = icono SVG +
  nombre + estado live + chevron; body mono truncado 200ch + expand); success
  `--ok` / error `--err` + Retry; running edge accent 1px + pulse ≤150ms (off bajo
  reduced-motion); 5+ tools → grupo compacto con contador; 1 row por call (dedupe
  en `store.svelte.ts`). Nunca un log que empuje al composer.
- **Sesiones**: agrupadas por fecha (Hoy/Ayer/Older) con caps; rows hover
  `--bg-hover`, activo `--accent-soft`; dot busy; empty "No hay conversaciones
  todavía." + CTA; "Refrescando…" preservando contenido.
- **Overlays (Settings/Usage/Help)**: drawers laterales con backdrop `--overlay`
  (≤0.20) que **no oscurecen** el chat; Esc/backdrop cierran y restauran foco al
  composer; `--shadow-lg`, radius-xl.
- **Paleta (Ctrl+K)**: 640px, radius-lg, `--shadow-lg`, `--overlay-strong`;
  agrupada (Ir a · Acciones · Tema); iconos SVG; fila activa `--accent-soft` +
  ring; highlight substring accent; footer hints; focus trap; Esc cierra → foco al
  composer.
- **Onboarding (ProviderSetup)**: split-screen brand 45% / form 55%; strings de
  test intactos ("Welcome to LetsGO", "Save & start chatting", "API Key").
- **Estados**: primitivas `empty-state` / `skeleton` (shimmer ≤150ms) /
  `error-banner` (icono + orientación + Retry que reproduce el último turno) en las
  7+ superficies; chat con streaming/error/empty/idle.

### 9.3 Espaciado / radio / sombra / foco / motion

- **Espaciado**: solo escala 4px (4/8/12/16/20/24/32/40/48) — G4; padding de panel
  16-24px; gap 8-12px; chat ≤760px; targets interactivos ≥44px.
- **Radio**: sm 6 (inputs/botones) · md 8 (cards de tool, sesiones) · lg 10
  (burbuja, composer, kpi) · xl 12 (drawers, paleta) · pill (chips/dot).
- **Sombras**: `--shadow-sm` composer/chips; `--shadow-md` flyouts; `--shadow-lg`
  overlays/paleta. La página y el transcript **no** llevan sombra.
- **Foco**: `:focus-visible { outline: var(--focus-ring); outline-offset: 2px }`
  en todo interactivo, ambos temas (G3); focus trap solo en overlays/paleta
  abiertos; restore de foco al composer.
- **Motion**: transiciones ≤150ms (`--dur-fast` 120ms, `--dur-normal` 150ms);
  excepción: progreso de stream. `prefers-reduced-motion: reduce` → sin pulse/
  shimmer/cursor blink/deslizamientos (G9).

### 9.4 Checklist smoke visual G10 (10 puntos, comparable contra captura Codex)

Captura Playwright (chat+code+tool, palette, onboarding, empty, dark+light) y
comparación a 8ft contra la referencia Codex:

1. **Composición general**: rail fino + status line inferior + transcript centered
   + composer docked; sin tarjetas flotantes ni pills en el chrome.
2. **Cero emoji en chrome**: rail, botones, headers, status line, paleta (G5).
3. **Transcript-canvas**: mensaje de usuario = burbuja derecha; mensaje de
   asistente = bloque full-width **sin borde**; separación por ritmo tipográfico.
4. **CodeBlock**: todo bloque de código es objeto elevado con header (lenguaje +
   Copy) y resaltado; sin `<pre>` desnudo (G8).
5. **Rail**: iconos SVG lineales monocromos; marcador activo único con acento;
   colapso 48↔56 con tooltip `Label Alt+n`.
6. **Composer shell**: textarea en shell con borde hairline, chip de modelo,
   Send⇄Stop en la misma ranura y hint row 11px; placeholder intacto.
7. **Status line**: proveedor · modelo · uso hoy · dot busy visible; sin
   "Streaming…/Ready" en el transcript.
8. **Paleta Ctrl+K**: 640px agrupada, fila activa con acento, footer hints, sin
   emoji.
9. **Overlays/drawers**: Settings/Usage/Help no oscurecen el chat; foco visible;
   onboarding split-screen con los strings bloqueados.
10. **Dark+light AA**: ambos temas con contraste AA (texto ≥4.5:1, no-texto ≥3:1,
    focus ≥3:1); escalas 100/125/150% sin roturas.

### 9.5 Reglas que no se negocian

- **Token purity (G1)**: cero hex literal en `.svelte`/`.ts`; todo color/tipografía/
  espaciado/radio/sombra vía `var(--token)` de `app.css`.
- **Cero emoji en el chrome (G5)**: emoji solo permitido en el contenido markdown
  del usuario/asistente; iconografía = SVG stroke monocromo.
- **Dark-first con acento único**: dark es la navegación principal; el único acento
  es indigo `#8b94f8`/`#4f46e5` (intercambiable solo en tokens, nunca por componente).
- **Sin regresión (UX-12..UX-14)**: tests vitest pre-existentes pasan sin cambios;
  `e2e/onboarding.spec.ts` intacto; `git diff` fuera de `specs/` solo toca
  `cmd/wails/frontend/**`; cero cambios en bindings Go, servicios bound, contrato
  de eventos, rutas hash, atajos Alt+1..9 y motor (FR-012).
- **Strings bloqueados**: "Welcome to LetsGO", "Save & start chatting", "API Key",
  placeholder del composer "Type your message..." se conservan salvo cambio
  deliberado documentado en el mismo commit.
- **WCAG AA + foco + motion**: contraste AA en ambos temas; focus visible en todo
  interactivo; transiciones ≤150ms; `prefers-reduced-motion` honrado (G2/G3/G9).
- **Una sola fuente**: este §9 es la fuente única de tokens y dirección de diseño;
  si difiere de 005 §8, manda este documento.

### 9.6 Gates (referencia)

Gates de restyle **G1–G10** y de UX **UX-01..UX-14** idénticos a 005 §8.2/§8.3
(verificables por grep, Vitest y Playwright), aplicados a los workstreams R1..R11
del [plan.md](../plan.md). Ejecución de la GUI: `letsgo.exe gui` (sin cambios).
