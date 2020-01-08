# NAD

NAD is a simple notebook for developers. It is forked from [Dnote](https://github.com/dnote/dnote)

## Features

- [ ] CLI
- [ ] Web application with full text search and WYSIWYG editor
- [ ] Optional encryption on CLI using PGP
- [ ] Note title, tag

## Design goal

The goal is to build and maintain a stable software that favors function over form.

- Simple command line interface with only the essential commands
- Minimalist, traditional server-rendered web interface with no front-end JavaScript framework.

## Example

CLI

```sh
# launches $EDITOR which saves a note on quit
nad add --book=php

# syncs with the server
nad sync
```

Note title support will be done through masthead using yaml or toml:

```
+++
title = "my title"
tag = ["myTag1"]
+++

my note content
```

Encryption will be done by specifying a path to key. The key path can be provided in the config file `~/.nad/config` or as a flag.

```sh
# encrypts the note with id 123
nad encrypt 123 --key=/my/key/path

# decrypts the note with id 123
nad decrypt 123 --key=/my/key/path
```

The benefit is that the user enjoys total privacy for some notes that they want to keep private. The server has zero knowledge of the encrypted data.
