FROM node:20-alpine AS builder

WORKDIR /app

ARG VITE_POLL_INTERVAL_MS=500
ENV VITE_POLL_INTERVAL_MS=$VITE_POLL_INTERVAL_MS

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

RUN npm run build

FROM nginx:1.27-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
