FROM php:8.3-cli-alpine

RUN apk add --no-cache \
    curl \
    curl-dev \
    git \
    libxml2-dev \
    oniguruma-dev \
    libzip-dev \
    sqlite-dev \
    && docker-php-ext-install -j"$(nproc)" \
        bcmath \
        curl \
        dom \
        mbstring \
        pcntl \
        pdo \
        pdo_sqlite \
        soap \
        xml \
        zip

COPY --from=composer:2 /usr/bin/composer /usr/bin/composer

WORKDIR /var/www/html

COPY . .

RUN composer install --no-interaction --no-progress --prefer-dist --optimize-autoloader --no-dev \
    && cp .env.example .env \
    && php artisan key:generate --force

EXPOSE 8000

CMD ["php", "artisan", "serve", "--host=0.0.0.0", "--port=8000"]
