import './SearchBar.css'
import { Button, Input } from "@fluentui/react-components";
import { DismissRegular, SearchRegular } from "@fluentui/react-icons";
import { useTranslation } from "../../i18n";

type SearchBarProps = {
    value: string;
    onChange: (value: string) => void;
};

function SearchBar({ value, onChange }: SearchBarProps) {
    const { t } = useTranslation();
    return (
        <div className="search-bar">
            <Input
                aria-label={t("searchSnippets")}
                className="search-input"
                contentBefore={<SearchRegular aria-hidden="true" />}
                contentAfter={value ? (
                    <Button
                        appearance="transparent"
                        aria-label={t("clearSearch")}
                        icon={<DismissRegular />}
                        onClick={() => onChange("")}
                        size="small"
                    />
                ) : undefined}
                onChange={(_, data) => onChange(data.value)}
                placeholder={t("searchSnippets")}
                value={value}
            />
        </div>
    );
}

export default SearchBar;
