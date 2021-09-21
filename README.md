## Logic
The app does following:
- Makes POST request to hardcoded URL and gets io.Reader of response body
- Reads response data to buffer
- Checks if response data is valid json; if not - returns error (exit code is 1 in case of any error)
- Performs simple string substitution against hardcoded template
- Format JSON since example output is multi-line formatted
- Outputs the result to STDOUT

## Benefits
- All stages are linked by interfaces - configurable, extendable, replaceable. 
It's possible to replace string substitution logic with real JSON parsing (if we really need to. For current task JSON actions will only slow us down.).
- 100% code coverage (`app/app_test.go` does all the testing stuff)
- Benchmark test `BenchmarkApp_Run` with real HTTP server to see the performance changes. 
Benchmark could be executed with memory/CPU profile from Goland IDE.

Current benchmark results:
```shell
goos: darwin
goarch: amd64
pkg: github.com/DaniilGo/challenge/internal/app
cpu: Intel(R) Core(TM) i9-9880H CPU @ 2.30GHz
BenchmarkApp_Run
BenchmarkApp_Run-16    	   16852	     69870 ns/op
```
