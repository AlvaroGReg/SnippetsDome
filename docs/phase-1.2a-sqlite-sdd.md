# SDD: Fase 1.2a - SQLite y colecciones

## 1. Estado y alcance

- Estado: aprobado para implementar
- Aplicación: SnippetsDome, Wails v2
- Bloque: 1.2a - SQLite y gestión de colecciones
- Roadmap relacionado: `docs/phase-1.2-roadmap.md`

Este documento define las decisiones de arquitectura y el comportamiento esperado del almacenamiento SQLite. El roadmap define el orden de ejecución por sesiones; no se repiten aquí sus tareas detalladas.

La fase 1.2a sustituye el almacenamiento JSON principal por una única base SQLite local que contiene varias colecciones de snippets y las preferencias de la aplicación.

## 2. Fuera de alcance

- Importación JSON: fase 1.2b.
- Exportación JSON: fase 1.2b.
- Migración automática desde el JSON existente.
- Compatibilidad completa de tray, autostart y validación operativa en Linux/macOS: fase 1.2c.
- Nuevas funciones avanzadas de búsqueda o etiquetado.

La importación y exportación JSON reutilizarán el modelo de datos de esta fase, pero no forman parte del bootstrap SQLite.

## 3. Estado inicial y decisiones de producto

- Una instalación nueva crea una base SQLite vacía.
- Se crea una colección inicial vacía para que la aplicación pueda arrancar sin contenido.
- La aplicación informa de que el JSON anterior puede recuperarse mediante importación explícita.
- El JSON existente no se lee, modifica, mueve ni elimina automáticamente.
- Existe una única base SQLite por usuario.
- La base contiene varias colecciones.
- Cambiar de lista significa cambiar de colección activa dentro de la misma base.
- El usuario no selecciona manualmente el archivo `.db` en esta fase.

Decisiones cerradas en `1.2a.0`:

- La base se llama `snippets.db`.
- La colección inicial se llama `General`.
- Los nombres de colección se normalizan con `strings.TrimSpace` y son únicos sin distinguir mayúsculas y minúsculas.
- Las colecciones se pueden renombrar, pero el nombre no puede quedar vacío ni duplicar otra colección.
- Siempre debe existir al menos una colección; no se puede eliminar la última.
- Al eliminar la colección activa, pasa a ser activa la colección restante con menor `position`.
- Los identificadores de colección son UUID v4 en formato textual.
- Los tags se recortan, se descartan los vacíos y se eliminan duplicados sin distinguir mayúsculas y minúsculas, conservando la primera grafía.
- SQLite usará `journal_mode=WAL`, `busy_timeout=5000` y una conexión abierta como máximo (`SetMaxOpenConns(1)`).
- La versión inicial del esquema será `1`; los índices definitivos serán `snippets(collection_id, position)` y `snippets(collection_id, favorite, position)`.
- Los errores públicos conservarán mensajes estables y orientados a la operación: `collection name is required`, `collection name already exists`, `collection not found`, `cannot delete the last collection` y `snippet not found`.

## 4. Contexto actual

La aplicación actual mantiene:

- preferencias en `config.json` dentro del directorio de configuración del usuario;
- snippets en un archivo JSON seleccionado por el usuario;
- un `SnippetService` acoplado a `JSONSnippetRepository`;
- operaciones que leen y reescriben la colección completa.

La fase 1.2a elimina ese acoplamiento. El frontend y la capa Wails no deben conocer sentencias SQL ni depender de una extensión `.json` para el almacenamiento principal.

## 5. Arquitectura propuesta

```text
Wails App
    |
    v
SnippetService
    |
    v
Repository interfaces
    |
    v
SQLite repository -> *sql.DB -> snippets.db
```

Reglas de separación:

- `internal/domain` contiene modelos y reglas de dominio.
- `internal/repository` contiene SQLite y el acceso a filesystem.
- `internal/service` coordina operaciones de negocio y validación.
- `App` expone adaptadores Wails y gestiona el ciclo de vida.
- El frontend solo usa los métodos generados de Wails.
- Ninguna capa de persistencia importa paquetes de Wails.

## 6. Driver SQLite

Usar `modernc.org/sqlite` mediante `database/sql`.

Motivos:

- SQLite embebido.
- Sin CGO.
- Evita depender de un compilador C o de librerías SQLite instaladas en el sistema.
- Facilita el mismo código Go para Windows, Linux y macOS.

La versión exacta debe fijarse en `go.mod` durante la sesión `1.2a.1` y mantenerse actualizada solo mediante una decisión explícita.

No usar `github.com/mattn/go-sqlite3` en esta fase porque introduce una dependencia de CGO en el proceso de compilación.

## 7. Ubicación de la base

La base se almacenará en el directorio de datos estándar de cada sistema operativo:

- Windows: `%APPDATA%/SnippetsDome/snippets.db`.
- Linux: `$XDG_DATA_HOME/SnippetsDome/snippets.db`.
- Linux sin `XDG_DATA_HOME`: `~/.local/share/SnippetsDome/snippets.db`.
- macOS: `~/Library/Application Support/SnippetsDome/snippets.db`.

La resolución debe usar `filepath.Join` y respetar las variables de entorno y convenciones del sistema. No se deben hardcodear separadores de ruta.

La implementación de las rutas multiplataforma y su validación operativa corresponden a 1.2c. La fase 1.2a debe usar un resolvedor aislado mediante una interfaz para poder probarlo y sustituirlo sin cambiar el repositorio.

El nombre de la base debe confirmarse en `1.2a.0`. Este documento usa `snippets.db` como nombre provisional.

## 8. Modelo de datos

El esquema debe permitir varias listas sin crear una base por lista.

### 8.1 Colecciones

```sql
CREATE TABLE collections (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    created_at TEXT NOT NULL,
    position   INTEGER NOT NULL DEFAULT 0
);
```

La unicidad se aplica con comparación `LOWER(name)` tras recortar espacios. La colección activa se guarda en `settings` con la clave `activeCollectionId`.

### 8.2 Snippets

```sql
CREATE TABLE snippets (
    id            TEXT PRIMARY KEY,
    collection_id TEXT NOT NULL,
    title         TEXT NOT NULL,
    language      TEXT NOT NULL DEFAULT '',
    code          TEXT NOT NULL,
    created_at    TEXT NOT NULL,
    favorite      INTEGER NOT NULL DEFAULT 0 CHECK (favorite IN (0, 1)),
    position      INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE
);
```

### 8.3 Tags

```sql
CREATE TABLE snippet_tags (
    snippet_id TEXT NOT NULL,
    tag        TEXT NOT NULL,
    PRIMARY KEY (snippet_id, tag),
    FOREIGN KEY (snippet_id) REFERENCES snippets(id) ON DELETE CASCADE
);
```

Los tags vacíos se descartan y los duplicados se comparan con `strings.ToLower` después de recortar espacios; se conserva la primera grafía no vacía. Esta regla también se aplicará a la futura importación/exportación.

### 8.4 Preferencias

```sql
CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
```

Las preferencias actuales se conservarán conceptualmente:

- `closeToTray` como booleano `0`/`1`;
- `traySnippetLimit` como entero decimal;
- `startAtLogin` como booleano `0`/`1`;
- `language` como texto;
- `activeCollectionId` como texto.

No se almacenará `snippetsFilePath`, porque el archivo JSON deja de ser el almacenamiento activo y no habrá migración automática.

### 8.5 Versión del esquema

```sql
CREATE TABLE schema_version (
    version INTEGER NOT NULL
);
```

La creación de tablas y cada cambio posterior se ejecutarán mediante migraciones versionadas e idempotentes. El esquema inicial debe crear la colección vacía inicial y guardar la versión en una única transacción.

Los nombres definitivos, restricciones e índices quedan cerrados por las decisiones de `1.2a.0` antes de implementar `1.2a.2`.

## 9. Contrato del repositorio

El servicio no dependerá de una implementación concreta. El contrato debe cubrir, como mínimo:

- inicialización y migración del esquema;
- cierre de la base;
- lectura y escritura de preferencias;
- creación, listado, renombrado y eliminación de colecciones;
- lectura del contenido de la colección activa;
- inserción, actualización y borrado de snippets;
- sustitución transaccional de tags;
- selección de colección activa.

Las operaciones compuestas deben ser atómicas. En particular:

- crear o actualizar un snippet y sus tags;
- eliminar una colección y sus snippets;
- crear una colección y convertirla en activa;
- importar datos en la fase 1.2b.

La API concreta del repositorio se define una sola vez en `1.2a.0` y se implementa en `1.2a.3`.

## 10. Concurrencia y ciclo de vida

- La aplicación usará una única instancia de `*sql.DB`.
- `database/sql` gestionará el pool de conexiones.
- El servicio protegerá las transiciones de estado en memoria con su mutex actual.
- Las escrituras relacionadas se protegerán mediante transacciones SQLite.
- La política de SQLite queda fijada en `journal_mode=WAL`, `busy_timeout=5000` y una conexión abierta como máximo.
- El arranque abrirá SQLite y ejecutará migraciones antes de servir llamadas del frontend.
- `shutdown` cerrará la conexión y registrará errores de cierre.
- No se crearán ni inicializarán archivos JSON durante el arranque.

## 11. Compatibilidad funcional

La fase debe conservar el comportamiento existente de snippets:

- IDs generados y persistidos.
- `CreatedAt` conservado durante actualizaciones.
- título y código obligatorios.
- normalización actual de campos de texto.
- favoritos persistidos.
- tags persistidos.
- favoritos primero en el listado.
- preferencias de idioma, tray, límite y autostart.

Las operaciones de snippets se aplicarán sobre la colección activa. Si no existe una colección activa válida, el servicio debe devolver un error controlado o seleccionar la colección inicial según la decisión de `1.2a.0`.

## 12. Frontend y API Wails

La interfaz Wails debe exponer operaciones para:

- obtener la colección activa;
- listar colecciones;
- crear una colección;
- renombrar una colección;
- eliminar una colección según las reglas aprobadas;
- seleccionar una colección;
- listar y modificar snippets de la colección activa.

El frontend debe:

- mostrar la colección activa;
- permitir cambiar de colección;
- mostrar el estado vacío inicial;
- informar de la importación JSON futura o disponible según la sesión 1.2b;
- dejar de tratar una ruta `.json` como almacenamiento principal.

Los bindings generados no se editarán manualmente. Se regenerarán mediante el flujo normal de Wails cuando sea necesario.

## 13. Pruebas requeridas

### Bootstrap y esquema

- crea la carpeta y la base en una ruta temporal;
- crea el esquema inicial;
- crea la colección inicial;
- reabre la misma base sin duplicar tablas ni colección inicial;
- aplica migraciones una sola vez;
- revierte una migración fallida.

### Colecciones

- crea una colección;
- lista colecciones en orden estable;
- selecciona una colección;
- conserva la colección activa tras reiniciar;
- aplica las reglas de nombre;
- aplica las reglas de eliminación;
- mantiene aislados los snippets entre colecciones.

### Snippets y tags

- persiste todos los campos;
- crea, actualiza y elimina snippets;
- conserva `CreatedAt` al actualizar;
- persiste cero, uno y varios tags;
- elimina tags al borrar un snippet;
- conserva favoritos y posición;
- revierte operaciones parciales.

### Preferencias

- persiste y recupera cada preferencia;
- persiste la colección activa;
- mantiene el rollback en memoria cuando una escritura falla.

### Aplicación

- inicializa el repositorio una sola vez;
- cierra la base durante `shutdown`;
- no intenta leer ni migrar el JSON antiguo;
- muestra una lista vacía en una instalación nueva.

La verificación obligatoria de 1.2a es:

```bash
go test ./...
```

Las pruebas frontend y el build frontend se ejecutan en `1.2a.5`:

```bash
npm test
npm run build
```

## 14. Criterios de aceptación de 1.2a

1. Una instalación nueva crea una base SQLite y una colección vacía.
2. La base se ubica en el directorio de datos estándar del sistema.
3. SQLite es la única fuente de datos principal.
4. La aplicación permite varias colecciones dentro de una única base.
5. Cambiar de colección no requiere mover ni seleccionar archivos `.db`.
6. Snippets, tags, favoritos y preferencias sobreviven a un reinicio.
7. No se ejecuta ninguna migración automática del JSON existente.
8. La interfaz informa de que la recuperación del JSON se hará mediante importación.
9. El servicio no depende de `JSONSnippetRepository`.
10. La persistencia no contiene imports de Wails.
11. El driver SQLite no requiere CGO.
12. `go test ./...` pasa antes de entregar la fase.

## 15. Decisiones cerradas en la sesión 1.2a.0

Estas decisiones no deben volver a discutirse en sesiones posteriores salvo que aparezca un bloqueo técnico:

- nombre definitivo de la base: `snippets.db`;
- nombre de la colección inicial: `General`;
- nombres únicos sin distinguir mayúsculas y minúsculas, después de recortar espacios;
- renombrado permitido con nombre no vacío y no duplicado;
- la colección activa sí se puede eliminar si queda otra colección;
- tras eliminar la activa se selecciona la restante con menor `position`;
- siempre debe existir al menos una colección;
- IDs de colección UUID v4;
- nombres recortados y tags recortados, sin vacíos ni duplicados case-insensitive;
- duplicados de tags resueltos conservando la primera grafía;
- `journal_mode=WAL`, `busy_timeout=5000`, una conexión máxima;
- esquema inicial versión `1`, con índices sobre colección/posición y colección/favorito/posición;
- mensajes públicos estables: `collection name is required`, `collection name already exists`, `collection not found`, `cannot delete the last collection` y `snippet not found`.

Una vez cerradas, estas decisiones deben registrarse en este documento y no duplicarse en los documentos de las sesiones `1.2a.1` a `1.2a.5`.
