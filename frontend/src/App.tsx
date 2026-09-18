import './App.css';
import SnippetsList from "./components/snippetsList/SnippetsList";
import SearchBar from "./components/searchbar/SearchBar";
import ConfirmDialog from "./components/dialogs/ConfirmDialog";
import ErrorDialog from "./components/dialogs/ErrorDialog";
import SnippetEditorDialog from "./components/dialogs/SnippetEditorDialog";
import StorageFileDialog from "./components/dialogs/StorageFileDialog";
import SettingsDialog from "./components/dialogs/SettingsDialog";
import { Button, Spinner } from "@fluentui/react-components";
import { useSnippets } from "./hooks/use-snippets";
import { useEffect, useMemo, useState } from "react";
import { AddRegular, BrightnessHighRegular, DarkThemeRegular, SettingsRegular } from "@fluentui/react-icons";
import type { CreateSnippetInput, SnippetModel } from "./models/Snippet";
import * as snippetsService from "./services/snippets-service";

type AppProps = {
    isDarkTheme: boolean;
    onToggleTheme: () => void;
};

function App({ isDarkTheme, onToggleTheme }: AppProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const [snippetPendingDeletion, setSnippetPendingDeletion] = useState<string | null>(null);
    const [snippetBeingEdited, setSnippetBeingEdited] = useState<SnippetModel | null | undefined>(undefined);
    const [isStorageFileDialogOpen, setIsStorageFileDialogOpen] = useState(false);
    const [isSettingsDialogOpen, setIsSettingsDialogOpen] = useState(false);
    const [closeToTrayEnabled, setCloseToTrayEnabled] = useState(false);
    const [startAtLoginEnabled, setStartAtLoginEnabled] = useState(false);
    const [startAtLoginSupported, setStartAtLoginSupported] = useState(false);
    const [traySnippetLimit, setTraySnippetLimit] = useState(5);
    const [settingsError, setSettingsError] = useState("");
    const {
        snippets,
        error,
        clearError,
        isLoading,
        storagePath,
        pickExistingStorageFile,
        createStorageFile,
        createSnippet,
        updateSnippet,
        deleteSnippet,
    } = useSnippets();

    useEffect(() => {
        void snippetsService.getCloseToTrayEnabled()
            .then(setCloseToTrayEnabled)
            .catch((requestError: unknown) => {
                setSettingsError(requestError instanceof Error ? requestError.message : "Unable to load settings.");
            });
        void snippetsService.getTraySnippetLimit()
            .then(setTraySnippetLimit)
            .catch((requestError: unknown) => {
                setSettingsError(requestError instanceof Error ? requestError.message : "Unable to load settings.");
            });
        void snippetsService.getStartAtLoginSupported()
            .then(setStartAtLoginSupported)
            .catch((requestError: unknown) => {
                setSettingsError(requestError instanceof Error ? requestError.message : "Unable to load settings.");
            });
        void snippetsService.getStartAtLoginEnabled()
            .then(setStartAtLoginEnabled)
            .catch((requestError: unknown) => {
                setSettingsError(requestError instanceof Error ? requestError.message : "Unable to load settings.");
            });
    }, []);

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
            setSettingsError(requestError instanceof Error ? requestError.message : "Unable to save settings.");
        }
    }

    async function handleTraySnippetLimitChange(limit: number): Promise<boolean> {
        try {
            setSettingsError("");
            await snippetsService.setTraySnippetLimit(limit);
            setTraySnippetLimit(limit);
            return true;
        } catch (requestError) {
            setSettingsError(requestError instanceof Error ? requestError.message : "Unable to save settings.");
            return false;
        }
    }

    async function handleStartAtLoginChange(enabled: boolean) {
        try {
            setSettingsError("");
            await snippetsService.setStartAtLoginEnabled(enabled);
            setStartAtLoginEnabled(enabled);
        } catch (requestError) {
            setSettingsError(requestError instanceof Error ? requestError.message : "Unable to save settings.");
        }
    }

    return (
        <main id="app" className="main-body">
            <header className="main-header">
                <SearchBar value={searchQuery} onChange={setSearchQuery} />
                <Button
                    appearance="primary"
                    aria-label="Create snippet"
                    icon={<AddRegular />}
                    onClick={() => setSnippetBeingEdited(null)}
                    title="Create snippet"
                />
            </header>
            {isLoading && <Spinner label="Loading snippets" />}
            {(!isLoading || snippets.length > 0) && (
                <SnippetsList
                    snippets={filteredSnippets}
                    onEdit={setSnippetBeingEdited}
                    onDelete={setSnippetPendingDeletion}
                    onToggleFavorite={(snippet) => void updateSnippet({ ...snippet, favorite: !snippet.favorite })}
                />
            )}
            <footer className='main-footer'>
                <Button
                    appearance="subtle"
                    className="settings-button"
                    icon={<SettingsRegular />}
                    onClick={() => setIsSettingsDialogOpen(true)}
                    aria-label="Open settings"
                    title="Settings"
                />
                <Button
                    appearance="subtle"
                    className="storage-file-button"
                    onClick={() => setIsStorageFileDialogOpen(true)}
                    title={storagePath || "No file selected."}
                >
                    {storagePath || "No file selected."}
                </Button>
                <Button
                    appearance="subtle"
                    className="theme-toggle-button"
                    icon={isDarkTheme ? <BrightnessHighRegular /> : <DarkThemeRegular />}
                    onClick={onToggleTheme}
                    aria-label={`Switch to ${isDarkTheme ? "light" : "dark"} theme`}
                    title={`Switch to ${isDarkTheme ? "light" : "dark"} theme`}
                />
            </footer>
            <ConfirmDialog
                open={snippetPendingDeletion !== null}
                title="Delete snippet"
                message="Are you sure you want to delete this snippet?"
                confirmLabel="Delete"
                onClose={handleDeleteConfirmation}
            />
            <SnippetEditorDialog
                open={snippetBeingEdited !== undefined}
                snippet={snippetBeingEdited ?? undefined}
                onClose={() => setSnippetBeingEdited(undefined)}
                onSave={saveSnippet}
            />
            <StorageFileDialog
                open={isStorageFileDialogOpen}
                onClose={() => setIsStorageFileDialogOpen(false)}
                onPickExisting={() => {
                    setIsStorageFileDialogOpen(false);
                    void pickExistingStorageFile();
                }}
                onCreateNew={() => {
                    setIsStorageFileDialogOpen(false);
                    void createStorageFile();
                }}
            />
            <SettingsDialog
                open={isSettingsDialogOpen}
                closeToTrayEnabled={closeToTrayEnabled}
                startAtLoginEnabled={startAtLoginEnabled}
                startAtLoginSupported={startAtLoginSupported}
                traySnippetLimit={traySnippetLimit}
                onClose={() => setIsSettingsDialogOpen(false)}
                onCloseToTrayChange={(enabled) => void handleCloseToTrayChange(enabled)}
                onStartAtLoginChange={(enabled) => void handleStartAtLoginChange(enabled)}
                onTraySnippetLimitChange={handleTraySnippetLimitChange}
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
