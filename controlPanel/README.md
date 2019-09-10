# PegNet ControlPanel

## Making Changes

    Using `docker-compose up` from within controlPanel directory will generate the static files (scripts fonts and styles) using node.
    
    Gulp will keep watch for changes on the `build` directory

## Compiling the Static files into the binary.
    
We are using https://github.com/gobuffalo/packr. All compiled statics are in the gitignore. If the compiled statics are not present, it will use the local files on disk. To compile in the static files:

```bash
packr install
```

Or
```bash
# Generate static files
packr
# Build pegnet
go install
# Clean up the static files
packr clean
```