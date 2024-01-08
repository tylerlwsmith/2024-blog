# Using a php:fpm image connected to a dedicated Apache container would be more
# high- performance and scalable. The php:apache image was used for simplicity.
FROM php:8.2-apache AS base

# WORKDIR is already set by the base image, but repeat here as documentation.
WORKDIR /var/www/html

# Bedrock's webroot is in the web/ directory, so Apache's configuration files 
# must be updated. Reference:
# https://stackoverflow.com/a/51457728/7759523
RUN sed -ri -e 's!/var/www/html!/var/www/html/web!g' /etc/apache2/sites-available/*.conf
RUN sed -ri -e 's!/var/www/html!/var/www/html/web!g' /etc/apache2/apache2.conf /etc/apache2/conf-available/*.conf

# Install the _recommended_ WordPress system dependencies. Reference:
# https://make.wordpress.org/hosting/handbook/server-environment/#system-packages
RUN apt-get update && apt-get install -y \
    curl \
    ghostscript \
    imagemagick \
    openssl

# Install the WordPress extensions that must be compiled in the Docker image,
# along with their dependencies. Extensions that are pre-compiled inside this 
# Docker image have been omitted from the installation list below. 

# Required and recommended extensions:
# https://make.wordpress.org/hosting/handbook/server-environment/#php-extensions

# Extensions pre-compiled into the Docker image:
# https://github.com/docker-library/php/issues/1049#issuecomment-673593583
RUN apt-get update && apt-get install -y \
    libicu-dev \
    libonig-dev \
    libzip-dev \
    && docker-php-ext-install \
    mysqli \
    exif \
    intl \
    opcache \
    zip

# The ImageMagick extension is installed through PECL instead of compiled
# into PHP itself, so it must be installed separately.
# https://discuss.circleci.com/t/how-to-install-php-imagick-php-extension/19051/7
RUN apt-get update && apt-get install -y \
    libmagickwand-dev --no-install-recommends \
    && pecl install imagick \
    && docker-php-ext-enable imagick

# Configure OPcache. I'm not sure that OPcache does anything without this.
# COPY misc/opcache.ini "$PHP_INI_DIR/conf.d/opcache.ini"

# Copy the WordPress CLI into the image. This could alternatively be installed
# as a Composer dependency, but we'll install it system-wide.
COPY --from=wordpress:cli /usr/local/bin/wp /usr/local/bin/wp

# Install `less` as WP CLI dependency. It's likely present on the base image,
# but it's also included here in case we use a base image that doesn't have it.
RUN apt-get update && apt-get install -y less

# Get the Composer package manager since it is not installed by default on PHP
# Docker images. Composer is required to version control the active plugins.
COPY --from=composer:2.6 /usr/bin/composer /usr/bin/composer

# MySQL isn't immediately available once its container starts, which can cause
# WP CLI failures when running ad hoc commands. The wait-for-it package allows
# us to wait to ensure the database is reachable before connecting to it.
RUN apt-get update && apt-get install -y wait-for-it

# Copy and set a custom entrypoint.
COPY wordpress/entrypoint.sh /var/www/html/entrypoint.sh
ENTRYPOINT [ "/var/www/html/entrypoint.sh" ]

# Setting ENTRYPOINT will reset CMD to an empty value if CMD is defined in
# the base image, so we must reset CMD to the base image value.
CMD ["apache2-foreground"]
