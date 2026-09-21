FROM postgres:16-bookworm@sha256:efedf3595f1d6f415c08568ba171029bf54052e754cc9f030e3f2412b21f3d67 AS postgis
RUN apt-get update && apt-get install -y --no-install-recommends \
    postgresql-16-postgis-3=3.6.4+dfsg-2.pgdg12+1 \
    postgresql-16-postgis-3-scripts=3.6.4+dfsg-2.pgdg12+1 \
    && rm -rf /var/lib/apt/lists/*
FROM postgis AS complete
RUN apt-get update && apt-get install -y --no-install-recommends \
    postgresql-16-h3=4.2.3-4.pgdg12+1 \
    && rm -rf /var/lib/apt/lists/*
