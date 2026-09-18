import { useCallback, useEffect, useState } from "react";
import type { CreateSnippetInput, SnippetModel } from "../models/Snippet";
import * as snippetsService from "../services/snippets-service";

function getErrorMessage(error: unknown): string {
    return error instanceof Error ? error.message : "An unexpected error occurred.";
}

function orderSnippets(snippets: SnippetModel[]): SnippetModel[] {
    return [...snippets].sort((left, right) => Number(Boolean(right.favorite)) - Number(Boolean(left.favorite)));
}

export function useSnippets() {
    const [snippets, setSnippets] = useState<SnippetModel[]>([]);
    const [error, setError] = useState("");
    const [storagePath, setStoragePath] = useState("");
    const [loadingOperations, setLoadingOperations] = useState(0);

    const clearError = useCallback(() => {
        setError("");
    }, []);

    const startLoading = useCallback(() => {
        setLoadingOperations((current) => current + 1);
    }, []);

    const stopLoading = useCallback(() => {
        setLoadingOperations((current) => Math.max(0, current - 1));
    }, []);

    const loadSnippets = useCallback(async () => {
        startLoading();
        try {
            setError("");
            setSnippets(orderSnippets(await snippetsService.getSnippets()));
        } catch (error) {
            setError(getErrorMessage(error));
        } finally {
            stopLoading();
        }
    }, [startLoading, stopLoading]);

    useEffect(() => {
        void loadSnippets();
        startLoading();
        void snippetsService.getSnippetsStoragePath()
            .then(setStoragePath)
            .catch((error: unknown) => {
                setError(getErrorMessage(error));
            })
            .finally(stopLoading);
    }, [loadSnippets, startLoading, stopLoading]);

    const selectStorageFile = useCallback(async (selectFile: () => Promise<string>) => {
        startLoading();
        try {
            setError("");
            const filePath = await selectFile();
            if (!filePath) {
                return;
            }
            if (!filePath.toLocaleLowerCase().endsWith(".json")) {
                throw new Error("The snippets file must use the .json extension.");
            }
            setStoragePath(await snippetsService.setSnippetsStoragePath(filePath));
            await loadSnippets();
        } catch (error) {
            setError(getErrorMessage(error));
        } finally {
            stopLoading();
        }
    }, [loadSnippets, startLoading, stopLoading]);

    const pickExistingStorageFile = useCallback(
        () => selectStorageFile(snippetsService.pickExistingSnippetsFile),
        [selectStorageFile],
    );

    const createStorageFile = useCallback(
        () => selectStorageFile(snippetsService.createSnippetsFile),
        [selectStorageFile],
    );

    const createSnippet = useCallback(async (input: CreateSnippetInput) => {
        startLoading();
        try {
            setError("");
            const snippet = await snippetsService.createSnippet(input);
            setSnippets((currentSnippets) => orderSnippets([...currentSnippets, snippet]));
        } catch (error) {
            setError(getErrorMessage(error));
        } finally {
            stopLoading();
        }
    }, [startLoading, stopLoading]);

    const updateSnippet = useCallback(async (snippet: SnippetModel) => {
        startLoading();
        try {
            setError("");
            const updatedSnippet = await snippetsService.updateSnippet(snippet);
            setSnippets((currentSnippets) => orderSnippets(currentSnippets.map((currentSnippet) =>
                currentSnippet.id === updatedSnippet.id ? updatedSnippet : currentSnippet,
            )));
        } catch (error) {
            setError(getErrorMessage(error));
        } finally {
            stopLoading();
        }
    }, [startLoading, stopLoading]);

    const deleteSnippet = useCallback(async (id: string) => {
        startLoading();
        try {
            setError("");
            await snippetsService.deleteSnippet(id);
            setSnippets((currentSnippets) =>
                currentSnippets.filter((snippet) => snippet.id !== id),
            );
        } catch (error) {
            setError(getErrorMessage(error));
        } finally {
            stopLoading();
        }
    }, [startLoading, stopLoading]);

    return {
        snippets,
        error,
        clearError,
        isLoading: loadingOperations > 0,
        storagePath,
        pickExistingStorageFile,
        createStorageFile,
        createSnippet,
        updateSnippet,
        deleteSnippet,
    };
}
