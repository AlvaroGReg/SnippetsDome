import { describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";
import { screen } from "@testing-library/react";
import CollectionManagementDialog from "./CollectionManagementDialog";
import { renderWithFluent } from "../../test/render";
import type { CollectionModel } from "../../models/Snippet";

const lockedEncryptedCollection: CollectionModel = {
    id: "collection-12345678",
    name: "Locked list (collecti)",
    createdAt: "2026-01-01T00:00:00Z",
    encrypted: true,
    locked: true,
};

function renderDialog(collection: CollectionModel, onUnlockForRename = vi.fn()) {
    renderWithFluent(
        <CollectionManagementDialog
            open
            collections={[collection]}
            defaultCollectionId={null}
            newCollectionName=""
            onClose={vi.fn()}
            onNewCollectionNameChange={vi.fn()}
            onCreate={vi.fn()}
            onRename={vi.fn().mockResolvedValue(true)}
            onUnlockForRename={onUnlockForRename}
            isJSONBusy={false}
            onImport={vi.fn()}
            onExport={vi.fn()}
            onDelete={vi.fn()}
            onSetDefault={vi.fn()}
            onSecurity={vi.fn()}
        />,
    );
    return onUnlockForRename;
}

describe("CollectionManagementDialog", () => {
    it("requests unlock before allowing a locked encrypted list to be renamed", async () => {
        const user = userEvent.setup();
        const onUnlockForRename = renderDialog(lockedEncryptedCollection);

        await user.click(screen.getByRole("button", { name: `Edit list ${lockedEncryptedCollection.name}` }));

        expect(onUnlockForRename).toHaveBeenCalledWith(lockedEncryptedCollection);
        expect(screen.queryByRole("textbox", { name: `Edit list ${lockedEncryptedCollection.name}` })).not.toBeInTheDocument();
    });

    it("allows renaming after an encrypted list has been unlocked", async () => {
        const user = userEvent.setup();
        renderDialog({ ...lockedEncryptedCollection, name: "Private notes", locked: false });

        await user.click(screen.getByRole("button", { name: "Edit list Private notes" }));

        expect(screen.getByRole("textbox", { name: "Edit list Private notes" })).toBeInTheDocument();
    });
});
