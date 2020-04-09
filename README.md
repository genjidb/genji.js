# Genji.js

Experimental wrapper around the [Genji](https://github.com/asdine/genji) database.

It's functional but, currently, the compiled WebAssembly file is too big (~3mb) to be objectively usable in production.

The code is compatible with [TinyGo](https://github.com/tinygo-org/tinygo) and produces a 400kb wasm file but there are too many bugs in the v0.12.0, I will wait for the next version to give a try.

## Build from source

```bash
yarn install
yarn build
```

## Running tests

```bash
yarn test
```
