FROM wordpress:cli AS cli

FROM wordpress AS base

# Copy the WordPress CLI into the image.
COPY --from=cli /usr/local/bin/wp /usr/local/bin/wp

# Remove the line that starts the server so that we can run commands in our own
# entrypoint file.
RUN sed 's/exec "$@"//g' /usr/local/bin/docker-entrypoint.sh > /usr/local/bin/docker-entrypoint.sh

# Copy our entrypoint into the container.
COPY wp-content/entrypoint.sh /usr/local/bin/entrypoint-stage2.sh
RUN chmod +x /usr/local/bin/entrypoint-stage2.sh

# Command from previous layers must be repeated because of surprising Docker
# implementation details.
# https://github.com/docker-library/wordpress/issues/194#issuecomment-477813257
# https://docs.docker.com/engine/reference/builder/#understand-how-cmd-and-entrypoint-interact
CMD ["apache2-foreground"]
ENTRYPOINT [ "entrypoint-stage2.sh" ]

RUN cp /usr/src/wordpress/wp-config-docker.php /usr/src/wordpress/wp-config.php

FROM base AS development

ARG USER_UID
ARG USER_GID

RUN if getent passwd $USER_UID; then userdel "$(getent passwd $USER_UID | cut -d: -f1)"; fi
RUN if getent group $USER_GID; then groupdel "$(getent group $USER_GID | cut -d: -f1)"; fi

RUN groupadd --gid $USER_GID app
RUN useradd --uid $USER_UID --gid $USER_GID --create-home app

USER $USER_UID
