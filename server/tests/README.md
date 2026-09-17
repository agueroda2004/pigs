# Tests

Esta guía explica cómo ejecutar los tests unitarios del proyecto.

Ejecuta los comandos desde la carpeta `server`.

## Requisitos

- Go instalado y disponible en el `PATH`.

No es necesario iniciar PostgreSQL ni Docker para ejecutar los tests unitarios
actuales.

## Ejecutar todos los tests

Ejecutar todos los tests del proyecto:

```bash
go test ./...
```

En PowerShell se utiliza el mismo comando:

```powershell
go test ./...
```

## Ver cada test en consola

Usar `-v` para mostrar el estado de cada test y subtest. Usar `-count=1` para
evitar que Go utilice resultados almacenados en caché:

```bash
go test -v -count=1 ./...
```

## Ejecutar tests del módulo user

```bash
go test -v -count=1 ./internal/modules/user/...
```

## Ejecutar un test específico

Filtrar por el nombre del test con `-run`:

```bash
go test -v -count=1 -run TestCreateUserServiceExecute ./internal/modules/user/...
```

Ejecutar un subtest específico:

```bash
go test -v -count=1 -run 'TestCreateUserServiceExecute/creates_a_user' ./internal/modules/user/...
```

## Ejecutar con cobertura

Mostrar el porcentaje de cobertura por paquete:

```bash
go test -cover ./...
```

Generar un archivo de cobertura:

```bash
go test -coverprofile=coverage.out ./...
```

Abrir un reporte HTML:

```bash
go tool cover -html=coverage.out -o coverage.html
```

En PowerShell:

```powershell
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Detectar condiciones de carrera

Ejecutar los tests con el detector de race conditions:

```bash
go test -race ./...
```

En PowerShell:

```powershell
go test -race ./...
```

## Comando recomendado

Para una ejecución completa con salida detallada, sin caché y cobertura:

```bash
go test -v -count=1 -cover ./...
```
