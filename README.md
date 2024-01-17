# 2024 Blog Site (WIP)

This repository is a work-in-progress. Its intention is to use features from Docker and the WordPress CLI to paper over missing features like a plugin manifest (example: `package.json` or `requirements.txt`), and predefined pages. Ideally, another developer should be able to clone this repo and have all the required plugins and static pages for this site. This is most important for a headless WordPress installation where another application may depend on certain pages already existing.

This was originally built on top of the [official WordPress image](https://github.com/docker-library/wordpress), but is now built on the PHP image because of limitations within the php image itself.

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
