// OPTIONS AND SETUP INFO
var
  PROJECT = './',
  OPTIONS = {

  },
  PATHS = {
    SCRIPTS: './build/scripts/',
    STYLES: './build/sass/',
    STYLES_EXTRAS: [
      './node_modules/font-awesome/scss',
    ],
    PUBLIC_STYLES: './static/styles/',
    PUBLIC_SCRIPTS: './static/scripts/',
    PUBLIC_FONTS: './static/fonts/'
  },
  MAIN_FILES = {
    STYLES: 'styles.scss', // the name for the main styles file
  };

module.exports = {
  OPTIONS: OPTIONS,
  PATHS: PATHS,
  MAIN_FILES: MAIN_FILES
};