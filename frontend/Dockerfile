FROM node:24-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
# The existing UI dependencies require npm's legacy peer resolution.
RUN npm ci --legacy-peer-deps
COPY . .
ARG VITE_API_URL=http://localhost:8096
ENV VITE_API_URL=${VITE_API_URL}
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
# BrowserRouter routes must survive a page refresh.
RUN printf 'server {\n  listen 80;\n  root /usr/share/nginx/html;\n  location / { try_files $uri $uri/ /index.html; }\n}\n' > /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
