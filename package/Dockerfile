FROM ubuntu:18.04
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -f /bin/sh && ln -s /bin/bash /bin/sh
COPY bin/rdns-server /usr/bin/
CMD ["rdns-server"]
