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

#EXPOSE 9033
#EXPOSE 19033

#COPY App /data/App/

WORKDIR /data/App/cmd/apisvr
#RUN chmod +x /data/App/cmd/apisvr

ENTRYPOINT ["./apisvr", "&"]
#ENTRYPOINT ["/data/App/docker-entrypoint.sh"]