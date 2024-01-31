# Headless WordPress with Go front-end

**This project was a huge mistake.** I built this as a relatively new Go programmer, thinking I could build a blog faster _and_ get more experience with Go if I built a Go front-end for headless WordPress via the [WordPress REST API](https://developer.wordpress.org/rest-api/). Instead, I ran into every sharp edge of WordPress. I completed the project for the sake of finishing it, but I certainly wouldn't recommend using this.

## Local setup

Before setting up locally, ensure that [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/install/) are installed on the development machine, then clone the project to a directory of your choice.

After cloning the project, copy the `.env.example` file to `.env` by running the following from the main project directory:

```sh
cp .env.example .env
```

**IMPORTANT: After copying the file, get salts from https://roots.io/salts.html and copy them into `.env`** (yes: this is jank and insecure, just like everything in WordPress)**.**

**If you're running Linux, update `USER_UID` and `USER_GID` to your `.env` file to negate Linux permission issues.** The appropriate values for these can be obtained by running `id -u` and `id -g` respectively.

Next, run the following commands from the main project directory to build the docker images:

```sh
docker compose build
```

After that, install the WordPress dependencies.

```sh
docker compose run --rm wordpress composer install
```

Run the following [WP-CLI](https://wp-cli.org/) commands to create the WordPress site, replacing the placeholder values with the desired username, password and email.

```sh
docker compose run --rm wordpress wp core install \
    --url="http://localhost" \
    --title="Blog" \
    --admin_user="<your_username>" \
    --admin_password="<your_password>" \
    --admin_email="<your_email>"
```

Override WordPress's default settings.

```sh
docker compose run --rm wordpress wp theme activate headless-wp-site
docker compose run --rm wordpress wp rewrite structure '/posts/%postname%/' \
    --tag-base='tags' \
    --hard
```

**Optional:** The Go web application uses the NPM package `prettier` and `prettier-plugin-go-template` to format the Go template files. If you have NPM installed on the host machine, run the following command from the project's main directory:

```sh
npm install --prefix webapp
```

To start the app, run the following command:

```sh
docker compose up
```

You can then visit the site a http://localhost. You can visit the WordPress admin at http://localhost/wp/wp-admin/

## Building production(ish) images

The Dockerfiles used in this project are multi-stage. For the `development` state, it only builds the images to the point where the files can be mounted from the host to the container, but it does not copy the source files into the container.

You can build production(ish) containers using the following command:

```sh
docker compose --file=docker-compose.prod.yml build
```

**The `docker-compose.prod.yml` file is not _completely_ suitable for building production containers: the `image` property must be set for each service in the Compose file in order to push the built images to a container registry.**

To run the production(ish) containers locally, run the following command:

```sh
docker compose --file=docker-compose.prod.yml up
```

## How Go is set up

The Go app uses [Gorilla Mux](https://github.com/gorilla/mux) for routing, and it embeds the CSS + template files into the compiled binary. The production stage of the app's Dockerfile puts the binary in a [scratch](https://hub.docker.com/_/scratch) image running as a non-root user to achieve the smallest container size possible.

## How Caddy is set up

Caddy acts a reverse proxy sitting in-front of both Go and WordPress. It proxies the `/app/*`, `/wp/*` and `/wp-json/*` paths to WordPress, and all other paths to the Go front-end. It redirects `www` paths to their non-`www` counterpart. When the `SITE_HOSTNAME` environment variable is set to a real domain (not `localhost`), Caddy will automatically provision TLS certificates.

## How WordPress is set up

WordPress has a famously bad architecture, and as such it is challenging to containerize. While Docker Hub provides an [official WordPress image](https://hub.docker.com/_/wordpress), its file structure made it challenging to run with [WP-CLI](https://wp-cli.org/) bundled in the same image.

I opted to suffer through creating my own WordPress image. I opted to use an Apache-based PHP image despite FPM variants being faster: I'm not experience setting up Apache images, and using WordPress with a webserver other than Apache is truly asking for a bad time. I used the WordPress docs [Server Environment page](https://make.wordpress.org/hosting/handbook/server-environment/) to determine what PHP packages needed to be present on the system, and painstakingly installed the packages' system level dependencies via Googling PHP compile errors. After that, copied the WP-CLI and Composer executables from their respective images. If I were containerize WordPress again in the future, I might opt for a [ServerSideUP PHP Image](https://serversideup.net/open-source/docker-php/) instead (thanks for the tip, [Tony](https://twitter.com/tonysmdev/status/1744003306576306208)).

[Roots Bedrock](https://roots.io/bedrock/) was used to install WordPress. Bedrock provides a more modern WordPress experience, with WP core backed by Composer and plugins backed by Composer + [WordPress Packagist](https://wpackagist.org/). When `WP_ENV=production`, users can't install new plugins. This combination of features makes WordPress more secure and easier to manage at the expense of having a Bedrock-incompatable theme/plugin break WordPress.

For new endpoints or modifications to WordPress, I opted to put this functionality in the theme's `functions.php` file. Storing this behavior in the theme is considered an anti-pattern in the WordPress community because functionality and presentation should be separate concerns. However, these concerns _are_ separate because this is a headless WordPress app, and the WordPress theme handles all of the functionality.

## Potential improvements

I realized about half way into this project's development cycle that this amalgamation was far too ridiculous to ever run in production. As a result, there are many features and improvements that I didn't pursue. They are listed here for posterity.

**Graceful template errors in Go.** By default, errors in Go templates crash the template mid-render. These templates can [be written to a buffer](https://medium.com/@leeprovoost/dealing-with-go-template-errors-at-runtime-1b429e8b854a) then flushed to the response only if they are successful, with a fallback error page if they are not.

**Implement WordPress pagination.** The web is fast enough to render a few hundred blog posts and tags as static HTML without problems, but a reasonably feature-complete blog should be able to include pagination.

**Implement WordPress query params.** WordPress accepts several [URL query parameters](https://codex.wordpress.org/WordPress_Query_Vars) to set [the WP_Query](https://developer.wordpress.org/reference/classes/wp_query/) and filter the results to what the user wants to see. These were not implemented.

**Add tests for WordPress and Go.** Surprisingly, WordPress _can_ actually scaffold [theme tests](https://developer.wordpress.org/cli/commands/scaffold/theme-tests/) and [plugin tests](https://developer.wordpress.org/cli/commands/scaffold/plugin-tests/) using the [WP-CLI](https://wp-cli.org/) (more info in this [Smashing Magazine tutorial](https://www.smashingmagazine.com/2017/12/automated-testing-wordpress-plugins-phpunit/)).

**Removing the `/wp/` prefix from `/wp/wp-admin/` that is added by Bedrock.** I tried this a few times a few different ways, but it seems to cause problems with login post request trying to redirect, which is not supported. A Roots community member seemed to [have some luck in this post](https://discourse.roots.io/t/recommended-subdomain-multisite-nginx-vhost-configuration-with-new-web-layout/1429/12?u=etc), but I couldn't get it to work. Changes to the admin url structure will likely require edits to `WP_SITEURL` in `.env`.

**WordPress caching.** WordPress is slow. The [W3 Total Cache](https://wordpress.org/plugins/w3-total-cache/) is webhost agnostic and seemingly recommended by WordPress, but I'm not sure if it would actually help with REST responses. The [WP REST Cache](https://wordpress.org/plugins/wp-rest-cache/) plugin purportedly helps with this, at least according to this [blog post](https://medium.com/@lodewijkm/our-headless-wordpress-journey-part-i-speeding-up-the-rest-api-aef76a898418).

**Automate WordPress setup with the `entrypoint.sh` script.** I had originally intended to automate much of the WordPress setup in the entrypoint script, but I gave up when I realized how painful and ridiculous this project was.
