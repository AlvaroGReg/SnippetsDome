import {
    Button,
    Dialog,
    DialogActions,
    DialogBody,
    DialogContent,
    DialogSurface,
    DialogTitle,
    Field,
    Input,
    Switch,
} from "@fluentui/react-components";
import { useEffect, useState } from "react";

type SettingsDialogProps = {
    open: boolean;
    closeToTrayEnabled: boolean;
    startAtLoginEnabled: boolean;
    startAtLoginSupported: boolean;
    traySnippetLimit: number;
    onClose: () => void;
    onCloseToTrayChange: (enabled: boolean) => void;
    onStartAtLoginChange: (enabled: boolean) => void;
    onTraySnippetLimitChange: (limit: number) => Promise<boolean>;
};

function SettingsDialog({
    open,
    closeToTrayEnabled,
    startAtLoginEnabled,
    startAtLoginSupported,
    traySnippetLimit,
    onClose,
    onCloseToTrayChange,
    onStartAtLoginChange,
    onTraySnippetLimitChange,
}: SettingsDialogProps) {
    const [traySnippetLimitInput, setTraySnippetLimitInput] = useState(String(traySnippetLimit));

    useEffect(() => {
        setTraySnippetLimitInput(String(traySnippetLimit));
    }, [traySnippetLimit]);

    async function handleTraySnippetLimitChange(value: string) {
        setTraySnippetLimitInput(value);
        const limit = Number(value);
        if (Number.isInteger(limit) && limit > 0) {
            const saved = await onTraySnippetLimitChange(limit);
            if (!saved) {
                setTraySnippetLimitInput(String(traySnippetLimit));
            }
        }
    }

    return (
        <Dialog open={open} onOpenChange={(_, data) => !data.open && onClose()}>
            <DialogSurface>
                <DialogBody>
                    <DialogTitle>Settings</DialogTitle>
                    <DialogContent>
                        <Switch
                            checked={closeToTrayEnabled}
                            label="Close to tray"
                            onChange={(_, data) => onCloseToTrayChange(data.checked)}
                        />
                        <p>Keep SnippetsDome running in the notification area when its window is closed.</p>
                        <Switch
                            checked={startAtLoginEnabled}
                            disabled={!startAtLoginSupported}
                            label="Start at login"
                            onChange={(_, data) => onStartAtLoginChange(data.checked)}
                        />
                        <p>{startAtLoginSupported
                            ? "Launch SnippetsDome automatically when you sign in."
                            : "Available on Windows, Linux, and macOS."}
                        </p>
                        <Field label="Snippets shown in tray">
                            <Input
                                aria-label="Snippets shown in tray"
                                min={1}
                                type="number"
                                value={traySnippetLimitInput}
                                onChange={(_, data) => void handleTraySnippetLimitChange(data.value)}
                            />
                        </Field>
                    </DialogContent>
                    <DialogActions>
                        <Button appearance="primary" onClick={onClose}>Done</Button>
                    </DialogActions>
                </DialogBody>
            </DialogSurface>
        </Dialog>
    );
}

export default SettingsDialog;
