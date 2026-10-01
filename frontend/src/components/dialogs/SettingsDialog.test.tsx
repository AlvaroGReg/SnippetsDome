import { describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";
import { screen } from "@testing-library/react";
import SettingsDialog from "./SettingsDialog";
import { renderWithFluent } from "../../test/render";

describe("SettingsDialog", () => {
    it("changes the close-to-tray preference", async () => {
        const user = userEvent.setup();
        const onCloseToTrayChange = vi.fn();
        const onTraySnippetLimitChange = vi.fn().mockResolvedValue(true);
        renderWithFluent(
            <SettingsDialog
                open
                closeToTrayEnabled={false}
                startAtLoginEnabled={false}
                startAtLoginSupported
                traySnippetLimit={5}
                onClose={vi.fn()}
                onCloseToTrayChange={onCloseToTrayChange}
                onStartAtLoginChange={vi.fn()}
                onTraySnippetLimitChange={onTraySnippetLimitChange}
            />,
        );

        await user.click(screen.getByRole("switch", { name: "Close to tray" }));

        expect(onCloseToTrayChange).toHaveBeenCalledWith(true);
    });

    it("changes the tray snippet limit", async () => {
        const user = userEvent.setup();
        let traySnippetLimit = 5;
        const onTraySnippetLimitChange = vi.fn((limit: number): Promise<boolean> => {
            traySnippetLimit = limit;
            rerenderDialog();
            return Promise.resolve(true);
        });
        const renderDialog = () => (
            <SettingsDialog
                open
                closeToTrayEnabled={false}
                startAtLoginEnabled={false}
                startAtLoginSupported
                traySnippetLimit={traySnippetLimit}
                onClose={vi.fn()}
                onCloseToTrayChange={vi.fn()}
                onStartAtLoginChange={vi.fn()}
                onTraySnippetLimitChange={onTraySnippetLimitChange}
            />
        );
        const { rerender } = renderWithFluent(renderDialog());
        const rerenderDialog = () => rerender(renderDialog());

        await user.clear(screen.getByRole("spinbutton", { name: "Snippets shown in tray" }));
        await user.type(screen.getByRole("spinbutton", { name: "Snippets shown in tray" }), "3");

        expect(onTraySnippetLimitChange).toHaveBeenCalledWith(3);
    });

    it("keeps the persisted tray snippet limit when saving fails", async () => {
        const user = userEvent.setup();
        const onTraySnippetLimitChange = vi.fn().mockResolvedValue(false);
        renderWithFluent(
            <SettingsDialog
                open
                closeToTrayEnabled={false}
                startAtLoginEnabled={false}
                startAtLoginSupported
                traySnippetLimit={5}
                onClose={vi.fn()}
                onCloseToTrayChange={vi.fn()}
                onStartAtLoginChange={vi.fn()}
                onTraySnippetLimitChange={onTraySnippetLimitChange}
            />,
        );

        await user.clear(screen.getByRole("spinbutton", { name: "Snippets shown in tray" }));
        await user.type(screen.getByRole("spinbutton", { name: "Snippets shown in tray" }), "3");

        expect(onTraySnippetLimitChange).toHaveBeenCalledWith(3);
        expect(screen.getByRole("spinbutton", { name: "Snippets shown in tray" })).toHaveValue(5);
    });

    it("changes the start-at-login preference when supported", async () => {
        const user = userEvent.setup();
        const onStartAtLoginChange = vi.fn();
        renderWithFluent(
            <SettingsDialog
                open
                closeToTrayEnabled={false}
                startAtLoginEnabled={false}
                startAtLoginSupported
                traySnippetLimit={5}
                onClose={vi.fn()}
                onCloseToTrayChange={vi.fn()}
                onStartAtLoginChange={onStartAtLoginChange}
                onTraySnippetLimitChange={vi.fn().mockResolvedValue(true)}
            />,
        );

        await user.click(screen.getByRole("switch", { name: "Start at login" }));

        expect(onStartAtLoginChange).toHaveBeenCalledWith(true);
    });

    it("changes the language preference", async () => {
        const user = userEvent.setup();
        const onLanguageChange = vi.fn();
        renderWithFluent(
            <SettingsDialog
                open
                closeToTrayEnabled={false}
                startAtLoginEnabled={false}
                startAtLoginSupported
                traySnippetLimit={5}
                language="en"
                onClose={vi.fn()}
                onCloseToTrayChange={vi.fn()}
                onStartAtLoginChange={vi.fn()}
                onTraySnippetLimitChange={vi.fn().mockResolvedValue(true)}
                onLanguageChange={onLanguageChange}
            />,
        );

        await user.selectOptions(screen.getByRole("combobox"), "es");

        expect(onLanguageChange).toHaveBeenCalledWith("es");
    });

    it("shows start at login as unavailable on unsupported systems", () => {
        renderWithFluent(
            <SettingsDialog
                open
                closeToTrayEnabled={false}
                startAtLoginEnabled={false}
                startAtLoginSupported={false}
                traySnippetLimit={5}
                onClose={vi.fn()}
                onCloseToTrayChange={vi.fn()}
                onStartAtLoginChange={vi.fn()}
                onTraySnippetLimitChange={vi.fn().mockResolvedValue(true)}
            />,
        );

        expect(screen.getByRole("switch", { name: "Start at login" })).toBeDisabled();
        expect(screen.getByText("Available on Windows, Linux, and macOS.")).toBeInTheDocument();
    });
});
