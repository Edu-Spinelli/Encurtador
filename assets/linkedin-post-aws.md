Lembra do encurtador de URLs em Go que eu publiquei semana passada? Pois e. Depois que terminei, assisti esse video do Renato Augusto sobre escalar uma arquitetura do zero a um milhao de usuarios:

https://www.youtube.com/watch?v=9g7twJrXqoY

E fiquei com uma pergunta na cabeca: quanto que a aplicacao que eu construi aguenta de verdade? Aquela versao no Railway com PostgreSQL e Redis gerenciados, como ela se comporta sob carga? E mais importante: como cada camada de infraestrutura impacta no numero final?

Resolvi descobrir na pratica. Subi a mesma aplicacao na AWS, mas em 5 etapas incrementais, rodando K6 em cada uma pra medir req/s, latencia e taxa de erro. Toda a infra foi provisionada com Terraform, pra conseguir subir e destruir rapido sem ficar gastando nuvem a toa.

O plano foi: depois de cada etapa, anotar os numeros. No final, comparar tudo. O K6 rodou dentro da mesma VPC pra nao ter latencia de rede externa mascarando o resultado.

Etapa 1: Go rodando sozinho numa EC2 + PostgreSQL RDS. Sem cache de leitura, sem load balancer, nada. O Redis ficou apenas pro INCR (gerador de IDs). Resultado: 4.402 req/s. Leitura p95 em 293ms. Escrita p95 em 504ms. 950 mil requests em 3.5 minutos. Baseline estabelecido.

Etapa 2: coloquei o ElastiCache Redis fazendo cache de leitura com TTL de 24h. Resultado: 4.003 req/s. A latencia nao caiu. O cache hit rate foi de 99.97%, com apenas 211 SELECTs de 818 mil leituras chegando no banco. O aprendizado aqui foi contra-intuitivo. O cache sozinho nao reduziu latencia, porque o round-trip de rede pro ElastiCache e parecido com o round-trip pro RDS. O real beneficio do cache e liberar o banco, nao acelerar a leitura individual. Isso so fica visivel quando voce escala horizontalmente.

Etapa 3: joguei o ECS Fargate com 3 tasks atras de um Application Load Balancer. Cada task com 1 vCPU e 2GB. Saltou pra 7.146 req/s. Leitura p95 caiu de 351ms pra 159ms. A API Go stateless tem uma vantagem absurda aqui, tres instancias absorvendo a carga que uma sozinha nao dava conta. O benefico do cache da Etapa 2 finalmente apareceu, porque o banco nao virou gargalo mesmo com 3x mais throughput chegando nele.

Etapa 4: criei uma read replica do PostgreSQL. O master ficou dedicado pra escrita, a replica pra leitura. Saltou pra 12.182 req/s. Leitura p95 caiu pra 102ms. 2.62 milhoes de requests processados. O banco finalmente parou de ser o gargalo.

Etapa 5: liguei o Auto Scaling do ECS com threshold de 50% de CPU. Aumentei a carga do K6 de 1100 pra 3300 VUs pra forcar o scaling. O sistema escalou automaticamente de 3 pra 6 tasks durante o teste. 12.908 req/s. 4.33 milhoes de requests em 5.5 minutos. Zero erro na leitura.

Do comeco ao fim, o throughput aumentou 193% e o total de requests processados cresceu 356%. A aplicacao absorveu 3x mais carga sem colapsar.

Algumas coisas que aprendi nesse processo:

Nao adianta jogar cache numa instancia unica esperando milagre. Cache bem colocado nao serve pra reduzir latencia, serve pra proteger o banco quando voce escalar horizontalmente.

Read replica e a unica coisa que realmente libera o banco quando a carga de leitura ultrapassa a capacidade de uma instancia de banco.

Auto scaling parece magico mas exige calibragem. Threshold muito alto nunca dispara, muito baixo fica escalando por qualquer coisa.

O Terraform foi o que tornou isso viavel. Cada etapa era trocar umas variaveis no tfvars, rodar apply, testar, rodar destroy. Total gasto na AWS em todas as etapas juntas: menos de 2 dolares.

O codigo e os resultados completos estao no repositorio (publico). Cada etapa tem os numeros detalhados, os scripts K6, e o Terraform que provisiona aquela configuracao especifica.

Repo: https://github.com/Edu-Spinelli/Encurtador

Agora testar ai e ver em acao: https://encurtador-peach.vercel.app

Stack: Go, PostgreSQL (RDS), Redis (ElastiCache), Next.js, Docker, Terraform
AWS: ECS Fargate, ALB, RDS com read replica, Auto Scaling
Load test: K6

E ja fica um gostinho do proximo passo: replicar exatamente a mesma infraestrutura pra um encurtador em .NET e outro em Java. Mesmo Terraform, mesmo K6, mesma carga. Depois comparar os 3 lado a lado e ver qual linguagem se sai melhor. Promete render.

#aws #terraform #systemdesign #golang #loadtesting #backend
