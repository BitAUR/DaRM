# 使用 golang:1.22.0 作为构建镜像
FROM golang:1.22.0 AS builder

# 设置工作目录
WORKDIR /app

# 仅复制 go.mod 和 go.sum 以利用缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制应用源代码
COPY . .

# 编译应用程序
RUN go build -o darm .

# 使用 alpine 作为运行时镜像
FROM alpine:latest

# 安装必要的依赖库和根证书
RUN apk --no-cache add \
    ca-certificates \
    libc6-compat \
    git

# 设置工作目录
WORKDIR /app

# 将编译好的应用程序从构建镜像中复制到运行镜像
COPY --from=builder /app/darm /app/darm

# 创建数据目录
RUN mkdir -p /data

# 创建符号链接
RUN ln -s /data /app/data

# 暴露端口
EXPOSE 9740

# 运行应用程序
CMD ["/app/darm"]
