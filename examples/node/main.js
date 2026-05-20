const Redis = require("ioredis");

async function main() {
  const redis = new Redis({ host: "localhost", port: 9090 });

  try {
    // PING
    const pong = await redis.ping();
    console.log(`PING: ${pong}`);

    // SET
    await redis.set("name", "nexo");
    console.log("SET name nexo: OK");

    // GET
    const val = await redis.get("name");
    console.log(`GET name: ${val}`);

    // SET with EX (seconds)
    await redis.set("temp", "expires-in-5s", "EX", 5);
    console.log("SET temp expires-in-5s EX 5: OK");

    // GET before expiry
    const tempVal = await redis.get("temp");
    console.log(`GET temp (before expiry): ${tempVal}`);

    // EXPIRE
    await redis.expire("name", 10);
    console.log("EXPIRE name 10: OK");

    // DEL
    await redis.del("temp");
    console.log("DEL temp: OK");

    // GET after DEL
    const afterDel = await redis.get("temp");
    if (afterDel === null) {
      console.log("GET temp (after DEL): key not found (expected)");
    }

    console.log("\nAll examples completed successfully!");
  } finally {
    await redis.quit();
  }
}

main().catch((err) => {
  console.error("Error:", err.message);
  process.exit(1);
});
