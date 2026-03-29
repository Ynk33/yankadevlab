# --- Build stage ---
FROM node:22-alpine AS build

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

ARG VITE_AUTH_API_URL
ENV VITE_AUTH_API_URL=$VITE_AUTH_API_URL

COPY . .
RUN npm run build

# --- Serve stage ---
FROM nginx:alpine

# SPA fallback: serve index.html for all routes react-router handles
RUN printf 'server {\n\
    listen 80;\n\
    root /usr/share/nginx/html;\n\
    index index.html;\n\
    location / {\n\
        try_files $uri $uri/ /index.html;\n\
    }\n\
}\n' > /etc/nginx/conf.d/default.conf

COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
