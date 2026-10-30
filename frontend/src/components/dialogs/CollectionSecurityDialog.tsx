import { Button, Dialog, DialogActions, DialogBody, DialogContent, DialogSurface, DialogTitle, Input } from "@fluentui/react-components";
import { useState } from "react";
import type { CollectionModel } from "../../models/Snippet";
import { useTranslation } from "../../i18n";
import { isValidCollectionPassword, limitCollectionPassword } from "../../utils/collection-password";

export type SecurityMode = "unlock" | "encrypt" | "decrypt";
type Props = { collection: CollectionModel; mode: SecurityMode; onClose: () => void; onSubmit: (password: string, confirmation: string) => Promise<void> };

export default function CollectionSecurityDialog({ collection, mode, onClose, onSubmit }: Props) {
    const { t } = useTranslation();
    const [password, setPassword] = useState("");
    const [confirmation, setConfirmation] = useState("");
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState("");
    const title = mode === "unlock" ? t("unlockList") : mode === "encrypt" ? t("encryptList") : t("decryptList");

    async function submit() {
        if (!isValidCollectionPassword(password) || (mode === "encrypt" && password !== confirmation)) {
            setError(mode === "encrypt" && password !== confirmation ? t("passwordMismatch") : t("passwordFormat"));
            return;
        }
        setBusy(true); setError("");
        try { await onSubmit(password, confirmation); onClose(); }
        catch { setError(t("collectionSecurityFailed")); }
        finally { setBusy(false); setPassword(""); setConfirmation(""); }
    }

    return <Dialog open onOpenChange={(_, data) => { if (!data.open && !busy) onClose(); }}>
        <DialogSurface><DialogBody><DialogTitle>{title}: {collection.name}</DialogTitle><DialogContent>
            <Input type="password" value={password} aria-label={t("collectionPassword")} placeholder={t("collectionPassword")} onChange={(_, data) => setPassword(limitCollectionPassword(data.value))} />
            {mode === "encrypt" && <Input type="password" value={confirmation} aria-label={t("confirmPassword")} placeholder={t("confirmPassword")} onChange={(_, data) => setConfirmation(limitCollectionPassword(data.value))} />}
            {error && <div role="alert">{error}</div>}
        </DialogContent><DialogActions>
            <Button onClick={onClose} disabled={busy}>{t("cancel")}</Button><Button appearance="primary" onClick={() => void submit()} disabled={busy}>{title}</Button>
        </DialogActions></DialogBody></DialogSurface>
    </Dialog>;
}
