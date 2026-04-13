# Benchmark: Escalando um Encurtador de URLs na AWS

Data: 2026-04-12
Objetivo: Medir o impacto de cada camada de infraestrutura na performance do sistema
Ferramenta: K6 (rodando dentro da mesma VPC na AWS)
Aplicacao: Encurtador de URLs em Go + Redis + PostgreSQL

---

## Configuracao do teste K6

Cenario de escrita (POST /shorten):
- Envia URLs aleatorias para encurtar
- Mede: req/s, latencia p50/p95/p99, taxa de erro

Cenario de leitura (GET /{code}):
- Acessa URLs curtas ja criadas
- Mede: req/s, latencia p50/p95/p99, taxa de erro, cache hit rate

Proporcao: 1 escrita : 10 leituras (realista conforme video)

---

## Etapa 1: Go sozinho + PostgreSQL + Redis local (sem cache de leitura, sem LB)

Infra:
- 1x EC2 t3.xlarge rodando Go + Redis (Docker)
- 1x RDS db.t3.micro PostgreSQL
- Redis local apenas para INCR (sem cache de leitura)
- Sem Load Balancer

Resultados K6:
- VUs (virtual users): 1100 max (100 write + 1000 read)
- Duracao do teste: 3m35s
- Total de requests: 950.232
- Requests/s (total): 4.402 req/s
- Requests/s (escrita): ~248 req/s (53.515 em ~3.5min)
- Requests/s (leitura): ~4.154 req/s (896.017 em ~3.5min)
- Latencia media (escrita): 234.71ms
- Latencia p50 (escrita): 202.84ms
- Latencia p90 (escrita): 434.46ms
- Latencia p95 (escrita): 504.78ms
- Latencia max (escrita): 1.38s
- Latencia media (leitura): 139.37ms
- Latencia p50 (leitura): 142.06ms
- Latencia p90 (leitura): 227.43ms
- Latencia p95 (leitura): 293.83ms
- Latencia max (leitura): 546.03ms
- Taxa de erro: 0.00%
- Observacoes: Sem cache de leitura, todas as leituras vao direto ao RDS. O gargalo e o banco (db.t3.micro). Escrita tem p95 alto (504ms) porque cada INSERT depende do RDS. Leitura p95 de 293ms tambem alto, cada SELECT bate no banco. Baseline solido pra comparar com cache.

---

## Etapa 2: + Redis cache (ElastiCache)

Infra:
- 1x EC2 t3.xlarge rodando o Go (Docker)
- 1x RDS db.t3.micro PostgreSQL
- 1x ElastiCache cache.t3.micro Redis
- Sem Load Balancer

Mudanca: leituras passam pelo cache Redis (ElastiCache) antes de ir ao banco

Resultados K6:
- VUs (virtual users): 1100 max (100 write + 1000 read)
- Duracao do teste: 3m35s
- Total de requests: 864.425
- Requests/s (total): 4.003 req/s
- Requests/s (escrita): ~209 req/s (45.160 em ~3.5min)
- Requests/s (leitura): ~3.794 req/s (818.565 em ~3.5min)
- Latencia media (escrita): 270.35ms
- Latencia p50 (escrita): 216.77ms
- Latencia p90 (escrita): 535.15ms
- Latencia p95 (escrita): 635.64ms
- Latencia max (escrita): 1.37s
- Latencia media (leitura): 145.15ms
- Latencia p50 (leitura): 131.55ms
- Latencia p90 (leitura): 288.56ms
- Latencia p95 (leitura): 351.29ms
- Latencia max (leitura): 1s
- Cache hit rate: 99.97% (818.354 hits / 211 misses)
- Taxa de erro: 0.00%
- Observacoes: Cache hit rate quase perfeito (99.97%). A latencia nao caiu significativamente porque o gargalo nao e mais o banco, e sim a EC2 unica processando 1100 VUs. Com cache, o banco quase nao e tocado nas leituras (apenas 211 SELECTs vs 818K requests). O real beneficio do cache aparecera quando escalarmos horizontalmente (Etapa 3), pois o banco nao sera mais o bottleneck.

Comparativo com Etapa 1:
- Melhoria req/s total: -9% (4.003 vs 4.402, variacao normal)
- Cache hit rate: 0% → 99.97%
- Carga no banco (leituras): 896.017 SELECTs → 211 SELECTs (reducao de 99.97%)
- Conclusao: o cache eliminou a carga no banco mas a EC2 unica e o gargalo agora

---

## Etapa 3: + ALB + 3 instancias ECS Fargate

Infra:
- 1x ALB (Application Load Balancer)
- 3x ECS Fargate tasks (1 vCPU + 2GB cada)
- 1x RDS db.t3.micro PostgreSQL
- 1x ElastiCache cache.t3.micro Redis

Mudanca: trafego distribuido entre 3 instancias via ALB

Resultados K6:
- VUs (virtual users): 1100 max (100 write + 1000 read)
- Duracao do teste: 3m36s
- Total de requests: 1.544.418
- Requests/s (total): 7.146 req/s
- Requests/s (escrita): ~366 req/s (79.224 em ~3.6min)
- Requests/s (leitura): ~6.780 req/s (1.464.494 em ~3.6min)
- Latencia media (escrita): 124.82ms
- Latencia p50 (escrita): 71.81ms
- Latencia p90 (escrita): 311.13ms
- Latencia p95 (escrita): 417.5ms
- Latencia max (escrita): 1.54s
- Latencia media (leitura): 57.91ms
- Latencia p50 (leitura): 43.82ms
- Latencia p90 (leitura): 110.49ms
- Latencia p95 (leitura): 159.79ms
- Latencia max (leitura): 1.41s
- Taxa de erro: 0.16% escrita (134 falhas), 0% leitura
- Observacoes: Salto massivo. 3 instancias com ALB distribuindo carga. Leitura p95 caiu de 351ms pra 159ms. req/s quase dobrou. O cache continua protegendo o banco.

Comparativo com Etapa 1 (baseline):
- req/s total: 4.402 → 7.146 (+62%)
- Leitura p95: 293ms → 159ms (-45%)
- Escrita p95: 504ms → 417ms (-17%)
- Total requests: 950K → 1.54M (+62%)

Comparativo com Etapa 2:
- req/s total: 4.003 → 7.146 (+78%)
- Leitura p95: 351ms → 159ms (-54%)
- Escrita p95: 635ms → 417ms (-34%)

---

## Etapa 4: + RDS read replica

Infra:
- 1x ALB
- 3x ECS Fargate tasks (Go API)
- 1x RDS db.t3.micro PostgreSQL (master, escrita)
- 1x RDS db.t3.micro PostgreSQL (read replica, leitura)
- 1x ElastiCache cache.t3.micro Redis

Mudanca: leituras que dao cache miss vao pra read replica, nao pro master

Resultados K6:
- VUs (virtual users): 1100 max (100 write + 1000 read)
- Duracao do teste: 3m35s
- Total de requests: 2.627.571
- Requests/s (total): 12.182 req/s
- Requests/s (escrita): ~395 req/s (85.469 em ~3.6min)
- Requests/s (leitura): ~11.787 req/s (2.541.402 em ~3.6min)
- Latencia media (escrita): 139.28ms
- Latencia p50 (escrita): 51.7ms
- Latencia p90 (escrita): 439.62ms
- Latencia p95 (escrita): 509.05ms
- Latencia max (escrita): 2.74s
- Latencia media (leitura): 40.32ms
- Latencia p50 (leitura): 34.64ms
- Latencia p90 (leitura): 66.42ms
- Latencia p95 (leitura): 102.3ms
- Latencia max (leitura): 980.28ms
- Taxa de erro: 0.13% escrita (114 falhas), 0% leitura
- Observacoes: Salto brutal. A read replica liberou o master pra focar em escritas. Leitura p95 caiu pra 102ms. req/s total quase triplicou vs baseline. 2.6M requests processados em 3.5min com 0% erro de leitura.

Comparativo com Etapa 3:
- req/s total: 7.146 → 12.182 (+70%)
- Leitura p95: 159ms → 102ms (-35%)
- Total requests: 1.54M → 2.62M (+70%)

Comparativo com Etapa 1 (baseline):
- req/s total: 4.402 → 12.182 (+176%)
- Leitura p95: 293ms → 102ms (-65%)
- Total requests: 950K → 2.62M (+176%)

---

## Etapa 5: + Auto Scaling (ate 10 instancias)

Infra:
- 1x ALB
- 3-10x ECS Fargate tasks (1 vCPU + 2GB) com auto scaling (CPU > 50%)
- 1x RDS db.t3.micro PostgreSQL (master)
- 1x RDS db.t3.micro PostgreSQL (read replica)
- 1x ElastiCache cache.t3.micro Redis

Mudanca: ECS escala automaticamente de 3 ate 10 instancias sob carga. Carga K6 aumentada de 1100 pra 3300 VUs (3000 leitura + 300 escrita).

Resultados K6:
- VUs (virtual users): 3300 max (300 write + 3000 read)
- Duracao do teste: 5m35s
- Total de requests: 4.335.438
- Requests/s (total): 12.908 req/s
- Requests/s (escrita): ~452 req/s (150.608 em ~5.5min)
- Requests/s (leitura): ~12.456 req/s (4.184.630 em ~5.5min)
- Latencia media (escrita): 417.21ms
- Latencia p50 (escrita): 166.17ms
- Latencia p90 (escrita): 1.23s
- Latencia p95 (escrita): 1.39s
- Latencia max (escrita): 8.6s
- Latencia media (leitura): 139.29ms
- Latencia p50 (leitura): 134ms
- Latencia p90 (leitura): 236.7ms
- Latencia p95 (leitura): 383.45ms
- Latencia max (leitura): 1.18s
- Instancias ativas no pico: 6 (escalou de 3 → 4 → 6 automaticamente)
- Taxa de erro: 0.05% escrita (85 falhas), 0% leitura
- Observacoes: Auto scaling disparou durante o teste. ECS escalou de 3 pra 6 tasks. 4.3 milhoes de requests processados em 5.5 minutos com 0% erro de leitura. A latencia aumentou um pouco porque a carga tambem aumentou (3x mais VUs que etapas anteriores), mas o sistema absorveu sem colapsar.

Comparativo com Etapa 4 (mesmo teste teria sido mais justo):
- Carga K6: 3x maior (3300 VUs vs 1100 VUs)
- req/s total: 12.182 → 12.908 (+6%, mas sob 3x mais carga)
- Total requests: 2.62M → 4.33M (+65%)

Comparativo com Etapa 1 (baseline):
- req/s total: 4.402 → 12.908 (+193%)
- Total requests: 950K → 4.33M (+356%)
- Instancias: 1 → 6 (auto scaling)
- Observacoes:

Comparativo com Etapa 4:
- Melhoria req/s total:
- Melhoria latencia p95:

---

## Resumo final

| Etapa | Infra | req/s | p95 leitura | p95 escrita | total requests | VUs |
|-------|-------|-------|-------------|-------------|----------------|-----|
| 1 | Go + PG (sem cache)      | 4.402  | 293ms | 504ms | 950K   | 1100 |
| 2 | + Redis cache            | 4.003  | 351ms | 635ms | 864K   | 1100 |
| 3 | + ALB + 3 Fargate        | 7.146  | 159ms | 417ms | 1.54M  | 1100 |
| 4 | + read replica           | 12.182 | 102ms | 509ms | 2.62M  | 1100 |
| 5 | + auto scaling (3→6)     | 12.908 | 383ms | 1.39s | 4.33M  | 3300 |

Fator de melhoria (Etapa 1 → 5):
- req/s: +193% (4.402 → 12.908)
- Total requests: +356% (950K → 4.33M)
- Instancias: 1 → 6 (automatico)
- Com 3x mais VUs na Etapa 5

Observacao: a Etapa 5 usou carga 3x maior (3300 VUs) pra forcar o auto scaling disparar. Se comparar apenas req/s sob mesma carga, a Etapa 4 (com 1100 VUs em 12.182 req/s) ja era proxima do throughput maximo sustentavel com 3 instancias. A Etapa 5 provou que o sistema escala automaticamente sob carga maior sem colapsar.

---

## Custo estimado AWS (por hora, tudo ligado)

| Servico | Tipo | Custo/h |
|---------|------|---------|
| EC2 K6 | t3.medium | |
| ECS Fargate | 0.25 vCPU x N | |
| RDS master | db.t3.micro | |
| RDS replica | db.t3.micro | |
| ElastiCache | cache.t3.micro | |
| ALB | | |
| Total | | |
