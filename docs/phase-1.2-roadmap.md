# Roadmap de sesiones agénticas: versión 1.2

## Objetivo general

La versión 1.2 sustituirá el almacenamiento JSON como sistema principal por SQLite, añadirá importación y exportación JSON y completará la compatibilidad operativa con Linux y macOS.

La ejecución se divide en tres bloques independientes y ordenados:

1. **1.2a - SQLite y gestión de colecciones**
2. **1.2b - Importación y exportación JSON**
3. **1.2c - Compatibilidad Linux y macOS**

Cada sesión debe mantener un alcance pequeño, ejecutar sus pruebas y dejar documentadas las decisiones o riesgos descubiertos.

## Progreso

- [x] `1.2a.0` - Cierre del diseño y aprobación del SDD.
- [x] `1.2a.1` - Resolución de rutas y bootstrap SQLite.
- [x] `1.2a.2` - Esquema y migraciones SQLite.
- [x] `1.2a.3` - Repositorio SQLite.
- [x] `1.2a.4` - Servicio y configuración.
- [x] `1.2a.5` - Integración Wails y frontend.
- [ ] `1.2b` - Importación y exportación JSON.
- [ ] `1.2c` - Compatibilidad Linux y macOS.

## Decisiones de producto

- SQLite será el almacenamiento principal de la aplicación.
- La aplicación usará una única base SQLite por usuario.
- La base se guardará en el directorio de datos estándar de cada sistema operativo.
- La base contendrá varias colecciones de snippets.
- Cambiar de lista significa cambiar de colección activa, no mover archivos manualmente.
- No habrá migración automática del JSON existente.
- Una instalación nueva comenzará con una colección vacía.
- La aplicación informará de que la lista anterior se puede recuperar importando su JSON.
- La importación y exportación JSON serán funciones explícitas de la versión 1.2b.

## Reglas comunes de las sesiones

- Revisar el estado del repositorio antes de editar.
- No revertir cambios ajenos.
- Mantener SQL y lógica de persistencia fuera de Wails y del frontend.
- Preferir cambios pequeños y verificables.
- Añadir o actualizar tests junto con cada cambio de comportamiento.
- Ejecutar `go test ./...` después de cada sesión con cambios Go.
- Ejecutar `npm test` y `npm run build` después de cada sesión con cambios frontend.
- No usar `wails build` ni `wails dev` como verificación automática.
- No crear commits automáticamente.

## Dependencias generales

```text
1.2a.0 -> 1.2a.1 -> 1.2a.2 -> 1.2a.3 -> 1.2a.4 -> 1.2a.5
                                                        |
                                                        v
1.2b.0 -> 1.2b.1 -> 1.2b.2 -> 1.2b.3
                                  |
                                  v
1.2c.0 -> 1.2c.1 -> 1.2c.2 -> 1.2c.3
```

# 1.2a - SQLite y colecciones

## 1.2a.0 - Cierre del diseño

### Objetivo

Cerrar el SDD antes de implementar la persistencia SQLite.

### Alcance

- Confirmar una base SQLite única con múltiples colecciones.
- Definir el esquema de colecciones, snippets, tags y preferencias.
- Definir la ubicación por sistema operativo.
- Confirmar que no existe migración automática de JSON.
- Definir el contrato del repositorio.
- Definir el estado vacío y la colección inicial.

### Salida

- SDD actualizado y aprobado.
- Esquema de datos cerrado.
- Criterios de aceptación claros.

### Verificación

Revisión documental. No requiere cambios de código.

## 1.2a.1 - Resolución de rutas y bootstrap SQLite

### Objetivo

Crear la infraestructura mínima para localizar y abrir la base de datos.

### Alcance

- Resolver el directorio de datos de SnippetsDome.
- Crear el directorio si no existe.
- Abrir y cerrar la conexión SQLite.
- Configurar pragmas básicos.
- Devolver errores claros ante rutas no utilizables.

### Fuera de alcance

- Colecciones.
- Snippets.
- Importación y exportación JSON.
- Integración completa con Linux y macOS.

### Verificación

```bash
go test ./...
```

Usar `t.TempDir()` para aislar las pruebas de filesystem.

## 1.2a.2 - Esquema y migraciones SQLite

### Objetivo

Implementar el esquema versionado de la base.

### Alcance

- Tabla de versión del esquema.
- Tabla de colecciones.
- Tabla de snippets.
- Tabla de tags.
- Tabla de preferencias.
- Claves foráneas e índices.
- Migraciones idempotentes.

### Verificación

- Crear una base nueva.
- Reabrir una base existente.
- Confirmar que una migración no se repite.
- Verificar rollback cuando falla una migración.

```bash
go test ./...
```

## 1.2a.3 - Repositorio SQLite

### Objetivo

Implementar el acceso a datos sin dependencias de Wails.

### Alcance

- Crear, listar, seleccionar y eliminar colecciones.
- CRUD de snippets.
- Persistir tags en una tabla relacionada.
- Persistir favoritos y orden estable.
- Ejecutar operaciones compuestas dentro de transacciones.
- Aislar los snippets de cada colección.

### Verificación

- Persistencia completa de todos los campos.
- Varias colecciones independientes.
- Cambio de colección activa.
- Eliminación en cascada de tags.
- Rollback de operaciones parciales.

```bash
go test ./...
```

## 1.2a.4 - Servicio y configuración

### Objetivo

Eliminar la dependencia del servicio respecto al repositorio JSON.

### Alcance

- Sustituir `JSONSnippetRepository` por una interfaz de repositorio.
- Persistir preferencias en SQLite.
- Mantener las validaciones actuales.
- Mantener el orden de favoritos.
- Mantener rollback de preferencias.
- Añadir el concepto de colección activa.
- Adaptar los tests existentes del servicio.

### Verificación

Cubrir como mínimo:

- Crear snippet.
- Editar snippet.
- Borrar snippet.
- Listar snippets.
- Marcar favoritos.
- Cambiar de colección.
- Persistir preferencias.
- Manejar errores de persistencia.

```bash
go test ./...
```

## 1.2a.5 - Integración Wails y frontend

### Objetivo

Exponer SQLite y las colecciones a la aplicación completa.

### Alcance backend

- Inicializar SQLite desde `NewApp`.
- Cerrar la conexión en `shutdown`.
- Exponer métodos Wails para colecciones.
- Eliminar la selección de un JSON como almacenamiento principal.
- Propagar errores al frontend.

### Alcance frontend

- Mostrar la colección activa.
- Cambiar de colección.
- Crear una colección.
- Mostrar el estado vacío inicial.
- Informar de que se puede importar el JSON antiguo.
- Eliminar validaciones que exijan la extensión `.json` para el almacenamiento principal.

### Verificación

Backend:

```bash
go test ./...
```

Frontend:

```bash
npm test
npm run build
```

# 1.2b - Importación y exportación JSON

## 1.2b.0 - Contrato de intercambio JSON

### Objetivo

Definir el comportamiento de importación y exportación antes de implementar dialogs y operaciones de filesystem.

### Alcance

- Mantener compatibilidad con el formato JSON actual.
- Importar una colección desde un archivo JSON.
- Exportar una colección concreta.
- Decidir el tratamiento de IDs repetidos.
- Validar campos obligatorios.
- Preservar IDs, fechas, favoritos, lenguajes, código y tags.
- Definir mensajes para JSON inválido o incompleto.

### Salida

- Contrato JSON documentado.
- Casos límite definidos.
- Tests de comportamiento preparados.

## 1.2b.1 - Importador JSON

### Objetivo

Importar un JSON como una nueva colección SQLite.

### Alcance

- Leer un archivo seleccionado.
- Validar su estructura.
- Crear una nueva colección.
- Insertar snippets y tags en una transacción.
- Seleccionar la colección importada cuando termine correctamente.
- No alterar el estado si la importación falla.

### Verificación

- JSON válido.
- JSON vacío.
- JSON inválido.
- Campos obligatorios ausentes.
- Tags vacíos y múltiples tags.
- Favoritos y fechas.
- Rollback completo ante error.

```bash
go test ./...
```

## 1.2b.2 - Exportador JSON

### Objetivo

Exportar la colección activa a un archivo JSON.

### Alcance

- Exportar únicamente la colección seleccionada.
- Usar el dialog nativo de guardado.
- Mantener un formato JSON estable.
- No modificar los datos internos.
- Gestionar errores de escritura.

### Verificación

- Exportación de una colección con datos.
- Exportación de una colección vacía.
- Preservación exacta de todos los campos.
- Error de destino no escribible.

```bash
go test ./...
```

## 1.2b.3 - UX de importación y exportación

### Objetivo

Conectar los flujos de importación y exportación con React.

### Alcance

- Botón de importar.
- Botón de exportar.
- Indicadores de carga.
- Mensajes de éxito y error.
- Recarga de la lista tras importar.
- Selección automática de la colección importada.
- Tests de interacción.

### Verificación

```bash
npm test
npm run build
```

# 1.2c - Compatibilidad Linux y macOS

## 1.2c.0 - Resolución multiplataforma de rutas

### Objetivo

Validar y completar la ubicación de datos según las convenciones de cada sistema.

### Alcance

- Windows: `%APPDATA%/SnippetsDome`.
- Linux: `$XDG_DATA_HOME/SnippetsDome`.
- Linux sin `XDG_DATA_HOME`: `~/.local/share/SnippetsDome`.
- macOS: `~/Library/Application Support/SnippetsDome`.
- Creación de directorios.
- Separadores de ruta portables.
- Gestión de permisos.
- Ausencia de rutas hardcodeadas.

### Verificación

- Tests de resolución de rutas.
- Tests condicionados por `runtime.GOOS`.
- Tests de variables de entorno.

```bash
go test ./...
```

## 1.2c.1 - Autostart multiplataforma

### Objetivo

Completar el inicio automático en Linux y macOS.

### Alcance

- Desktop entry de Linux.
- Launch Agent de macOS.
- Activación y desactivación.
- Arranque minimizado cuando sea compatible.
- Limpieza de configuraciones desactivadas.
- Errores claros cuando la plataforma no lo permite.

### Verificación

- Tests de contenido de los archivos generados.
- Tests de rutas específicas por plataforma.
- Tests de plataformas no soportadas.

```bash
go test ./...
```

## 1.2c.2 - System tray multiplataforma

### Objetivo

Completar el comportamiento de bandeja del sistema donde la plataforma lo soporte.

### Alcance

- Mostrar y ocultar la ventana.
- Mostrar snippets de la colección activa.
- Copiar snippets desde la bandeja.
- Salir explícitamente desde la bandeja.
- Mantener el comportamiento de close-to-tray.
- Definir fallback cuando la bandeja no esté disponible.

### Verificación

- Tests de selección de controlador por plataforma.
- Tests del comportamiento de salida.
- Tests de fallback.

```bash
go test ./...
```

## 1.2c.3 - Integración y cierre de la versión 1.2

### Objetivo

Validar de extremo a extremo la versión 1.2.

### Casos mínimos

- Arranque con base vacía.
- Creación de una colección.
- Cambio entre colecciones.
- Creación, edición y borrado de snippets.
- Importación JSON.
- Exportación JSON.
- Persistencia de preferencias.
- Autostart.
- System tray.
- Errores de permisos.
- Reinicio de la aplicación sin pérdida de datos.

### Verificación final

Desde la raíz del repositorio:

```bash
go test ./...
```

Desde `frontend/`:

```bash
npm test
npm run build
```

La verificación interactiva en cada sistema operativo se realizará manualmente. No se usará `wails build` ni `wails dev` como sustituto de los tests automatizados.

## Criterios de finalización de 1.2

- SQLite es la única fuente de datos principal.
- La aplicación permite trabajar con varias colecciones.
- Una instalación nueva empieza vacía y muestra cómo importar JSON.
- Importar y exportar JSON funciona sin pérdida de información.
- Linux y macOS usan ubicaciones de datos apropiadas.
- Autostart y system tray tienen comportamiento definido por plataforma.
- Los tests Go pasan con `go test ./...`.
- Los tests frontend y el build frontend pasan.
- La documentación de almacenamiento y uso está actualizada.
- No quedan referencias que presenten el JSON como almacenamiento principal.
