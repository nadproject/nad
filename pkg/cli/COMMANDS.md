# Commands

- [add](#nad-add)
- [view](#nad-view)
- [edit](#nad-edit)
- [remove](#nad-remove)
- [find](#nad-find)
- [sync](#nad-sync)
- [login](#nad-login)
- [logout](#nad-logout)

## nad add

_alias: a, n, new_

Add a new note to a book.

```bash
# Launch a text editor to add a new note to the specified book.
nad add linux

# Write a new note with a content to the specified book.
nad add linux -c "find - recursively walk the directory"
```

## nad view

_alias: v_

- List books or notes.
- View a note detail.

```bash
# List all books.
nad view

# List all notes in a book.
nad view golang

# See details of a note
nad view 12
```

## nad edit

_alias: e_

Edit a note or a book.

```bash
# Launch a text editor to edit a note with the given id.
nad edit 12

# Edit a note with the given id in the specified book with a content.
nad edit 12 -c "New Content"

# Launch a text editor to edit a book name.
nad edit js

# Edit a book name by using a flag.
nad edit js -n "javascript"
```

## nad remove

_alias: rm, d_

Remove either a note or a book.

```bash
# Remove a note with an id.
nad remove 1

# Remove a book with the `book name`.
nad remove js
```

## nad find

_alias: f_

Find notes by keywords.

```bash
# find notes by a keyword
nad find rpoplpush

# find notes by multiple keywords
nad find "building a heap"

# find notes within a book
nad find "merge sort" -b algorithm
```

## nad sync

_NAD Pro only_

_alias: s_

Sync notes with NAD server. All your data is encrypted before being sent to the server.

## nad login

_NAD Pro only_

Start a login prompt.

## nad logout

_NAD Pro only_

Log out of NAD.
