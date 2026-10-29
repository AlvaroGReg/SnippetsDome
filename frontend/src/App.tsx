import './App.css';
import SnippetsList from "./components/snippetsList/SnippetsList";
import SearchBar from "./components/searchbar/SearchBar";
import ConfirmDialog from "./components/dialogs/ConfirmDialog";
import ErrorDialog from "./components/dialogs/ErrorDialog";
import SnippetEditorDialog from "./components/dialogs/SnippetEditorDialog";
import SettingsDialog from "./components/dialogs/SettingsDialog";
import CollectionManagementDialog from "./components/dialogs/CollectionManagementDialog";
import CollectionSecurityDialog, { type SecurityMode } from "./components/dialogs/CollectionSecurityDialog";
import CollectionExchangeDialog from "./components/dialogs/CollectionExchangeDialog";
import { Button, Select, Spinner } from "@fluentui/react-components";
import { useSnippets } from "./hooks/use-snippets";
import { useEffect, useMemo, useState } from "react";
import { AddRegular, BrightnessHighRegular, DarkThemeRegular, LockClosedRegular, LockOpenRegular, SettingsRegular } from "@fluentui/react-icons";
import type { CollectionModel, CreateSnippetInput, SnippetModel } from "./models/Snippet";
import * as snippetsService from "./services/snippets-service";
import { useTranslation, type LanguagePreference } from "./i18n";

type AppProps = {
    isDarkTheme: boolean;
    onToggleTheme: () => void;
    language: LanguagePreference;
    onLanguageChange: (language: LanguagePreference) => void;
};

const MAX_COLLECTION_NAME_LENGTH = 25;

function getCollectionDisplayName(name: string) {
    return name.length > MAX_COLLECTION_NAME_LENGTH
        ? `${name.slice(0, MAX_COLLECTION_NAME_LENGTH - 1)}…`
        : name;
}

function App({ isDarkTheme, onToggleTheme, language, onLanguageChange }: AppProps) {
    const { t } = useTranslation();
    const [searchQuery, setSearchQuery] = useState("");
    const [snippetPendingDeletion, setSnippetPendingDeletion] = useState<string | null>(null);
    const [snippetBeingEdited, setSnippetBeingEdited] = useState<SnippetModel | null | undefined>(undefined);
    const [collections, setCollections] = useState<CollectionModel[]>([]);
    const [activeCollection, setActiveCollection] = useState<CollectionModel | null>(null);
    const [defaultCollection, setDefaultCollection] = useState<CollectionModel | null>(null);
    const [collectionPendingDeletion, setCollectionPendingDeletion] = useState<CollectionModel | null>(null);
    const [isCollectionManagementOpen, setIsCollectionManagementOpen] = useState(false);
    const [newCollectionName, setNewCollectionName] = useState("");
    const [isSettingsDialogOpen, setIsSettingsDialogOpen] = useState(false);
    const [closeToTrayEnabled, setCloseToTrayEnabled] = useState(false);
    const [startAtLoginEnabled, setStartAtLoginEnabled] = useState(false);
    const [startAtLoginSupported, setStartAtLoginSupported] = useState(false);
    const [traySnippetLimit, setTraySnippetLimit] = useState(5);
    const [settingsError, setSettingsError] = useState("");
    const [isJSONBusy, setIsJSONBusy] = useState(false);
    const [securityRequest, setSecurityRequest] = useState<{ collection: CollectionModel; mode: SecurityMode } | null>(null);
    const [exchangeRequest, setExchangeRequest] = useState<{ mode: "import" | "export"; collection?: CollectionModel } | null>(null);
    const {
        snippets,
        error,
        clearError,
        clearSnippets,
        isLoading,
        reload,
        createSnippet,
        updateSnippet,
        deleteSnippet,
    } = useSnippets();

    useEffect(() => {
        void Promise.all([snippetsService.getCollections(), snippetsService.getActiveCollection(), snippetsService.getDefaultCollection()])
            .then(([loadedCollections, loadedActiveCollection, loadedDefaultCollection]) => {
                setCollections(loadedCollections);
                const active = loadedActiveCollection.encrypted ? { ...loadedActiveCollection, locked: true } : loadedActiveCollection;
                setActiveCollection(active);
                setDefaultCollection(loadedDefaultCollection);
                if (active.encrypted) setSecurityRequest({ collection: active, mode: "unlock" });
            })
            .catch((requestError: unknown) => {
                setSettingsError(requestError instanceof Error ? requestError.message : t("unableToLoadCollections"));
            });
        void snippetsService.getCloseToTrayEnabled()
            .then(setCloseToTrayEnabled)
            .catch((requestError: unknown) => {
                    setSettingsError(requestError instanceof Error ? requestError.message : t("unableToLoadSettings"));
            });
        void snippetsService.getTraySnippetLimit()
            .then(setTraySnippetLimit)
            .catch((requestError: unknown) => {
                    setSettingsError(requestError instanceof Error ? requestError.message : t("unableToLoadSettings"));
            });
        void snippetsService.getStartAtLoginSupported()
            .then(setStartAtLoginSupported)
            .catch((requestError: unknown) => {
                    setSettingsError(requestError instanceof Error ? requestError.message : t("unableToLoadSettings"));
            });
        void snippetsService.getStartAtLoginEnabled()
            .then(setStartAtLoginEnabled)
            .catch((requestError: unknown) => {
                    setSettingsError(requestError instanceof Error ? requestError.message : t("unableToLoadSettings"));
            });
    }, []);

    async function changeCollection(id: string) {
        const selected = collections.find((collection) => collection.id === id) ?? null;
        if (!selected) return;
        clearSnippets();
        setActiveCollection(selected.encrypted ? { ...selected, name: `${t("lockedCollection")} (${selected.id.slice(0, 8)})`, locked: true } : selected);
        setCollections((current) => current.map((item) => item.encrypted ? { ...item, name: `${t("lockedCollection")} (${item.id.slice(0, 8)})`, locked: true } : item));
        try {
            // Selecting first clears the previous collection key before a PIN
            // dialog is shown, including when the selected collection is locked.
            await snippetsService.selectCollection(id);
            if (selected.encrypted) {
                const lockedCollection = { ...selected, name: `${t("lockedCollection")} (${selected.id.slice(0, 8)})`, locked: true };
                setActiveCollection(lockedCollection);
                setSecurityRequest({ collection: lockedCollection, mode: "unlock" });
            } else {
                setCollections(await snippetsService.getCollections());
                await reload();
            }
        } catch (requestError) {
            setSettingsError(requestError instanceof Error ? requestError.message : t("unableToSelectCollection"));
        }
    }

    async function addCollection() {
        try {
            const collection = await snippetsService.createCollection(newCollectionName);
            setCollections((current) => [...current, collection]);
            setActiveCollection(collection);
            setNewCollectionName("");
            await reload();
        } catch (requestError) {
            setSettingsError(requestError instanceof Error ? requestError.message : t("unableToCreateCollection"));
        }
    }

    const filteredSnippets = useMemo(() => {
        const query = searchQuery.trim().toLocaleLowerCase();

        if (!query) {
            return snippets;
        }

        return snippets.filter((snippet) =>
            [snippet.title, snippet.language, snippet.code, ...snippet.tags]
                .some((field) => field.toLocaleLowerCase().includes(query)),
        );
    }, [searchQuery, snippets]);

    function saveSnippet(input: CreateSnippetInput) {
        if (snippetBeingEdited) {
            void updateSnippet({ ...snippetBeingEdited, ...input });
        } else {
            void createSnippet(input);
        }

        setSnippetBeingEdited(undefined);
    }

    function handleDeleteConfirmation(confirmed: boolean) {
        if (confirmed && snippetPendingDeletion) {
            void deleteSnippet(snippetPendingDeletion);
        }

        setSnippetPendingDeletion(null);
    }

    async function handleCloseToTrayChange(enabled: boolean) {
        try {
            setSettingsError("");
            await snippetsService.setCloseToTrayEnabled(enabled);
            setCloseToTrayEnabled(enabled);
        } catch (requestError) {
            setSettingsError(requestError instanceof Error ? requestError.message : t("unableToSaveSettings"));
        }
    }

    async function handleTraySnippetLimitChange(limit: number): Promise<boolean> {
        try {
            setSettingsError("");
            await snippetsService.setTraySnippetLimit(limit);
            setTraySnippetLimit(limit);
            return true;
        } catch (requestError) {
            setSettingsError(requestError instanceof Error ? requestError.message : t("unableToSaveSettings"));
            return false;
        }
    }

    async function handleStartAtLoginChange(enabled: boolean) {
        try {
            setSettingsError("");
            await snippetsService.setStartAtLoginEnabled(enabled);
            setStartAtLoginEnabled(enabled);
        } catch (requestError) {
            setSettingsError(requestError instanceof Error ? requestError.message : t("unableToSaveSettings"));
        }
    }

    async function handleLanguageChange(nextLanguage: LanguagePreference) {
        try {
            setSettingsError("");
            await snippetsService.setLanguage(nextLanguage);
            onLanguageChange(nextLanguage);
        } catch (requestError) {
            setSettingsError(requestError instanceof Error ? requestError.message : t("unableToSaveSettings"));
        }
    }

    async function beginJSONImport() {
        try {
            setIsJSONBusy(true);
            setSettingsError("");
            const result = await snippetsService.beginJSONImport();
            if (result.needsPassword) {
                setExchangeRequest({ mode: "import" });
                return;
            }
            const collection = result.collection;
            if (!collection?.id) {
                return;
            }
            const loadedCollections = await snippetsService.getCollections();
            setCollections(loadedCollections);
            setActiveCollection(collection);
            await reload();
        } catch (requestError) {
            throw requestError;
        } finally {
            setIsJSONBusy(false);
        }
    }

    async function importEncryptedJSON(password: string) {
        try {
            setIsJSONBusy(true);
            setSettingsError("");
            const collection = await snippetsService.importPendingEncryptedJSON(password);
            const loadedCollections = await snippetsService.getCollections();
            setCollections(loadedCollections);
            setActiveCollection(collection);
            await reload();
        } finally {
            setIsJSONBusy(false);
        }
    }

    async function exportCollectionJSON(collection: CollectionModel, collectionPassword: string, password: string) {
        try {
            setIsJSONBusy(true);
            setSettingsError("");
            if (collection.encrypted) await snippetsService.exportEncryptedCollectionJSON(collection.id, collectionPassword, password);
            else await snippetsService.exportCollectionJSON(collection.id);
        } catch (requestError) {
            throw requestError;
        } finally {
            setIsJSONBusy(false);
        }
    }

    async function submitSecurity(pin: string, confirmation: string) {
        if (!securityRequest) return;
        const { collection, mode } = securityRequest;
        if (mode === "unlock") {
            const unlocked = await snippetsService.unlockCollection(collection.id, pin);
            setActiveCollection({ ...unlocked, locked: false });
            const all = await snippetsService.getCollections(); setCollections(all);
            await reload();
        } else if (mode === "encrypt") {
            await snippetsService.encryptCollection(collection.id, pin, confirmation);
            const all = await snippetsService.getCollections(); setCollections(all);
            if (activeCollection?.id === collection.id) { setActiveCollection({ ...collection, encrypted: true }); await reload(); }
        } else {
            await snippetsService.decryptCollection(collection.id, pin);
            const all = await snippetsService.getCollections(); setCollections(all);
            if (activeCollection?.id === collection.id) { setActiveCollection({ ...collection, encrypted: false }); await reload(); }
        }
        setSecurityRequest(null);
    }

    async function lockActive() {
        if (!activeCollection) return;
        clearSnippets();
        await snippetsService.lockActiveCollection();
        setActiveCollection({ ...activeCollection, name: `${t("lockedCollection")} (${activeCollection.id.slice(0, 8)})`, locked: true });
        setCollections((current) => current.map((item) => item.id === activeCollection.id ? { ...item, name: `${t("lockedCollection")} (${item.id.slice(0, 8)})`, locked: true } : item));
    }

    async function markCollectionDefault(collection: CollectionModel) {
        try {
            await snippetsService.setDefaultCollection(collection.id);
            setDefaultCollection(collection);
        } catch (requestError) {
            setSettingsError(requestError instanceof Error ? requestError.message : t("unableToSetDefaultCollection"));
        }
    }

    async function renameCollection(collection: CollectionModel, name: string): Promise<boolean> {
        try {
            await snippetsService.renameCollection(collection.id, name);
            const renamed = { ...collection, name: name.trim() };
            setCollections((current) => current.map((item) => item.id === collection.id ? renamed : item));
            setActiveCollection((current) => current?.id === collection.id ? renamed : current);
            setDefaultCollection((current) => current?.id === collection.id ? renamed : current);
            return true;
        } catch (requestError) {
            setSettingsError(requestError instanceof Error ? requestError.message : t("unableToRenameCollection"));
            return false;
        }
    }

    async function confirmCollectionDeletion(confirmed: boolean) {
        const collection = collectionPendingDeletion;
        setCollectionPendingDeletion(null);
        if (!confirmed || !collection) return;
        try {
            await snippetsService.deleteCollection(collection.id);
            const [loadedCollections, loadedActiveCollection, loadedDefaultCollection] = await Promise.all([
                snippetsService.getCollections(), snippetsService.getActiveCollection(), snippetsService.getDefaultCollection(),
            ]);
            setCollections(loadedCollections);
            setActiveCollection(loadedActiveCollection);
            setDefaultCollection(loadedDefaultCollection);
            if (activeCollection?.id === collection.id) await reload();
        } catch (requestError) {
            setSettingsError(requestError instanceof Error ? requestError.message : t("unableToDeleteCollection"));
        }
    }

    return (
        <main id="app" className="main-body">
            <header className="main-header">
                <SearchBar value={searchQuery} onChange={setSearchQuery} />
                <Select value={activeCollection?.id ?? ""} onChange={(event) => void changeCollection(event.target.value)} aria-label={t("activeCollection")}>
                    {collections.map((collection) => <option key={collection.id} value={collection.id} title={collection.name}>
                        {collection.encrypted ? `🔒 ${getCollectionDisplayName(collection.name)}` : getCollectionDisplayName(collection.name)}
                    </option>)}
                </Select>
                {activeCollection?.encrypted && <Button
                    appearance="subtle"
                    icon={activeCollection.locked ? <LockOpenRegular /> : <LockClosedRegular />}
                    aria-label={activeCollection.locked ? t("unlockList") : t("lockList")}
                    title={activeCollection.locked ? t("unlockList") : t("lockList")}
                    onClick={() => activeCollection.locked
                        ? setSecurityRequest({ collection: activeCollection, mode: "unlock" })
                        : void lockActive()}
                />}
                <Button
                    appearance="primary"
                    aria-label={t("createSnippet")}
                    icon={<AddRegular />}
                    onClick={() => setSnippetBeingEdited(null)}
                    title={t("createSnippet")}
                />
            </header>
            {isLoading && <Spinner label={t("loadingSnippets")} />}
            {((!activeCollection?.encrypted && (!isLoading || snippets.length > 0)) || (activeCollection?.encrypted && securityRequest === null && snippets.length > 0)) && (
                <SnippetsList
                    snippets={filteredSnippets}
                    onEdit={setSnippetBeingEdited}
                    onDelete={setSnippetPendingDeletion}
                    onToggleFavorite={(snippet) => void updateSnippet({ ...snippet, favorite: !snippet.favorite })}
                />
            )}
            <footer className="main-footer">
                <div className="footer-group footer-options">
                    <Button
                        appearance="subtle"
                        className="settings-button"
                        icon={<SettingsRegular />}
                        onClick={() => setIsSettingsDialogOpen(true)}
                        aria-label={t("settings")}
                        title={t("settings")}
                    />
                </div>
                <div className="footer-group footer-collection">
                    <Button appearance="subtle" onClick={() => setIsCollectionManagementOpen(true)}>{t("manageCollections")}</Button>
                </div>
                <div className="footer-group footer-theme">
                    <Button
                        appearance="subtle"
                        className="theme-toggle-button"
                        icon={isDarkTheme ? <BrightnessHighRegular /> : <DarkThemeRegular />}
                        onClick={onToggleTheme}
                        aria-label={isDarkTheme ? t("switchToLight") : t("switchToDark")}
                        title={isDarkTheme ? t("switchToLight") : t("switchToDark")}
                    />
                </div>
            </footer>
            <ConfirmDialog
                open={snippetPendingDeletion !== null}
                title={t("deleteSnippet")}
                message={t("deleteConfirmation")}
                confirmLabel={t("delete")}
                onClose={handleDeleteConfirmation}
            />
            <CollectionManagementDialog
                open={isCollectionManagementOpen}
                collections={collections}
                defaultCollectionId={defaultCollection?.id ?? null}
                newCollectionName={newCollectionName}
                onClose={() => setIsCollectionManagementOpen(false)}
                onNewCollectionNameChange={setNewCollectionName}
                onCreate={() => void addCollection()}
                onRename={renameCollection}
                onUnlockForRename={(collection) => setSecurityRequest({ collection, mode: "unlock" })}
                isJSONBusy={isJSONBusy}
                onImport={() => void beginJSONImport()}
                onExport={(collection) => setExchangeRequest({ mode: "export", collection })}
                onDelete={setCollectionPendingDeletion}
                onSetDefault={(collection) => void markCollectionDefault(collection)}
                onSecurity={(collection) => setSecurityRequest({ collection, mode: collection.encrypted ? "decrypt" : "encrypt" })}
            />
            {exchangeRequest && <CollectionExchangeDialog
                mode={exchangeRequest.mode}
                collection={exchangeRequest.collection}
                onClose={() => {
                    if (exchangeRequest.mode === "import") snippetsService.cancelPendingJSONImport();
                    setExchangeRequest(null);
                }}
                onSubmit={(collectionPassword, password) => exchangeRequest.mode === "import"
                    ? importEncryptedJSON(password)
                    : exportCollectionJSON(exchangeRequest.collection!, collectionPassword, password)}
            />}
            {securityRequest && <CollectionSecurityDialog collection={securityRequest.collection} mode={securityRequest.mode} onClose={() => setSecurityRequest(null)} onSubmit={submitSecurity} />}
            <ConfirmDialog
                open={collectionPendingDeletion !== null}
                title={t("deleteCollection")}
                message={t("deleteCollectionConfirmation")}
                confirmLabel={t("delete")}
                onClose={(confirmed) => void confirmCollectionDeletion(confirmed)}
            />
            <SnippetEditorDialog
                open={snippetBeingEdited !== undefined}
                snippet={snippetBeingEdited ?? undefined}
                onClose={() => setSnippetBeingEdited(undefined)}
                onSave={saveSnippet}
            />
            <SettingsDialog
                open={isSettingsDialogOpen}
                closeToTrayEnabled={closeToTrayEnabled}
                startAtLoginEnabled={startAtLoginEnabled}
                startAtLoginSupported={startAtLoginSupported}
                traySnippetLimit={traySnippetLimit}
                language={language}
                onClose={() => setIsSettingsDialogOpen(false)}
                onCloseToTrayChange={(enabled) => void handleCloseToTrayChange(enabled)}
                onStartAtLoginChange={(enabled) => void handleStartAtLoginChange(enabled)}
                onTraySnippetLimitChange={handleTraySnippetLimitChange}
                onLanguageChange={(nextLanguage) => void handleLanguageChange(nextLanguage)}
            />
            <ErrorDialog
                error={error || settingsError}
                onClose={() => {
                    clearError();
                    setSettingsError("");
                }}
            />
        </main>
    );
}

export default App;
