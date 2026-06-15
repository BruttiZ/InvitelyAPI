# Envio de Lembretes por E-mail

Este documento explica como usar a rota de lembretes da API Go do Invitely e como configurar o Brevo para envio real de e-mails.

## Visao Geral

A API possui uma rota autenticada para o organizador enviar lembretes de RSVP para convidados de um evento.

Rota:

```http
POST /events/{id}/reminders
```

Ela faz:

- valida o token Bearer;
- busca o evento pelo `{id}`;
- garante que o evento pertence a organizacao autenticada;
- valida remetente, destinatarios, assunto e mensagem;
- registra uma campanha na tabela `event_reminder_campaigns`;
- retorna `202 Accepted` rapidamente;
- envia o e-mail pelo Brevo em background;
- marca a campanha como `sent` quando o envio e aceito;
- marca a campanha como `failed` quando o Brevo retorna erro.

## Configuracao do Brevo

O projeto suporta dois modos:

- SMTP Brevo, usando chave `xsmtpsib-...`;
- API HTTP Brevo, usando chave `xkeysib-...`.

Como a chave atual comeca com `xsmtpsib-`, use o modo SMTP.

Variaveis no `.env`:

```env
BREVO_FROM_NAME=Invitely
BREVO_SMTP_HOST=smtp-relay.brevo.com
BREVO_SMTP_PORT=587
BREVO_SMTP_USERNAME=seu-smtp-login-do-brevo
BREVO_SMTP_KEY=sua-smtp-key-do-brevo
```

O `BREVO_SMTP_USERNAME` deve ser copiado no painel do Brevo, na area de SMTP. Ele nao e necessariamente o e-mail remetente verificado.

O remetente exibido aos destinatarios vem do payload da rota:

```json
{
  "from_email": "vbrutti02@gmail.com"
}
```

Esse `from_email` precisa estar verificado no Brevo como remetente.

## Autenticacao

A rota exige Bearer Token:

```http
Authorization: Bearer <access_token>
```

Sem token, a API retorna:

```http
401 Unauthorized
```

## Request

```http
POST /events/{id}/reminders
Content-Type: application/json
Authorization: Bearer <access_token>
```

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

Campos:

| Campo | Tipo | Obrigatorio | Descricao |
| --- | --- | --- | --- |
| `from_email` | string | Sim | E-mail remetente verificado no Brevo |
| `recipients` | array string | Sim | Lista de e-mails destinatarios |
| `subject` | string | Sim | Assunto do e-mail |
| `message` | string | Sim | Corpo da mensagem |

## Validacoes

- `{id}` precisa ser um evento existente.
- O evento precisa pertencer a organizacao autenticada.
- `from_email` precisa ser um e-mail valido.
- `recipients` precisa ter pelo menos 1 e-mail valido.
- `recipients` aceita no maximo 200 e-mails por requisicao.
- `subject` e obrigatorio.
- `message` e obrigatoria.
- E-mails duplicados em `recipients` sao removidos antes do envio.

## Response de Sucesso

Quando a campanha e aceita para processamento:

```http
202 Accepted
```

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

Observacao: o campo `queued` representa a quantidade de destinatarios aceitos para envio naquela requisicao. O envio real acontece em background e a campanha e atualizada para `sent` ou `failed`.

## Erros

### JSON invalido

```http
400 Bad Request
```

```json
{
  "error": "invalid request body"
}
```

### Sem token

```http
401 Unauthorized
```

```json
{
  "error": "missing bearer token"
}
```

### Evento de outra organizacao

```http
403 Forbidden
```

```json
{
  "error": "forbidden"
}
```

### Evento inexistente

```http
404 Not Found
```

```json
{
  "error": "event not found"
}
```

### Erro de validacao

```http
422 Unprocessable Entity
```

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

### Brevo respondeu com erro

```http
502 Bad Gateway
```

```json
{
  "error": "email provider failed"
}
```

### Brevo sem configuracao ou indisponivel

```http
503 Service Unavailable
```

```json
{
  "error": "email provider unavailable"
}
```

## Exemplo com cURL

```bash
curl -X POST "http://localhost:8080/events/{event_id}/reminders" \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "from_email": "vbrutti02@gmail.com",
    "recipients": ["convidado1@example.com", "convidado2@example.com"],
    "subject": "Lembrete: confirme sua presenca no evento",
    "message": "Oi! Passando para lembrar voce de confirmar presenca."
  }'
```

## Banco de Dados

A campanha e registrada na tabela:

```text
event_reminder_campaigns
```

Campos principais:

- `id`
- `event_id`
- `from_email`
- `recipients`
- `recipient_count`
- `subject`
- `message`
- `status`
- `provider_message_id`
- `error_message`
- `created_at`
- `updated_at`

Status possiveis:

- `queued`: campanha criada antes da tentativa de envio;
- `sent`: Brevo aceitou o envio;
- `failed`: Brevo ou configuracao falhou.

## Laravel Proxy

No outro projeto, liberar a rota no proxy:

```php
'POST' => [
    // ...
    'events/*/reminders',
],
```

Depois o frontend pode chamar:

```http
POST /api/v1/go/events/{id}/reminders
```

Payload igual ao da API Go.

## Deploy Render/Laravel

Se o frontend chamar:

```http
POST /api/v1/go/events/{id}/reminders
```

e receber `go_api_unavailable`, o erro provavelmente esta no proxy Laravel, antes mesmo do envio pelo Brevo. Confira:

- a API Go esta online no Render;
- a URL base da API Go no Laravel esta correta;
- a rota `events/*/reminders` esta liberada no proxy;
- o Laravel esta encaminhando o header `Authorization`;
- o servico Go no Render possui as variaveis Brevo.

Variaveis que precisam existir no servico Go do Render:

```env
BREVO_FROM_NAME=Invitely
BREVO_SMTP_HOST=smtp-relay.brevo.com
BREVO_SMTP_PORT=587
BREVO_SMTP_USERNAME=aed787001@smtp-brevo.com
BREVO_SMTP_KEY=sua-smtp-key-do-brevo
```

No arquivo `render.yaml`, as chaves sensiveis ficam como `sync: false`, entao precisam ser preenchidas manualmente no painel do Render.

Se a API Go responder erro do Brevo, ela retorna:

```json
{
  "error": "email provider failed",
  "data": {
    "provider_status": 0,
    "provider_error": "detalhe seguro do erro"
  }
}
```

`provider_status: 0` significa erro de conexao SMTP, autenticacao SMTP ou indisponibilidade antes de existir um status HTTP.

## Checklist Para Funcionar

- Remetente verificado no Brevo.
- `BREVO_SMTP_USERNAME` preenchido com o SMTP login do Brevo.
- `BREVO_SMTP_KEY` preenchido com uma SMTP key valida.
- Variaveis Brevo preenchidas no ambiente de deploy da API Go, nao apenas no `.env` local.
- Laravel apontando para a URL publica correta da API Go.
- API reiniciada ou watcher ativo.
- Token Bearer valido no request.
- Evento pertence ao usuario autenticado.
- Destinatarios validos.
