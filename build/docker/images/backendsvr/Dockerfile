FROM alpine:3.10

# author
MAINTAINER admin

#更新Alpine的软件源为国内站点
RUN echo "https://mirror.tuna.tsinghua.edu.cn/alpine/v3.10/main/" > /etc/apk/repositories

RUN apk update \
        && apk upgrade \
        && apk add --no-cache bash \
        bash-doc \
        bash-completion \
		ca-certificates \
        && rm -rf /var/cache/apk/* \
        && /bin/bash

#EXPOSE 9030
#EXPOSE 19030

#COPY App /data/App/
WORKDIR /data/App/cmd/backendsvr
#RUN chmod +x /data/App/cmd/backendsvr

ENTRYPOINT ["./backendsvr", "&"]
#ENTRYPOINT ["/data/App/docker-entrypoint.sh"]