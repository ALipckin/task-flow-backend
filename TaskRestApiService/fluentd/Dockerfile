FROM fluent/fluentd:v1.17-1

USER root

RUN apk add --no-cache --update --virtual .build-deps \
    build-base \
    ruby-dev \
    libffi-dev \
    && gem install fluent-plugin-elasticsearch --no-document \
    && apk del .build-deps

USER fluent