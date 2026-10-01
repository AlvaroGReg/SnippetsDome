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
import { useTranslation, type LanguagePreference } from "../../i18n";

type SettingsDialogProps = {
    open: boolean;
    closeToTrayEnabled: boolean;
    startAtLoginEnabled: boolean;
    startAtLoginSupported: boolean;
    traySnippetLimit: number;
    language?: LanguagePreference;
    onClose: () => void;
    onCloseToTrayChange: (enabled: boolean) => void;
    onStartAtLoginChange: (enabled: boolean) => void;
    onTraySnippetLimitChange: (limit: number) => Promise<boolean>;
    onLanguageChange?: (language: LanguagePreference) => void;
};

function SettingsDialog({
    open,
    closeToTrayEnabled,
    startAtLoginEnabled,
    startAtLoginSupported,
    traySnippetLimit,
    language = "en",
    onClose,
    onCloseToTrayChange,
    onStartAtLoginChange,
    onTraySnippetLimitChange,
    onLanguageChange = () => undefined,
}: SettingsDialogProps) {
    const { t } = useTranslation();
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
                    <DialogTitle>{t("settings")}</DialogTitle>
                    <DialogContent>
                        <Switch
                            checked={closeToTrayEnabled}
                            label={t("closeToTray")}
                            onChange={(_, data) => onCloseToTrayChange(data.checked)}
                        />
                        <p>{t("closeToTrayDescription")}</p>
                        <Switch
                            checked={startAtLoginEnabled}
                            disabled={!startAtLoginSupported}
                            label={t("startAtLogin")}
                            onChange={(_, data) => onStartAtLoginChange(data.checked)}
                        />
                        <p>{startAtLoginSupported ? t("startAtLoginDescription") : t("unavailablePlatforms")}
                        </p>
                        <Field label={t("snippetsShownInTray")}>
                            <Input
                                aria-label={t("snippetsShownInTray")}
                                min={1}
                                type="number"
                                value={traySnippetLimitInput}
                                onChange={(_, data) => void handleTraySnippetLimitChange(data.value)}
                            />
                        </Field>
                        <Field label={t("language")}>
                            <select value={language} onChange={(event) => onLanguageChange(event.target.value as LanguagePreference)}>
                                <option value="en">{t("english")}</option>
                                <option value="es">{t("spanish")}</option>
                            </select>
                        </Field>
                    </DialogContent>
                    <DialogActions>
                        <Button appearance="primary" onClick={onClose}>{t("done")}</Button>
                    </DialogActions>
                </DialogBody>
            </DialogSurface>
        </Dialog>
    );
}

export default SettingsDialog;
