import React, { useState } from "react";

interface DiffViewerProps {
    diff: string;
    filePath?: string;
    maxHeight?: string | number;
}

interface ParsedLine {
    type: "add" | "del" | "hunk" | "meta" | "normal";
    text: string;
    oldNum?: number;
    newNum?: number;
}

export const DiffViewer: React.FC<DiffViewerProps> = ({ diff, filePath, maxHeight }) => {
    const [copied, setCopied] = useState(false);

    if (!diff || !diff.trim()) {
        return (
            <div className="diff-empty-state">
                <span>Sin diferencias que mostrar.</span>
            </div>
        );
    }

    const lines = diff.split("\n");
    let oldLine = 0;
    let newLine = 0;
    const parsedLines: ParsedLine[] = [];

    for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        if (line.startsWith("@@")) {
            // Parse hunk header: @@ -14,7 +14,9 @@
            const match = line.match(/@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@/);
            if (match) {
                oldLine = parseInt(match[1], 10);
                newLine = parseInt(match[2], 10);
            }
            parsedLines.push({ type: "hunk", text: line });
        } else if (line.startsWith("+") && !line.startsWith("+++")) {
            parsedLines.push({
                type: "add",
                text: line,
                newNum: newLine++,
            });
        } else if (line.startsWith("-") && !line.startsWith("---")) {
            parsedLines.push({
                type: "del",
                text: line,
                oldNum: oldLine++,
            });
        } else if (line.startsWith("diff --git") || line.startsWith("index ") || line.startsWith("--- ") || line.startsWith("+++ ")) {
            parsedLines.push({ type: "meta", text: line });
        } else {
            parsedLines.push({
                type: "normal",
                text: line,
                oldNum: oldLine > 0 ? oldLine++ : undefined,
                newNum: newLine > 0 ? newLine++ : undefined,
            });
        }
    }

    const copyDiff = () => {
        navigator.clipboard.writeText(diff);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div className="diff-viewer-wrapper" style={{ maxHeight }}>
            <div className="diff-viewer-bar">
                <span className="diff-viewer-filename" title={filePath}>
                    {filePath || "Diff"}
                </span>
                <button
                    type="button"
                    className="diff-copy-btn"
                    onClick={copyDiff}
                    title="Copiar diff completo"
                >
                    {copied ? "✓ Copiado" : "Copiar"}
                </button>
            </div>
            <div className="diff-viewer-content">
                {parsedLines.map((pl, idx) => (
                    <div key={idx} className={`diff-row ${pl.type}`}>
                        <span className="diff-num old-num">
                            {pl.oldNum != null ? pl.oldNum : ""}
                        </span>
                        <span className="diff-num new-num">
                            {pl.newNum != null ? pl.newNum : ""}
                        </span>
                        <span className="diff-sign">
                            {pl.type === "add" ? "+" : pl.type === "del" ? "-" : pl.type === "hunk" ? "@" : " "}
                        </span>
                        <span className="diff-code-text">
                            {pl.text.startsWith("+") || pl.text.startsWith("-") ? pl.text.substring(1) : pl.text}
                        </span>
                    </div>
                ))}
            </div>
        </div>
    );
};
