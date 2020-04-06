import { Observable } from 'rxjs';
import './bindings/wasm_exec';

export async function initDatabase({ url }) {
  const go = new Go();
  const result = await WebAssembly.instantiateStreaming(fetch(url), go.importObject);
  go.run(result.instance);
}

export async function loadDatabase() {
  await ready;

  return new Promise((resolve, reject) => {
    runDB((err, id) => {
      if (err) {
        reject(err);
        return
      }

      resolve(new Genji(id));
    });
  })
}

class Genji {
  async Database() {
    return new Promise((resolve, reject) => {
      runDB((err, id) => {
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
  constructor(id) {
    this.id = id;
  }

  exec(query, ...args) {
    return new Promise((resolve, reject) => {
      dbExec(this.id, query, ...args, (err) => {
        if (err) {
          reject(err);
        }

        resolve();
      })
    })
  }

  query(query, ...args) {
    return new Observable(observer => {
      dbQuery(this.id, query, ...args, (err, document) => {
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
