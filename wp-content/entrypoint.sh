#! /usr/bin/env bash
set -e

/usr/local/bin/docker-entrypoint.sh

# TODO: Replace with wait script.
sleep 5

if ! wp core is-installed > /dev/null; then
    # WP is not installed. Let's try installing it.
    wp core install \
        --path=/usr/src/wordpress \
        --url=$WORDPRESS_URL \
        --title=WordPress \
        --admin_user=$WORDPRESS_USERNAME \
        --admin_password=$WORDPRESS_PASSWORD \
        --admin_email=$WORDPRESS_EMAIL \
        --skip-email
fi

exec "$@"
