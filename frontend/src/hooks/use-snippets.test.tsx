import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CreateSnippetInput, SnippetModel } from "../models/Snippet";
import * as snippetsService from "../services/snippets-service";
import { useSnippets } from "./use-snippets";

vi.mock("../services/snippets-service", () => ({
    getSnippets: vi.fn(),
    createSnippet: vi.fn(),
    updateSnippet: vi.fn(),
    deleteSnippet: vi.fn(),
}));

const regularSnippet: SnippetModel = {
    id: "regular",
    title: "Regular",
    language: "Go",
    code: "regular()",
    tags: [],
    createdAt: "2026-01-01T00:00:00Z",
    favorite: false,
};

const favoriteSnippet: SnippetModel = {
    ...regularSnippet,
    id: "favorite",
    title: "Favorite",
    favorite: true,
};

describe("useSnippets", () => {
    beforeEach(() => {
        vi.resetAllMocks();
        vi.mocked(snippetsService.getSnippets).mockResolvedValue([regularSnippet, favoriteSnippet]);
        vi.mocked(snippetsService.createSnippet).mockResolvedValue({ ...regularSnippet, id: "created" });
        vi.mocked(snippetsService.updateSnippet).mockImplementation(async (snippet) => snippet);
        vi.mocked(snippetsService.deleteSnippet).mockResolvedValue(undefined);
    });

    it("loads snippets, places favorites first, and ends loading", async () => {
        const { result } = renderHook(() => useSnippets());

        expect(result.current.isLoading).toBe(true);
        await waitFor(() => expect(result.current.isLoading).toBe(false));
        expect(result.current.snippets.map(({ id }) => id)).toEqual(["favorite", "regular"]);
        expect(result.current.error).toBe("");
    });

    it("keeps existing snippets when loading fails and supports clearing the error", async () => {
        vi.mocked(snippetsService.getSnippets)
            .mockResolvedValueOnce([regularSnippet])
            .mockRejectedValueOnce(new Error("database unavailable"));
        const { result } = renderHook(() => useSnippets());
        await waitFor(() => expect(result.current.snippets).toHaveLength(1));

        await act(async () => result.current.reload());

        expect(result.current.error).toBe("database unavailable");
        expect(result.current.snippets).toEqual([regularSnippet]);
        act(() => result.current.clearError());
        expect(result.current.error).toBe("");
    });

    it("updates, reorders and deletes snippets after successful operations", async () => {
        vi.mocked(snippetsService.getSnippets).mockResolvedValueOnce([
            regularSnippet,
            { ...regularSnippet, id: "second", title: "Second" },
        ]);
        const { result } = renderHook(() => useSnippets());
        await waitFor(() => expect(result.current.isLoading).toBe(false));
        const secondSnippet = { ...regularSnippet, id: "second", title: "Second" };
        const promoted = { ...secondSnippet, favorite: true };

        await act(async () => result.current.updateSnippet(promoted));
        expect(result.current.snippets.map(({ id }) => id)).toEqual(["second", "regular"]);
        expect(result.current.snippets[0].favorite).toBe(true);

        await act(async () => result.current.deleteSnippet("regular"));
        expect(snippetsService.deleteSnippet).toHaveBeenCalledWith("regular");
        expect(result.current.snippets.map(({ id }) => id)).toEqual(["second"]);
    });

    it("reports non-Error failures and does not add a failed creation", async () => {
        vi.mocked(snippetsService.createSnippet).mockRejectedValue("offline");
        const { result } = renderHook(() => useSnippets());
        await waitFor(() => expect(result.current.isLoading).toBe(false));
        const input: CreateSnippetInput = { title: "New", language: "Go", code: "new()", tags: [] };

        await act(async () => result.current.createSnippet(input));

        expect(result.current.error).toBe("An unexpected error occurred.");
        expect(result.current.snippets.map(({ id }) => id)).toEqual(["favorite", "regular"]);
        expect(result.current.isLoading).toBe(false);
    });
});
