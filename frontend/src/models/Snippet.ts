export type SnippetModel = {
    id: string
    title: string
    language: string
    code: string
    tags: string[]
    createdAt: string
    favorite: boolean
}

export type CreateSnippetInput = Omit<SnippetModel, "id" | "createdAt" | "favorite">;

export type CollectionModel = {
    id: string
    name: string
    createdAt: string
}
