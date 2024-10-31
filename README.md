# CountNSort

Small cli utility to pipe into to sort based on selection frequency.
Usefull for tools like fzf, rofi, dmenu.

We save how often specific entries/lines have been choosen and use this information to sort the input.
After a new entry has been choosen, we update the underlying database (for now only json, later sqlite).

The idea is this:

```bash
#!/bin/bash
CHOSE_ELEMENT="$(cat <list-of-choices> | countnsort -name my-usecase-db | dmenu)"
countnsort -name my-usecase-db -inc "$CHOSE_ELEMENT"
# Do what you want with the chosen entry
...
```




## Usage


```
Usage of countnsort:
  -delimiter string
        Delimiter to split the line into parts, requires -f.
  -field int
        Position of the ID part of the line, requires -d. (default -1)
  -help
        Print this message.
  -inc string
        Increment the line with the specified identifier.
  -name string
        REQUIRED! Name for the database.
  -path
        Print path to database save location.
  -remove string
        Remove specified line from database.
```

