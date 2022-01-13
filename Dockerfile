ARG app_name=foobar

FROM golang:1.17-alpine3.15 as builder

ARG app_name

COPY ./scripts ./scripts
COPY ./src ./src

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

RUN ./scripts/get_deps.sh

RUN ./scripts/build.sh ${app_name}



FROM alpine:3.15 as runner

ARG app_name
ENV app_name ${app_name} 

# Note we are following the unix file-system standard by placing our third-party app in the /opt folder
COPY ./migrations /opt/${app_name}/migrations
COPY --from=builder /go/bin/${app_name} /opt/${app_name}/${app_name}

# Need workdir because otherwise docker runs command in root, and app can't find migrations folder
WORKDIR /opt/${app_name}

# Despite CMD being in "shell form" and not "exec form," 
# it appears that the app still responds to a SIGTERM and runs with PID=1 contrary to documentation
CMD /opt/${app_name}/${app_name}
