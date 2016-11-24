# Toggl

```go
import (
    "fmt"
    "github.com/en30/toggl"
)

func main() {
    interval = 30
    dashboardId = 1
    token = "..."
    onStart := func(a *toggl.Activity) {
        fmt.Println(a)
    }
    onStop := func(a *toggl.Activity) {
        fmt.Println(a)
    }
    onError := func(e error) {
        fmt.Println(e)
    }
    toggle.NewHook(interval, dashboardId, token, onStart, onStop, onError)
    select{}
}
```

Dirty polling solution till [implementation of web hooks](https://github.com/toggl/toggl_api_docs/issues/64)
