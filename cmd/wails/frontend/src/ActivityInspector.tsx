import React, { useState } from "react";
import { ActivityTab, CommandLogItem, GitDiffFile, GitDiffSummary, BackgroundTask } from "./types";
import { DiffViewer } from "./DiffViewer";

interface ActivityInspectorProps {
    open: boolean;
    activeTab: ActivityTab;
    onTabChange: (tab: ActivityTab) => void;
    maximized: boolean;
    onToggleMaximize: () => void;
    onClose: () => void;
    gitSummary: GitDiffSummary | null;
    selectedFile: string | null;
    fileDiff: string | null;
    gitLoading: boolean;
    onSelectFile: (filePath: string) => void;
    onStageFile: (filePath: string) => void;
    onUnstageFile: (filePath: string) => void;
    onStageAll: () => void;
    onUnstageAll: () => void;
    onCommit: (msg: string) => Promise<void>;
    onRefreshGit: () => void;
    commandLogs: CommandLogItem[];
    backgroundTasks: BackgroundTask[];
    skillsUsed: string[];
    lang: string;
}

export const ActivityInspector: React.FC<ActivityInspectorProps> = ({
    open,
    activeTab,
    onTabChange,
    maximized,
    onToggleMaximize,
    onClose,
    gitSummary,
    selectedFile,
    fileDiff,
    gitLoading,
    onSelectFile,
    onStageFile,
    onUnstageFile,
    onStageAll,
    onUnstageAll,
    onCommit,
    onRefreshGit,
    commandLogs,
    backgroundTasks,
    skillsUsed,
    lang,
}) => {
    const [commitMsg, setCommitMsg] = useState("");
    const [committing, setCommitting] = useState(false);
    const [expandedCommandId, setExpandedCommandId] = useState<string | null>(null);

    if (!open) return null;

    const files = gitSummary?.files || [];
    const unstagedFiles = files.filter((f) => !f.staged);
    const stagedFiles = files.filter((f) => f.staged);

    const handleCommit = async () => {
        if (!commitMsg.trim()) return;
        setCommitting(true);
        try {
            await onCommit(commitMsg.trim());
            setCommitMsg("");
        } finally {
            setCommitting(false);
        }
    };

    return (
        <aside
            id="activity-inspector"
            className={`activity-inspector ${maximized ? "maximized" : ""}`}
            aria-label="Activity Inspector"
        >
            {/* Inspector Header */}
            <div className="inspector-header">
                <div className="inspector-tabs" role="tablist">
                    <button
                        type="button"
                        role="tab"
                        aria-selected={activeTab === "overview"}
                        className={`inspector-tab-btn ${activeTab === "overview" ? "active" : ""}`}
                        onClick={() => onTabChange("overview")}
                        title="Overview: sistema, cambios y tareas"
                    >
                        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <rect width="7" height="9" x="3" y="3" rx="1"/>
                            <rect width="7" height="5" x="14" y="3" rx="1"/>
                            <rect width="7" height="9" x="14" y="12" rx="1"/>
                            <rect width="7" height="5" x="3" y="16" rx="1"/>
                        </svg>
                        <span>Overview</span>
                    </button>

                    <button
                        type="button"
                        role="tab"
                        aria-selected={activeTab === "git"}
                        className={`inspector-tab-btn ${activeTab === "git" ? "active" : ""}`}
                        onClick={() => onTabChange("git")}
                        title="Git Review: staged, unstaged y diff"
                    >
                        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <line x1="6" x2="6" y1="3" y2="15"/>
                            <circle cx="18" cy="6" r="3"/>
                            <circle cx="6" cy="18" r="3"/>
                            <path d="M18 9a9 9 0 0 1-9 9"/>
                        </svg>
                        <span>Git Review</span>
                        {files.length > 0 && <span className="tab-count-badge">{files.length}</span>}
                    </button>

                    <button
                        type="button"
                        role="tab"
                        aria-selected={activeTab === "commands"}
                        className={`inspector-tab-btn ${activeTab === "commands" ? "active" : ""}`}
                        onClick={() => onTabChange("commands")}
                        title="Comandos y llamadas a herramientas"
                    >
                        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <polyline points="4 17 10 11 4 5"/>
                            <line x1="12" x2="20" y1="19" y2="19"/>
                        </svg>
                        <span>Comandos</span>
                        {commandLogs.length > 0 && <span className="tab-count-badge">{commandLogs.length}</span>}
                    </button>
                </div>

                {/* Center File Title */}
                <div className="inspector-center-title" title={activeTab === "git" && selectedFile ? selectedFile : (gitSummary?.branch ? `Rama: ${gitSummary.branch}` : "")}>
                    {activeTab === "git" && selectedFile ? (
                        <span className="file-chip">
                            📄 {selectedFile.split(/[/\\]/).pop() || selectedFile}
                        </span>
                    ) : (
                        <span className="context-chip">
                            {gitSummary?.branch ? `🌿 ${gitSummary.branch}` : ""}
                        </span>
                    )}
                </div>

                {/* Right Controls */}
                <div className="inspector-actions">
                    <button
                        type="button"
                        className="inspector-icon-btn"
                        onClick={onRefreshGit}
                        title="Actualizar estado Git"
                    >
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/>
                            <path d="M3 3v5h5"/>
                            <path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16"/>
                            <path d="M16 21h5v-5"/>
                        </svg>
                    </button>

                    <button
                        type="button"
                        className={`inspector-icon-btn ${maximized ? "active" : ""}`}
                        onClick={onToggleMaximize}
                        title={maximized ? "Restaurar tamaño (🗗)" : "Maximizar panel (🗖)"}
                    >
                        {maximized ? (
                            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                <polyline points="4 14 10 14 10 20"/>
                                <polyline points="20 10 14 10 14 4"/>
                                <line x1="14" y1="10" x2="21" y2="3"/>
                                <line x1="3" y1="21" x2="10" y2="14"/>
                            </svg>
                        ) : (
                            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                <polyline points="15 3 21 3 21 9"/>
                                <polyline points="9 21 3 21 3 15"/>
                                <line x1="21" y1="3" x2="14" y2="10"/>
                                <line x1="3" y1="21" x2="10" y2="14"/>
                            </svg>
                        )}
                    </button>

                    <button
                        type="button"
                        className="inspector-icon-btn close-btn"
                        onClick={onClose}
                        title="Cerrar panel (Ctrl+B)"
                    >
                        ✕
                    </button>
                </div>
            </div>

            {/* Inspector Body */}
            <div className="inspector-body">
                {/* ─── TAB 1: OVERVIEW ────────────────────────────────────────── */}
                {activeTab === "overview" && (
                    <div className="overview-container">
                        {/* Summary Metrics Banner */}
                        <div className="overview-stats-grid">
                            <div className="overview-stat-card uncommitted">
                                <span className="stat-label">Uncommitted</span>
                                <span className="stat-value">{gitSummary?.uncommitted_count ?? files.length}</span>
                                <span className="stat-sub">archivos modificados</span>
                            </div>
                            <div className="overview-stat-card committed">
                                <span className="stat-label">Committed</span>
                                <span className="stat-value">{gitSummary?.committed_count ?? 0}</span>
                                <span className="stat-sub">commits en branch</span>
                            </div>
                        </div>

                        {/* Changed Files Section */}
                        <div className="overview-section">
                            <div className="overview-section-header">
                                <span className="section-title">Archivos cambiados</span>
                                <span className="section-badge">{files.length}</span>
                            </div>
                            <div className="overview-file-list">
                                {!gitSummary?.is_repo ? (
                                    <div className="panel-empty-hint">No se detectó repositorio Git en este directorio.</div>
                                ) : files.length === 0 ? (
                                    <div className="panel-empty-hint">Árbol de trabajo limpio (sin cambios).</div>
                                ) : (
                                    files.map((f) => (
                                        <button
                                            type="button"
                                            key={f.path}
                                            className={`overview-file-row ${selectedFile === f.path ? "active" : ""}`}
                                            onClick={() => {
                                                onSelectFile(f.path);
                                                onTabChange("git");
                                            }}
                                        >
                                            <div className="file-info-left">
                                                <span className={`file-status-tag ${f.status}`}>
                                                    {f.status}
                                                </span>
                                                <span className="file-basename">{f.name || f.path.split("/").pop()}</span>
                                                {f.dir && <span className="file-dirname">{f.dir}/</span>}
                                            </div>
                                            <div className="file-metrics-right">
                                                {f.additions > 0 && (
                                                    <span className="metric-add">+{f.additions}</span>
                                                )}
                                                {f.deletions > 0 && (
                                                    <span className="metric-del">-{f.deletions}</span>
                                                )}
                                                {f.staged && <span className="staged-dot" title="Staged en git" />}
                                            </div>
                                        </button>
                                    ))
                                )}
                            </div>
                        </div>

                        {/* Background Tasks Section */}
                        <div className="overview-section">
                            <div className="overview-section-header">
                                <span className="section-title">Tareas en segundo plano</span>
                                <span className="section-badge">{backgroundTasks.length}</span>
                            </div>
                            <div className="overview-tasks-list">
                                {backgroundTasks.length === 0 ? (
                                    <div className="panel-empty-hint">No hay tareas activas en segundo plano.</div>
                                ) : (
                                    backgroundTasks.map((t) => (
                                        <div key={t.id} className="task-row">
                                            <span className={`task-status-dot ${t.status}`} />
                                            <span className="task-name">{t.title}</span>
                                            <span className="task-id">#{t.id.substring(0, 6)}</span>
                                        </div>
                                    ))
                                )}
                            </div>
                        </div>

                        {/* Skills Used Section */}
                        <div className="overview-section">
                            <div className="overview-section-header">
                                <span className="section-title">Skills activas / usadas</span>
                                <span className="section-badge">{skillsUsed.length}</span>
                            </div>
                            <div className="overview-skills-list">
                                {skillsUsed.length === 0 ? (
                                    <div className="panel-empty-hint">Ninguna skill invocada en esta sesión.</div>
                                ) : (
                                    <div className="skills-tags-wrap">
                                        {skillsUsed.map((sk) => (
                                            <span key={sk} className="skill-pill">
                                                ⚡ {sk}
                                            </span>
                                        ))}
                                    </div>
                                )}
                            </div>
                        </div>

                        {/* Terminals Section */}
                        <div className="overview-section">
                            <div className="overview-section-header">
                                <span className="section-title">Terminales y ejecuciones</span>
                            </div>
                            <div className="overview-terminal-box">
                                <div className="term-info">
                                    <span className="term-dot live" />
                                    <span>Sesión de shell integrada activa</span>
                                </div>
                            </div>
                        </div>
                    </div>
                )}

                {/* ─── TAB 2: GIT REVIEW ──────────────────────────────────────── */}
                {activeTab === "git" && (
                    <div className="git-review-container">
                        <div className="git-review-sidebar">
                            {/* Unstaged section */}
                            <div className="git-subgroup">
                                <div className="subgroup-head">
                                    <span>Cambios sin preparar ({unstagedFiles.length})</span>
                                    {unstagedFiles.length > 0 && (
                                        <button
                                            type="button"
                                            className="mini-action-btn"
                                            onClick={onStageAll}
                                            title="Preparar todos los cambios"
                                        >
                                            + Todo
                                        </button>
                                    )}
                                </div>
                                <div className="subgroup-list">
                                    {unstagedFiles.length === 0 ? (
                                        <div className="subgroup-empty">Sin cambios pendientes</div>
                                    ) : (
                                        unstagedFiles.map((f) => (
                                            <div
                                                key={f.path}
                                                className={`git-item-row ${selectedFile === f.path ? "selected" : ""}`}
                                                onClick={() => onSelectFile(f.path)}
                                            >
                                                <span className={`status-pill ${f.status}`}>{f.status}</span>
                                                <span className="git-item-path" title={f.path}>
                                                    {f.name || f.path.split("/").pop()}
                                                </span>
                                                <button
                                                    type="button"
                                                    className="stage-toggle-btn"
                                                    onClick={(e) => {
                                                        e.stopPropagation();
                                                        onStageFile(f.path);
                                                    }}
                                                    title="Preparar archivo (stage)"
                                                >
                                                    +
                                                </button>
                                            </div>
                                        ))
                                    )}
                                </div>
                            </div>

                            {/* Staged section */}
                            <div className="git-subgroup">
                                <div className="subgroup-head">
                                    <span>Cambios preparados / Staged ({stagedFiles.length})</span>
                                    {stagedFiles.length > 0 && (
                                        <button
                                            type="button"
                                            className="mini-action-btn"
                                            onClick={onUnstageAll}
                                            title="Despreparar todos"
                                        >
                                            - Todo
                                        </button>
                                    )}
                                </div>
                                <div className="subgroup-list">
                                    {stagedFiles.length === 0 ? (
                                        <div className="subgroup-empty">Ningún archivo preparado</div>
                                    ) : (
                                        stagedFiles.map((f) => (
                                            <div
                                                key={f.path}
                                                className={`git-item-row staged ${selectedFile === f.path ? "selected" : ""}`}
                                                onClick={() => onSelectFile(f.path)}
                                            >
                                                <span className={`status-pill ${f.status}`}>{f.status}</span>
                                                <span className="git-item-path" title={f.path}>
                                                    {f.name || f.path.split("/").pop()}
                                                </span>
                                                <button
                                                    type="button"
                                                    className="stage-toggle-btn unstage"
                                                    onClick={(e) => {
                                                        e.stopPropagation();
                                                        onUnstageFile(f.path);
                                                    }}
                                                    title="Despreparar archivo (unstage)"
                                                >
                                                    -
                                                </button>
                                            </div>
                                        ))
                                    )}
                                </div>
                            </div>

                            {/* Commit Composer */}
                            <div className="git-commit-box">
                                <textarea
                                    className="commit-textarea"
                                    placeholder="Mensaje de commit..."
                                    value={commitMsg}
                                    onChange={(e) => setCommitMsg(e.target.value)}
                                    rows={2}
                                />
                                <button
                                    type="button"
                                    className="commit-submit-btn"
                                    disabled={!commitMsg.trim() || committing || files.length === 0}
                                    onClick={handleCommit}
                                >
                                    {committing ? "Guardando..." : "✓ Confirmar cambios"}
                                </button>
                            </div>
                        </div>

                        {/* Center Diff View */}
                        <div className="git-review-diff-center">
                            {gitLoading ? (
                                <div className="diff-loading">Cargando diferencias de Git...</div>
                            ) : (
                                <DiffViewer
                                    diff={fileDiff || gitSummary?.raw_diff || ""}
                                    filePath={selectedFile || (files.length > 0 ? files[0].path : undefined)}
                                />
                            )}
                        </div>
                    </div>
                )}

                {/* ─── TAB 3: COMMANDS & TOOLS ────────────────────────────────── */}
                {activeTab === "commands" && (
                    <div className="commands-container">
                        <div className="commands-header-note">
                            Historial de herramientas, ejecuciones MCP y comandos de terminal.
                        </div>
                        {commandLogs.length === 0 ? (
                            <div className="panel-empty-hint">Aún no se han ejecutado comandos o herramientas.</div>
                        ) : (
                            <div className="commands-list">
                                {commandLogs.map((cmd) => {
                                    const isExpanded = expandedCommandId === cmd.id;
                                    return (
                                        <div key={cmd.id} className={`command-card ${cmd.status}`}>
                                            <div
                                                className="command-card-header"
                                                onClick={() => setExpandedCommandId(isExpanded ? null : cmd.id)}
                                            >
                                                <div className="command-title-row">
                                                    <span className={`cmd-type-tag ${cmd.kind}`}>{cmd.kind}</span>
                                                    <span className="cmd-name">{cmd.name}</span>
                                                    <span className={`cmd-status-badge ${cmd.status}`}>{cmd.status}</span>
                                                </div>
                                                <div className="command-meta-row">
                                                    {cmd.duration_ms != null && (
                                                        <span className="cmd-duration">{cmd.duration_ms}ms</span>
                                                    )}
                                                    <span className="expand-indicator">{isExpanded ? "▲" : "▼"}</span>
                                                </div>
                                            </div>

                                            {cmd.args && (
                                                <div className="cmd-args-preview">
                                                    <code>{cmd.args}</code>
                                                </div>
                                            )}

                                            {isExpanded && (
                                                <div className="command-expanded-content">
                                                    {cmd.output && (
                                                        <div className="command-output-box">
                                                            <div className="box-title">Salida / Resultado:</div>
                                                            <pre>{cmd.output}</pre>
                                                        </div>
                                                    )}

                                                    {cmd.diff && (
                                                        <div className="command-embedded-diff">
                                                            <div className="box-title">Modificaciones realizadas (Diff):</div>
                                                            <DiffViewer diff={cmd.diff} />
                                                        </div>
                                                    )}
                                                </div>
                                            )}
                                        </div>
                                    );
                                })}
                            </div>
                        )}
                    </div>
                )}
            </div>
        </aside>
    );
};
