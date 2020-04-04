import Genji from '../src';

describe('blah', () => {
  it('works', (done) => {
    const db = new Genji(":memory:");

    let index = 0;

    const args = ["SELECT * FROM foo", "a", "b", 3];
    db.query("SELECT * FROM foo", "a", "b", 3).subscribe({
      next: val => {
        expect(val).toBe(args[index])
        index++;
      },
      complete: () => {
        expect(index).toEqual(4);
        done();
      },
    });

  });
});
