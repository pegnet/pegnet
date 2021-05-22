
"use strict";

// gulp and plugins
var
  gulp = require('gulp'),
  plumber = require('gulp-plumber'),
  sass = require('gulp-sass'),
  autoprefix = require('gulp-autoprefixer'),
  sourcemaps = require('gulp-sourcemaps'),
  concat = require('gulp-concat'),
  log = require('fancy-log');


// CONFIG options
// These can be changed in .gulp-config.js
var 
  CONFIG = require('./gulp-config.js'),
  OPTIONS = CONFIG.OPTIONS,
  PATHS = CONFIG.PATHS,
  STYLES_MAIN_FILE = CONFIG.MAIN_FILES.STYLES;


// Compile Sass, autoprefix CSS3,
// and save to target CSS directory
function styles() {
  return gulp.src(PATHS.STYLES + STYLES_MAIN_FILE)
    .pipe(plumber())
    //.pipe(sourcemaps.init())
    .pipe(sass({
      outputStyle: 'compressed',
      includePaths: PATHS.STYLES_EXTRAS
    }).on('error', function(e) { log.error('SASS error: ' + e.message) }))
    .pipe(autoprefix({cascade: false }))
    //.pipe(sourcemaps.write('.', {
    //  addComment: true,
    //  debug: true
    //}))
    .pipe(gulp.dest(PATHS.PUBLIC_STYLES))
    .on('end', function() { log('the files were compiled and minified') });
}

function scripts() {
  return gulp.src([
      './node_modules/jquery/dist/jquery.min.js',
      './node_modules/bootstrap/dist/js/bootstrap.min.js',
      './node_modules/datatables/media/js/jquery.dataTables.min.js',
      './node_modules/moment/min/moment.min.js',
      PATHS.SCRIPTS + '/*.js'
    ])
    .pipe(sourcemaps.init())
    .pipe(concat('app.js'))
    .pipe(sourcemaps.write('.'))
    .pipe(gulp.dest(PATHS.PUBLIC_SCRIPTS))
    .on('end', function() { log('the files were compiled and minified') });
}

function fonts() {
  return gulp.src([
      './node_modules/@fortawesome/fontawesome-free/webfonts/*',
    ])
    .pipe(gulp.dest(PATHS.PUBLIC_FONTS))
    .on('end', function() { log('the files were copied and updated') });
}

// Keep an eye on changes...
function watch() {
  gulp.watch(['**/*.+(scss|sass)'], {cwd: PATHS.STYLES}, styles);
  gulp.watch(['*.js'], {cwd: PATHS.SCRIPTS}, scripts);
}

var build = gulp.series(styles, gulp.parallel(scripts, fonts, watch));

exports.styles = styles;
exports.scripts = scripts;
exports.fonts = fonts;
exports.watch = watch;
exports.build = build;

/*
 * Define default task that can be called by just running `gulp` from cli
 */
exports.default = build;
