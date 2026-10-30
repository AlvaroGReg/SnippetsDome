import { fireEvent, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import CollectionSecurityDialog from "./CollectionSecurityDialog";
import { renderWithFluent } from "../../test/render";

describe("CollectionSecurityDialog", () => {
    it("accepts a letter-and-symbol password and matching confirmation", async () => {
        const submit = vi.fn().mockResolvedValue(undefined);
        renderWithFluent(<CollectionSecurityDialog
            collection={{ id: "c1", name: "Private", createdAt: "", encrypted: false }}
            mode="encrypt" onClose={vi.fn()} onSubmit={submit}
        />);
        fireEvent.change(screen.getByLabelText("Collection password (1–10 characters)"), { target: { value: "A!b9x" } });
        fireEvent.change(screen.getByLabelText("Confirm password"), { target: { value: "A!b9x" } });
        fireEvent.click(screen.getByRole("button", { name: "Encrypt list" }));
        await waitFor(() => expect(submit).toHaveBeenCalledWith("A!b9x", "A!b9x"));
    });

    it("limits passwords to ten characters", async () => {
        const submit = vi.fn().mockResolvedValue(undefined);
        renderWithFluent(<CollectionSecurityDialog
            collection={{ id: "c1", name: "Private", createdAt: "", encrypted: false }}
            mode="encrypt" onClose={vi.fn()} onSubmit={submit}
        />);
        fireEvent.change(screen.getByLabelText("Collection password (1–10 characters)"), { target: { value: "A!b9x123456" } });
        fireEvent.change(screen.getByLabelText("Confirm password"), { target: { value: "A!b9x12345" } });
        fireEvent.click(screen.getByRole("button", { name: "Encrypt list" }));
        await waitFor(() => expect(submit).toHaveBeenCalledWith("A!b9x12345", "A!b9x12345"));
    });
});
