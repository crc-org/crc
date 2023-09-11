FROM docker.io/antora/antora
LABEL org.opencontainers.image.source="https://github.com/crc-org/crc"
RUN yarn global add --ignore-optional --silent \
    @antora/atlas-extension \
    @antora/cli \
    @antora/collector-extension \
    @antora/lunr-extension \
    @antora/pdf-extension \
    @antora/site-generator \
    asciidoctor-kroki
RUN antora --version
