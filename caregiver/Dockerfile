FROM node:24.6-bookworm AS build

WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM ubuntu:24.04
RUN apt-get update && apt-get install -y nginx && rm -rf /var/lib/apt/lists/*

COPY --from=build /app/dist /var/www/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
