# Invitely API

API em Go para autenticação, eventos, convidados, RSVP, dashboard e analytics do InvitelyApp.

## Variáveis

Copie `.env.example` para `.env` e preencha:

```env
APP_ENV=development
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/invitely?sslmode=disable
SUPABASE_URL=https://seu-projeto.supabase.co
SUPABASE_ANON_KEY=sua_anon_key
SUPABASE_SERVICE_ROLE_KEY=sua_service_role_key
CORS_ALLOWED_ORIGINS=http://localhost:5173
```

Use `SUPABASE_SERVICE_ROLE_KEY` no backend para cadastrar usuários com e-mail já confirmado.

## Rodar local

```sh
docker compose up --build
```

Healthcheck:

```sh
curl http://localhost:8080/health
```

## Endpoints

- `POST /auth/register`
- `POST /auth/login`
- `GET /auth/me`
- `GET /events`
- `POST /events`
- `GET /events/{id}`
- `PUT /events/{id}`
- `DELETE /events/{id}`
- `POST /events/{id}/reminders`
- `GET /events/{id}/budget`
- `POST /events/{id}/budget`
- `PUT /budget/{id}`
- `DELETE /budget/{id}`
- `GET /events/{id}/gifts`
- `POST /events/{id}/gifts`
- `PUT /gifts/{id}`
- `DELETE /gifts/{id}`
- `GET /guests?event_id={id}`
- `POST /guests`
- `POST /rsvp`
- `GET /dashboard`
- `GET /analytics/events/{event_id}`

Rotas de eventos, convidados, dashboard e analytics exigem `Authorization: Bearer <access_token>`.
