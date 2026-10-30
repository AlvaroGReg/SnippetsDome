export const MAX_COLLECTION_PASSWORD_LENGTH = 10;

export function limitCollectionPassword(value: string): string {
    return Array.from(value).slice(0, MAX_COLLECTION_PASSWORD_LENGTH).join("");
}

export function isValidCollectionPassword(value: string): boolean {
    const length = Array.from(value).length;
    return length >= 1 && length <= MAX_COLLECTION_PASSWORD_LENGTH;
}
