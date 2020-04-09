import path from 'path';
import fs from 'fs';
import './gopolyfill';
import Go from '../src/go';
import { Genji } from '../src/index';

export async function initDatabase() {
    const go = new Go();
    const wasmPath = path.resolve(__dirname, '../', 'dist', 'genji.wasm');
    const buffer = fs.readFileSync(wasmPath);

    // @ts-ignore
    const result = await WebAssembly.instantiate(buffer, go.importObject);
    go.run(result.instance);
    return new Genji();
}
