# =========================================================
# 阶段 1：构建 React + TypeScript 前端静态文件
# =========================================================
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend

# 复制依赖清单并安装
COPY frontend/package*.json ./
RUN npm install

# 复制前端源码并执行编译 (生成 dist 文件夹)
COPY frontend/ ./
RUN npm run build

# =========================================================
# 阶段 2：编译 Go 后端可执行二进制文件
# =========================================================
FROM golang:1.25-alpine AS backend-builder
# 安装 CGO 编译 SQLite 所需的 GCC 编译器
RUN apk add --no-cache gcc musl-dev

WORKDIR /app/backend

# 复制依赖清单并下载
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制后端源码并编译
COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o main .

# =========================================================
# 阶段 3：最终轻量级运行镜像 (Alpine Linux)
# =========================================================
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata sqlite

# 设置时区为中国标准时间
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从前两个阶段复制编译产物
COPY --from=backend-builder /app/backend/main ./main
COPY --from=frontend-builder /app/frontend/dist ./dist

EXPOSE 8899

CMD ["./main"]