# FROM golang:1.10 as builder
# WORKDIR /go/src/github.com/devclub-iitd/DeployBot/
# RUN go get -d -v golang.org/x/net/html  
# COPY app.go .
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM docker:18.06
LABEL maintainer="ozym4nd145@outlook.com"
RUN wget https://github.com/docker/machine/releases/download/v0.14.0/docker-machine-`uname -s`-`uname -m` -O /usr/local/bin/docker-machine && \
        chmod +x /usr/local/bin/docker-machine
#RUN wget https://github.com/docker/machine/releases/download/v0.14.0/docker-machine-`uname -s`-`uname -m` -O /usr/local/bin/docker-machine && \
        #chmod +x /usr/local/bin/docker-machine && \
        #wget https://github.com/docker/compose/releases/download/1.22.0/run.sh -O /usr/local/bin/docker-compose && \
        #chmod +x /usr/local/bin/docker-compose
RUN apk upgrade --update-cache --available
RUN apk add --no-cache \
        ca-certificates\
        git \
        openssh-client\
        gnupg\
        bash\
        wget\
        perl\
        make\
        gawk

RUN git clone https://github.com/sobolevn/git-secret.git git-secret && cd git-secret && make build && PREFIX="/usr/local" make install

VOLUME ["/config","/keys"]

WORKDIR /usr/local/bin/
# COPY --from=builder /go/src/github.com/devclub-iitd/DeployBot/app .
COPY ./scripts/* /usr/local/bin/

EXPOSE 7777/tcp

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]  
