import time
import redis

def main():
    r = redis.Redis(host="localhost", port=9090, decode_responses=True)

    # PING
    pong = r.ping()
    print(f"PING: {pong}")

    # SET
    r.set("name", "nexo")
    print("SET name nexo: OK")

    # GET
    val = r.get("name")
    print(f"GET name: {val}")

    # SET with EX (seconds)
    r.set("temp", "expires-in-5s", ex=5)
    print("SET temp expires-in-5s EX 5: OK")

    # GET before expiry
    val = r.get("temp")
    print(f"GET temp (before expiry): {val}")

    # EXPIRE
    r.expire("name", 10)
    print("EXPIRE name 10: OK")

    # DEL
    r.delete("temp")
    print("DEL temp: OK")

    # GET after DEL
    val = r.get("temp")
    if val is None:
        print("GET temp (after DEL): key not found (expected)")

    print("\nAll examples completed successfully!")

if __name__ == "__main__":
    main()
