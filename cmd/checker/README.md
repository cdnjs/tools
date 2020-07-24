# Checker

Tools for our CI.

## `lint`

Checks that a package is correctly configured based on its JSON.

## `show-files`

Output how many package files match and whether they will be ignored for a number of latest npm/git versions.

## `meta`

Outputs the distribution of packages that contain a particular JSON property (or sub-property).

```
make checker && ./bin/checker meta author email
```

This will find the the distribution of packages that contain both properties:

```
{"author": {"email" : <>}}
```

ones that contain no `email`:

```
{"author": <>}
```

and ones that do not have `author` at all:

```
{}
```

## `meta-list`

Lists all the unique JSON keys for this particular JSON object across all packages.

```
make checker && ./bin/checker meta-list author
```

This will find the summary of keys found in the `author` JSON object.
