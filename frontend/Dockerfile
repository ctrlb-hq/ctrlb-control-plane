FROM node:24-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
# The existing UI dependencies require npm's legacy peer resolution.
RUN npm ci --legacy-peer-deps
COPY . .
ARG VITE_BASE_PATH=/control-plane/
ARG VITE_API_BASE_URL=
ARG VITE_API_URL=
ENV VITE_BASE_PATH=${VITE_BASE_PATH}
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_API_URL=${VITE_API_URL}
RUN npm run build

FROM nginx:alpine
ARG VITE_BASE_PATH=/control-plane/
COPY --from=build /app/dist /usr/share/nginx/html${VITE_BASE_PATH}
# BrowserRouter routes must survive a page refresh.
RUN ctrlb_base_path="${VITE_BASE_PATH%/}/"; \
    if [ "${ctrlb_base_path}" = "/" ]; then \
      printf 'server {\n  listen 80;\n  root /usr/share/nginx/html;\n  location / { try_files $uri $uri/ /index.html; }\n}\n' > /etc/nginx/conf.d/default.conf; \
    else \
      printf 'server {\n  listen 80;\n  root /usr/share/nginx/html;\n  location = / { return 302 %s; }\n  location / { try_files $uri $uri/ %sindex.html; }\n}\n' "${ctrlb_base_path}" "${ctrlb_base_path}" > /etc/nginx/conf.d/default.conf; \
    fi
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
