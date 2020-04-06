import path from 'path';
import fs from 'fs';
import { map } from 'rxjs/operators';
import { Genji } from '../src/index';
import '../src/bindings/wasm_exec.js';
import '../src/globals.d';

async function initDatabase() {
    const go = new Go();
    const wasmPath = path.resolve(__dirname, '..', 'dist', 'genji.wasm');
    const buffer = fs.readFileSync(wasmPath);

    const result = await WebAssembly.instantiate(buffer, go.importObject);
    go.run(result.instance);
    return new Genji();
}

describe('db', function () {
    it('should run queries', async function () {
        const genji = await initDatabase();
        const db = await genji.Database();
        let count = 1;
        await db.exec("CREATE TABLE foo");
        await db.exec("INSERT INTO foo (a) VALUES (1), (2), (3)");
        await db.query("SELECT * FROM foo").
            pipe(map((val) => {
                expect(val).toEqual({ a: count });
                count++;
            })).
            subscribe();

        expect(count).toEqual(4);
    })
});
