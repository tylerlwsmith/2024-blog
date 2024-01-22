<?php
/**
 * headless-wp-site functions and definitions
 *
 * @link https://developer.wordpress.org/themes/basics/theme-functions/
 *
 * @package headless-wp-site
 */

if ( ! defined( '_S_VERSION' ) ) {
	// Replace the version number of the theme on each release.
	define( '_S_VERSION', '1.0.0' );
}

/**
 * Sets up theme defaults and registers support for various WordPress features.
 *
 * Note that this function is hooked into the after_setup_theme hook, which
 * runs before the init hook. The init hook is too late for some features, such
 * as indicating support for post thumbnails.
 */
function headless_wp_site_setup() {
	/*
		* Make theme available for translation.
		* Translations can be filed in the /languages/ directory.
		* If you're building a theme based on headless-wp-site, use a find and replace
		* to change 'headless-wp-site' to the name of your theme in all the template files.
		*/
	load_theme_textdomain( 'headless-wp-site', get_template_directory() . '/languages' );

	/*
		* Enable support for Post Thumbnails on posts and pages.
		*
		* @link https://developer.wordpress.org/themes/functionality/featured-images-post-thumbnails/
		*/
	add_theme_support( 'post-thumbnails' );

	// This theme uses wp_nav_menu() in one location.
	register_nav_menus(
		array(
			'menu-1' => esc_html__( 'Primary', 'headless-wp-site' ),
		)
	);

	/*
		* Switch default core markup for search form, comment form, and comments
		* to output valid HTML5.
		*/
	add_theme_support(
		'html5',
		array(
			'gallery',
			'caption',
			'style',
			'script',
		)
	);
}
add_action( 'after_setup_theme', 'headless_wp_site_setup' );

/**
 * Custom template tags for this theme.
 */
require get_template_directory() . '/inc/template-tags.php';

// This endpoint is a potential security vulnerability so it should be
// disabled by our webserver to outside traffic.
add_action( 'rest_api_init', function () {
	// For reasons I don't understand, this function cannot be inlined below
	// and still work. It must be called in this scope, and then passed into
	// the callback below. My guess is that the wp_create_nonce() function
	// was not designed to be nested and the callback expects a named function.
	$nonce = wp_create_nonce( 'wp_rest' );

	register_rest_route( 'nonce/v1', 'nonce', [
		'methods' => 'GET',
		'callback' => function () use ( $nonce ) {
			return [
				'nonce' => $nonce,
			];
		},
	] );
} );