# Documentacao Completa da API Invitely

Este documento descreve a API Go do Invitely: tecnologias usadas, como executar o projeto, autenticacao, padrao de envio e retorno de dados, erros, Swagger e todas as rotas disponiveis.

## Visao Geral

A API Invitely gerencia eventos, convidados, confirmacoes de presenca, painel de metricas, analises, orcamento e lista de presentes. Ela usa autenticacao via Supabase e organiza os dados por `tenant`, ou seja, por organizacao/conta.

Base URL local:

```http
http://localhost:8080
```

Documentacao interativa:

```http
http://localhost:8080/swagger
```

## Tecnologias

- Go `1.22`
- Gin `v1.10.1` para roteamento HTTP
- PostgreSQL como banco de dados
- Supabase Auth para cadastro, login e validacao de usuario
- Docker e Docker Compose para ambiente local
- OpenAPI `3.0.3` para documentacao Swagger
- `pgx` como driver PostgreSQL
- `gopkg.in/yaml.v3` para gerar o JSON do Swagger a partir do YAML embutido
- Brevo para envio de e-mails transacionais de lembrete

## Estrutura do Projeto

```text
cmd/api/main.go              Entrada da aplicacao
routes/routes.go             Registro das rotas Gin
internal/config              Leitura de variaveis de ambiente
internal/database            Conexao PostgreSQL e migrations
internal/middleware          Auth, CORS e autorizacao por papel
internal/auth                Login, cadastro e usuario atual
internal/events              Eventos
internal/reminders           Campanhas de lembrete de RSVP
internal/guests              Convidados
internal/rsvp                Confirmacao de presenca
internal/dashboard           Visao geral do painel
internal/analytics           Resumo de metricas por evento
internal/budget              Orcamento do evento
internal/gifts               Presentes do evento
internal/apidocs             Swagger embutido no binario
docs/swagger.yaml            OpenAPI em YAML
```

## Ambiente

Variaveis principais:

```env
APP_ENV=development
PORT=8080
DATABASE_URL=postgres://postgres:postgres@postgres:5432/invitely?sslmode=disable
SUPABASE_URL=https://seu-projeto.supabase.co
SUPABASE_ANON_KEY=sua_anon_key
SUPABASE_SERVICE_ROLE_KEY=sua_service_role_key
CORS_ALLOWED_ORIGINS=http://localhost:5173
BREVO_FROM_NAME=Invitely
BREVO_API_KEY=
BREVO_SMTP_HOST=smtp-relay.brevo.com
BREVO_SMTP_PORT=587
BREVO_SMTP_USERNAME=
BREVO_SMTP_KEY=
```

Para Brevo existem dois modos suportados:

- API HTTP: use `BREVO_API_KEY`, normalmente com prefixo `xkeysib-`.
- SMTP: use `BREVO_SMTP_USERNAME` e `BREVO_SMTP_KEY`, normalmente com chave `xsmtpsib-`.

Se a chave comecar com `xsmtpsib-`, ela e uma chave SMTP. Nesse caso tambem e obrigatorio informar o SMTP login do Brevo em `BREVO_SMTP_USERNAME`.

No Docker Compose, a API usa o banco interno:

```env
DATABASE_URL=postgres://postgres:postgres@postgres:5432/invitely?sslmode=disable
```

## Executando

Subir a API e o PostgreSQL:

```bash
docker compose up -d
```

Ver logs da API:

```bash
docker compose logs -f api
```

Rodar testes:

```bash
docker compose exec -T api go test ./...
```

Health check:

```bash
curl http://localhost:8080/health
```

O ambiente local usa watcher de desenvolvimento. Ao alterar arquivos `.go`, `.yaml`, `go.mod` ou `go.sum`, o container recompila e reinicia a API automaticamente, sem rebuild manual do Docker.

## Banco de Dados

As migrations rodam automaticamente ao iniciar a API.

Tabelas principais:

- `tenants`
- `users`
- `events`
- `guests`
- `rsvps`
- `event_budget_items`
- `event_gifts`

Eventos removidos apagam dados relacionados em cascata, como convidados, confirmacoes, itens de orcamento e presentes.

## Autenticacao

A API usa Bearer Token no header `Authorization`.

Formato:

```http
Authorization: Bearer <access_token>
```

Rotas publicas:

- `GET /`
- `GET /health`
- `GET /swagger`
- `GET /swagger/doc.json`
- `GET /swagger/doc.yaml`
- `POST /auth/login`
- `POST /auth/register`
- `POST /rsvp`

Rotas protegidas exigem token:

- Eventos
- Lembretes
- Convidados
- Orcamento
- Presentes
- Dashboard
- Analytics
- `GET /auth/me`

## Padrao de Envio

A maioria das rotas recebe JSON:

```http
Content-Type: application/json
```

Exemplo:

```json
{
  "title": "Lancamento do Produto",
  "description": "Evento exclusivo para convidados.",
  "starts_at": "2026-07-20T19:00:00Z",
  "ends_at": "2026-07-20T23:00:00Z",
  "location": "Sao Paulo"
}
```

Datas usam formato ISO 8601/RFC3339.

## Padrao de Retorno

Sucesso com dados:

```json
{
  "data": {
    "id": "uuid",
    "title": "Evento"
  }
}
```

Sucesso com lista:

```json
{
  "data": []
}
```

Erro:

```json
{
  "error": "mensagem do erro"
}
```

Exclusao com sucesso retorna `204 No Content`, sem corpo.

## Codigos HTTP

- `200 OK`: consulta ou atualizacao feita com sucesso
- `201 Created`: recurso criado com sucesso
- `204 No Content`: recurso deletado com sucesso
- `400 Bad Request`: corpo invalido ou regra de validacao falhou
- `401 Unauthorized`: token ausente ou invalido
- `404 Not Found`: recurso nao encontrado ou nao pertence ao tenant autenticado
- `500 Internal Server Error`: erro inesperado no servidor

## Swagger

Rotas da documentacao:

```http
GET /swagger
GET /swagger/
GET /swagger/index.html
GET /swagger/doc.json
GET /swagger/doc.yaml
```

O arquivo fonte da documentacao fica em:

```text
docs/swagger.yaml
internal/apidocs/swagger.yaml
```

`internal/apidocs/swagger.yaml` e embutido no binario Go. O endpoint `GET /swagger/doc.json` gera JSON a partir desse YAML.

## Rotas

### Raiz

#### `GET /`

Redireciona para a documentacao Swagger.

Resposta:

```http
307 Temporary Redirect
Location: /swagger
```

### Saude

#### `GET /health`

Verifica se a API esta online.

Resposta `200`:

```json
{
  "data": {
    "status": "ok"
  }
}
```

## Autenticacao

### `POST /auth/register`

Cadastra um usuario usando Supabase e retorna tokens de autenticacao.

Autenticacao: publica.

Body:

```json
{
  "email": "admin@example.com",
  "password": "secret123",
  "name": "Usuario Admin",
  "role": "tenant_admin"
}
```

Resposta `201`:

```json
{
  "data": {
    "access_token": "jwt",
    "refresh_token": "refresh_token",
    "token_type": "bearer",
    "expires_in": 3600,
    "user": {
      "id": "uuid",
      "tenant_id": "uuid",
      "supabase_user_id": "uuid",
      "email": "admin@example.com",
      "name": "Usuario Admin",
      "role": "tenant_admin"
    }
  }
}
```

### `POST /auth/login`

Autentica usuario existente no Supabase.

Autenticacao: publica.

Body:

```json
{
  "email": "admin@example.com",
  "password": "secret123"
}
```

Resposta `200`: mesmo formato de `POST /auth/register`.

### `GET /auth/me`

Retorna o usuario autenticado a partir do Bearer Token.

Autenticacao: obrigatoria.

Headers:

```http
Authorization: Bearer <access_token>
```

Resposta `200`:

```json
{
  "data": {
    "id": "uuid",
    "tenant_id": "uuid",
    "supabase_user_id": "uuid",
    "email": "admin@example.com",
    "name": "Usuario Admin",
    "role": "tenant_admin"
  }
}
```

## Eventos

### `GET /events`

Lista eventos do tenant autenticado.

Autenticacao: obrigatoria.

Resposta `200`:

```json
{
  "data": [
    {
      "id": "uuid",
      "tenant_id": "uuid",
      "title": "Lancamento do Produto",
      "description": "Evento exclusivo para convidados.",
      "starts_at": "2026-07-20T19:00:00Z",
      "ends_at": "2026-07-20T23:00:00Z",
      "location": "Sao Paulo",
      "slug": "lancamento-do-produto-12345678",
      "created_at": "2026-06-15T12:00:00Z",
      "updated_at": "2026-06-15T12:00:00Z"
    }
  ]
}
```

### `POST /events`

Cria evento no tenant autenticado.

Autenticacao: obrigatoria.

Body:

```json
{
  "title": "Lancamento do Produto",
  "description": "Evento exclusivo para convidados.",
  "starts_at": "2026-07-20T19:00:00Z",
  "ends_at": "2026-07-20T23:00:00Z",
  "location": "Sao Paulo"
}
```

Resposta `201`: objeto `Evento` dentro de `data`.

### `GET /events/{id}`

Busca um evento por ID, desde que ele pertenca ao tenant autenticado. Usuario `platform_admin` tambem pode acessar.

Autenticacao: obrigatoria.

Resposta `200`: objeto `Evento` dentro de `data`.

### `PUT /events/{id}`

Atualiza dados principais de um evento.

Autenticacao: obrigatoria.

Body:

```json
{
  "title": "Lancamento Atualizado",
  "description": "Descricao atualizada.",
  "starts_at": "2026-07-21T19:00:00Z",
  "ends_at": "2026-07-21T23:00:00Z",
  "location": "Rio de Janeiro"
}
```

Resposta `200`: objeto `Evento` atualizado dentro de `data`.

### `DELETE /events/{id}`

Remove um evento do tenant autenticado.

Autenticacao: obrigatoria.

Resposta:

```http
204 No Content
```

## Lembretes

Lembretes permitem que o organizador dispare uma campanha de e-mail para convidados de um evento. A API valida os dados, garante que o evento pertence ao tenant autenticado, registra a campanha e envia pelo Brevo em background.

A rota retorna rapidamente com status `queued`. Quando o Brevo aceita o envio, a campanha fica com status `sent`. Se o provedor falhar, a campanha fica com status `failed`.

Guia especifico desta integracao:

```text
docs/LEMBRETES_EMAIL.md
```

### `POST /events/{id}/reminders`

Cria uma campanha de lembrete de RSVP para um evento.

Autenticacao: obrigatoria.

Limites e validacoes:

- `id` precisa ser um evento existente.
- O evento precisa pertencer ao tenant autenticado.
- `from_email` precisa ser um e-mail valido.
- `recipients` precisa ter pelo menos 1 e-mail valido.
- `recipients` aceita no maximo 200 e-mails por requisicao.
- `subject` e obrigatorio.
- `message` e obrigatoria.
- E-mails duplicados em `recipients` sao ignorados na contagem final.

Body:

```json
{
  "from_email": "organizador@example.com",
  "recipients": [
    "convidado1@example.com",
    "convidado2@example.com"
  ],
  "subject": "Lembrete: confirme sua presenca no evento",
  "message": "Oi! Passando para lembrar voce de confirmar presenca."
}
```

Resposta `202`:

```json
{
  "data": {
    "campaign_id": "uuid",
    "event_id": "uuid",
    "queued": 2,
    "status": "queued"
  }
}
```

Erro de validacao `422`:

```json
{
  "error": "validation failed",
  "data": [
    {
      "field": "recipients",
      "message": "recipients must contain at least one email"
    }
  ]
}
```

Possiveis erros:

- `400`: corpo JSON invalido.
- `401`: token ausente ou invalido.
- `403`: evento pertence a outra organizacao.
- `404`: evento nao encontrado.
- `422`: campos invalidos.
- `502`: Brevo respondeu com erro.
- `503`: Brevo indisponivel ou credenciais ausentes.

## Convidados

### `GET /guests?event_id={id}`

Lista convidados de um evento.

Autenticacao: obrigatoria.

Query params:

- `event_id`: ID do evento

Resposta `200`:

```json
{
  "data": [
    {
      "id": "uuid",
      "event_id": "uuid",
      "name": "Maria Silva",
      "email": "maria@example.com",
      "phone": "+5511999999999",
      "status": "invited",
      "created_at": "2026-06-15T12:00:00Z",
      "updated_at": "2026-06-15T12:00:00Z"
    }
  ]
}
```

### `POST /guests`

Cria convidado para um evento do tenant autenticado.

Autenticacao: obrigatoria.

Body:

```json
{
  "event_id": "uuid",
  "name": "Maria Silva",
  "email": "maria@example.com",
  "phone": "+5511999999999"
}
```

Resposta `201`: objeto `Convidado` dentro de `data`.

## Confirmacao de Presenca

### `POST /rsvp`

Registra a confirmacao de presenca de um convidado.

Autenticacao: publica.

Body:

```json
{
  "guest_id": "uuid",
  "event_id": "uuid",
  "status": "confirmed"
}
```

Status aceitos pela documentacao:

- `confirmed`
- `declined`
- `pending`

Resposta `200`:

```json
{
  "data": {
    "id": "uuid",
    "guest_id": "uuid",
    "event_id": "uuid",
    "status": "confirmed",
    "created_at": "2026-06-15T12:00:00Z",
    "updated_at": "2026-06-15T12:00:00Z"
  }
}
```

## Orcamento

Itens de orcamento pertencem a um evento e so podem ser acessados por usuarios do tenant dono do evento.

### `GET /events/{id}/budget`

Lista itens de orcamento do evento.

Autenticacao: obrigatoria.

Resposta `200`:

```json
{
  "data": [
    {
      "id": "uuid",
      "event_id": "uuid",
      "description": "Buffet",
      "category": "Alimentacao",
      "amount": 2500,
      "paid": false,
      "created_at": "2026-06-15T12:00:00Z",
      "updated_at": "2026-06-15T12:00:00Z"
    }
  ]
}
```

### `POST /events/{id}/budget`

Cria item de orcamento para o evento.

Autenticacao: obrigatoria.

Body:

```json
{
  "description": "Buffet",
  "category": "Alimentacao",
  "amount": 2500,
  "paid": false
}
```

Resposta `201`: item de orcamento dentro de `data`.

### `PUT /budget/{id}`

Atualiza item de orcamento.

Autenticacao: obrigatoria.

Body:

```json
{
  "description": "Buffet completo",
  "category": "Alimentacao",
  "amount": 3000,
  "paid": true
}
```

Resposta `200`: item atualizado dentro de `data`.

### `DELETE /budget/{id}`

Remove item de orcamento.

Autenticacao: obrigatoria.

Resposta:

```http
204 No Content
```

## Presentes

Presentes pertencem a um evento e so podem ser acessados por usuarios do tenant dono do evento.

### `GET /events/{id}/gifts`

Lista presentes do evento.

Autenticacao: obrigatoria.

Resposta `200`:

```json
{
  "data": [
    {
      "id": "uuid",
      "event_id": "uuid",
      "name": "Jogo de panelas",
      "description": "Presente sugerido para a casa nova.",
      "price": 399.9,
      "url": "https://loja.example.com/produto",
      "reserved": false,
      "reserved_by": "",
      "created_at": "2026-06-15T12:00:00Z",
      "updated_at": "2026-06-15T12:00:00Z"
    }
  ]
}
```

### `POST /events/{id}/gifts`

Cria presente para o evento.

Autenticacao: obrigatoria.

Body:

```json
{
  "name": "Jogo de panelas",
  "description": "Presente sugerido para a casa nova.",
  "price": 399.9,
  "url": "https://loja.example.com/produto",
  "reserved": false,
  "reserved_by": ""
}
```

Resposta `201`: presente dentro de `data`.

### `PUT /gifts/{id}`

Atualiza presente.

Autenticacao: obrigatoria.

Body:

```json
{
  "name": "Jogo de panelas premium",
  "description": "Conjunto completo.",
  "price": 499.9,
  "url": "https://loja.example.com/produto-premium",
  "reserved": true,
  "reserved_by": "Maria Silva"
}
```

Resposta `200`: presente atualizado dentro de `data`.

### `DELETE /gifts/{id}`

Remove presente.

Autenticacao: obrigatoria.

Resposta:

```http
204 No Content
```

## Dashboard

### `GET /dashboard`

Retorna metricas gerais do tenant autenticado.

Autenticacao: obrigatoria.

Resposta `200`:

```json
{
  "data": {
    "total_events": 10,
    "total_guests": 200,
    "total_confirmed": 120
  }
}
```

## Analytics

### `GET /analytics/events/{eventID}`

Retorna resumo de confirmacoes de presenca de um evento.

Autenticacao: obrigatoria.

Resposta `200`:

```json
{
  "data": {
    "event_id": "uuid",
    "guest_count": 200,
    "confirmed_count": 120,
    "declined_count": 20,
    "pending_count": 60
  }
}
```

## CORS

O CORS e configurado por `CORS_ALLOWED_ORIGINS`.

Exemplo:

```env
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

Se nao for configurado, o fallback atual e `*`.

## Resumo de Rotas

| Metodo | Rota | Auth | Descricao |
| --- | --- | --- | --- |
| GET | `/` | Nao | Redireciona para Swagger |
| GET | `/health` | Nao | Verifica saude da API |
| GET | `/swagger` | Nao | Swagger UI |
| GET | `/swagger/doc.json` | Nao | OpenAPI em JSON |
| GET | `/swagger/doc.yaml` | Nao | OpenAPI em YAML |
| POST | `/auth/register` | Nao | Cadastra usuario |
| POST | `/auth/login` | Nao | Autentica usuario |
| GET | `/auth/me` | Sim | Retorna usuario atual |
| GET | `/events` | Sim | Lista eventos |
| POST | `/events` | Sim | Cria evento |
| GET | `/events/{id}` | Sim | Busca evento |
| PUT | `/events/{id}` | Sim | Atualiza evento |
| DELETE | `/events/{id}` | Sim | Remove evento |
| POST | `/events/{id}/reminders` | Sim | Enfileira lembretes de RSVP |
| GET | `/guests?event_id={id}` | Sim | Lista convidados |
| POST | `/guests` | Sim | Cria convidado |
| POST | `/rsvp` | Nao | Envia confirmacao |
| GET | `/events/{id}/budget` | Sim | Lista orcamento |
| POST | `/events/{id}/budget` | Sim | Cria item de orcamento |
| PUT | `/budget/{id}` | Sim | Atualiza item de orcamento |
| DELETE | `/budget/{id}` | Sim | Remove item de orcamento |
| GET | `/events/{id}/gifts` | Sim | Lista presentes |
| POST | `/events/{id}/gifts` | Sim | Cria presente |
| PUT | `/gifts/{id}` | Sim | Atualiza presente |
| DELETE | `/gifts/{id}` | Sim | Remove presente |
| GET | `/dashboard` | Sim | Visao geral |
| GET | `/analytics/events/{eventID}` | Sim | Analise do evento |

## Observacoes Para Frontend

- Use `Authorization: Bearer <access_token>` em todas as rotas protegidas.
- Use `Content-Type: application/json` em `POST` e `PUT`.
- Ao receber `401`, redirecione o usuario para login ou renove a sessao.
- Ao receber `404` em recurso protegido, trate como inexistente para o usuario atual.
- `DELETE` com sucesso retorna `204` e nao deve tentar ler JSON no corpo.
- Orcamento e presentes agora sao persistidos na API, nao precisam mais ficar apenas no navegador.
- Lembretes de RSVP agora podem ser enviados pelo backend em `POST /events/{id}/reminders`; no app Laravel, liberar o proxy `events/*/reminders`.
