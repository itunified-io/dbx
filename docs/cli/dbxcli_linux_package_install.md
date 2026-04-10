## dbxcli linux package install

Install a package via DNF (confirm-gated)

### Synopsis

Install a package using dnf. Requires confirmation before execution.

```
dbxcli linux package install [flags]
```

### Examples

```
  dbxcli linux package install name=oracle-database-preinstall-19c --target db01
```

### Options

```
  -h, --help   help for install
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli linux package](dbxcli_linux_package.md)	 - RPM/DNF package management

