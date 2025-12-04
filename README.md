# Servicio de Usuarios - Crabi

Servicio REST en Go para gestión de usuarios con integración PLD, JWT y RabbitMQ.

## Requisitos

- Docker y Docker Compose

## Inicio Rápido

1. Clonar el repositorio (si aplica)

2. **Obtener el archivo `.env`** enviado por correo y colocarlo en la raíz del proyecto

3. Levantar todos los servicios:
```bash
docker-compose up -d
```

Esto iniciará automáticamente:
- PostgreSQL en puerto 5432
- RabbitMQ en puerto 5672 (Management UI en 15672)
- API en puerto 8080

4. Verificar que todo esté corriendo:
```bash
docker-compose ps
```

5. Probar el servicio:
```bash
curl http://localhost:8080/health
```

Deberías recibir: `{"status":"ok"}`

## Configuración

### Archivo `.env`

El proyecto requiere un archivo `.env` en la raíz del proyecto con las variables de entorno necesarias. **Este archivo contiene información sensible y será enviado por correo electrónico** por seguridad.

El archivo `.env` debe contener las siguientes variables:

```
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=<contraseña-segura>
DB_NAME=crabi_db
DB_SSLMODE=disable

JWT_SECRET_KEY=<clave-secreta-minimo-32-caracteres>
JWT_EXPIRES_IN=24

PLD_BASE_URL=http://98.81.235.22
PLD_TIMEOUT=10

RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=<contraseña-segura>

ENV=production
```

**Nota:** El archivo `.env` está en `.gitignore` y no se sube al repositorio por seguridad. Si no lo recibiste, solicítalo por correo.

## Endpoints

### 1. Crear Usuario
```http
POST /api/v1/users
Content-Type: application/json

{
  "email": "usuario@example.com",
  "password": "password123",
  "name": "Gustavo Hernández"
}
```

**Respuesta exitosa (201):**
```json
{
  "user": {
    "id": "uuid",
    "email": "usuario@example.com",
    "name": "Gustavo Hernández",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "token": "jwt-token"
}
```

**Errores posibles:**
- 400: Datos inválidos (email mal formado, password corto, etc.)
- 403: Usuario en lista negra PLD
- 409: Usuario ya existe
- 500: Error interno

### 2. Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "usuario@example.com",
  "password": "password123"
}
```

**Respuesta exitosa (200):**
```json
{
  "token": "jwt-token"
}
```

**Errores posibles:**
- 400: Datos inválidos
- 401: Credenciales inválidas
- 500: Error interno

### 3. Obtener Usuario
```http
GET /api/v1/users/me
Authorization: Bearer <jwt-token>
```

**Respuesta exitosa (200):**
```json
{
  "user": {
    "id": "uuid",
    "email": "usuario@example.com",
    "name": "Gustavo Hernández",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**Errores posibles:**
- 401: Token no proporcionado o inválido
- 404: Usuario no encontrado
- 500: Error interno

## Comandos Útiles

### Ver logs del API
```bash
docker-compose logs api -f
```

### Ver logs de todos los servicios
```bash
docker-compose logs -f
```

### Detener todos los servicios
```bash
docker-compose down
```

### Detener y eliminar volúmenes (elimina datos)
```bash
docker-compose down -v
```

### Reconstruir y levantar
```bash
docker-compose up --build -d
```

## Integración PLD

El servicio consulta automáticamente el servicio PLD externo al crear un usuario. El endpoint es:
- **URL:** `http://98.81.235.22/check-blacklist`
- **Método:** POST
- **Body:** `{"first_name": "...", "last_name": "...", "email": "..."}`
- **Response:** `{"is_in_blacklist": true/false}`

Si el usuario está en lista negra, se rechaza la creación con código 403.

**Ejemplos para probar lista negra:**
- Nombre: "Pablo", Apellido: "Escobar", Email: "pablo@escobar.com"
- Nombre: "Joaquín", Apellido: "Guzmán", Email: "joaquin@guzman.com"

## RabbitMQ

### Eventos Publicados

Al crear un usuario exitosamente, se publica un evento en la cola `user.created`:

```json
{
  "user_id": "uuid",
  "email": "usuario@example.com",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Consumidor

El servicio incluye un consumidor que procesa eventos de `user.created` automáticamente:
- Registra un log: "Enviando email de bienvenida a <email>"
- Guarda el evento en la tabla `user_events` para auditoría

### RabbitMQ Management UI

Accede a la interfaz de administración:
- **URL:** http://localhost:15672
- **Usuario:** guest
- **Contraseña:** guest

Puedes ver las colas, mensajes y estadísticas desde ahí.

## Colección Postman

Importa el archivo `Crabi_API.postman_collection.json` en Postman para probar los endpoints fácilmente.

La colección incluye:
- Ejemplos de todos los endpoints
- Variables de colección (el token JWT se guarda automáticamente)
- Ejemplos de respuestas exitosas y errores

## Troubleshooting

### Error: "Error al conectar a la base de datos"
**Solución:**
```bash
# Verificar que PostgreSQL esté corriendo
docker-compose ps postgres

# Ver logs
docker-compose logs postgres

# Reiniciar servicios
docker-compose restart
```

### Error: "Error al inicializar publisher de RabbitMQ"
**Solución:**
```bash
# Verificar que RabbitMQ esté corriendo
docker-compose ps rabbitmq

# Ver logs
docker-compose logs rabbitmq

# Acceder a Management UI
# http://localhost:15672 (guest/guest)
```

### Error: "Usuario ya existe" (409)
**Solución:** Usar un email diferente o hacer login con las credenciales existentes.

### Error: "Usuario en lista negra" (403)
**Solución:** El usuario está en la lista negra del servicio PLD. No se puede crear. Si es un error, contactar al administrador.

### Error: "Token inválido o expirado" (401)
**Solución:** Hacer login nuevamente para obtener un token fresco.

### Puerto ya en uso
**Solución:** 
```bash
# Cambiar puertos en docker-compose.yml
# O detener el servicio que usa el puerto
```

## Arquitectura

El proyecto sigue **Clean Architecture** con las siguientes capas:

- **Domain**: Entidades e interfaces (puertos)
- **UseCase**: Lógica de negocio (casos de uso)
- **Infrastructure**: Implementaciones concretas (repositorios, servicios externos)
- **Interfaces**: Handlers HTTP, DTOs, middleware

## Estructura del Proyecto

```
user-service/
├── cmd/
│   └── api/
│       └── main.go              # Punto de entrada
├── internal/
│   ├── domain/                  # Entidades e interfaces
│   ├── usecase/                 # Casos de uso
│   ├── infrastructure/          # Implementaciones
│   │   ├── repository/
│   │   ├── pld/
│   │   ├── rabbitmq/
│   │   ├── jwt/
│   │   └── logger/
│   └── interfaces/
│       └── http/                # Handlers, DTOs, middleware
├── configs/                     # Configuración
├── pkg/                         # Paquetes compartidos
├── docker-compose.yml
├── Dockerfile
└── go.mod
```

## Tests

Ejecutar todos los tests:
```bash
go test ./...
```

Ejecutar tests con cobertura:
```bash
go test -cover ./...
```

## Notas Importantes

- Las contraseñas se hashean con bcrypt (cost 10)
- Los JWT expiran según configuración (default: 24 horas)
- El servicio PLD debe estar accesible en la URL configurada
- El servicio implementa estrategia "fail-open" para PLD: si el servicio falla, se permite el registro (se registra en logs)
