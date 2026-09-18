import { afterEach, describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";
import { screen } from "@testing-library/react";
import SnippetsList from "./SnippetsList";
import { renderWithFluent } from "../../test/render";
import type { SnippetModel } from "../../models/Snippet";

const snippet: SnippetModel = {
    id: "snippet-1",
    title: "Fetch data",
    language: "TypeScript",
    code: "await fetch('/api/data')",
    tags: ["http", "async"],
    createdAt: "2026-01-01T00:00:00Z",
    favorite: false,
};

afterEach(() => {
    vi.restoreAllMocks();
});

describe("SnippetsList", () => {
    it("shows an empty-state message", () => {
        renderWithFluent(<SnippetsList snippets={[]} onEdit={vi.fn()} onDelete={vi.fn()} onToggleFavorite={vi.fn()} />);

        expect(screen.getByText("Empty list")).toBeInTheDocument();
    });

    it("renders snippets and sends edit, delete and copy actions", async () => {
        const user = userEvent.setup();
        const onEdit = vi.fn();
        const onDelete = vi.fn();
        const onToggleFavorite = vi.fn();
        const writeText = vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue(undefined);
        renderWithFluent(<SnippetsList snippets={[snippet]} onEdit={onEdit} onDelete={onDelete} onToggleFavorite={onToggleFavorite} />);

        expect(screen.getByText("Fetch data")).toBeInTheDocument();
        expect(screen.getByText("http")).toBeInTheDocument();

        await user.click(screen.getByRole("button", { name: "Copy" }));
        expect(writeText).toHaveBeenCalledWith(snippet.code);
        expect(await screen.findByText("Code copied to clipboard")).toBeInTheDocument();

        await user.click(screen.getByRole("button", { name: "Edit" }));
        expect(onEdit).toHaveBeenCalledWith(snippet);

        await user.click(screen.getByRole("button", { name: "Delete" }));
        expect(onDelete).toHaveBeenCalledWith(snippet.id);

        await user.click(screen.getByRole("button", { name: "Add to favorites" }));
        expect(onToggleFavorite).toHaveBeenCalledWith(snippet);
    });
});
