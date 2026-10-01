import { createRoot } from "react-dom/client";
import { useEffect, useState } from "react";
import "./style.css";
import App from "./App";
import { FluentProvider, webDarkTheme, webLightTheme } from "@fluentui/react-components";
import * as snippetsService from "./services/snippets-service";
import { I18nProvider, resolveLocale, resolveSystemLocale, type LanguagePreference } from "./i18n";

function Root() {
    const [isDarkTheme, setIsDarkTheme] = useState(true);
    const [language, setLanguage] = useState<LanguagePreference>(resolveSystemLocale());

    useEffect(() => {
        void snippetsService.getLanguage().then((value) => {
            if (value === "en" || value === "es") setLanguage(value);
        });
    }, []);

    return (
        <FluentProvider theme={isDarkTheme ? webDarkTheme : webLightTheme} style={{ minHeight: "100vh" }}>
            <I18nProvider locale={resolveLocale(language)}>
            <App
                isDarkTheme={isDarkTheme}
                onToggleTheme={() => setIsDarkTheme((currentTheme) => !currentTheme)}
                language={language}
                onLanguageChange={setLanguage}
            />
            </I18nProvider>
        </FluentProvider>
    );
}

const container = document.getElementById("root");
createRoot(container!).render(<Root />);
