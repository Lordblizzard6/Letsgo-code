import { describe, it, expect, vi } from "vitest";
import { render, fireEvent } from "@testing-library/svelte";
import { tick } from "svelte";
import Palette, { type PaletteItem } from "../components/palette.svelte";

const items: PaletteItem[] = [
  { id: "chat", label: "Chat", hint: "Alt+1" },
  { id: "git", label: "Git", hint: "Alt+3" },
  { id: "tasks", label: "Tasks", hint: "Alt+4" },
  { id: "settings", label: "Settings", hint: "Alt+7" },
];

describe("Paleta teclado (US5 U5-A I-3)", () => {
  it("navega con ↓/↑ y selecciona el item resaltado con Enter", async () => {
    const onSelect = vi.fn();
    const onClose = vi.fn();
    const { container } = render(Palette, {
      props: { items, onSelect, onClose },
    });
    const input = container.querySelector("input") as HTMLInputElement;
    await tick();
    input.focus();

    await fireEvent.keyDown(input, { key: "ArrowDown" });
    await fireEvent.keyDown(input, { key: "ArrowDown" });
    await fireEvent.keyDown(input, { key: "Enter" });
    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(onSelect).toHaveBeenCalledWith(items[2]);

    await fireEvent.keyDown(input, { key: "ArrowUp" });
    await fireEvent.keyDown(input, { key: "Enter" });
    expect(onSelect).toHaveBeenCalledTimes(2);
    expect(onSelect).toHaveBeenLastCalledWith(items[1]);
  });

  it("recorre en carrusel al llegar al final de la lista", async () => {
    const onSelect = vi.fn();
    const onClose = vi.fn();
    const { container } = render(Palette, {
      props: { items, onSelect, onClose },
    });
    const input = container.querySelector("input") as HTMLInputElement;
    await tick();
    input.focus();

    await fireEvent.keyDown(input, { key: "ArrowDown" });
    await fireEvent.keyDown(input, { key: "ArrowDown" });
    await fireEvent.keyDown(input, { key: "ArrowDown" });
    await fireEvent.keyDown(input, { key: "ArrowDown" });
    await fireEvent.keyDown(input, { key: "Enter" });
    expect(onSelect).toHaveBeenLastCalledWith(items[0]);
  });

  it("Esc cierra la paleta", async () => {
    const onSelect = vi.fn();
    const onClose = vi.fn();
    const { container } = render(Palette, {
      props: { items, onSelect, onClose },
    });
    const input = container.querySelector("input") as HTMLInputElement;
    await tick();
    await fireEvent.keyDown(input, { key: "Escape" });
    expect(onClose).toHaveBeenCalledTimes(1);
  });
});