import * as services from "../bindings/github.com/user/go-claude-code/cmd/wails/services";
import { Call } from "@wailsio/runtime";

export * from "../bindings/github.com/user/go-claude-code/cmd/wails/services";

export const ChatService = {
    ...services.ChatService,
    ApproveScope: (callID: string, scope: "once" | "chat" | "project" | "deny", toolName?: string) => {
        if ((services.ChatService as any).ApproveScope) {
            return (services.ChatService as any).ApproveScope(callID, scope, toolName || "");
        }
        return services.ChatService.Approve(callID, scope !== "deny");
    },
};

export const GitService = {
    ...services.GitService,
    StageFile: (filePath: string): Promise<string> => {
        return Call.ByName("github.com/user/go-claude-code/cmd/wails/services.GitService.StageFile", filePath);
    },
    UnstageFile: (filePath: string): Promise<string> => {
        return Call.ByName("github.com/user/go-claude-code/cmd/wails/services.GitService.UnstageFile", filePath);
    },
    StageAll: (): Promise<string> => {
        return Call.ByName("github.com/user/go-claude-code/cmd/wails/services.GitService.StageAll");
    },
    UnstageAll: (): Promise<string> => {
        return Call.ByName("github.com/user/go-claude-code/cmd/wails/services.GitService.UnstageAll");
    },
};