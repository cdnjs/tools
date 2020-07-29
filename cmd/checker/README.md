# Checker

Tools for our CI.
Pass `-no-path-validation` to allow all package file paths to be accepted. Otherwise, the path will be validated against a regex.

## `lint`

Checks that a package is correctly configured based on its JSON.

## `show-files`

Output how many package files match and whether they will be ignored for a number of latest npm/git versions.
