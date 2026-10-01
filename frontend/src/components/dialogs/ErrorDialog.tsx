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

type ErrorDialogProps = {
    error: string;
    onClose: () => void;
};

function ErrorDialog({ error, onClose }: ErrorDialogProps) {
    const { t } = useTranslation();
    return (
        <Dialog
            open={Boolean(error)}
            modalType="alert"
            onOpenChange={(_, data) => {
                if (!data.open) {
                    onClose();
                }
            }}
        >
            <DialogSurface>
                <DialogBody>
                    <DialogTitle>{t("error")}</DialogTitle>
                    <DialogContent>{error}</DialogContent>
                    <DialogActions>
                        <Button appearance="primary" onClick={onClose}>{t("close")}</Button>
                    </DialogActions>
                </DialogBody>
            </DialogSurface>
        </Dialog>
    );
}

export default ErrorDialog;
