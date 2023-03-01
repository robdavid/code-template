# Generate code from Go template

A simple command line tool for generating code from Go template source files.

```text
Usage: code-template [options] files...
  -h, --help                  Display help
  -i, --include stringArray   Include specified files in each template execution
  -o, --output string         Send output to specified file, - for standard out
      --repeat string         Numeric range to iterator over in format var.name=n..m
      --set stringToString    Set a value to place within template values (default [])
```

