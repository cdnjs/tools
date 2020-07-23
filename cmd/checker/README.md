# Checker

Tools for our CI.

## `lint`

Checks that a package is correctly configured based on its JSON.

## `show-files`

Output how many package files match and whether they will be ignored for a number of latest npm/git versions.

## `print-meta`

Outputs the distribution of packages that contain a particular JSON property (or sub-property).

For example:

```make checker && ./bin/checker print-meta author email```

This will find the the distribution of packages that contain `{"author": {"email" : <>}}`, as
well as note the ones that contain no `email` (`{"author": <>}`) and ones that do not have `author` whatsoever (`{}`).
