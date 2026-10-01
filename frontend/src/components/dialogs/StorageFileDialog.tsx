import {
    Button,
    Dialog,
    DialogActions,
    DialogBody,
    DialogContent,
    DialogSurface,
    DialogTitle,
} from "@fluentui/react-components";
import { useTranslation } from "../../i18n";

type StorageFileDialogProps = {
    open: boolean;
    onClose: () => void;
    onPickExisting: () => void;
    onCreateNew: () => void;
};

function StorageFileDialog({ open, onClose, onPickExisting, onCreateNew }: StorageFileDialogProps) {
    const { t } = useTranslation();
    return (
        <Dialog open={open} onOpenChange={(_, data) => !data.open && onClose()}>
            <DialogSurface>
                <DialogBody>
                    <DialogTitle>{t("snippetsFile")}</DialogTitle>
                    <DialogContent>
                        {t("chooseSnippetsFile")}
                    </DialogContent>
                    <DialogActions>
                        <Button appearance="secondary" onClick={onClose}>{t("cancel")}</Button>
                        <Button appearance="secondary" onClick={onPickExisting}>{t("chooseFile")}</Button>
                        <Button appearance="primary" onClick={onCreateNew}>{t("createFile")}</Button>
                    </DialogActions>
                </DialogBody>
            </DialogSurface>
        </Dialog>
    );
}

export default StorageFileDialog;
