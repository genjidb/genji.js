// @ts-ignore
if (!global.crypto) {
    const nodeCrypto = require("crypto");
    // @ts-ignore
    global.crypto = {
        // @ts-ignore
        getRandomValues(b) {
            nodeCrypto.randomFillSync(b);
        },
    };
}

// @ts-ignore
if (!global.performance) {
    // @ts-ignore
    global.performance = {
        now() {
            const [sec, nsec] = process.hrtime();
            return sec * 1000 + nsec / 1000000;
        },
    };
}

// @ts-ignore
if (!global.TextEncoder) {
    // @ts-ignore
    global.TextEncoder = require("util").TextEncoder;
}

// @ts-ignore
if (!global.TextDecoder) {
    // @ts-ignore
    global.TextDecoder = require("util").TextDecoder;
}
