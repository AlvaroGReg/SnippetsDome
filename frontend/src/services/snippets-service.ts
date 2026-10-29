import {
	CreateSnippet,
    DeleteSnippet,
    CreateCollection,
    DeleteCollection,
    GetActiveCollection,
    GetDefaultCollection,
    GetCollections,
    GetCloseToTrayEnabled,
    GetStartAtLoginEnabled,
    GetStartAtLoginSupported,
    GetSnippets,
    GetTraySnippetLimit,
    GetLanguage,
    SelectCollection,
    SetCloseToTrayEnabled,
    SetDefaultCollection,
    SetStartAtLoginEnabled,
    SetTraySnippetLimit,
    SetLanguage,
    RenameCollection,
    UpdateSnippet,
    BeginJSONImport,
    CancelPendingJSONImport,
    ExportJSON,
    ExportCollectionJSON,
    ExportEncryptedCollectionJSON,
    ImportPendingEncryptedJSON,
    EncryptCollection,
    UnlockCollection,
    LockActiveCollection,
    DecryptCollection,
} from "../../wailsjs/go/main/App";
import type { CollectionModel, CreateSnippetInput, SnippetModel } from "../models/Snippet";
import type { main } from "../../wailsjs/go/models";

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
    return runRequest(GetSnippets, "Unable to load snippets.");
}

export function getCollections(): Promise<CollectionModel[]> {
    return runRequest(GetCollections, "Unable to load collections.");
}

export function getActiveCollection(): Promise<CollectionModel> {
    return runRequest(GetActiveCollection, "Unable to load the active collection.");
}

export function getDefaultCollection(): Promise<CollectionModel> {
    return runRequest(GetDefaultCollection, "Unable to load the default collection.");
}

export function createCollection(name: string): Promise<CollectionModel> {
    return runRequest(() => CreateCollection(name), "Unable to create the collection.");
}

export function selectCollection(id: string): Promise<void> {
    return runRequest(() => SelectCollection(id), "Unable to select the collection.");
}

export function setDefaultCollection(id: string): Promise<void> {
    return runRequest(() => SetDefaultCollection(id), "Unable to set the default collection.");
}

export function deleteCollection(id: string): Promise<void> {
    return runRequest(() => DeleteCollection(id), "Unable to delete the collection.");
}

export function renameCollection(id: string, name: string): Promise<void> {
    return runRequest(() => RenameCollection(id, name), "Unable to rename the collection.");
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

export function getLanguage(): Promise<string> {
    return runRequest(GetLanguage, "Unable to get the language preference.");
}

export function setLanguage(language: string): Promise<void> {
    return runRequest(() => SetLanguage(language), "Unable to save the language preference.");
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

export function beginJSONImport(): Promise<main.JSONImportResult> {
    return runRequest(BeginJSONImport, "Unable to import the JSON file.");
}

export function exportJSON(): Promise<void> {
    return runRequest(ExportJSON, "Unable to export the JSON file.");
}

export function cancelPendingJSONImport(): void {
    CancelPendingJSONImport();
}

export function importPendingEncryptedJSON(password: string): Promise<CollectionModel> {
    return runRequest(() => ImportPendingEncryptedJSON(password), "Unable to import the JSON file.");
}

export function exportCollectionJSON(id: string): Promise<void> {
    return runRequest(() => ExportCollectionJSON(id), "Unable to export the collection.");
}

export function exportEncryptedCollectionJSON(id: string, collectionPassword: string, password: string): Promise<void> {
    return runRequest(() => ExportEncryptedCollectionJSON(id, collectionPassword, password), "Unable to export the encrypted list.");
}

export function encryptCollection(id: string, pin: string, confirmation: string): Promise<void> {
    return runRequest(() => EncryptCollection(id, pin, confirmation), "Unable to encrypt the list.");
}

export function unlockCollection(id: string, pin: string): Promise<CollectionModel> {
    return runRequest(() => UnlockCollection(id, pin), "Unable to unlock the list.");
}

export function lockActiveCollection(): Promise<void> {
    return runRequest(LockActiveCollection, "Unable to lock the list.");
}

export function decryptCollection(id: string, pin: string): Promise<void> {
    return runRequest(() => DecryptCollection(id, pin), "Unable to decrypt the list.");
}
