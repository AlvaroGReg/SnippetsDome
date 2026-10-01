import { createContext, useContext } from "react";

export type LanguagePreference = "en" | "es";
export type Locale = "en" | "es";

type TranslationKey = keyof typeof translations.en;

const translations = {
    en: {
        settings: "Settings", closeToTray: "Close to tray", closeToTrayDescription: "Keep SnippetsDome running in the notification area when its window is closed.",
        startAtLogin: "Start at login", startAtLoginDescription: "Launch SnippetsDome automatically when you sign in.", unavailablePlatforms: "Available on Windows, Linux, and macOS.",
        snippetsShownInTray: "Snippets shown in tray", done: "Done", language: "Language", english: "English", spanish: "Español",
        searchSnippets: "Search snippets", clearSearch: "Clear search", createSnippet: "Create snippet", loadingSnippets: "Loading snippets", noFileSelected: "No file selected.", snippetsFile: "Snippets file", chooseSnippetsFile: "Choose an existing JSON file or create a new one with the name you want.", chooseFile: "Choose file", createFile: "Create file",
        switchToLight: "Switch to light theme", switchToDark: "Switch to dark theme", deleteSnippet: "Delete snippet", deleteConfirmation: "Are you sure you want to delete this snippet?", delete: "Delete",
        cancel: "Cancel", confirm: "Confirm", error: "Error", close: "Close", emptyList: "Empty list", copy: "Copy", edit: "Edit", addToFavorites: "Add to favorites", removeFromFavorites: "Remove from favorites",
        codeCopied: "Code copied to clipboard", codeCopyFailed: "Could not copy the code", editSnippet: "Edit snippet", title: "Title", languageField: "Language", code: "Code", tags: "Tags", tagsHint: "Separate tags with commas", save: "Save", create: "Create",
        unableToLoadSettings: "Unable to load settings.", unableToSaveSettings: "Unable to save settings.",
    },
    es: {
        settings: "Opciones", closeToTray: "Cerrar en la bandeja", closeToTrayDescription: "Mantener SnippetsDome en el área de notificación al cerrar su ventana.",
        startAtLogin: "Iniciar al iniciar sesión", startAtLoginDescription: "Iniciar SnippetsDome automáticamente al iniciar sesión.", unavailablePlatforms: "Disponible en Windows, Linux y macOS.",
        snippetsShownInTray: "Snippets mostrados en la bandeja", done: "Listo", language: "Idioma", english: "English", spanish: "Español",
        searchSnippets: "Buscar snippets", clearSearch: "Borrar búsqueda", createSnippet: "Crear snippet", loadingSnippets: "Cargando snippets", noFileSelected: "No hay ningún archivo seleccionado.", snippetsFile: "Archivo de snippets", chooseSnippetsFile: "Elige un archivo JSON existente o crea uno nuevo con el nombre que quieras.", chooseFile: "Elegir archivo", createFile: "Crear archivo",
        switchToLight: "Cambiar al tema claro", switchToDark: "Cambiar al tema oscuro", deleteSnippet: "Eliminar snippet", deleteConfirmation: "¿Seguro que quieres eliminar este snippet?", delete: "Eliminar",
        cancel: "Cancelar", confirm: "Confirmar", error: "Error", close: "Cerrar", emptyList: "Lista vacía", copy: "Copiar", edit: "Editar", addToFavorites: "Añadir a favoritos", removeFromFavorites: "Quitar de favoritos",
        codeCopied: "Código copiado al portapapeles", codeCopyFailed: "No se pudo copiar el código", editSnippet: "Editar snippet", title: "Título", languageField: "Lenguaje", code: "Código", tags: "Etiquetas", tagsHint: "Separa las etiquetas con comas", save: "Guardar", create: "Crear",
        unableToLoadSettings: "No se pudieron cargar las opciones.", unableToSaveSettings: "No se pudieron guardar las opciones.",
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
