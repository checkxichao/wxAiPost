FROM golang:1.21.0-alpine3.18

# 设置代理
ENV GOPROXY="https://goproxy.cn"

# 抓取源代码
ADD . /workdir

# 创建工作目录运行编译
RUN cd /workdir/src \
    && go get \
    && go build -o fxmMain main.go \
    && chmod +x fxmMain

FROM alpine:3.18
WORKDIR /outdir
# 从第一个容器里面复制东西
RUN echo "https://mirrors.aliyun.com/alpine/v3.18/main/" > /etc/apk/repositories \
    && apk add -U tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY --from=0 /workdir/src/fxmMain /outdir/
COPY --from=0 /workdir/template.txt /outdir/template.txt
COPY --from=0 /workdir/富豪传奇.txt /outdir/富豪传奇.txt

CMD ["./fxmMain"]