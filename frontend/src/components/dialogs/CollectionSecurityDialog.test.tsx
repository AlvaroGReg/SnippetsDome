import { fireEvent, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import CollectionSecurityDialog from "./CollectionSecurityDialog";
import { renderWithFluent } from "../../test/render";

describe("CollectionSecurityDialog", () => {
    it("requires a four-digit PIN confirmation before encrypting", async () => {
        const submit = vi.fn().mockResolvedValue(undefined);
        renderWithFluent(<CollectionSecurityDialog
            collection={{ id: "c1", name: "Private", createdAt: "", encrypted: false }}
            mode="encrypt" onClose={vi.fn()} onSubmit={submit}
        />);
        fireEvent.change(screen.getByLabelText("Four-digit PIN"), { target: { value: "0042" } });
        fireEvent.change(screen.getByLabelText("Confirm PIN"), { target: { value: "0042" } });
        fireEvent.click(screen.getByRole("button", { name: "Encrypt list" }));
        await waitFor(() => expect(submit).toHaveBeenCalledWith("0042", "0042"));
    });
});
