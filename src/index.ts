import Go from './go';

export async function initDatabase(
  cfg: { url: string } = { url: 'genji.wasm' }
) {
  const go = new Go();
  const result = await WebAssembly.instantiateStreaming(
    fetch(cfg.url),
    // @ts-ignore
    go.importObject
  );
  go.run(result.instance);
  return new Genji();
}

export class Genji {
  async Database() {
    return new Promise<Database>((resolve, reject) => {
      openDB((err: any, id: number) => {
        if (err) {
          reject(err);
          return;
        }

        resolve(new Database(id));
      });
    });
  }
}

class Database {
  id: number;

  constructor(id: number) {
    this.id = id;
  }

  exec(query: string, ...args: any[]) {
    return new Promise<void>((resolve, reject) => {
      dbExec(this.id, query, args.map(jsValueToGenjiValue), (err: any) => {
        if (err) {
          reject(err);
        }

        resolve();
      });
    });
  }

  query(query: string, ...args: any): Stream {
    return new Stream(this.id, query, args.map(jsValueToGenjiValue));
  }
}

function jsValueToGenjiValue(v: any) {
  if (Object.prototype.toString.call(v) === '[object Object]') {
    Object.keys(v).forEach((k, _) => {
      v[k] = jsValueToGenjiValue(v[k]);
    });

    return {
      _keys: Object.keys(v),
      _object: v,
    };
  }

  if (Array.isArray(v)) {
    return v.map(val => jsValueToGenjiValue(val));
  }

  return v;
}

class Stream {
  id: number;
  query: string;
  args: any[];
  pipeline: ((document: Object) => Object | null)[];

  constructor(id: number, query: string, args: any[]) {
    this.id = id;
    this.query = query;
    this.args = args;
    this.pipeline = [];
  }

  callback(err: any, document: Object) {
    if (err) {
      return;
    }

    if (!document) {
      return;
    }
  }

  forEach(cb: (document: Object) => void) {
    return new Promise<void>((resolve, reject) => {
      dbQuery(this.id, this.query, this.args, (err: any, document: Object) => {
        if (err) {
          reject(err);
          return;
        }

        if (!document) {
          resolve();
          return;
        }

        for (const fn of this.pipeline) {
          const ret = fn(document);
          if (!ret) {
            return;
          }
          document = ret;
        }

        cb(document);
      });
    });
  }

  map(cb: (document: Object) => Object) {
    this.pipeline.push(cb);
  }

  filter(cb: (document: Object) => Boolean) {
    this.pipeline.push(v => (cb(v) ? v : null));
  }
}
