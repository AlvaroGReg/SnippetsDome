# SDD: Phase 1.3 - Encrypted Collections

## 1. Status and Scope

- Status: Completed.
- Application: SnippetsDome, Wails v2.
- Phase: 1.3 - Collection encryption.
- Dependency: phase 1.2, SQLite, collection management, and JSON exchange.

This phase protects a SQLite collection with a PIN. Encryption and decryption
are performed in Go; the UI never receives or stores derived keys. When a
collection is encrypted, its name and the `title`, `language`, `code`, and `tags`
fields of its snippets are encrypted.

Implementation is divided into ordered sessions. Each session keeps its scope
small, updates its tests, and documents new technical decisions.

## 2. Original Scope Exclusions (Session 1.3.0)

- Importing or exporting encrypted collections. This was the agreed scope at
  design finalization; the later scope update below records the subsequent
  implementation.
- Encrypting the entire SQLite file or metadata outside the defined scope.
- Recovering a forgotten PIN.
- Synchronizing collections or keys with external services.
- Keeping a collection unlocked when switching to another collection.

At the original scope freeze, JSON import created unencrypted collections and
export was available only for unencrypted collections.

### 2.1 Later Scope Update: Password-Protected JSON Exchange

After the original 1.3.0 scope decision, encrypted JSON exchange was implemented
in follow-up commits `f5f3075` and `df5a0f7`. An encrypted collection can be
exported to a password-protected JSON envelope using a separate export password.
Import accepts that envelope and creates a new, unencrypted SQLite collection;
it does not restore collection encryption or reuse the original collection PIN.
The original exclusion above is retained as historical context, while this later
update defines the delivered behavior.

## 3. Product and Security Decisions

### 3.1 Agreed Decisions

- Users can mark an existing collection as encrypted.
- The collection name and its snippets' `title`, `language`, `code`, and each
  tag are encrypted.
- Selecting an encrypted collection prompts for the PIN. Switching collections
  locks the previous collection and discards its decrypted state.
- Argon2id derives an encryption key; AES-GCM encrypts and authenticates each
  value.
- Neither the PIN nor derived key is stored in SQLite. Go keeps the key only
  while the selected collection is unlocked.
- The tray does not show snippets from a locked collection.
- Unencrypted collections use the legacy JSON export. Encrypted collections
  require authenticated, password-protected JSON export. Import accepts legacy
  JSON and the password-protected encrypted envelope; imported collections are
  always unencrypted.

### 3.2 Decisions Finalized in Session 1.3.0

- Removing encryption requires unlocking the collection with its current PIN.
  The operation decrypts the name, snippets, and tags and persists them in
  plaintext in a single transaction; failure cannot leave a partial conversion.
- If the default collection is encrypted, it remains locked at startup and the
  PIN is requested before its data is displayed.
- The PIN is exactly four ASCII digits (`0`-`9`), including possible leading
  zeroes. A forgotten PIN cannot be reset; the collection's data cannot be
  recovered without the correct PIN.
- Uniqueness is maintained for names available in plaintext. Uniqueness between
  the actual name of a locked encrypted collection and other collections cannot
  be guaranteed. Such duplicate visible names are allowed; collections are
  identified by their stable IDs and display a generic name while locked.
- The public API uses `EncryptCollection(collectionID, pin, pinConfirmation)`,
  `UnlockCollection(collectionID, pin)`, `LockActiveCollection()`,
  `DecryptCollection(collectionID, pin)`, and `ListCollections()`. Response
  types expose status and a visible name when available, never a PIN or key.
- Errors are classified as invalid collection password, collection not found, invalid encryption
  state (already encrypted/not encrypted), collection locked, duplicate name,
  authentication failure, and persistence failure. An incorrect PIN and altered
  ciphertext produce the same generic authentication error.

A four-digit PIN has only 10,000 possible combinations. Argon2id increases the
cost of each attempt, but does not make this PIN resistant to offline dictionary
attacks if someone obtains the database. This limitation is accepted as a
product decision.

### 3.3 Later Scope Update: Collection Passwords

The collection credential was subsequently changed from a four-digit PIN to a
password of 1–10 Unicode characters. Letters, digits, and symbols are accepted;
the password is treated as an exact string and is not trimmed or normalized.
This update supersedes the PIN length and character restrictions above while
retaining the established Argon2id and AES-GCM format. Existing encrypted
collections remain compatible: their previous numeric PINs still unlock them.
The separate password used for encrypted JSON exports is not subject to this
collection-password length limit.

## 4. Architecture

```text
React UI
   |
   v
Wails App (adapters and lifecycle)
   |
   v
SnippetService (locked/unlocked state and operations)
   |                         |
   v                         v
SQLite repository       internal/cryptography
```

- `internal/domain` exposes a collection's encrypted state without exposing
  secrets.
- `internal/cryptography` encapsulates Argon2id, AES-GCM, ciphertext formats,
  and password/invalid-data errors. It does not import Wails or SQLite.
- `internal/repository` stores and transforms records transactionally. It does
  not derive keys or retain passwords.
- `internal/service` controls unlocking, locking, valid operations, and temporary
  ownership of the in-memory key.
- `App` exposes Wails adapters and does not log passwords, keys, or decrypted
  content.
- React requests the PIN and sends it only to Wails methods for unlocking or
  encryption. It does not persist the password in durable state, local storage,
  or logs.

## 5. Cryptographic Design

### 5.1 Key Derivation

- Use Argon2id to derive an AES-256 key from the PIN and a random salt unique to
  each collection.
- Persist the salt and Argon2id parameters needed to repeat the derivation.
- Validate persisted parameters against supported limits before running
  Argon2id, preventing unbounded memory/CPU use from manipulated data.
- Use 64 MiB (65536 KiB) of memory, 3 iterations, and parallelism 1. Verify that
  unlock time is reasonable on the development machine.
- Add and pin the `golang.org/x/crypto` dependency required for Argon2id.

### 5.2 Value Encryption

- Use AES-GCM with a unique random nonce for each encrypted value.
- Generate a 12-byte nonce using a cryptographically secure random generator
  for each value. The nonce is not secret and must never be reused with the same
  key.
- Persist each value as `v1:` followed by unpadded Base64URL encoding of the
  concatenated bytes `nonce || ciphertext || tag`. AES-GCM uses a 16-byte
  authentication tag. The prefix enables unambiguous format versioning.
- Provide associated data (AAD) binding ciphertext to its collection, record,
  and field, for example the collection ID, snippet ID, and field name. For the
  collection name, use the collection ID and `collection.name`.
- Encrypt each tag separately after applying current normalization (trim
  whitespace, discard empty values and case-insensitive duplicates, preserving
  the first spelling).
- Never write temporary plaintext to logs or error messages.
- Treat AES-GCM authentication errors as an incorrect password or damaged data;
  never return partial plaintext.

### 5.3 In-Memory Key Management

- The service retains at most the derived key for the unlocked collection.
- Before switching collections, lock the current one and overwrite the key
  buffer on a best-effort basis.
- When locking or switching collections, clear snippets and the decrypted name
  from React state before loading another selection.
- Do not include the key or PIN in domain structs, Wails responses, serialized
  errors, logs, or preferences.

## 6. Data Model and Migration

The current schema version is 1. This phase introduces version 2 through a
transactional migration compatible with existing collections.

The `collections` table needs metadata equivalent to:

```sql
encrypted       INTEGER NOT NULL DEFAULT 0 CHECK (encrypted IN (0, 1)),
encryption_salt BLOB,
kdf_parameters  TEXT,
key_check       TEXT
```

- `encrypted` indicates whether `name` and the collection's snippet fields
  contain ciphertext.
- `encryption_salt` and `kdf_parameters` enable key derivation when unlocking.
- The Argon2id salt contains 16 random bytes per collection. Persisted parameters
  record memory (65536 KiB), iterations (3), and parallelism (1); validation
  applies limits before key derivation.
- `key_check` contains a value encrypted and authenticated with the derived key.
  It never contains a password or derived key.
- Cryptographic metadata remains null for unencrypted collections.
- The `name` column contains plaintext when `encrypted = 0` and its encrypted
  representation when `encrypted = 1`.
- Current title, language, code, and tag columns store plaintext for unencrypted
  collections and the ciphertext format defined in section 5 for encrypted
  collections. The collection flag determines how rows are interpreted.

The migration does not automatically encrypt existing data. All existing
collections remain `encrypted = 0`. Encrypting or decrypting a collection must
transform its name, snippets, and tags in a single transaction; failure must not
leave partially encrypted rows.

The current unique-name index guarantees uniqueness among names stored in
plaintext. It cannot guarantee uniqueness against the real name of a locked
encrypted collection because that name is stored as random ciphertext. Duplicate
visible names are therefore allowed in this case; the UX and operations use each
collection's stable ID.

## 7. Service and Wails Contracts

The service interface must support:

- listing collections with their ID, visible name only when available, and the
  `encrypted` flag;
- displaying a generic name for locked encrypted collections;
- encrypting a collection with a password and confirmation;
- unlocking/selecting an encrypted collection with a password;
- locking the active collection and discarding its key;
- requiring a locked encrypted collection to be unlocked before renaming it,
  then persisting the replacement name as ciphertext;
- decrypting an unlocked collection after authenticating again with its current
  password;
- returning a stable collection-locked error when reading or writing encrypted
  content that is not unlocked.

The Wails methods are `EncryptCollection(collectionID, password, confirmation)`,
`UnlockCollection(collectionID, password)`, `LockActiveCollection()`,
`DecryptCollection(collectionID, password)`, and `ListCollections()`. The
`CollectionSummary` response type contains `ID string`, `Name string`,
`Encrypted bool`, and `Locked bool`; `ListCollections` returns
`[]CollectionSummary` and `UnlockCollection` returns the unlocked summary. The
remaining methods return errors. `Name` is empty/unavailable when it cannot be
shown. No response contains a password or derived key. Go validates arguments even
if the UI already validates fields. A collection password must contain 1–10
Unicode characters; all characters are accepted and the entered string is used
exactly as-is. Existing four-digit PINs remain valid passwords for collections
encrypted before this change. Public errors distinguish invalid collection
password, collection not found, invalid encryption state, collection locked,
duplicate name, authentication failure, and persistence failure. Incorrect PIN
and altered content share a generic authentication error.

When selecting an encrypted collection, validate the password before publishing
the decrypted name or snippets in a Wails response. Selecting another
collection must invalidate the previous key and decrypted state immediately,
before showing its PIN prompt; cancelling the prompt leaves the newly selected
collection locked. Operations that create or modify snippets in the active
collection must encrypt fields in Go before persisting them and return decrypted
results only while the collection remains unlocked. Renaming a locked encrypted
collection first requires unlocking it; rename operations must never write its
name in plaintext.

## 8. UX and Integration

- Add a visual encrypted-collection indicator to the collection selector.
- In collection management, allow enabling encryption through a PIN and
  confirmation dialog; request the current PIN to remove encryption.
- Request the collection PIN before allowing a locked encrypted collection to
  be renamed. Persist its renamed value as ciphertext; never write the name in
  plaintext while the collection remains encrypted.
- While a collection is locked, show a stable generic name based on part of its
  ID, never its real name.
- Request the PIN when opening/selecting an encrypted collection and show
  authentication errors without exposing sensitive information.
- Lock immediately when switching collections: clear visible data from the
  previous collection and invalidate its Go key before showing the next
  collection's PIN prompt.
- Offer password-protected JSON export for encrypted collections and keep the
  legacy JSON export for unencrypted collections. Import both supported formats
  while always creating an unencrypted SQLite collection.
- If the active collection is encrypted and locked, the tray must not show its
  snippets. Do not retrieve encrypted content to build the tray menu.
- Review and remove existing logs that print complete snippets, particularly
  `SnippetService.List`, `CreateSnippet`, and `UpdateSnippet`.
- Add UI and error translations in both English and Spanish.

## 9. Implementation Sessions

### 1.3.0 - Design Finalization

**Status: Completed.** The decisions for this session are recorded in sections
3, 5, 6, and 7 of this document.

**Goal:** confirm cryptographic decisions, contracts, and user experience before
changing code.

**Scope:**

- Argon2id: 64 MiB, 3 iterations, parallelism 1; random 16-byte salt.
- AES-GCM: random 12-byte nonce per value, 16-byte tag, and format `v1:` plus
  unpadded Base64URL of `nonce || ciphertext || tag`.
- Original four-digit PIN contract, superseded by the later password update:
  confirm the collection password when enabling encryption.
- Removing encryption transactionally decrypts and persists all fields in
  plaintext; startup with an encrypted default collection requests the PIN
  before displaying data.
- The visible name of a locked encrypted collection may duplicate another name;
  the stable ID distinguishes collections.
- Public contract and error categories are fixed in section 7.

**Deliverables:** approved SDD, fixed cryptographic parameters, and agreed
service contract. Completed.

**Verification:** documentation review; no code changes.

### 1.3.1 - Cryptographic Primitives

**Goal:** implement and test key derivation and value encryption without SQLite
or Wails dependencies.

**Scope:**

- Create an internal cryptography package with Argon2id and AES-GCM.
- Serialize/deserialize salt, parameters, nonce, and ciphertext.
- Implement AAD binding values to their fields and records.
- Validate Argon2id parameters before deriving keys.
- Define incorrect-password and invalid-ciphertext errors without including
  secrets.
- Overwrite key buffers after use on a best-effort basis.

**Tests:** round trip, incorrect password, truncated nonce/ciphertext, altered
ciphertext, incorrect AAD, out-of-range parameters, and empty/large values.

**Verification:**

```bash
go test ./...
```

### 1.3.2 - SQLite Migration and Transactional Encryption

**Goal:** persist encryption metadata and transform a collection atomically.

**Scope:**

- Migrate schema version 1 to 2 within the existing migration system.
- Add cryptographic metadata and `encrypted` state to collections.
- Update the collection model to expose encryption state.
- Implement transactional collection encryption/decryption in the repository,
  including its name, snippets, and tags.
- Preserve IDs, dates, favorites, positions, and relationships.
- Guarantee rollback if any field fails.
- Do not automatically encrypt data during migration.

**Tests:** migration of an existing database, reopen/idempotence, empty and
populated collections, multiple tags, atomicity on failure, and preservation of
all fields outside the encryption scope.

**Verification:**

```bash
go test ./...
```

### 1.3.3 - Service, Unlocking, and Lifecycle

**Goal:** integrate cryptography with current operations and enforce locked
state from Go.

**Scope:**

- Keep the derived key in the service only for the active unlocked collection.
- Add encryption, unlocking, locking, and decryption operations.
- Request a PIN when selecting each encrypted collection.
- Lock and clear the key when switching collections.
- Reject CRUD operations on a locked collection.
- Decrypt fields and tags when reading; encrypt them when creating/updating.
- Prevent startup from automatically loading encrypted content without
  authentication.
- Ensure errors/logs contain no passwords, keys, or plaintext.

**Tests:** locked/unlocked transitions, encrypted CRUD, collection switching,
incorrect PIN, encrypted default collection, protected concurrency, and
repository failures.

**Verification:**

```bash
go test ./...
```

### 1.3.4 - Wails and React UX

**Goal:** expose password and encryption flows to the UI without persisting
secrets in the frontend.

**Backend scope:**

- Add Wails adapters for encrypting, unlocking, locking, and decrypting.
- Return encryption state and visible names only when appropriate.
- Regenerate bindings through the normal workflow; do not edit generated files
  manually.

**Frontend scope:**

- Add PIN/confirmation dialogs and translated errors.
- Show an indicator and generic name while a collection is locked.
- Request the PIN when selecting an encrypted collection, including an
  encrypted default collection shown at startup.
- Clear snippets and name from memory when switching collections or locking.
- Hide export for encrypted rows and leave import unchanged.
- Prevent loading state from briefly showing the previous collection's data.

**Tests:** unlock interaction and error, encrypt/decrypt from management,
clearing on switch, contextual export visibility, and translations.

**Verification:**

```bash
npm test
npm run build
```

### 1.3.5 - Tray and Security Integration

**Goal:** ensure the tray and logs respect locked state and close integration
cases.

**Scope:**

- Do not show snippets from a locked collection in the tray menu.
- Verify tray behavior for unencrypted collections is preserved.
- Remove logs of snippet content and review new errors/logs for secret exposure.
- Ensure unencrypted JSON export cannot export an encrypted collection even if
  the Wails method is invoked directly; encrypted export requires the collection
  PIN and a separate export password.
- Import password-protected encrypted JSON as a new unencrypted collection;
  preserve legacy JSON import behavior.

**Tests:** tray with locked/unlocked encrypted collections, backend protection
for the legacy exporter, encrypted JSON export/import and password failures,
unchanged legacy import, and regressions for unencrypted collections.

**Final verification:**

```bash
go test ./...
```

From `frontend/`:

```bash
npm test
npm run build
```

Interactive manual verification is performed by the user. Do not use `wails
build`, `wails dev`, or frontend production builds as substitutes for these
tests.

## 10. Session Dependencies

```text
1.3.0 -> 1.3.1 -> 1.3.2 -> 1.3.3 -> 1.3.4 -> 1.3.5
```

## 11. Phase Acceptance Criteria

1. Existing SQLite databases migrate to the new version without changing
   content or state for unencrypted collections.
2. The name, title, language, code, and tags of an encrypted collection are not
   stored in plaintext in SQLite.
3. An incorrect password or altered ciphertext never returns partial content or
   causes a partial collection transformation.
4. The password and derived key are not stored on disk or exposed in logs or
   read bindings.
5. Every selection of an encrypted collection requires a PIN; switching
   collections locks it and discards previously decrypted content.
6. Snippet operations on an encrypted collection work only while it is unlocked
   and persist fields as ciphertext.
7. Renaming a locked encrypted collection requires successful PIN authentication,
   and its name remains encrypted in SQLite.
8. A locked collection's name is not shown in the UI and is replaced with a
   generic label.
9. The tray does not expose snippets from a locked collection.
10. Encrypted collections cannot be exported as legacy plaintext JSON; password-
   protected JSON export requires the collection PIN and a separate password.
11. Legacy JSON import remains available, encrypted JSON envelopes can be
    imported with their export password, and both formats create unencrypted
    collections.
12. Logs contain no snippet title, code, language, or tags.
13. `go test ./...`, `npm test`, and `npm run build` pass before the phase is
    considered complete.

## 12. Known Risks and Limitations

- Anyone with a copy of the database can attempt offline password guessing.
  Argon2id raises the cost per attempt but does not replace a strong password.
- Unencrypted metadata (IDs, dates, favorites, positions, and relationships)
  reveals some usage structure.
- Losing the password means practical loss of access to the data.
- Short or predictable collection passwords can still be guessed offline against
  a database copy; Argon2id increases the cost of each attempt.
- The real name's uniqueness cannot be guaranteed against locked encrypted
  collections.
- The key and decrypted data exist in memory while a collection is open. Memory
  clearing from Go is best-effort, and the frontend cannot guarantee physical
  removal of internal runtime copies.
- The tray may show content only while the active collection is unlocked; its
  menu must be rebuilt or cleared when the collection is locked.
