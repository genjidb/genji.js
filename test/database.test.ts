import { initDatabase } from './init';

describe('db', function() {
  it('should run queries', async function() {
    const genji = await initDatabase();
    const db = await genji.Database();
    let count = 1;
    await db.exec('CREATE TABLE foo');
    await db.exec('INSERT INTO foo (a) VALUES (1), (2), (3)');
    await db.query('SELECT * FROM foo').forEach(val => {
      expect(val).toEqual({ a: count });
      count++;
    });

    expect(count).toEqual(4);
  });

  it('should support all types', async function() {
    const genji = await initDatabase();
    const db = await genji.Database();
    let count = 0;
    await db
      .query(
        'SELECT null, 1, 1.5, true, "hello", [1, "foo"], {a: 10, b: [true]}'
      )
      .forEach(val => {
        expect(val).toEqual({
          null: null,
          '1': 1,
          '1.5': 1.5,
          true: true,
          '"hello"': 'hello',
          '[1, "foo"]': [1, 'foo'],
          '{a: 10, b: [true]}': { a: 10, b: [true] },
        });
        count++;
      });

    expect(count).toEqual(1);
  });

  it('should support array and object params', async function() {
    const genji = await initDatabase();
    const db = await genji.Database();
    let count = 0;
    await db.exec(
      'CREATE TABLE foo; INSERT INTO foo (a, b) VALUES (?, ?)',
      { a: 1, b: ['foo'] },
      [1, true, [1], { b: false }]
    );
    await db.query('SELECT a, b FROM foo').forEach(val => {
      expect(val).toEqual({
        a: { a: 1, b: ['foo'] },
        b: [1, true, [1], { b: false }],
      });
      count++;
    });
    expect(count).toEqual(1);
  });
});
