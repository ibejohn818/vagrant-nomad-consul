FROM alpine:3.20

RUN apk add --update --no-cache \
  postgresql postgresql-timescaledb \
  python3 py3-pip py3-psycopg py3-psutil

RUN pip3 install \
  --no-cache-dir --break-system-packages \
  patroni[consul]

RUN mkdir /run/postgresql && \
  chown postgres:postgres /run/postgresql

WORKDIR /var/lib/postgresql/

USER postgres

RUN echo -e 'somepassword\n' > /var/lib/postgresql/pg.pass \
    && /usr/libexec/postgresql/initdb -U postgres --pwfile=/var/lib/postgresql/pg.pass data 
    # && rm -f /var/lib/postgresql/pg.pass

COPY data/ data/

# USER root
