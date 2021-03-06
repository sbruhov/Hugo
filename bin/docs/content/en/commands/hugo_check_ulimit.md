---
title: "hugo check ulimit"
slug: hugo_check_ulimit
url: /commands/hugo_check_ulimit/
---
## hugo check ulimit

Check system ulimit settings

### Synopsis

Hugo will inspect the current ulimit settings on the system.
This is primarily to ensure that Hugo can watch enough files on some OSs

```
hugo check ulimit [flags]
```

### Options

```
  -h, --help   help for ulimit
```

### Options inherited from parent commands

```
      --config string              config file (default is path/config.yaml|json|toml)
      --configDir string           config dir (default "config")
      --debug                      debug output
  -e, --environment string         build environment
      --ignoreVendorPaths string   ignores any _vendor for module paths matching the given Glob pattern
      --log                        enable Logging
      --logFile string             log File path (if set, logging enabled automatically)
      --quiet                      build in quiet mode
  -s, --source string              filesystem path to read files relative from
      --themesDir string           filesystem path to themes directory
  -v, --verbose                    verbose output
      --verboseLog                 verbose logging
```

### SEE ALSO

* [hugo check](/commands/hugo_check/)	 - Contains some verification checks

