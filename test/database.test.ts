import { initDatabase } from './init';

describe('db', function () {
    it('should run queries', async function () {
        const genji = await initDatabase();
        const db = await genji.Database();
        let count = 1;
        await db.exec("CREATE TABLE foo");
        await db.exec("INSERT INTO foo (a) VALUES (1), (2), (3)");
        await db.query("SELECT * FROM foo").
            forEach((val) => {
                expect(val).toEqual({ a: count });
                count++;
            });

        expect(count).toEqual(4);
    })
});
