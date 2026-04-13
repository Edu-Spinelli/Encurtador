Continuando o meu hiperfoco na playlist do Lucas Montano (inclusive recomendo demais), dessa vez assisti o vídeo sobre System Design de um Encurtador de URLs, no estilo Bitly, TinyURL.

https://www.youtube.com/watch?v=m_anIoKW7Jg

E como sempre, de estudo virou projeto.

Construi um encurtador de URLs do zero em Go, aplicando cada conceito que o Lucas discutiu no video. O fluxo ficou assim:

O usuario manda uma URL longa. A API Go pede ao Redis um ID unico (comando INCR, atomico, sem colisao mesmo com varios servidores). Esse ID numerico passa por um encoding Base62 com HashIDs usando salt e pepper (guardados em variaveis de ambiente, nunca no codigo), e vira um codigo tipo "4QBzJW". Salva no PostgreSQL e devolve o link curto.

Quando alguem acessa o link curto, primeiro checa o cache Redis (TTL de 24h). Se achar, redireciona na hora. Se nao, busca no PostgreSQL, salva no cache pra proxima vez, e redireciona com status 302 (nao 301, pra cada request passar pelo servidor e permitir analytics).

Algumas decisoes que vieram direto do video:

O counter do Redis comeca em 14 milhoes. Por que? Pra garantir que os codigos ja nascem com 4 ou 5 caracteres. Codigos curtos demais (tipo "a", "b") facilitam brute force.

CQRS natural. O fluxo de escrita (encurtar) e o de leitura (redirecionar) tem necessidades completamente diferentes. Escrita precisa de consistencia. Leitura precisa de velocidade. Separar faz sentido.

A proporção leitura/escrita e de 10:1. O cache Redis absorve a maioria das leituras sem tocar no banco.

E como escala horizontalmente?

A API Go e stateless. Qualquer instancia responde qualquer request. Coloca um load balancer na frente, sobe 10, 50, 100 instancias, o Redis continua garantindo IDs unicos pra todas. PostgreSQL com read replicas distribui a carga de leitura. O cache absorve os picos (aquele link que viraliza).

Uma coisa que me fez pensar bastante: uma vez ouvi a pergunta "se voce deixasse seu projeto open source hoje, seu produto estaria protegido e seguro?". Isso mudou minha forma de codar. Nesse projeto, desde o primeiro commit o repositorio ja era publico. Salt, pepper, connection strings, tudo via variaveis de ambiente. Zero segredo no codigo. Se alguem clonar o repo agora, nao encontra nenhuma credencial. E isso te forca a trabalhar do jeito certo desde o inicio.

Testa ai: https://encurtador-peach.vercel.app

Stack: Go, Redis, PostgreSQL, Next.js, Docker, Prometheus, Grafana
Arquitetura: Vertical Slice (VSA) + CQRS
Deploy: Railway (backend) + Vercel (frontend)

Repo (publico): https://github.com/Edu-Spinelli/Encurtador

#golang #systemdesign #redis #postgresql #backend #opensource
