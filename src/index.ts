import { Observable } from 'rxjs';

export default class Genji {
  path: string;

  constructor(path: string) {
    this.path = path;
  }

  async exec(query: string, ...args: any[]) {
    console.log(`Run ${query} with args ${args}`);
  }

  query(query: string, ...args: any[]): Observable<Object> {
    return new Observable(observer => {
      observer.next(query);
      args.forEach(arg => observer.next(arg));
      observer.complete();
    })
  }
};
