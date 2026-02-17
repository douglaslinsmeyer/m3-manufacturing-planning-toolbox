# Base stage - common dependencies
FROM node:18-alpine AS base

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install ALL dependencies
RUN npm ci

# Copy source code
COPY . .

# Development stage - Vite dev server with HMR
FROM base AS development

EXPOSE 5173

ENV NODE_ENV=development

# Run Vite dev server
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

# Build stage - create production bundle
FROM base AS builder

# Build production bundle
RUN npm run build

# Production stage - nginx serving static files
FROM nginx:alpine AS production

# Copy built assets from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
