
FROM    postgres:9.4.8

WORKDIR /init
COPY    load.sh /init/
COPY    export.sql /init/
RUN     ./load.sh
