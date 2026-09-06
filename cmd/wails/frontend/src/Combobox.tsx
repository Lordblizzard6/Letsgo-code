import {useEffect, useMemo, useRef, useState} from "react";
import {createPortal} from "react-dom";

interface Props {
    value: string;
    onChange: (v: string) => void;
    options: string[];
    placeholder?: string;
    maxSuggestions?: number;
}

// Combobox: input de texto con lista de sugerencias filtradas al escribir.
// Permite escribir libremente (modelo personalizado) y navegar con teclado.
// El menú se renderiza en un portal para no ser recortado por overflow: hidden.
export function Combobox({value, onChange, options, placeholder, maxSuggestions = 120}: Props) {
    const inputRef = useRef<HTMLInputElement>(null);
    const [open, setOpen] = useState(false);
    const [hi, setHi] = useState(0);
    const [pos, setPos] = useState<{left: number; top: number; width: number} | null>(null);

    const filtered = useMemo(() => {
        const q = value.trim().toLowerCase();
        const list = options.filter((o) => o.toLowerCase().includes(q));
        if (q !== "") {
            list.sort((a, b) => {
                const ap = a.toLowerCase().startsWith(q) ? 0 : 1;
                const bp = b.toLowerCase().startsWith(q) ? 0 : 1;
                if (ap !== bp) return ap - bp;
                return a.localeCompare(b);
            });
        }
        return list.slice(0, maxSuggestions);
    }, [value, options, maxSuggestions]);

    useEffect(() => {
        const onDoc = (e: MouseEvent) => {
            const target = e.target as Element;
            if (!inputRef.current?.contains(target) && !target.closest(".combobox-menu")) {
                setOpen(false);
            }
        };
        document.addEventListener("mousedown", onDoc);
        return () => document.removeEventListener("mousedown", onDoc);
    }, []);

    useEffect(() => {
        if (!open) {
            setPos(null);
            return;
        }
        let raf = 0;
        const update = () => {
            const el = inputRef.current;
            if (!el) return;
            const r = el.getBoundingClientRect();
            setPos({left: r.left, top: r.bottom + 4, width: r.width});
            raf = requestAnimationFrame(update);
        };
        raf = requestAnimationFrame(update);
        return () => cancelAnimationFrame(raf);
    }, [open]);

    const select = (m: string) => {
        onChange(m);
        setOpen(false);
        setHi(0);
    };

    const onInputChange = (v: string) => {
        onChange(v);
        setHi(0);
    };

    const onKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === "ArrowDown") {
            e.preventDefault();
            setOpen(true);
            setHi((h) => (filtered.length ? Math.min(h + 1, filtered.length - 1) : 0));
        } else if (e.key === "ArrowUp") {
            e.preventDefault();
            setHi((h) => Math.max(h - 1, 0));
        } else if (e.key === "Enter") {
            if (open && filtered[hi]) {
                e.preventDefault();
                select(filtered[hi]);
            }
        } else if (e.key === "Escape") {
            setOpen(false);
        }
    };

    return (
        <>
            <input
                ref={inputRef}
                type="text"
                className="combobox-input"
                value={value}
                placeholder={placeholder}
                onChange={(e) => onInputChange(e.target.value)}
                onFocus={() => setOpen(true)}
                onKeyDown={onKeyDown}
            />
            {open && filtered.length > 0 && pos && createPortal(
                <div className="combobox-menu" style={{left: pos.left, top: pos.top, width: pos.width}}>
                    {filtered.map((m, i) => (
                        <button
                            type="button"
                            key={m}
                            role="option"
                            aria-selected={i === hi}
                            className={"combobox-option" + (i === hi ? " selected" : "")}
                            onMouseDown={(e) => {
                                e.preventDefault();
                                select(m);
                            }}
                            onMouseEnter={() => setHi(i)}
                        >
                            {m}
                        </button>
                    ))}
                </div>,
                document.body,
            )}
        </>
    );
}
