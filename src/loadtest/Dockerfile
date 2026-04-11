FROM mostafamoradian/xk6-kafka:1.0.0

USER root

RUN mkdir -p /scripts && \
    addgroup -S k6group 2>/dev/null || true && \
    adduser  -S k6 -G k6group 2>/dev/null || true && \
    chown -R k6:k6group /scripts

COPY --chown=k6:k6group loadtest.js /scripts/loadtest.js

USER k6
WORKDIR /scripts

ENTRYPOINT ["k6"]

CMD ["run", "/scripts/loadtest.js"]