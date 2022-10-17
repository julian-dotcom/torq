# Error handling in Go

Example of how to handle errors

```go
import (
	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
)

func main() {
	err := doSomeCalculations()
	if err != nil {
		// This is the top level, log the problem. Can include words such as "error, problem with x, failed, etc"
		// Use zerolog and not fmt or log packages
    log.Error().Err(err).Msg("Problem doing some calculations")
		return
	}
}

func doSomeCalculations() error {
	result, err := divide(4, 0)
	if err != nil {

		// if you want to detect and respond to specific error conditions use errors.Is or errors.As
		// errors.Is checks the VALUE of the error matches, errors.As checks that the TYPE of the error is the same
		// In this example, the error we want to check for is a generic error value so we can't use errors.As and have to use errors.Is
    // If you want calling code to be able to easily respond to specific errors it would be better to create a dedicated error type and use errors.As
    var divideByZero = errors.New("Divide by zero")
		if errors.Is(err, divideByZero) {
			log.Debug().Msg("We have had another one of those pesky divide by zero errors")
		}
		// Bubble up error to caller. No need to log, the caller will either log OR bubble it up further with it's own Wrap
		// No need for words like "error, problem with x. Just give context to what was just attempted"
		return errors.Wrap(err, "Running division")
	}
	log.Debug().Msgf("Result was: %v", result)
	return nil
}

func divide(first int, second int) (int, error) {
	if second == 0 {
		// If we don't have an err to Wrap such as in this example create a new error describing what the error is and send it to caller.
		// No need to log as we are sending error back
		// When returning error, set the return value to it's zero value which for int is zero and string would be "" etc etc
		return 0, errors.New("Divide by zero")
	}
  // Explicitly return nil when there is no error
	return first / second, nil
}
```

The output of the above program:

```json
{"level":"debug","time":"2022-10-17T17:13:17+01:00","message":"We have had another one of those pesky divide by zero errors"}
{"level":"error","error":"Running division: Divide by zero","time":"2022-10-17T17:13:17+01:00","message":"Problem doing some calculations"}
```

Using the above lines, we can determine not only what went wrong but also the path the execution took through the code.

For further reading, look at Cockroach DB’s guide on error handling: [https://wiki.crdb.io/wiki/spaces/CRDB/pages/1676640931/Error+handling+basics](https://wiki.crdb.io/wiki/spaces/CRDB/pages/1676640931/Error+handling+basics)
