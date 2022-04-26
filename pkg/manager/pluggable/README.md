# Bhojpur ISO - Pluggable Framework

The `pluggable` is a light Bus-event driven plugin library. It implements the event/sub pattern
to extend your project with external binary plugins that can be written in any language.

```golang
import "github.com/bhojpur/iso/pkg/manager/pluggable"

func main() {

    var myEv pluggableEventType = "something.to.hook.on"
    temp := "/usr/custom/bin"

    m = pluggable.NewManager(
        []pluggable.EventType{
            myEv,
        },
    )
        
    // Load plugins
    m.Autoload("test", temp) // Scan for binary plugins with the "test" prefix. E.g. 'test-foo'
    m.Plugin = append(m.Plugin, pluggable.Plugin{ Name: "foo" , Executable: "path" }) // manually add a Plugin
    m.Load("my-binary", "my-binary-2"...) // Load individually, scanning $PATH

    // Register to events and initialize the manager
    m.Register()

    // Optionally process plugin results response
    // The plugins has to return as output a json in stdout in the format { 'state': "somestate", data: "some data", error: "some error" }
    // e.g. with jq:  
    // jq --arg key0   'state' \
    // --arg value0 '' \
    // --arg key1   'data' \
    // --arg value1 "" \
    // --arg key2   'error' \
    // --arg value2 '' \
    // '. | .[$key0]=$value0 | .[$key1]=$value1 | .[$key2]=$value2' \
    // <<<'{}'
    m.Response(myEv, func(p *pluggable.Plugin, r *pluggable.EventResponse) { ... }) 

    // Emit events, they are encoded and passed as JSON payloads to the plugins.
    // In our case, test-foo will receive the map as JSON
    m.Publish(myEv,  map[string]string{"foo": "bar"})
}
```

## Plugin processed data

The interface passed to `Publish` gets marshalled in JSON in a event struct of the following form:

```go
type Event struct {
	Name EventType `json:"name"`
	Data string    `json:"data"`
	File string    `json:"file"`
}
```

An example bash plugin could be, for example:

```bash
#!/bin/bash

event="$1"
payload="$2"
if [ "$event" == "something.to.hook.on" ]; then
  custom_data=$(echo "$payload" | jq -r .data | jq -r .foo )
  ...
fi
```

which could be called by

```golang
m.Publish(myEv,  map[string]string{"foo": "bar"})
```

NOTE: when the payload exceeds the `threshold size` the payload published with `Publish` is written
into a temporary file and the file location is sent to the plugin with the Event `file` field, so
for example, a plugin should expect data in a file if the publisher expects to send big chunk of
data:

```bash
#!/bin/bash
data_file="$(echo $2 | jq -r .file)"
if [ -n "${data_file}" ]; then
    payload="$(cat $data_file)"

...
fi
```

## Writing Plugin

It is present a `FactoryPlugin`, which allows to create plugins, consider:

```golang
import "github.com/bhojpur/iso/pkg/manager/pluggable"

func main() {
    var myEv pluggableEventType = "event"

    factory := pluggable.NewPluginFactory()
    factory.Add(myEv, func(e *Event) EventResponse { return EventResponse{ ... })

    factory.Run(os.Args[1], os.Args[2], os.Stdout)
}
```
