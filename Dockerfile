# Runtime-only image — build artifacts must exist locally before running compose:
#   cd web && npm run build          → produces web/dist/
#   cd server && go build -o main ./cmd/start_server  → produces server/main
#
# Then: docker compose -f docker-compose.prod.yml up --build

FROM nginx:stable-alpine

# Supervisor to manage nginx + backend processes
RUN apk --no-cache add supervisor ca-certificates tzdata

# Frontend static files (pre-built)
COPY web/dist /usr/share/nginx/html

# Backend binary (pre-built)
COPY server/main /app/main
RUN chmod +x /app/main

# Nginx config
COPY web/nginx.conf /etc/nginx/conf.d/default.conf

# Supervisor config
COPY supervisord.conf /etc/supervisord.conf

EXPOSE 80

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]
