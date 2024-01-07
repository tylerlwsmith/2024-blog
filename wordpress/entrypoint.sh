#! /usr/bin/env bash
set -e

# TODO: Replace with wait script.
wait-for-it ${DB_HOST:-database}:3306 -- echo "connected to database"

# if ! wp core is-installed  > /dev/null; then
#     # WP is not installed. Let's try installing it.
#     wp core install \
#         --url=$WORDPRESS_URL \
#         --title=WordPress \
#         --admin_user=$WORDPRESS_USERNAME \
#         --admin_password=$WORDPRESS_PASSWORD \
#         --admin_email=$WORDPRESS_EMAIL \
#         --skip-email
# fi

# wp post create --post_type=post --post_title="$(date)" --post_status="publish"
# wp theme list

exec "$@"
