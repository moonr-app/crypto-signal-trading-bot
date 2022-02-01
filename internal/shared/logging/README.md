# logging

Tiny [zap](https://github.com/uber-go/zap) wrapper to log like a champ.

## how to use

```go
package main

import (
	"context"

	"go.uber.org/zap"
	
	"github.com/moonr-app/crypto-signal-trading-bot/internal/shared/logging"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Print something
	logging.Info(ctx, "hello", zap.String("myKey", "myValue"))
	
	// Create your logger
	noopLogger := zap.NewNop()
	ctx    = logging.With(context.Background(), noopLogger)
	
	// Customise your logger
	ctx = logging.WithOptions(ctx, zap.Fields(zap.String("coolKey", "coolValue")))
}
```
