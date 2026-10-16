FROM node:24.6-bookworm AS build

# Build from the repo root (docker build -f caregiver/Dockerfile .).
# The app sits in /app/caregiver so its "../shared" resolves to /app/shared.
WORKDIR /app/caregiver
ARG VITE_API_URL
ENV VITE_API_URL=$VITE_API_URL
ARG VITE_GOOGLE_CLIENT_ID
ENV VITE_GOOGLE_CLIENT_ID=$VITE_GOOGLE_CLIENT_ID
COPY caregiver/package*.json ./
RUN npm install
COPY shared/ /app/shared/
COPY caregiver/ ./
RUN npm run build

FROM nginx:stable
COPY caregiver/nginx.conf /etc/nginx/conf.d/default.conf

COPY --from=build /app/caregiver/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]