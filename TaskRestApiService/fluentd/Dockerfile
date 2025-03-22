FROM fluent/fluentd:v1.16-1

USER root

# Устанавливаем пакеты, нужные для сборки плагинов
RUN apk add --no-cache --update --virtual .build-deps \
    build-base \
    ruby-dev \
    libffi-dev \
    && gem install fluent-plugin-kafka fluent-plugin-elasticsearch --no-document \
    && apk del .build-deps  # Удаляем временные пакеты после установки

USER fluent
