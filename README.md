# Gopthermart

Индивидуальный дипломный проект курса «Go-разработчик»

## DB

Для развертывания PostgreSQL используется:
```bash
docker run -d \
  --name local-postgres \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  -e PGDATA=/var/lib/postgresql/data/pgdata \
  -v ./.db:/var/lib/postgresql/data \
  postgres:15-alpine
```
