import { createContext, useContext } from "react";

export type LanguagePreference = "en" | "es";
export type Locale = "en" | "es";

type TranslationKey = keyof typeof translations.en;

const translations = {
    en: {
        settings: "Settings", closeToTray: "Close to tray", closeToTrayDescription: "Keep SnippetsDome running in the notification area when its window is closed.",
        startAtLogin: "Start at login", startAtLoginDescription: "Launch SnippetsDome automatically when you sign in.", unavailablePlatforms: "Available on Windows, Linux, and macOS.",
        snippetsShownInTray: "Snippets shown in tray", done: "Done", language: "Language", english: "English", spanish: "Español",
        searchSnippets: "Search snippets", clearSearch: "Clear search", createSnippet: "Create snippet", loadingSnippets: "Loading snippets", activeCollection: "Active collection", newCollection: "New collection", createCollection: "Create collection", importJson: "Import JSON", exportJson: "Export JSON", importingJson: "Importing JSON...", exportingJson: "Exporting JSON...", snippetsFile: "Snippets file", chooseSnippetsFile: "Choose an existing JSON file or create a new one with the name you want.", chooseFile: "Choose file", createFile: "Create file",
        manageCollections: "Manage lists", defaultCollection: "Default", setAsDefault: "Set as default", editCollection: "Edit list", deleteCollection: "Delete list", deleteCollectionConfirmation: "Are you sure you want to delete this list and all its snippets?", unableToDeleteCollection: "Unable to delete the list.", unableToSetDefaultCollection: "Unable to set the default list.", unableToRenameCollection: "Unable to rename the list.", unableToExportCollection: "Unable to export the list.",
        switchToLight: "Switch to light theme", switchToDark: "Switch to dark theme", deleteSnippet: "Delete snippet", deleteConfirmation: "Are you sure you want to delete this snippet?", delete: "Delete",
        cancel: "Cancel", confirm: "Confirm", error: "Error", close: "Close", emptyList: "Empty list", copy: "Copy", edit: "Edit", addToFavorites: "Add to favorites", removeFromFavorites: "Remove from favorites",
        codeCopied: "Code copied to clipboard", codeCopyFailed: "Could not copy the code", editSnippet: "Edit snippet", title: "Title", languageField: "Language", code: "Code", tags: "Tags", tagsHint: "Separate tags with commas", save: "Save", create: "Create",
        unableToLoadSettings: "Unable to load settings.", unableToSaveSettings: "Unable to save settings.", unableToLoadCollections: "Unable to load collections.", unableToSelectCollection: "Unable to select the collection.", unableToCreateCollection: "Unable to create the collection.", unableToImportJson: "Unable to import the JSON file.", unableToExportJson: "Unable to export the JSON file.",
        encryptList: "Encrypt list", unlockList: "Unlock list", decryptList: "Remove encryption", pin: "Four-digit PIN", pinPlaceholder: "0000", confirmPin: "Confirm PIN", pinFormat: "Enter exactly four digits.", pinMismatch: "PINs do not match.", collectionSecurityFailed: "The list could not be secured or unlocked. Check the PIN and try again.", lockedCollection: "Locked list", encryptedCollection: "Encrypted list", lockList: "Lock list",
        currentListPassword: "Current list password", exportPassword: "New export password", importPassword: "Password for encrypted JSON", exchangeCredentialsRequired: "Enter the current four-digit PIN and a new export password.",
    },
    es: {
        settings: "Opciones", closeToTray: "Cerrar en la bandeja", closeToTrayDescription: "Mantener SnippetsDome en el área de notificación al cerrar su ventana.",
        startAtLogin: "Iniciar al iniciar sesión", startAtLoginDescription: "Iniciar SnippetsDome automáticamente al iniciar sesión.", unavailablePlatforms: "Disponible en Windows, Linux y macOS.",
        snippetsShownInTray: "Snippets mostrados en la bandeja", done: "Listo", language: "Idioma", english: "English", spanish: "Español",
        searchSnippets: "Buscar snippets", clearSearch: "Borrar búsqueda", createSnippet: "Crear snippet", loadingSnippets: "Cargando snippets", activeCollection: "Colección activa", newCollection: "Nueva colección", createCollection: "Crear colección", importJson: "Importar JSON", exportJson: "Exportar JSON", importingJson: "Importando JSON...", exportingJson: "Exportando JSON...", snippetsFile: "Archivo de snippets", chooseSnippetsFile: "Elige un archivo JSON existente o crea uno nuevo con el nombre que quieras.", chooseFile: "Elegir archivo", createFile: "Crear archivo",
        manageCollections: "Gestionar listas", defaultCollection: "Por defecto", setAsDefault: "Marcar por defecto", editCollection: "Editar lista", deleteCollection: "Eliminar lista", deleteCollectionConfirmation: "¿Seguro que quieres eliminar esta lista y todos sus snippets?", unableToDeleteCollection: "No se pudo eliminar la lista.", unableToSetDefaultCollection: "No se pudo establecer la lista por defecto.", unableToRenameCollection: "No se pudo cambiar el nombre de la lista.", unableToExportCollection: "No se pudo exportar la lista.",
        switchToLight: "Cambiar al tema claro", switchToDark: "Cambiar al tema oscuro", deleteSnippet: "Eliminar snippet", deleteConfirmation: "¿Seguro que quieres eliminar este snippet?", delete: "Eliminar",
        cancel: "Cancelar", confirm: "Confirmar", error: "Error", close: "Cerrar", emptyList: "Lista vacía", copy: "Copiar", edit: "Editar", addToFavorites: "Añadir a favoritos", removeFromFavorites: "Quitar de favoritos",
        codeCopied: "Código copiado al portapapeles", codeCopyFailed: "No se pudo copiar el código", editSnippet: "Editar snippet", title: "Título", languageField: "Lenguaje", code: "Código", tags: "Etiquetas", tagsHint: "Separa las etiquetas con comas", save: "Guardar", create: "Crear",
        unableToLoadSettings: "No se pudieron cargar las opciones.", unableToSaveSettings: "No se pudieron guardar las opciones.", unableToLoadCollections: "No se pudieron cargar las colecciones.", unableToSelectCollection: "No se pudo seleccionar la colección.", unableToCreateCollection: "No se pudo crear la colección.", unableToImportJson: "No se pudo importar el JSON.", unableToExportJson: "No se pudo exportar el JSON.",
        encryptList: "Cifrar lista", unlockList: "Desbloquear lista", decryptList: "Quitar cifrado", pin: "PIN de cuatro dígitos", pinPlaceholder: "0000", confirmPin: "Confirmar PIN", pinFormat: "Introduce exactamente cuatro dígitos.", pinMismatch: "Los PIN no coinciden.", collectionSecurityFailed: "No se pudo proteger o desbloquear la lista. Comprueba el PIN e inténtalo de nuevo.", lockedCollection: "Lista bloqueada", encryptedCollection: "Lista cifrada", lockList: "Bloquear lista",
        currentListPassword: "Contraseña actual de la lista", exportPassword: "Nueva contraseña de exportación", importPassword: "Contraseña del JSON cifrado", exchangeCredentialsRequired: "Introduce el PIN actual de cuatro dígitos y una nueva contraseña de exportación.",
    },
} as const;

const I18nContext = createContext<{ locale: Locale; t: (key: TranslationKey) => string }>({
    locale: "en",
    t: (key) => translations.en[key],
});

export function resolveSystemLocale(): Locale {
    return typeof navigator !== "undefined" && navigator.language.toLowerCase().startsWith("es") ? "es" : "en";
}

export function resolveLocale(preference: LanguagePreference): Locale {
    return preference;
}

export function I18nProvider({ locale, children }: { locale: Locale; children: React.ReactNode }) {
    return <I18nContext.Provider value={{ locale, t: (key) => translations[locale][key] }}>{children}</I18nContext.Provider>;
}

export function useTranslation() {
    return useContext(I18nContext);
}
