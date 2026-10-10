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

## Reset de datos (conservando usuarios)

Elimina todos los datos de negocio pero conserva la tabla `users`, sus sesiones
(`refresh_tokens`) y el control de migraciones (`schema_migrations`).

En PowerShell (desde `server`):

```powershell
Get-Content .\scripts\reset-db.sql | docker compose exec -T postgres psql -U pig_farm -d pig_farm
```

En Bash (desde `server`):

```bash
docker compose exec -T postgres psql -U pig_farm -d pig_farm < ./scripts/reset-db.sql
```

La bandera `-T` es necesaria para que `psql` acepte la entrada por stdin.

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
