# Guía para Verificar RabbitMQ - Cola user.created

Esta guía explica cómo verificar que la implementación de RabbitMQ está funcionando correctamente en el proyecto.

## Prerequisitos

1. Todos los servicios deben estar corriendo:
```bash
docker-compose up -d
```

2. Verificar que los servicios estén activos:
```bash
docker-compose ps
```

Deberías ver: `crabi_api`, `crabi_postgres`, y `crabi_rabbitmq` todos corriendo.

## Paso 1: Acceder a RabbitMQ Management UI

1. Abre tu navegador
2. Ve a: http://localhost:15672
3. Inicia sesión con:
   - **Usuario:** `guest`
   - **Contraseña:** `guest`

## Paso 2: Verificar la Cola `user.created`

1. En la barra de navegación superior, haz clic en **"Queues and Streams"**
2. Busca la cola llamada `user.created`
3. Si no aparece, no hay problema. La cola se crea automáticamente cuando se publica el primer mensaje

## Paso 3: Crear un Usuario para Generar un Evento

Para que se publique un mensaje en RabbitMQ, necesitas crear un usuario:

1. Usa Postman o curl para crear un usuario:
```bash
POST http://localhost:8080/api/v1/users
Content-Type: application/json

{
  "email": "gustavo.hernandez@example.com",
  "password": "password123",
  "name": "Gustavo Hernández"
}
```

2. Si el usuario se crea exitosamente (código 201), automáticamente:
   - Se guarda en PostgreSQL
   - Se publica un evento en la cola `user.created` de RabbitMQ

## Paso 4: Ver el Mensaje en RabbitMQ

1. Vuelve a RabbitMQ Management UI (http://localhost:15672)
2. Ve a **"Queues and Streams"**
3. Haz clic en la cola `user.created`
4. Verás información de la cola:
   - **Messages**: Cantidad de mensajes en la cola
   - **Ready**: Mensajes listos para consumir
   - **Unacked**: Mensajes siendo procesados

### Ver el Contenido del Mensaje

1. En la página de la cola `user.created`, desplázate hacia abajo
2. En la sección **"Get messages"**, haz clic para expandirla
3. Deja los valores por defecto
4. Haz clic en **"Get Message(s)"**
5. Verás el contenido del mensaje en formato JSON:
```json
{
  "user_id": "uuid-del-usuario",
  "email": "gustavo.hernandez@example.com",
  "created_at": "2024-12-03T21:11:23Z"
}
```

**Nota:** Si el consumidor está activo (lo cual es normal), los mensajes se procesan inmediatamente y la cola puede estar vacía. Esto es correcto.

## Paso 5: Verificar el Consumidor

El consumidor procesa los mensajes automáticamente cuando llegan. Para verificar que está funcionando:

1. Ver los logs de la API:
```bash
docker-compose logs api -f
```

2. Crear un usuario (Paso 3)

3. Deberías ver en los logs algo como:
```
Consumiendo mensajes de la cola: user.created
Procesando evento user.created
Enviando email de bienvenida a gustavo.hernandez@example.com
Evento procesado: user_id=..., email=...
```

## Paso 6: Ver Mensajes Acumulados (Opcional)

Si quieres ver mensajes acumulados en la cola (para inspección):

1. Detén temporalmente el consumidor:
```bash
docker-compose stop api
```

2. Crea varios usuarios usando Postman

3. Vuelve a RabbitMQ Management UI y verás los mensajes acumulados en la cola

4. Reanuda el consumidor:
```bash
docker-compose start api
```

Los mensajes se procesarán automáticamente.

## Flujo Completo de Verificación

1. ✅ RabbitMQ Management UI accesible (http://localhost:15672)
2. ✅ Crear un usuario → se publica evento en `user.created`
3. ✅ Ver el mensaje en la cola (o confirmar que se procesó inmediatamente)
4. ✅ Ver en los logs que el consumidor procesó el evento
5. ✅ Verificar en la base de datos que se guardó el evento en `user_events`

## Verificar en la Base de Datos

Para confirmar que todo el flujo funcionó:

```bash
docker exec -it crabi_postgres psql -U postgres -d crabi_db -c "SELECT * FROM user_events ORDER BY created_at DESC LIMIT 5;"
```

Deberías ver registros con:
- `user_id`: ID del usuario creado
- `event_type`: "user.created"
- `payload`: JSON con la información del evento

## Troubleshooting

### No veo la cola `user.created`
- La cola se crea cuando se publica el primer mensaje
- Crea un usuario primero (Paso 3)

### La cola está vacía
- Esto es normal si el consumidor está activo (procesa mensajes inmediatamente)
- Verifica los logs para confirmar que se procesaron: `docker-compose logs api`
- Si quieres ver mensajes acumulados, detén temporalmente el consumidor

### Error al crear usuario
- Verifica que PostgreSQL esté corriendo: `docker-compose ps postgres`
- Verifica que el API esté corriendo: `docker-compose ps api`
- Revisa los logs: `docker-compose logs api`

### No veo logs del consumidor
- Verifica que el consumidor se haya iniciado en los logs: `docker-compose logs api | Select-String -Pattern "Consumer"`
- Si no aparece, reinicia el servicio: `docker-compose restart api`
