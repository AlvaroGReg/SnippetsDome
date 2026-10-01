import { Badge, Button, Toast, ToastTitle, Toaster, useToastController } from "@fluentui/react-components";
import { StarFilled, StarRegular } from "@fluentui/react-icons";
import type { SnippetModel } from "../../models/Snippet";
import "./SnippetsList.css";
import { useTranslation } from "../../i18n";

type SnippetsListProps = {
    snippets: SnippetModel[];
    onEdit: (snippet: SnippetModel) => void;
    onDelete: (id: string) => void;
    onToggleFavorite: (snippet: SnippetModel) => void;
};

function SnippetsList({ snippets, onEdit, onDelete, onToggleFavorite }: SnippetsListProps) {
    const toasterId = "snippets-list";
    const { t } = useTranslation();
    const { dispatchToast } = useToastController(toasterId);

    async function addToClipboard(code: string) {
        try {
            await navigator.clipboard.writeText(code);
            dispatchToast(
                <Toast>
                     <ToastTitle>{t("codeCopied")}</ToastTitle>
                </Toast>,
                { intent: "success" },
            );
        } catch {
            dispatchToast(
                <Toast>
                     <ToastTitle>{t("codeCopyFailed")}</ToastTitle>
                </Toast>,
                { intent: "error" },
            );
        }
    }

    return (
        <div className="snippets-list">
            <Toaster toasterId={toasterId} position="bottom-end" />
            {snippets.length === 0 ? (
                <p className="snippets-list-empty">{t("emptyList")}</p>
            ) : snippets.map((snippet) => (
                <article key={snippet.id} className="snippet-item">
                    <div className="snippet-head">
                        <div className="snippet-title-group">
                            <Button
                                appearance="subtle"
                                className="favorite-button"
                                data-favorite={snippet.favorite ? "true" : undefined}
                                 aria-label={snippet.favorite ? t("removeFromFavorites") : t("addToFavorites")}
                                aria-pressed={Boolean(snippet.favorite)}
                                icon={snippet.favorite ? <StarFilled /> : <StarRegular />}
                                onClick={() => onToggleFavorite(snippet)}
                                 title={snippet.favorite ? t("removeFromFavorites") : t("addToFavorites")}
                            />
                            <span className="snippet-title">{snippet.title}</span>
                        </div>
                         <Button appearance="primary" onClick={() => void addToClipboard(snippet.code)}>{t("copy")}</Button>
                    </div>
                    <div className="snippet-subtitle">
                        <span className="snippet-lang">{snippet.language}</span>
                        <div className="snippet-actions">
                             <Button onClick={() => onDelete(snippet.id)}>{t("delete")}</Button>
                             <Button onClick={() => onEdit(snippet)}>{t("edit")}</Button>
                        </div>
                    </div>
                    <div className="snippet-body">
                        <pre className="snippet-code"><code>{snippet.code}</code></pre>
                    </div>
                    <ul className="snippet-tags">
                        {snippet.tags.map((tag) => (
                            <li key={`${snippet.id}-${tag}`}>
                                <Badge appearance="tint" color="brand" className="tag-item">{tag}</Badge>
                            </li>
                        ))}
                    </ul>
                </article>
            ))}
        </div>
    );
}

export default SnippetsList;
