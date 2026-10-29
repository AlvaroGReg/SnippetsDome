# Phase 1.2b JSON Contract

The exchange format is compatible with the previous JSON file: the root value
is an array of snippets. The fields `id`, `title`, `language`, `code`, `tags`,
`createdAt`, and `favorite` are preserved.

Import creates a new collection named after the file without its extension and
selects it on completion. `id`, `title`, `code`, and `createdAt` are required; an
empty array is valid. Duplicate IDs within the file or already present in
SQLite reject the entire operation. The collection and its snippets are
inserted in a single transaction, so an error cannot leave partial changes.

Export contains only the active collection and never modifies SQLite.
