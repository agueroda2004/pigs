# Migraciones

Esta guia explica como administrar PostgreSQL y ejecutar las migraciones del
proyecto con `golang-migrate`.

Ejecuta los comandos desde la carpeta `server`.

## Requisitos

- Docker y Docker Compose.
- Go.
- `golang-migrate` instalado y disponible en el `PATH`.

Para instalar `golang-migrate`:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Levantar el contenedor

Iniciar PostgreSQL en segundo plano:

```bash
docker compose up -d postgres
```

Iniciar PostgreSQL mostrando los logs:

```bash
docker compose up postgres
```

## Revisar el contenedor

Consultar el estado de los servicios:

```bash
docker compose ps
```

Ver los logs de PostgreSQL:

```bash
docker compose logs -f postgres
```

Comprobar que PostgreSQL acepta conexiones:

```bash
docker compose exec postgres pg_isready -U pig_farm -d pig_farm
```

## Configurar la conexion

El CLI de `golang-migrate` no carga automáticamente el archivo `.env`. Define
la variable `DATABASE_URL` antes de ejecutar cualquier migración.

En Bash:

```bash
export DATABASE_URL="DATABASE_URL"
```

En PowerShell:

```powershell
$env:DATABASE_URL = "DATABASE_URL"
```

Verificar la variable configurada:

```powershell
$env:DATABASE_URL
```

## Aplicar migraciones

Aplicar todas las migraciones pendientes:

```bash
migrate -path ./migrations -database "$DATABASE_URL" up
```

En PowerShell:

```powershell
migrate -path ./migrations -database $env:DATABASE_URL up
```

Consultar la versión aplicada:

```bash
migrate -path ./migrations -database "$DATABASE_URL" version
```

En PowerShell:

```powershell
migrate -path ./migrations -database $env:DATABASE_URL version
```

Crear una nueva migración:

```bash
migrate create -ext sql -dir ./migrations -seq describe_change
```

Esto genera dos archivos:

```text
NNN_describe_change.up.sql
NNN_describe_change.down.sql
```

## Revisar el esquema

Listar las tablas:

```bash
docker compose exec postgres psql -U pig_farm -d pig_farm -c "\\dt"
```

Inspeccionar una tabla:

```bash
docker compose exec postgres psql -U pig_farm -d pig_farm -c "\\d nombre_de_tabla"
```

## Revertir migraciones

Revertir la última migración:

```bash
migrate -path ./migrations -database "$DATABASE_URL" down 1
```

En PowerShell:

```powershell
migrate -path ./migrations -database $env:DATABASE_URL down 1
```

## Reset de datos (conservando catálogos y usuarios)

Elimina los registros operativos (animales, servicios, montas, partos, bajas,
destetes y sus tablas de asociación) pero conserva intactas las tablas
`users`, `refresh_tokens`, `operators`, `medications` y `breeds`, junto con el
control de migraciones (`schema_migrations`).

El script `scripts/reset-db.sql`:

- Lista explícitamente cada tabla operativa que se va a vaciar.
- No usa `CASCADE`: si una migración futura agrega una tabla que referencia
  datos operativos, el reset falla en lugar de vaciar tablas no previstas.
- Envuelve el `TRUNCATE` en `BEGIN`/`COMMIT`. `TRUNCATE` es transaccional en
  PostgreSQL, así que cualquier error deja la base sin cambios.

> Advertencia: el script borra datos de forma irreversible. Ejecútalo solo
> contra una base de desarrollo local y después de revisar el destino.

Antes de ejecutarlo, verifica a qué base te vas a conectar:

```powershell
docker compose exec -T postgres psql -U pig_farm -d pig_farm -c "SELECT current_database(), current_user;"
```

Se recomienda respaldar antes:

```powershell
docker compose exec -T postgres pg_dump -U pig_farm -d pig_farm -Fc > .\backup-before-reset.dump
```

En PowerShell (desde `server`):

```powershell
Get-Content .\scripts\reset-db.sql | docker compose exec -T postgres psql -U pig_farm -d pig_farm
```

En Bash (desde `server`):

```bash
docker compose exec -T postgres psql -U pig_farm -d pig_farm < ./scripts/reset-db.sql
```

La bandera `-T` es necesaria para que `psql` acepte la entrada por stdin.

Después del reset, confirma que los catálogos conservaron sus filas y que las
tablas operativas quedaron vacías:

```sql
SELECT 'users' AS tabla, count(*) FROM users
UNION ALL SELECT 'refresh_tokens', count(*) FROM refresh_tokens
UNION ALL SELECT 'operators', count(*) FROM operators
UNION ALL SELECT 'medications', count(*) FROM medications
UNION ALL SELECT 'breeds', count(*) FROM breeds;

SELECT 'boars' AS tabla, count(*) FROM boars
UNION ALL SELECT 'sows', count(*) FROM sows
UNION ALL SELECT 'services', count(*) FROM services
UNION ALL SELECT 'mounts', count(*) FROM mounts
UNION ALL SELECT 'abortions', count(*) FROM abortions
UNION ALL SELECT 'sow_removals', count(*) FROM sow_removals
UNION ALL SELECT 'boar_removals', count(*) FROM boar_removals
UNION ALL SELECT 'farrowings', count(*) FROM farrowings
UNION ALL SELECT 'farrowing_operators', count(*) FROM farrowing_operators
UNION ALL SELECT 'farrowing_medications', count(*) FROM farrowing_medications
UNION ALL SELECT 'piglet_deaths', count(*) FROM piglet_deaths
UNION ALL SELECT 'piglet_fosterings', count(*) FROM piglet_fosterings
UNION ALL SELECT 'partial_weagings', count(*) FROM partial_weagings
UNION ALL SELECT 'weagings', count(*) FROM weagings;
```

Valida primero el script en una base descartable (por ejemplo un contenedor
temporal) antes de ejecutarlo contra tu base de desarrollo.

## Detener el contenedor

Detener los servicios conservando los datos:

```bash
docker compose down
```

Detener los servicios y eliminar los volúmenes:

```bash
docker compose down -v
```

El segundo comando elimina todos los datos almacenados en PostgreSQL.
