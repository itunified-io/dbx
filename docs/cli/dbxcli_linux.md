## dbxcli linux

Oracle Linux host management over SSH

### Synopsis

Oracle Linux system management operations executed over SSH.
Covers package management, kernel parameters, storage/LVM, network, and security.

Requires a target with SSH endpoint configured (oracle_host entity_type).

### Options

```
  -h, --help            help for linux
      --target string   target name (from ~/.dbx/targets/)
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
```

### SEE ALSO

* [dbxcli](dbxcli.md)	 - dbx — multi-database management platform
* [dbxcli linux kernel](dbxcli_linux_kernel.md)	 - Kernel parameter management
* [dbxcli linux network](dbxcli_linux_network.md)	 - Network diagnostics
* [dbxcli linux package](dbxcli_linux_package.md)	 - RPM/DNF package management
* [dbxcli linux security](dbxcli_linux_security.md)	 - Security status checks
* [dbxcli linux storage](dbxcli_linux_storage.md)	 - Storage and LVM management

