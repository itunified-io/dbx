## dbxcli license activate

Activate a license file

### Synopsis

Activate a license JWT file. The file is validated, copied to ~/.dbx/license.jwt, and EE features are unlocked.

```
dbxcli license activate [flags]
```

### Examples

```
  dbxcli license activate --file /path/to/license.jwt
```

### Options

```
      --file string   path to license JWT file
  -h, --help          help for activate
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
```

### SEE ALSO

* [dbxcli license](dbxcli_license.md)	 - Manage dbx license

