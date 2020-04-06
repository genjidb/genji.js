import { Observable } from 'rxjs';
import './bindings/wasm_exec.js';

export async function initDatabase(cfg: { url: string } = { url: 'genji.wasm' }) {
  const go = new Go();
  const result = await WebAssembly.instantiateStreaming(fetch(cfg.url), go.importObject);
  go.run(result.instance);
  return new Genji();
}

export class Genji {
  async Database() {
    return new Promise<Database>((resolve, reject) => {
      runDB((err: any, id: number) => {
        if (err) {
          reject(err);
          return
        }

        resolve(new Database(id));
      });
    })
  }
}

class Database {
  id: number;

  constructor(id: number) {
    this.id = id;
  }

  exec(query: string, ...args: any) {
    return new Promise((resolve, reject) => {
      dbExec(this.id, query, ...args, (err: any) => {
        if (err) {
          reject(err);
        }

        resolve();
      })
    })
  }

  query(query: string, ...args: any) {
    return new Observable(observer => {
      dbQuery(this.id, query, ...args, (err: any, document: Object) => {
        if (err) {
          observer.error(err)
          return
        }

        if (!document) {
          observer.complete();
          return;
        }

        observer.next(document);
      })
    })
  }
};
