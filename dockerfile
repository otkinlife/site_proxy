# 使用官方的 Go 镜像作为构建阶段
FROM golang AS builder

# 设置工作目录
WORKDIR /build

# 将当前项目的所有文件复制到工作目录
COPY . .

# 下载依赖
RUN go mod download

# 编译 Go 代码
RUN go build -o proxy-server .

# 使用一个更小的基础镜像来运行应用程序
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 将编译后的二进制文件复制到新镜像
COPY --from=builder /build/proxy-server .

# 将配置文件和 HTML 模板复制到新镜像
COPY --from=builder /build/config.json .
COPY --from=builder /build/template.html .

# 暴露应用程序的端口
EXPOSE 8080

# 运行二进制文件
CMD ["./proxy-server"]
