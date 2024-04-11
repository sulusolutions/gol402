FROM lightninglabs/lnd:v0.17.3-beta as lnd

# Start a new, final image to reduce size.
FROM alpine as final

# Copy the binaries and entrypoint from the builder image.
COPY --from=lnd /bin/lncli /bin/
COPY --from=lnd /bin/lnd /bin/
COPY scripts ./scripts

ENV PATH="/scripts:${PATH}"

# Add bash.
RUN apk add --no-cache \
    bash \
    ca-certificates \
    bitcoin-cli \
    jq

ENTRYPOINT ["./scripts/init-script.sh"]
