FROM caddy:2.7.6-alpine AS base

FROM base AS development

CMD ["caddy" "run" "--watch" "--config" "/etc/caddy/Caddyfile" "--adapter" "caddyfile"]

FROM base AS production

COPY "caddy/Caddyfile" "/etc/caddy/Caddyfile"

CMD ["caddy" "run" "--config" "/etc/caddy/Caddyfile" "--adapter" "caddyfile"]
