# Estado de la v1

Última actualización: 2026-11-05

## Arquitectura

- Frontend Angular desplegado en Vercel.
- Proxy `/api/v1` de Vercel hacia la API Go en Railway.
- PostgreSQL administrado en Neon.
- Migraciones aplicadas con `golang-migrate` (versión 25).

## Estado

| Elemento | Estado |
|---|---|
| Base de datos Neon | Migrada, versión 25, `dirty=false` |
| API Go en Railway | Desplegada y conectada a Neon |
| Usuario administrador inicial | Creado |
| Cliente Angular en Vercel | Desplegado con proxy `/api/v1` |
| Login y cookies | Funcionando |
| CORS | Configurado para desarrollo local |

## Medidas de seguridad implementadas

- TLS en Railway, Neon y Vercel.
- Autenticación JWT con cookies `HttpOnly`, `Secure` y `SameSite=Lax`.
- Contraseñas almacenadas con bcrypt.
- Secretos en variables de entorno de Railway/Vercel, fuera de Git.
- Límite de tamaño del cuerpo de las peticiones (1 MB).
- Validación de entrada en los casos de uso.

## Deuda técnica y seguridad pendiente

1. Añadir rate limiting a `/api/v1/auth/login` y `/api/v1/auth/refresh`.
2. Evaluar un WAF o protección de borde (por ejemplo, Cloudflare) para la API.
3. Documentar el procedimiento de respaldo y restauración de Neon, incluida la ventana PITR.
4. Configurar monitoreo y alertas de uso/logs.
5. Revisar las credenciales locales de pgAdmin en `server/docker-compose.yml`; no deben commitearse.

## Notas operativas

- CORS solo aplica a navegadores; no es un control de acceso para clientes HTTP.
- El proxy de Vercel evita CORS y cookies de terceros, pero no restringe el acceso directo a la API.
- Las migraciones se aplican manualmente según `server/Database_Deploy.md`; no se ejecutan al iniciar la API.
- Antes de un cambio de esquema, crear una rama de respaldo en Neon.

## Reporte de problemas

Abre un issue en el repositorio o contacta al responsable del proyecto.
No incluyas secretos, cadenas de conexión ni datos personales en el reporte.
