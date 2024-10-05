# Playground

To build the playground for web, run the following command inside `docs` directory:

```
GOOS=js GOARCH=wasm go build -o playground.wasm
```

And optionally (if not present yet) copy `wasm_exec.js` to `docs` directory:

```
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./
```

### VSCode syscall/js error

VSCode might give a compiler error in the line `import "syscall/js"` of the file `playground.go`:

> could not import syscall/js (no required module provides package "syscall/js") compiler(BrokenImport)
>
> error while importing syscall/js: build constraints exclude all Go files in /usr/local/Cellar/go/1.19/libexec/src/syscall/jscompiler

I don't know how to solve it, but can be ignored as the Go compiler compiles successfully.
