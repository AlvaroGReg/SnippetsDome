import { Button, Dialog, DialogActions, DialogBody, DialogContent, DialogSurface, DialogTitle, Input } from "@fluentui/react-components";
import { useState } from "react";
import type { CollectionModel } from "../../models/Snippet";
import { useTranslation } from "../../i18n";

type Props = {
    mode: "import" | "export";
    collection?: CollectionModel;
    onClose: () => void;
    onSubmit: (collectionPassword: string, password: string) => Promise<void>;
};

export default function CollectionExchangeDialog({ mode, collection, onClose, onSubmit }: Props) {
    const { t } = useTranslation();
    const [collectionPassword, setCollectionPassword] = useState("");
    const [password, setPassword] = useState("");
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState("");
    const title = mode === "import" ? t("importJson") : t("exportJson");

    async function submit() {
        if (mode === "export" && (!/^\d{4}$/.test(collectionPassword) || !password)) {
            setError(t("exchangeCredentialsRequired"));
            return;
        }
        setBusy(true);
        setError("");
        try {
            await onSubmit(collectionPassword, password);
            onClose();
        } catch {
            setError(mode === "export" ? t("collectionSecurityFailed") : t("unableToImportJson"));
        } finally {
            setBusy(false);
            setCollectionPassword("");
            setPassword("");
        }
    }

    return <Dialog open onOpenChange={(_, data) => { if (!data.open && !busy) onClose(); }}>
        <DialogSurface><DialogBody><DialogTitle>{title}{collection ? `: ${collection.name}` : ""}</DialogTitle><DialogContent>
            {mode === "export" && <>
                <Input type="password" value={collectionPassword} maxLength={4} inputMode="numeric" aria-label={t("currentListPassword")} placeholder={t("currentListPassword")} onChange={(_, data) => setCollectionPassword(data.value)} />
                <Input type="password" value={password} aria-label={t("exportPassword")} placeholder={t("exportPassword")} onChange={(_, data) => setPassword(data.value)} />
            </>}
            {mode === "import" && <Input type="password" value={password} aria-label={t("importPassword")} placeholder={t("importPassword")} onChange={(_, data) => setPassword(data.value)} />}
            {error && <div role="alert">{error}</div>}
        </DialogContent><DialogActions>
            <Button onClick={onClose} disabled={busy}>{t("cancel")}</Button>
            <Button appearance="primary" onClick={() => void submit()} disabled={busy}>{title}</Button>
        </DialogActions></DialogBody></DialogSurface>
    </Dialog>;
}
