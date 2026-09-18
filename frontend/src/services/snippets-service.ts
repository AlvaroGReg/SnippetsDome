import {
	CreateSnippet,
	CreateSnippetsFile,
    DeleteSnippet,
    GetCloseToTrayEnabled,
    GetStartAtLoginEnabled,
    GetStartAtLoginSupported,
    GetSnippets,
	GetSnippetsStoragePath,
	GetTraySnippetLimit,
	PickExistingSnippetsFile,
    SetSnippetsStoragePath,
    SetCloseToTrayEnabled,
    SetStartAtLoginEnabled,
	SetTraySnippetLimit,
    UpdateSnippet,
} from "../../wailsjs/go/main/App";
import type { CreateSnippetInput, SnippetModel } from "../models/Snippet";

class SnippetsServiceError extends Error {
    constructor(message: string, cause: unknown) {
        super(message);
        this.name = "SnippetsServiceError";
        this.cause = cause;
    }
}

async function runRequest<T>(request: () => Promise<T>, errorMessage: string): Promise<T> {
    try {
        return await request();
    } catch (error) {
        throw new SnippetsServiceError(errorMessage, error);
    }
}

export function getSnippets(): Promise<SnippetModel[]> {
    return runRequest(GetSnippets, "No snippets file is configured. Choose a file first.");
}

export function getCloseToTrayEnabled(): Promise<boolean> {
    return runRequest(GetCloseToTrayEnabled, "Unable to get the close-to-tray preference.");
}

export function setCloseToTrayEnabled(enabled: boolean): Promise<void> {
    return runRequest(() => SetCloseToTrayEnabled(enabled), "Unable to save the close-to-tray preference.");
}

export function getStartAtLoginSupported(): Promise<boolean> {
    return runRequest(GetStartAtLoginSupported, "Unable to determine whether start at login is available.");
}

export function getStartAtLoginEnabled(): Promise<boolean> {
    return runRequest(GetStartAtLoginEnabled, "Unable to get the start-at-login preference.");
}

export function setStartAtLoginEnabled(enabled: boolean): Promise<void> {
    return runRequest(() => SetStartAtLoginEnabled(enabled), "Unable to save the start-at-login preference.");
}

export function getTraySnippetLimit(): Promise<number> {
    return runRequest(GetTraySnippetLimit, "Unable to get the tray snippet limit.");
}

export function setTraySnippetLimit(limit: number): Promise<void> {
    return runRequest(() => SetTraySnippetLimit(limit), "Unable to save the tray snippet limit.");
}

export function createSnippet(input: CreateSnippetInput): Promise<SnippetModel> {
    return runRequest(() => CreateSnippet(input), "Unable to create the snippet.");
}

export function updateSnippet(snippet: SnippetModel): Promise<SnippetModel> {
    return runRequest(() => UpdateSnippet(snippet), "Unable to update the snippet.");
}

export function deleteSnippet(id: string): Promise<void> {
    return runRequest(() => DeleteSnippet(id), "Unable to delete the snippet.");
}

export function getSnippetsStoragePath(): Promise<string> {
    return runRequest(GetSnippetsStoragePath, "Unable to get the snippets file path.");
}

export function pickExistingSnippetsFile(): Promise<string> {
    return runRequest(PickExistingSnippetsFile, "Unable to choose the snippets file.");
}

export function createSnippetsFile(): Promise<string> {
    return runRequest(CreateSnippetsFile, "Unable to create the snippets file.");
}

export function setSnippetsStoragePath(filePath: string): Promise<string> {
    return runRequest(() => SetSnippetsStoragePath(filePath), "Unable to configure the snippets file.");
}
