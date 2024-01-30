#! /usr/bin/env bash
set -e

wait-for-it ${DB_HOST:-database}:3306 -- echo "connected to database"

if ! wp core is-installed > /dev/null; then
    echo "WordPress is not yet installed..."

    # TODO: Maybe setup WordPress site if it does not exist. Example:

    # wp core install \
    #     --url=$WORDPRESS_URL \
    #     --title=WordPress \
    #     --admin_user=$WORDPRESS_USERNAME \
    #     --admin_password=$WORDPRESS_PASSWORD \
    #     --admin_email=$WORDPRESS_EMAIL \
    #     --skip-email
fi

exec "$@"
