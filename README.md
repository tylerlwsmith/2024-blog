# Headless WordPress with Go front-end (WIP)

**This project was a huge mistake.** I built this as a relatively new Go programmer, thinking I could build a blog faster _and_ get more experience with Go if I built a Go front-end for headless WordPress. Instead, I ran into every sharp edge of WordPress. I completed the project for the sake of finishing it, but I certainly wouldn't recommend using this.

## Local setup

Before setting up locally, ensure that [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/install/) are installed on the development machine, then clone the project to a directory of your choice.

After cloning the project, copy the `.env.example` file to `.env` by running the following from the main project directory:

```sh
cp .env.example .env
```

Once copied, navigate to the following URL to get salts for WordPress (this is jank and insecure, just like everything in WordPress) and copy them into the `.env` file:

https://roots.io/salts.html

Next, run the following commands from the main project directory to build the docker images:

```sh
docker compose build
```

After that, install the WordPress dependencies.

```sh
docker compose run --rm wordpress composer install
```

**Optional:** The Go web application uses the NPM package `prettier` and `prettier-plugin-go-template` to format the Go template files. If you have NPM installed on the host machine, run the following command from the project's main directory:

```sh
npm install --prefix webapp
```

To start the app, run the following command:

```sh
docker compose up
```

You can then visit the site a http://localhost.

## Building production(ish) images

The Dockerfiles used in this app are multi-stage. For the `development` state, it only builds the images to the point where the files can be mounted from the host to the container, but it does not copy the source files into the container.

You can build full production(ish) containers using the following command:

```sh
BUILD_TARGET=production docker compose build
```

**The `docker-compose.yml` file is not completely suitable for building or running production containers.**

## How WordPress is set up

WordPress is used

This repository is a work-in-progress. Its intention is to use features from Docker and the WordPress CLI to paper over missing features like a plugin manifest (example: `package.json` or `requirements.txt`), and predefined pages. Ideally, another developer should be able to clone this repo and have all the required plugins and static pages for this site. This is most important for a headless WordPress installation where another application may depend on certain pages already existing.

This was originally built on top of the [official WordPress image](https://github.com/docker-library/wordpress), but is now built on the PHP image because of limitations within the php image itself.

## Installing WordPress plugins with Composer

Install plugins:
https://wpackagist.org/

Remove capabilities from roles:
https://developer.wordpress.org/reference/classes/wp_role/remove_cap/

All capabilities:
https://wordpress.org/documentation/article/roles-and-capabilities/

## WordPress requirements

https://make.wordpress.org/hosting/handbook/server-environment/

## WP REST API Docs

https://developer.wordpress.org/rest-api/

## Caching?

Seems to be WP recommended and web host agnostic:
https://wordpress.org/plugins/w3-total-cache/

## Rewrite URLs to remove /wp/ from /wp/wp-admin/

https://discourse.roots.io/t/recommended-subdomain-multisite-nginx-vhost-configuration-with-new-web-layout/1429/12?u=etc

## Automated testing in WordPress:

https://www.smashingmagazine.com/2017/12/automated-testing-wordpress-plugins-phpunit/

## Things this app doesn't handle:

I realized about half way into this project's development cycle that I would never actually run this project in production, so there are a number of features that I didn't attempt to build

- Graceful template errors: it'll try to render until it explodes.
- Pagination
- Query params
- Tests
