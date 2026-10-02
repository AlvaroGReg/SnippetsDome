# Contrato JSON de la fase 1.2b

El formato de intercambio es compatible con el archivo JSON anterior: la raíz es
un array de snippets. Se conservan `id`, `title`, `language`, `code`, `tags`,
`createdAt` y `favorite`.

La importación crea una nueva colección con el nombre del archivo sin extensión
y la selecciona al terminar. `id`, `title`, `code` y `createdAt` son obligatorios;
un array vacío es válido. Los IDs repetidos dentro del archivo o ya existentes
en SQLite rechazan toda la operación. La colección y sus snippets se insertan
en una única transacción, por lo que un error no deja cambios parciales.

La exportación contiene únicamente la colección activa y nunca modifica SQLite.
