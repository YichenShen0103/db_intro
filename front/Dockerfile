FROM node:18-alpine AS builder

RUN npm config set registry https://registry.npmmirror.com/

WORKDIR /app

COPY package.json ./
RUN npm install

COPY . .
RUN npm run build

# Stage 2: Serve with Nginx
FROM nginx:stable-alpine

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Timezone
ENV TZ=Asia/Shanghai
