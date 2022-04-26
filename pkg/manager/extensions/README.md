# Bhojpur ISO - Extension Framework

It creates `Git`-like extensions for your Bhojpur ISO projects.

## simple Usage

```golang
import "github.com/bhojpur/iso/pkg/manager/extensions"

// Detect my-awesome-cli-foo, my-awesome-cli-bar in $PATH and extensiopath1 (relative to the bin)
// it also accepts abspath
exts := extensions.Discover("my-awesome-cli", "extensiopath1", "extensiopath2")

fmt.Println("Detected extensions:", exts)

for _, ex := range exts {
  name := ex.Short()
  cobraCmd := ex.CobraCommand()
}
```
