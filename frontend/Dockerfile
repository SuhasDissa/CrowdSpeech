# ── Build stage ───────────────────────────────────────────────────────────────
FROM docker.io/library/node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci --prefer-offline

COPY . .
RUN npm run build

# ── Runtime stage: serve with nginx ───────────────────────────────────────────
FROM docker.io/library/nginx:1.25-alpine

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Custom minimal nginx config for SPA
COPY --from=builder /app/dist /usr/share/nginx/html

RUN cat > /etc/nginx/conf.d/default.conf << 'EOF'
server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    # SPA fallback
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot|webp|webmanifest)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/javascript;
}
EOF

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
