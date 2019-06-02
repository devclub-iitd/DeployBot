FROM golang:1.10 as builder
WORKDIR /go/src/github.com/devclub-iitd/DeployBot/
RUN go get -v github.com/sirupsen/logrus
COPY ./src/*.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o DeployBot -a -ldflags '-extldflags "-static"' .

FROM docker:18.06
LABEL maintainer="ozym4nd145@outlook.com"
RUN wget https://github.com/docker/machine/releases/download/v0.14.0/docker-machine-`uname -s`-`uname -m` -O /usr/local/bin/docker-machine && \
        chmod +x /usr/local/bin/docker-machine
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
        gawk\
        gettext\
        curl\
        coreutils
RUN git clone https://github.com/sobolevn/git-secret.git git-secret && cd git-secret && make build && PREFIX="/usr/local" make install

VOLUME ["/root/.docker","/keys"]

WORKDIR /usr/local/bin/
COPY --from=builder /go/src/github.com/devclub-iitd/DeployBot/DeployBot .
COPY ./scripts/* /usr/local/bin/

EXPOSE 7777/tcp

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
