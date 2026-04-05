# ✂️ Encurtador de URLs

Encurtador de URLs de alta performance construído com **Go**, **Redis**, **PostgreSQL** e **Next.js**.

Projeto desenvolvido para estudo de **System Design**, baseado no vídeo [Arquitetando um Encurtador de URL](https://www.youtube.com/watch?v=m_anIoKW7Jg&t=593s). O objetivo foi aplicar na prática os conceitos de arquitetura discutidos no vídeo, implementando tudo do zero em Go.

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-DC382D?style=flat&logo=redis&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?style=flat&logo=postgresql&logoColor=white)
![Next.js](https://img.shields.io/badge/Next.js-000000?style=flat&logo=next.js&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat&logo=docker&logoColor=white)

---

## Arquitetura

```
┌─────────────┐       ┌──────────────────────────────────────────┐
│   Next.js   │       │              Go API (:8080)              │
│  (Vercel)   │──────▶│                                          │
└─────────────┘       │  ┌─────────────┐    ┌────────────────┐  │
                      │  │ POST /shorten│    │GET /{shortCode}│  │
                      │  │  (Command)   │    │    (Query)     │  │
                      │  └──────┬───────┘    └───────┬────────┘  │
                      │         │                    │           │
                      │         ▼                    ▼           │
                      │  ┌─────────────┐    ┌────────────────┐  │
                      │  │ Redis INCR  │    │  Redis Cache   │  │
                      │  │ (ID único)  │    │  (leitura)     │  │
                      │  └──────┬───────┘    └───────┬────────┘  │
                      │         │                    │           │
                      │         ▼                    │ miss      │
                      │  ┌─────────────┐             ▼           │
                      │  │  HashIDs    │    ┌────────────────┐  │
                      │  │ (Base62)    │    │  PostgreSQL    │  │
                      │  └──────┬───────┘    └────────────────┘  │
                      │         │                                │
                      │         ▼                                │
                      │  ┌─────────────┐                        │
                      │  │ PostgreSQL  │                        │
                      │  │  (salvar)   │                        │
                      │  └─────────────┘                        │
                      └──────────────────────────────────────────┘
```

## Conceitos aplicados do vídeo

| Conceito | Implementação |
|----------|---------------|
| **IDs únicos em sistema distribuído** | Redis `INCR` com counter atômico, iniciando em 14M |
| **Conversão Base62 + ofuscação** | HashIDs com salt + pepper via variáveis de ambiente |
| **Proporção leitura/escrita 10:1** | Cache Redis nas leituras com TTL de 24h |
| **Redirecionamento HTTP 302** | Mantém analytics (cada request passa pelo servidor) |
| **CQRS** | Command (shorten) e Query (redirect) completamente separados |
| **VSA (Vertical Slice Architecture)** | Cada feature isolada em seu próprio pacote |

## Stack

| Camada | Tecnologia | Função |
|--------|-----------|--------|
| **API** | Go (net/http) | Servidor HTTP, lógica de negócio |
| **ID Generation** | Redis INCR | IDs únicos e atômicos |
| **Encoding** | go-hashids (Base62) | Ofuscação dos IDs sequenciais |
| **Persistência** | PostgreSQL | Armazenamento das URLs |
| **Cache** | Redis | Cache de leitura (TTL 24h) |
| **Frontend** | Next.js + Tailwind | Interface com métricas em tempo real |
| **Observabilidade** | Prometheus + Grafana | Métricas HTTP, cache hit rate, latência |
| **Documentação** | Swagger | API docs auto-geradas |

## Estrutura do projeto

```
├── main.go                          # Ponto de entrada
├── features/
│   ├── shorten/                     # Command: encurtar URL
│   │   ├── command.go               # DTO de entrada/saída
│   │   ├── handler.go               # HTTP POST /shorten
│   │   └── service.go               # Redis INCR → HashID → PostgreSQL
│   ├── redirect/                    # Query: redirecionar
│   │   ├── query.go                 # DTO de entrada
│   │   ├── handler.go               # HTTP GET /{shortCode} → 302
│   │   └── service.go               # Cache Redis → PostgreSQL → redirect
│   └── stats/                       # Query: métricas
│       ├── query.go                 # DTO de resposta
│       └── handler.go               # HTTP GET /stats
├── shared/
│   ├── config/                      # Carrega variáveis do .env
│   ├── encoder/                     # HashIDs (Base62 + salt + pepper)
│   ├── idgen/                       # Redis INCR para IDs únicos
│   ├── metrics/                     # Prometheus counters + middleware
│   ├── middleware/                   # CORS com origens configuráveis
│   ├── model/                       # Entidade URL
│   ├── dto/                         # ErrorResponse compartilhado
│   └── repository/                  # Interface + implementação PostgreSQL
├── web/                             # Frontend Next.js
├── infra/                           # Prometheus + Grafana configs
├── docker-compose.yml               # Dev local
├── Dockerfile                       # Build de produção
├── .env.example                     # Template de variáveis
└── docs/                            # Swagger (auto-gerado)
```

## Rodando localmente

**Pré-requisitos:** Go 1.26+, Docker, Node.js 18+

```bash
# 1. Clone o repositório
git clone https://github.com/Edu-Spinelli/Encurtador.git
cd Encurtador

# 2. Configure as variáveis de ambiente
cp .env.example .env

# 3. Suba Redis + PostgreSQL
docker compose up -d redis postgres

# 4. Rode a API
go run main.go

# 5. Rode o frontend (em outro terminal)
cd web
cp .env.example .env.local
npm install
npm run dev
```

| Serviço | URL |
|---------|-----|
| API | http://localhost:8080 |
| Frontend | http://localhost:3001 |
| Swagger | http://localhost:8080/swagger/index.html |
| Prometheus Metrics | http://localhost:8080/metrics |

### Observabilidade (opcional)

```bash
docker compose up -d prometheus grafana
```

| Serviço | URL | Credenciais |
|---------|-----|-------------|
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | — |

## API

### `POST /shorten`

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.google.com"}'
```

```json
{ "short_url": "http://localhost:8080/4QBzJW" }
```

### `GET /{shortCode}`

```bash
curl -I http://localhost:8080/4QBzJW
```

```
HTTP/1.1 302 Found
Location: https://www.google.com
```

### `GET /stats`

```json
{
  "urls_shortened": 42,
  "urls_redirected": 318,
  "cache_hits": 280,
  "cache_misses": 38,
  "cache_hit_rate": 88.05
}
```

## Deploy

| Serviço | Plataforma |
|---------|-----------|
| Backend Go | [Railway](https://railway.app) |
| Frontend Next.js | [Vercel](https://vercel.app) |
| PostgreSQL | Railway |
| Redis | Railway |

## Variáveis de ambiente

```env
SERVER_PORT=:8080
BASE_URL=https://seu-backend.railway.app
ALLOWED_ORIGINS=https://seu-frontend.vercel.app

REDIS_URL=redis://...
REDIS_START_OFFSET=14000000

DATABASE_URL=postgresql://...

HASH_SALT=seu-salt-secreto
HASH_PEPPER=seu-pepper-secreto
HASH_MIN_LENGTH=6
```

## Referência

Este projeto foi construído como exercício de aprendizagem, aplicando os conceitos de System Design apresentados por [Lucas Montano](https://www.youtube.com/@LucasMontano) no vídeo **"Arquitetando um Encurtador de URL"**:

[![Arquitetando um Encurtador de URL](https://img.youtube.com/vi/m_anIoKW7Jg/maxresdefault.jpg)](https://www.youtube.com/watch?v=m_anIoKW7Jg&t=593s)
