# Invitely API

Documentacao completa:

- `docs/API_COMPLETA.md`

The API exposes OpenAPI documentation through the running Gin server.

## Swagger

- UI: `GET /swagger`
- JSON: `GET /swagger/doc.json`
- YAML: `GET /swagger/doc.yaml`

The YAML document is also kept in this repository:

- `docs/swagger.yaml`

## Auth

Protected routes use a bearer token:

```http
Authorization: Bearer <access_token>
```

## Rotas principais

- `GET /events`
- `POST /events`
- `GET /events/{id}`
- `PUT /events/{id}`
- `DELETE /events/{id}`
- `GET /events/{id}/budget`
- `POST /events/{id}/budget`
- `PUT /budget/{id}`
- `DELETE /budget/{id}`
- `GET /events/{id}/gifts`
- `POST /events/{id}/gifts`
- `PUT /gifts/{id}`
- `DELETE /gifts/{id}`
