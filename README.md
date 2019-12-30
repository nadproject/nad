# NAD

NAD is a simple notebook for developers. It is forked from [Dnote](https://github.com/dnote/dnote)

## Features

- [ ] CLI
- [ ] Web application with full text search and WYSIWYG editor
- [ ] Optional encryption on CLI using PGP
- [ ] Note title, tag

## Example

CLI

```
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

