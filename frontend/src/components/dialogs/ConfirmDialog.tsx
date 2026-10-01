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

type ConfirmDialogProps = {
    open: boolean;
    title: string;
    message: string;
    confirmLabel?: string;
    cancelLabel?: string;
    onClose: (confirmed: boolean) => void;
};

function ConfirmDialog({
    open,
    title,
    message,
    confirmLabel,
    cancelLabel,
    onClose,
}: ConfirmDialogProps) {
    const { t } = useTranslation();
    confirmLabel ??= t("confirm");
    cancelLabel ??= t("cancel");
    return (
        <Dialog open={open} onOpenChange={(_, data) => {
            if (!data.open) {
                onClose(false);
            }
        }}>
            <DialogSurface>
                <DialogBody>
                    <DialogTitle>{title}</DialogTitle>
                    <DialogContent>{message}</DialogContent>
                    <DialogActions>
                        <Button onClick={() => onClose(false)}>{cancelLabel}</Button>
                        <Button appearance="primary" onClick={() => onClose(true)}>{confirmLabel}</Button>
                    </DialogActions>
                </DialogBody>
            </DialogSurface>
        </Dialog>
    );
}

export default ConfirmDialog;
