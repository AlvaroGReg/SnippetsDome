import {
    Button,
    Dialog,
    DialogActions,
    DialogBody,
    DialogContent,
    DialogSurface,
    DialogTitle,
    Input,
} from "@fluentui/react-components";
import { CheckmarkRegular, DeleteRegular, DismissRegular, DocumentArrowDownRegular, DocumentArrowUpRegular, EditRegular, LockClosedRegular, LockOpenRegular, StarFilled, StarRegular } from "@fluentui/react-icons";
import { useState } from "react";
import type { CollectionModel } from "../../models/Snippet";
import { useTranslation } from "../../i18n";

type CollectionManagementDialogProps = {
    open: boolean;
    collections: CollectionModel[];
    defaultCollectionId: string | null;
    newCollectionName: string;
    onClose: () => void;
    onNewCollectionNameChange: (name: string) => void;
    onCreate: () => void;
    onRename: (collection: CollectionModel, name: string) => Promise<boolean>;
    isJSONBusy: boolean;
    onImport: () => void;
    onExport: (collection: CollectionModel) => void;
    onDelete: (collection: CollectionModel) => void;
    onSetDefault: (collection: CollectionModel) => void;
    onSecurity: (collection: CollectionModel) => void;
};

function CollectionManagementDialog({ open, collections, defaultCollectionId, newCollectionName, isJSONBusy, onClose, onNewCollectionNameChange, onCreate, onRename, onImport, onExport, onDelete, onSetDefault, onSecurity }: CollectionManagementDialogProps) {
    const { t } = useTranslation();
    const [editingCollectionId, setEditingCollectionId] = useState<string | null>(null);
    const [editedCollectionName, setEditedCollectionName] = useState("");

    function beginEditing(collection: CollectionModel) {
        setEditingCollectionId(collection.id);
        setEditedCollectionName(collection.name);
    }

    async function saveName(collection: CollectionModel) {
        if (await onRename(collection, editedCollectionName)) {
            setEditingCollectionId(null);
            setEditedCollectionName("");
        }
    }

    return (
        <Dialog open={open} onOpenChange={(_, data) => { if (!data.open) onClose(); }}>
            <DialogSurface className="collection-management-surface">
                <DialogBody>
                    <DialogTitle
                        action={(
                            <Button
                                appearance="subtle"
                                icon={<DocumentArrowDownRegular />}
                                iconPosition="after"
                                aria-label={t("importJson")}
                                title={isJSONBusy ? t("importingJson") : t("importJson")}
                                disabled={isJSONBusy}
                                onClick={onImport}
                            >
                                {t("importJson")}
                            </Button>
                        )}
                    >
                        {t("manageCollections")}
                    </DialogTitle>
                    <DialogContent>
                        <div className="collection-management-create">
                            <Input
                                value={newCollectionName}
                                onChange={(event) => onNewCollectionNameChange(event.target.value)}
                                placeholder={t("newCollection")}
                                aria-label={t("newCollection")}
                            />
                            <Button appearance="primary" onClick={onCreate} disabled={!newCollectionName.trim()}>{t("createCollection")}</Button>
                        </div>
                        <table className="collection-management-table">
                            <tbody>
                            {collections.map((collection) => {
                                const isDefault = collection.id === defaultCollectionId;
                                return (
                                    <tr key={collection.id}>
                                        <td className="collection-management-name">
                                            {editingCollectionId === collection.id ? (
                                                <Input
                                                    value={editedCollectionName}
                                                    onChange={(_, data) => setEditedCollectionName(data.value)}
                                                    aria-label={`${t("editCollection")} ${collection.name}`}
                                                    autoFocus
                                                />
                                            ) : collection.name}
                                        </td>
                                        <td className="collection-management-actions">
                                            {editingCollectionId === collection.id ? (
                                                <>
                                                    <Button appearance="subtle" icon={<CheckmarkRegular />} aria-label={t("save")} title={t("save")} disabled={!editedCollectionName.trim()} onClick={() => void saveName(collection)} />
                                                    <Button appearance="subtle" icon={<DismissRegular />} aria-label={t("cancel")} title={t("cancel")} onClick={() => setEditingCollectionId(null)} />
                                                </>
                                            ) : (
                                                <>
                                                    <Button appearance="subtle" icon={<EditRegular />} aria-label={`${t("editCollection")} ${collection.name}`} title={t("editCollection")} onClick={() => beginEditing(collection)} />
                                                    <Button
                                                        appearance="subtle"
                                                        icon={collection.encrypted ? <LockOpenRegular /> : <LockClosedRegular />}
                                                        aria-label={collection.encrypted ? t("decryptList") : t("encryptList")}
                                                        title={collection.encrypted ? t("decryptList") : t("encryptList")}
                                                        onClick={() => onSecurity(collection)}
                                                    />
                                                    <Button
                                                        appearance="subtle"
                                                        icon={<DocumentArrowUpRegular />}
                                                        aria-label={`${t("exportJson")} ${collection.name}`}
                                                        title={isJSONBusy ? t("exportingJson") : t("exportJson")}
                                                        disabled={isJSONBusy}
                                                        onClick={() => onExport(collection)}
                                                    />
                                                    <Button
                                                        appearance="subtle"
                                                        className={isDefault ? "collection-default-button" : undefined}
                                                        icon={isDefault ? <StarFilled /> : <StarRegular />}
                                                        aria-label={`${isDefault ? t("defaultCollection") : t("setAsDefault")} ${collection.name}`}
                                                        title={isDefault ? t("defaultCollection") : t("setAsDefault")}
                                                        disabled={isDefault}
                                                        onClick={() => onSetDefault(collection)}
                                                    />
                                                    <Button
                                                        appearance="subtle"
                                                        icon={<DeleteRegular />}
                                                        aria-label={`${t("deleteCollection")} ${collection.name}`}
                                                        title={t("deleteCollection")}
                                                        disabled={collections.length <= 1}
                                                        onClick={() => onDelete(collection)}
                                                    />
                                                </>
                                            )}
                                        </td>
                                    </tr>
                                );
                            })}
                            </tbody>
                        </table>
                    </DialogContent>
                    <DialogActions>
                        <Button appearance="primary" onClick={onClose}>{t("done")}</Button>
                    </DialogActions>
                </DialogBody>
            </DialogSurface>
        </Dialog>
    );
}

export default CollectionManagementDialog;
