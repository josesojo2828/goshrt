# goshrt ⚡

**Acortador de URLs ultrarrápido, observable y listo para producción.**

Go + PostgreSQL + Redis + Prometheus + Grafana — todo en un `docker-compose.yml`.

---

## 🎯 **¿Qué hace?**

- Acorta URLs con **alias personalizado** o **código aleatorio base62 (6 chars = 56B combinaciones)**.
- **TTL opcional** para links temporales.
- **Cache-first**: Redis para redirecciones casi instantáneas (~sub-ms).
- **Persistencia real**: PostgreSQL como *source of truth*.
- **Métricas Prometheus** out-of-the-box + **dashboard Grafana** listo.
- **Seed por configuración**: carga URLs/DNS desde YAML al arranque.
- **Arquitectura limpia**: handlers → service → store (PostgreSQL/Redis interfaces).

---

## 📊 **Stack técnico**

| Capa | Tecnología |
|-------|------------|
| **Lenguaje** | Go 1.25+ |
| **Base de datos** | PostgreSQL 16 (source of truth) |
| **Cache** | Redis 7 (cache L1) |
| **Monitoring** | Prometheus + Grafana |
| **Router** | chi/v5 |
| **Deployment** | Docker + docker-compose |

---

## 🚀 **Quick Start (30 segundos)**

```bash
# 1. Clonar y entrar
git clone https://github.com/tuusuario/goshrt.git
cd goshrt

# 2. Crear volúmenes reutilizables (una sola vez)
docker volume create goshrt_seed
docker volume create goshrt_config

# 3. Cargar tus URLs/DNS al volumen de seed
docker run --rm \
  -v goshrt_seed:/seed \
  -v $(pwd)/config:/host_config \
  alpine sh -c "cp /host_config/urls.yaml /seed/"

# 4. Levantar TODO el stack
docker-compose up -d

# 5. Verificar
curl http://localhost:8080/health
# {"status":"ok"}

# 6. Abrir Grafana
open http://localhost:3000
# user: admin | pass: ysecreto123
```

---

## 📁 **Configuración del Seed**

Presurta URLs y DNS desde `config/urls.yaml` al volumen `goshrt_seed`:

```yaml
# config/urls.yaml (ejemplo)
urls:
  - short_code: "google"
    original_url: "https://google.com"
    ttl: 0
  - short_code: "github"
    original_url: "https://github.com/tuusuario"
  - short_code: "promo-verano"
    original_url: "https://tienda.com/promo-verano-2025"
    ttl: 2592000

dns:
  - name: "api.local"
    target: "http://host.docker.internal:8080/api"
```

> **Tip:** El archivo se monta **read-only** en `/seed/urls.yaml` dentro del contenedor. Cambiá el archivo en el host y hacé `docker-compose restart goshrt` para recargar.

---

## 🔌 **API Endpoints**

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| `POST` | `/api/url` | Crear URL corta |
| `GET` | `/{shortCode}` | Redirigir (301) |
| `GET` | `/api/url/{shortCode}/stats` | Stats de una URL |
| `GET` | `/api/urls` | Listar URLs (paginado) |
| `DELETE` | `/api/url/{shortCode}` | Eliminar (soft delete) |
| `GET` | `/health` | Health check simple |
| `GET` | `/metrics` | Prometheus metrics |

### **Ejemplos**

```bash
# Crear URL
curl -X POST http://localhost:8080/api/url \
  -H "Content-Type: application/json" \
  -d '{"url":"https://ejemplo.com","custom_alias":"mi-link","ttl_seconds":86400}'

# Response
{
  "short_code": "mi-link",
  "original_url": "https://ejemplo.com",
  "created_at": "2026-07-14T10:30:00Z"
}

# Redirigir
curl -I http://localhost:8080/mi-link
# HTTP/1.1 301 Moved Permanently
# Location: https://ejemplo.com

# Stats
curl http://localhost:8080/api/url/mi-link/stats

# Listar (paginado)
curl "http://localhost:8080/api/urls?page=1&limit=20"

# Eliminar
curl -X DELETE http://localhost:8080/api/url/mi-link
```

---

## 📊 **Monitoreo: Prometheus + Grafana**

### **Métricas expuestas (`/metrics`)**
- `goshrt_urls_created_total` (contador)
- `goshrt_redirect_duration_seconds` (histograma)
- `goshrt_cache_hit_ratio` (promedio del ratio de hits en cache)

### **Dashboard en Grafana**
- **URLs Lean**: Gráfico de barras con URLs recientes
- **Latencia P95**: Distribución de tiempos de redirección
- **Ratio de Cache Hit**: Porcentaje de redirecciones servidas desde Redis

Accede a Grafana en `http://localhost:3000` con usuario `admin` y contraseña `ysecreto123` (cambiable en producción).

---

## 🧪 **Testing**

```bash
go test ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
```

### **Makefile targets**
```bash
make test
make lint
make migrate-up
```

---

## 🚢 **Deployment a Producción**

### 🔐 Secrets
```bash
POSTGRES_PASSWORD=supersecreto123
REDIS_PASSWORD=redis-secreto
GF_SECURITY_ADMIN_PASSWORD=admin-seguro123
```

### ⚙️ `docker-compose.yml`
```yaml
services:
  goshrt:
    environment:
      POSTGRES_DSN: "postgres://goshrt:${POSTGRES_PASSWORD}@postgres:5432/goshrt"
      REDIS_PASSWORD: "${REDIS_PASSWORD}"
    deploy:
      replicas: 3
```

---

## 🔒 **Seguridad**
- **Rate limiting**: middleware `chi` + `tollbooth`
- **Auth**: JWT en endpoints críticos
- **Validación**: Sanitizar URLs
- **CORS**: Solo orígenes permitidos

---

## 📦 **Extensiones Futuras**
- Dominios personalizados
- Generador de QR codes
- Dashboard de analytics avanzado
- Webhooks
- API Keys con rate limits

---

## 🙏 **Agradecimientos**
- **chi**: Router HTTP minimalista y rápido
- **sqlx**: Extensiones de `database/sql`
- **go-redis/v9**: Cliente Redis robusto
- **prometheus/client_golang**: Métricas nativas
- **grafana**: Dashboards visuales
- **MaxMind GeoLite2**: Geolocalización (para Proxi-pulse)

---

> **Hecho por https://www.quanticarch.com/** 🇻🇪

**Tags:** `go` `url-shortener` `postgresql` `redis` `prometheus` `grafana` `docker` `observability` `clean-architecture` `golang`
