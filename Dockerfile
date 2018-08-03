# FROM golang:1.10 as builder
# WORKDIR /go/src/github.com/devclub-iitd/DeployBot/
# RUN go get -d -v golang.org/x/net/html  
# COPY app.go .
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM docker:18.06
LABEL maintainer="ozym4nd145@outlook.com"
RUN sed -i -e 's/v[[:digit:]]\.[[:digit:]]/edge/g' /etc/apk/repositories
RUN apk update && apk add --no-cache \
        ca-certificates\
        git \
        #git-secret\
        openssh-client\
        gnupg\
        bash\
        wget\
        perl
RUN wget https://github.com/docker/machine/releases/download/v0.14.0/docker-machine-`uname -s`-`uname -m` -O /usr/local/bin/docker-machine && \
        chmod +x /usr/local/bin/docker-machine && \
        wget https://github.com/docker/compose/releases/download/1.22.0/run.sh -O /usr/local/bin/docker-compose && \
        chmod +x /usr/local/bin/docker-compose

VOLUME ["/root/.ssh","/config"]

WORKDIR /usr/local/bin/
# COPY --from=builder /go/src/github.com/devclub-iitd/DeployBot/app .
COPY ./scripts/* /usr/local/bin/

EXPOSE 7777/tcp

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]  


