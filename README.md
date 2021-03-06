<h1 align="center"> Genji.js </h1>
<p align="center">
  <a href="https://genji.dev">
    <img alt="Genji" title="Genji" src="https://raw.githubusercontent.com/genjidb/docs/master/assets/icons/logo.svg?sanitize=true" width="100">
  </a>
</p>

<p align="center">
  Document-oriented, embedded, SQL database
</p>

Experimental wrapper around the [Genji](https://github.com/genjidb/genji) database.

## Getting started

Install Genji

```bash
yarn add @genjidb/genji
```

Copy the wasm file from `node_modules` into your public directory

```bash
cp node_modules/@genjidb/genji/genji.wasm public/
```

Or if you are using Webpack, add this to your config, after installing the `copy-webpack-plugin` loader :

```bash
yarn add --dev copy-webpack-plugin
```

```javascript
const CopyWebpackPlugin = require('copy-webpack-plugin');

module.exports = {
    ...
    plugins: [
        new CopyWebpackPlugin({
            patterns: [
                { from: 'node_modules/@genjidb/genji/dist/genji.wasm' }
            ]
        })
    ]
}
```

## Usage

```javascript
import { initDatabase } from '@genjidb/genji';

async function run() {
  const genji = await initDatabase();
  const db = await genji.Database();
  await db.exec('CREATE TABLE foo');
  await db.exec('INSERT INTO foo (a) VALUES (1), (2), (3)');

  db.query('SELECT * FROM foo').forEach(v => console.log(v));
}

run();
```

## Build from source

Requires [Go](https://golang.org/dl/) >= 1.16 and [Node](https://nodejs.org/en/download/) >= 10

```bash
yarn install
yarn build
```

## Running tests

```bash
yarn test
```
