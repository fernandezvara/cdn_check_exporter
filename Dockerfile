ARG ARCH

FROM multiarch/alpine:${ARCH}-latest-stable

ARG USER=exporter
ENV HOME /home/${USER}

RUN apk add --no-cache bash
RUN adduser -D -s /bin/bash ${USER}

USER ${USER}
WORKDIR ${HOME}

# binary
COPY cdn_check_exporter /usr/local/bin

# cdn_check_exporter variables
ENV DEBUG "0"
ENV DOMAIN ""
ENV EVERY "10"
ENV IPPORT "0.0.0.0:9000"
ENV PATH "/metrics"
ENV RESOURCE ""

# bash prompt
ENV PS1 "$(whoami):$(pwd)> "

EXPOSE 9000

ENTRYPOINT ["/usr/local/bin/cdn_check_exporter"]
CMD [ "-h" ]