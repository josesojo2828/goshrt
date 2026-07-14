# goshrt ⚡

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?logo=postgresql&logoColor=white)](https://postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-DC382D?logo=redis&logoColor=white)](https://redis.io)
[![Prometheus](https://img.shields.io/badge/Metrics-Prometheus-E6522C?logo=prometheus)](https://prometheus.io)
[![Grafana](https://img.shields.io/badge/Dashboard-Grafana-F46800?logo=grafana)](https://grafana.com)

**Acortador de URLs** con caché Redis, métricas Prometheus y dashboard Grafana. Pensado para autohosting — rápido, liviano y con todo el stack de monitoreo incluido.

## Features

- ⚡ **Cache-first**: Las URLs se sirven desde Redis, no desde PostgreSQL
- 🔗 **Links temporales**: TTL configurable por URL (expiran automáticamente)
- 📊 **Métricas**: Prometheus + Grafana dashboard pre-configurado
- 📦 **Seed YAML**: Cargá URLs desde un archivo YAML al iniciar
- 🐳 **Docker Compose**: Stack completo (app + postgres + redis + prometheus + grafana)
- 🔄 **Click tracking**: Contador de clics sincronizado asincrónicamente

## Stack

| Capa | Tecnología |
|------|-----------|
| Router | chi/v5 |
| Store | PostgreSQL (pgx/v5) |
| Cache | Redis (go-redis/redis) |
| Métricas | prometheus/client_golang |
| Dashboard | Grafana (pre-provisionado) |

## Quick Start

```bash
cp .env.example .env  # o configurar vars de entorno
docker compose up -d
```

### Servicios

| Puerto | Servicio |
|--------|----------|
| :8080 | goshrt API |
| :5432 | PostgreSQL |
| :6379 | Redis |
| :9090 | Prometheus |
| :3000 | Grafana (admin:admin) |

## API

### Crear URL corta

```bash
curl -X POST http://localhost:8080/api/url \
  -H "Content-Type: application/json" \
  -d '{"url": "https://ejemplo.com/muy/larga/ruta"}'

# Response: {"short_code":"Ab3XyZ","original_url":"...","created_at":"..."}
```

### Redirección

```bash
curl -L http://localhost:8080/Ab3XyZ
# → Redirecciona 301 a la URL original
```

### Más endpoints

| Método | Ruta | Descripción |
|--------|------|-------------|
| POST | /api/url | Crear URL corta |
| GET | /api/urls?page=1&limit=20 | Listar URLs |
| GET | /api/url/{shortCode}/stats | Stats de una URL |
| DELETE | /api/url/{shortCode} | Eliminar URL |
| GET | /health | Health check |
| GET | /metrics | Métricas Prometheus |

### Parámetros adicionales

```bash
# Con alias personalizado y TTL de 1 hora
curl -X POST http://localhost:8080/api/url \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://ejemplo.com",
    "custom_alias": "mi-link",
    "ttl_seconds": 3600
  }'
```

## Dashboard Grafana

Incluye dashboard pre-configurado con:
- URLs activas, creadas y eliminadas
- Tasa de redirecciones (cache hit vs miss)
- Latencia de requests
- Operaciones de caché (hit/miss/set)
- Errores de base de datos

## Variables de Entorno

| Variable | Default | Descripción |
|----------|---------|-------------|
| PORT | 8080 | Puerto del servidor |
| POSTGRES_DSN | postgres://goshrt:goshrt@localhost:5432/goshrt?sslmode=disable | DSN de PostgreSQL |
| REDIS_ADDR | localhost:6379 | Dirección de Redis |
| REDIS_PASSWORD | — | Password de Redis |
| CLICK_SYNC_INTERVAL | 5s | Intervalo de sync de clics |

## Desarrollo

```bash
make build      # Compilar
make run        # Ejecutar
make test       # Tests
make docker     # Build imagen
```

## Licencia

MIT
