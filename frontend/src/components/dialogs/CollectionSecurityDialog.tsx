import { Button, Dialog, DialogActions, DialogBody, DialogContent, DialogSurface, DialogTitle, Input } from "@fluentui/react-components";
import { useState } from "react";
import type { CollectionModel } from "../../models/Snippet";
import { useTranslation } from "../../i18n";

export type SecurityMode = "unlock" | "encrypt" | "decrypt";
type Props = { collection: CollectionModel; mode: SecurityMode; onClose: () => void; onSubmit: (pin: string, confirmation: string) => Promise<void> };

export default function CollectionSecurityDialog({ collection, mode, onClose, onSubmit }: Props) {
    const { t } = useTranslation();
    const [pin, setPin] = useState("");
    const [confirmation, setConfirmation] = useState("");
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState("");
    const title = mode === "unlock" ? t("unlockList") : mode === "encrypt" ? t("encryptList") : t("decryptList");

    async function submit() {
        if (!/^\d{4}$/.test(pin) || (mode === "encrypt" && pin !== confirmation)) {
            setError(mode === "encrypt" && pin !== confirmation ? t("pinMismatch") : t("pinFormat"));
            return;
        }
        setBusy(true); setError("");
        try { await onSubmit(pin, confirmation); onClose(); }
        catch { setError(t("collectionSecurityFailed")); }
        finally { setBusy(false); setPin(""); setConfirmation(""); }
    }

    return <Dialog open onOpenChange={(_, data) => { if (!data.open && !busy) onClose(); }}>
        <DialogSurface><DialogBody><DialogTitle>{title}: {collection.name}</DialogTitle><DialogContent>
            <Input type="password" value={pin} maxLength={4} inputMode="numeric" aria-label={t("pin")} placeholder={t("pinPlaceholder")} onChange={(_, data) => setPin(data.value)} />
            {mode === "encrypt" && <Input type="password" value={confirmation} maxLength={4} inputMode="numeric" aria-label={t("confirmPin")} placeholder={t("confirmPin")} onChange={(_, data) => setConfirmation(data.value)} />}
            {error && <div role="alert">{error}</div>}
        </DialogContent><DialogActions>
            <Button onClick={onClose} disabled={busy}>{t("cancel")}</Button><Button appearance="primary" onClick={() => void submit()} disabled={busy}>{title}</Button>
        </DialogActions></DialogBody></DialogSurface>
    </Dialog>;
}
