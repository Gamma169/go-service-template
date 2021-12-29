FROM golang:1.17-alpine as builder

MAINTAINER rienzi-gokea

STOPSIGNAL SIGTERM

USER root

ARG app_name=foobar
ARG service_root=/opt/service
ARG app_root=${service_root}/${app_name}

RUN if [ ! -d $app_root}/src ]; then mkdir -p $app_root; fi

WORKDIR $app_root
RUN if [ ! -d src ]; then mkdir src ; fi

COPY ./scripts ./scripts
COPY ./src ./src
COPY ./migrations ./migrations

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

RUN ./scripts/get_deps


RUN ./scripts/build ${app_name}


FROM golang:1.17-alpine as runner

ARG app_name
ENV app_name ${app_name} 
ARG service_root
ARG app_root

WORKDIR $app_root

COPY --from=builder ${app_root}/bin/${app_name} ${app_root}/bin/${app_name}

CMD ["./bin/${app_name}"]
