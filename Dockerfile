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

# It seems I have to copy these
# TODO: Would be good if I didn't have to copy them
# And would be good if I could configure it in the CMD line as well
# https://github.com/Gamma169/go-service-template/issues/1

ARG app_name=foobar
ARG service_root=/opt/service
ARG app_root=${service_root}/${app_name}
WORKDIR $app_root


COPY --from=builder ${app_root}/bin/${app_name} ${app_root}/bin/${app_name}

CMD ["./bin/foobar"]
