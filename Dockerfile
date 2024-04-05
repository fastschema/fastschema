FROM alpine:3 AS fastschemadownloader

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG VERSION

ENV BUILDX_ARCH="${TARGETOS:-linux}_${TARGETARCH:-amd64}${TARGETVARIANT}"

RUN wget https://github.com/fastschema/fastschema/releases/download/v${VERSION}/fastschema_${VERSION}_${BUILDX_ARCH}.zip \
    && unzip fastschema_${VERSION}_${BUILDX_ARCH}.zip \
    && chmod +x /fastschema

FROM alpine:3

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

EXPOSE 8000

WORKDIR /fastschema

COPY --from=fastschemadownloader /fastschema /fastschema/fastschema

CMD ["/fastschema/fastschema", "start"]
