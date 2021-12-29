ARG app_name=foobar

FROM golang:1.17-alpine3.15 as builder

MAINTAINER rienzi-gokea

STOPSIGNAL SIGTERM

USER root

ARG app_name

COPY ./scripts ./scripts
COPY ./src ./src
COPY ./migrations ./migrations

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

RUN ./scripts/get_deps


RUN ./scripts/build ${app_name}


FROM alpine:3.15 as runner

ARG app_name
ENV app_name ${app_name} 

# Note we are following the unix file-system standard by placing our third-party app in the /opt folder
COPY --from=builder /go/bin/${app_name} /opt/${app_name}

CMD /opt/${app_name}
