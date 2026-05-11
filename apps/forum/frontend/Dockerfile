FROM node:22-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci --ignore-scripts
COPY . .
ARG VITE_API_URL
ARG VITE_ZITADEL_URL
ARG VITE_ZITADEL_CLIENT_ID
ARG VITE_IDP_HINT_TELEGRAM
ARG VITE_IDP_HINT_GITHUB
ARG VITE_IDP_HINT_GITLAB
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 3000
