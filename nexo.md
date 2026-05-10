# 📦 Project: Nexo - Learning Journey

## 🎯 Overview
Nexo is a high-performance, type-safe, network-accessible in-memory key-value store. The goal is to master Go Generics, Networking (TCP), and advanced memory management (LRU).

## 🏗 Architecture Blueprint
`Client (TCP)` $\rightarrow$ `Protocol Parser (Codec)` $\rightarrow$ `Generic Store (RWMutex)` $\rightarrow$ `Eviction Engine (LRU)`

## 🗺 Roadmap

### Stage 1: The Type-Safe Core
- **Goal:** Implement a generic, thread-safe store.
- **Requirements:**
    - Use Generics (`[V any]`) for type safety.
    - Implement `sync.RWMutex` for concurrent access.
    - Provide `Set`, `Get`, and `Delete` operations.
- **Status:** ✅ Completed
- **Notes:** 
    - *Focus: Generics and Synchronization.*

### Stage 2: The Wire (Networking)
- **Goal:** Make the store accessible over TCP.
- **Status:** ✅ Completed
- **Notes:** 
    - *Focus: TCP Sockets, Protocol Design, and Network Concurrency.*
    - *Key Learning: Using `bufio.Scanner` for line-by-line command reading.*
    - *Key Learning: Implementing a text-based protocol (SET/GET/DEL) using `strings.Fields`.*
    - *Crucial Discovery: The difference between `return` (disconnects client) and `continue` (allows client to retry) in connection handlers.*

### Stage 3: The Memory Limit (LRU)
- **Goal:** Implement an LRU eviction policy to prevent memory overflow.
- **Status:** ✅ Completed
- **Notes:** 
    - *Focus: Doubly Linked Lists and O(1) Cache Eviction.*
    - *Key Learning: `container/list` provides a Doubly Linked List where nodes can be moved to the front or removed in O(1) time regardless of list size.*
    - *Architectural Pattern: The "Map + List" combo. The map provides instant lookup (O(1)), while the list maintains the chronological order of access.*
    - *LRU Logic: `GET` triggers `MoveToFront`; `SET` triggers `PushFront` and potentially a `Remove(Back)` if capacity is exceeded.*

### Stage 4: The Cluster (Distribution)
- **Goal:** Allow multiple Nexo nodes to sync data.
- **Status:** ✅ Completed
- **Notes:** 
    - *Focus: Consistent Hashing and Distributed Routing.*
    - *Key Learning: Implementing a Hash Ring to distribute keys across multiple servers.*
    - *Architectural Pattern: The Coordinator-Worker model. The Coordinator acts as a proxy that forwards requests to the correct worker based on the ring.*
    - *Concurrency: Managing multiple TCP connections (Client $\rightarrow$ Coordinator $\rightarrow$ Worker).*
    - *Enhancement: Virtual Nodes — Each physical server is represented by multiple virtual nodes on the ring (configurable via `replicas`). This improves key distribution uniformity and reduces the impact of node failures.*
    - *Key Learning: `slices.Sort` vs `sort.Slice` — The modern `slices.Sort` is cleaner and requires no comparison function for primitive types.*
    - *Key Learning: `range` over integers (Go 1.22+) — `for i := range replicas` replaces the manual `i < replicas` pattern.*

### Stage 5: Graceful Shutdown
- **Goal:** Allow the cluster to shut down cleanly on SIGINT/SIGTERM without losing in-flight data.
- **Status:** ✅ Completed
- **Notes:**
    - *Technique: `signal.NotifyContext` binds OS signals directly to a cancellable context — no manual channel handling needed.*
    - *Technique: `listener.Close()` unblocks `Accept()` and causes it to return an error, allowing the accept loop to exit.*
    - *Pattern: A goroutine sits on `<-ctx.Done()` and calls `listener.Close()` — the goroutine is spawned once, outside the accept loop.*
    - *Pattern: `wg.Add(1)` in `Start()` **before** `go handleConnection(...)` — never inside the goroutine itself.*
    - *Key Learning: Counter tracing — walk every `Add` and `Done` from main.go to verify the WaitGroup reaches zero.*

### Stage 6: Testing & Bug Hunt
- **Goal:** Achieve near-100% test coverage and catch latent bugs.
- **Status:** ✅ Completed
- **Notes:**
    - *Technique: `net.Pipe()` creates in-memory connection pairs for testing TCP handlers without opening real ports.*
    - *Technique: Injectable `Dial` function allows the coordinator to be tested without real worker servers.*
    - *Coverage: **94.9%** total across all packages — only the `net.Listen` error path remains uncovered.*
    - *Tests written: 38 unit tests across hashring (11), store (9), server (8), coordinator (10).*

### Stage 7: RESP Protocol
- **Goal:** Replace ad-hoc plain text with a proper protocol (RESP — REdis Serialization Protocol).
- **Status:** ✅ Completed
- **Notes:**
    - *Focus: Protocol Design, Dual-Mode Parsing, Framing.*
    - *Key Learning: RESP uses a type system encoded in the first byte — `+` for simple strings, `-` for errors, `$` for bulk strings, `*` for arrays.*
    - *Key Learning: Bulk strings are length-prefixed (`$3\r\nbar\r\n`) — you read exactly N bytes, no delimiter needed, no escaping issues.*
    - *Technique: `Peek(1)` allows a single `bufio.Reader` to dispatch between RESP and plain text without consuming data.*
    - *Architecture: Coordinator acts as a protocol bridge — clients can use RESP or plain text, internal worker communication is always RESP.*
- **Files:**
    - `internal/resp/resp.go` — Value type (Kind, Str, Integer, Array), Writer (`Write`), Reader (`Read`).
    - Changed: `internal/network/server.go` — dual-mode connection handler.
    - Changed: `internal/network/coordinator.go` — dual-mode client handler, RESP-only worker forwarding.

---

## 📝 Session Logs

### [2026-05-06] Stage 1: The Generic Foundation
The goal was to create a store that could hold any value type while maintaining strict type safety.

#### 🚩 The Struggle & The Solutions

**1. The "Zero Value" Return**
- **Error:** Attempted to return `return , false` in the `Get` method.
- **Cause:** Go requires all return parameters to be explicitly provided. Since `V` is generic, I couldn't return `0` or `""`.
- **Fix:** Declared a variable `var zero V` and returned it. This returns the "zero value" of whatever type the user chose for the store.

**2. The Nil Map Panic**
- **Error:** The store crashed on the first `Set` operation.
- **Cause:** The map was declared in the struct but never initialized with `make()`.
- **Fix:** Implemented the **Constructor Pattern** (`func New[V any]()`), ensuring the map is allocated before any operations occur.

**3. Concurrency Strategy**
- **Implementation:** Chose `sync.RWMutex` over a standard `Mutex`.
- **Key Learning:** `RLock` allows multiple concurrent readers, which is critical for a cache where reads happen significantly more often than writes.

#### 🏆 Final Result of Stage 1
- A fully functional, type-safe generic store.
- **Success Metric:** Verified with `main.go` using both `Store[string]` and `Store[int]`.
- **Key takeaway:** Mastered Go Generics and the correct implementation of a thread-safe shared resource.

### 🧠 Conceptual Breakthroughs: Stage 1
- **The Blueprint Analogy**: Understood that generics are blueprints, not real types, and must be "instantiated" (given a concrete type) before they can be used to allocate memory.

### [2026-05-06] Stage 2: The Wire (Network Access)
The goal was to transform the local store into a network server that accepts TCP connections and processes text commands.

#### 🚩 The Struggle & The Solutions

**1. The "Case Sensitivity" Trap**
- **Error:** Compilation errors when accessing struct fields (e.g., `s.port` vs `s.Port`).
- **Fix:** Ensured consistent use of exported (Uppercase) fields when accessing them from other packages.

**2. The "One-and-Done" Disconnect**
- **Error:** The server would close the connection immediately if the user made a typo in a command.
- **Cause:** Used `return` instead of `continue` in the command switch block.
- **Fix:** Changed `return` to `continue`, allowing the loop to persist and the client to try again.

**3. The "Silent Server" Problem**
- **Error:** Errors were printed to the server's terminal but not sent to the client.
- **Fix:** Replaced `fmt.Printf` with `io.WriteString(conn, ...)` to send error messages back over the network.

**4. The "Missing Newline" Hang**
- **Error:** Client (netcat) didn't display responses until the connection closed.
- **Cause:** Missing `\n` at the end of network responses.
- **Fix:** Added `\n` to all `io.WriteString` and `fmt.Fprintf` calls to signal the end of a response line.

#### 🏆 Final Result of Stage 2
- A concurrent TCP server that supports SET, GET, and DEL commands.
- **Success Metric:** Verified via `nc localhost 9090` with successful data persistence and error handling.
- **Key takeaway:** Mastered the basics of TCP networking, stream parsing with `bufio`, and the importance of the "Request-Response" cycle in network protocols.

### 🧠 Conceptual Breakthroughs: Stage 2
- **The "Sipping" Analogy**: Learned that `bufio` acts as a "glass" for a data "tank," reducing expensive system calls by reading chunks of data into memory.
- **The "Boundary" Concept**: Understood that TCP is a stream, not a packet system, making delimiters like `\n` essential for the receiver to identify a complete message.
- **The "Attentive Worker"**: Discovered the power of `for { select { ... } }` to allow a goroutine to be interrupted by a control signal while waiting for data.

### [2026-05-06] Stage 3: The Memory Limit (LRU Engine)
The goal was to prevent memory overflow by implementing a Least Recently Used (LRU) eviction policy.

#### 🚩 The Struggle & The Solutions

**1. The "Map vs List" Paradox**
- **Challenge:** Understanding how to maintain both a map forT fast lookup and a list for chronological order.
- **Solution:** Storing `*list.Element` in the map instead of the value itself. This allowed O(1) access to the node in the list to move it to the front.

**2. The "Type Assertion" Mystery**
- **Error:** Compiler errors when accessing values from the `list.Element`.
- **Cause:** `list.Element.Value` is of type `any`.
- **Fix:** Used type assertions (`element.Value.(*entry[V])`) to cast the "mystery box" back into the specific `entry` struct.

**3. The "Generic Instantiation" Error**
- **Error:** `cannot use generic type entry[V any] without instantiation`.
- **Cause:** Trying to use the generic struct in a non-generic context.
- **Fix:** Ensured all operations were performed within the `Store[V]` methods where the type `V` is defined.

**4. The "Wrong Eviction Order" Bug**
- **Error:** Cache was exceeding capacity by one item.
- **Cause:** Eviction check was happening *before* the new item was added.
- **Fix:** Moved the capacity check to occur *after* the `PushFront` operation.

#### 🏆 Final Result of Stage 3
- A fully functional LRU cache that automatically evicts the least recently used items.
- **Success Metric:** Verified that adding items beyond capacity correctly removes the oldest item.
- **Key takeaway:** Mastered the use of `container/list` and the pattern of combining two data structures to optimize for both speed and order.

### 🧠 Conceptual Breakthroughs: Stage 3
- **The "Mystery Box" (Reflection)**: Learned that `interface{}` (any) is a box that hides the underlying type, requiring a "Type Assertion" to safely retrieve the original struct.
- **The "Library Catalog" (LRU)**: Understood the hybrid Map+List architecture—using the map as a GPS for instant lookup and the list as a timeline for access order.
- **Symmetry in Concurrency**: Realized that any change to a shared resource (like the LRU list) must be applied consistently across all methods (`Set`, `Get`, `Delete`) to prevent state corruption.

### [2026-05-06] Stage 4: The Cluster (Distributed Routing)
The goal was to evolve Nexo from a single server to a distributed cluster where requests are routed to specific nodes using Consistent Hashing.

#### 🚩 The Struggle & The Solutions

**1. The "Symmetry" of the Ring**
- **Challenge:** Correcting the `sort.Slice` logic to sort by hash values rather than indices.
- **Fix:** Changed the comparison to `r.nodes[i] < r.nodes[j]`, ensuring the ring is logically ordered.

**2. The "Proxy" Connection Leak**
- **Error:** Using `defer workerConn.Close()` inside a loop in the Coordinator.
- **Cause:** Defers only run when the function returns, causing thousands of open connections to workers.
- **Fix:** Manually called `workerConn.Close()` immediately after the response was relayed to the client.

**3. The "Fake Success" Response**
- **Error:** Sending `OK` to the client before the worker had actually confirmed the operation.
- **Fix:** Moved the response logic to the end of the relay process, forwarding the worker's actual response back to the client.

**4. The "Address vs Port" Confusion**
- **Error:** Attempting to `net.Dial` using only a port number.
- **Fix:** Standardized on using full addresses (e.g., `localhost:9091`) in the Ring, while using port integers for the Server listeners.

#### 🏆 Final Result of Stage 4
- A fully functioning distributed system with a Coordinator and multiple Worker nodes.
- **Success Metric:** Verified via `nc` that keys are correctly routed to different workers and retrieved accurately.
- **Key takeaway:** Mastered the "Coordinator" pattern and the implementation of Consistent Hashing to achieve scalability and stability.

### 🧠 Conceptual Breakthroughs: Stage 4
- **The "Hash Ring"**: Understood that consistent hashing isn't about distributing keys evenly — it's about minimizing remapping when nodes join or leave. Only the keys on the affected arc need to move.
- **The "Virtual Node"**: Realized that adding N virtual copies of each physical node on the ring smooths out distribution skew without changing the hash function.

### [2026-05-09] Stage 5: Graceful Shutdown

#### 🚩 The Struggle & The Solutions

**1. The "Single-Select" Trap**
- **Error:** The shutdown goroutine was spawned inside the accept loop, creating a new goroutine on every connection.
- **Fix:** Moved the ctx goroutine outside the loop — one goroutine for the listener's lifetime.

**2. WaitGroup: The "Too Late" Add**
- **Error:** `wg.Add(1)` was inside `handleConnection`, not in `Start()` before `go handleConnection(...)`.
- **Result:** A race window where a signal could arrive between the `go` and the `Add`, making that connection invisible to the WaitGroup.
- **Fix:** Moved `wg.Add(1)` into `Start()` before the `go` statement.

**3. WaitGroup: The Orphaned Add**
- **Error:** An extra `wg.Add(1)` in `Coordinator.Start` with no matching `Done()`.
- **Result:** WaitGroup counter never reached zero — `wg.Wait()` blocked forever.
- **Fix:** Traced every Add/Done from main.go and removed the orphan.

### 🧠 Conceptual Breakthroughs: Stage 5
- *No new conceptual breakthroughs — this stage was purely about technique and shutdown mechanics.*

### [2026-05-10] Stage 6: Testing & Bug Hunt
The goal was to lock down the codebase with thorough tests and catch latent bugs before adding new features.

#### 🚩 The Struggle & The Solutions

**1. The "Pointer vs Value" Store Bug**
- **Problem:** Store `Set` update path panicked on `Get`. Failed assertion.
- **Cause:** `Set` retyped `0` entry from `entry[V]` to `*entry[V]`, pointing `list.Element.Value` at a new pointer while the map still held the old non-pointer value.
- **Fix:** Replaced `entry[V]` with `*entry[V]` – both map and list now point at the same struct.

**2. Testable TCP Without Real Ports**
- **Problem:** Tests requiring live TCP connections are flaky and slow, plus tests conflict when ports overlap.
- **Solution:** `net.Pipe()` creates in-memory `(conn1, conn2)` pairs. Both sides behave like real TCP `net.Conn` – perfect for testing connection handlers.

**3. Testable Coordinator Without Real Workers**
- **Problem:** Coordinator tests need worker servers running.
- **Solution:** Made `Dial` an injectable field on `Coordinator` (`Dial func(network, addr string) (net.Conn, error)`). Tests replace it with `net.Pipe()` and mock worker goroutines.

**4. The Orphaned WaitGroup Add**
- **Problem:** Shutdown blocked forever in coordinator tests.
- **Cause:** Coordinator's `Start` had a `wg.Add(1)` for a goroutine that called `Done()` only on error, but on success neither `Done` nor `Add` ran.
- **Fix:** Traced the exact Add-Done path and removed the orphaned Add.

#### 🏆 Final Result of Stage 6
- **38 tests** across hashring (11), store (9), server (8), coordinator (10).
- **94.9% coverage** – only `net.Listen` error path uncovered.

### 🧠 Conceptual Breakthroughs: Stage 6
- *No conceptual breakthroughs this stage — it was purely about technique and testing discipline.*

### [2026-05-10] Stage 7: RESP Protocol
The goal was to replace the ad-hoc plain text protocol with a proper framing protocol (RESP) so Nexo speaks a standard wire format.

#### 🚩 The Struggle & The Solutions

**1. The "Peek vs Read" Confusion**
- **Problem:** `bufio.Scanner` doesn't support `Peek`. Switched to `bufio.Reader`, but `Peek(1)` was initially confusing — it *inspects* without *consuming*.
- **Solution:** Peek is a window into the buffer. Call `Read` or `ReadString` afterwards to advance past the peeked data. This is exactly what enables dual-mode dispatch.

**2. The "Empty Line" Panic**
- **Problem:** Plain text path crashed with index out of range on `listCmd[0]` when `ReadString('\n')` returned an empty string.
- **Cause:** Leftover newlines in the TCP stream after a prior RESP message created a blank line.
- **Fix:** Added `if len(listCmd) == 0 { continue }` as a guard.

**3. The "Coordinator Protocol Bridge"**
- **Problem:** Coordinator forwarded commands to workers as plain text, but workers now write RESP responses. The coordinator was parsing plain text responses from RESP writers.
- **Solution:** Coordinator always sends RESP arrays to workers and always reads RESP responses with `resp.Read()`. Client-facing side retains dual-mode — the coordinator translates.

**4. The "RESP Error Format"**
- **Problem:** RESP errors need a `-` prefix (`-ERR message\r\n`), but simple string errors were being sent with the wrong format.
- **Fix:** Used `resp.Value{Kind: '-', Str: msg}` for errors and `resp.Value{Kind: '+', Str: "OK"}` for success.

#### 🏆 Final Result of Stage 7
- A working RESP implementation with Reader and Writer.
- Dual-mode servers (coordinator + workers) that accept RESP or plain text.
- Coordinator translates between client format and internal RESP.
- **Success Metric:** `printf '*2\r\n$3\r\nGET\r\n$1\r\na\r\n' | nc 9090` returns `$1\r\n1\r\n`.

### 🧠 Conceptual Breakthroughs: Stage 7
- **The "First Byte" Dispatch**: Realized that a protocol's type system can be encoded entirely in the first byte. Peek(1) is enough to know how to parse the rest — no multi-byte header needed.
- **Length-Prefixed > Delimiter-Escaped**: Bulk strings with length prefixes (`$3\r\nbar\r\n`) are simpler than delimiter escaping — you know exactly how many bytes to read, no need to escape delimiters inside the data.
- **The "Protocol Bridge" Pattern**: The coordinator isn't just routing data — it's translating between two formats. Internal consistency (always RESP) simplifies the worker-side code; the complexity is isolated in the bridge.

---

## 📉 The Mistake Log (Lessons Learned)

This section tracks the recurring patterns of errors encountered and the architectural shifts needed to fix them.

### 1. Control Flow: `return` vs `continue`
- **The Mistake:** Using `return` inside a `select` or `for` loop when an error occurs.
- **The Result:** The entire worker goroutine died, closing the connection to the client immediately.
- **The Fix:** Use `continue`. This skips the current faulty request but keeps the worker alive to process the next one.

### 2. I/O Direction: `fmt.Printf` vs `io.WriteString`
- **The Mistake:** Printing error messages to the server console instead of the network socket.
- **The Result:** The server operator saw the error, but the client was left hanging with no response.
- **The Fix:** Always use the `net.Conn` (via `io.WriteString`) to communicate errors back to the client.

### 3. Map Lifecycle: Declaration vs Initialization
- **The Mistake:** Declaring a map in a struct (`data map[string]V`) but forgetting to `make()` it.
- **The Result:** Immediate panic on first write (assignment to entry in nil map).
- **The Fix:** Implement a `New()` constructor to ensure all internal data structures are allocated before use.

### 4. Generics: The "Blueprint" Error
- **The Mistake:** Trying to use a generic type `entry[V]` in a non-generic context.
- **The Result:** `cannot use generic type ... without instantiation`.
- **The Fix:** Realized that generic types must be "instantiated" with a concrete type (like `entry[string]`) or used within a function that is also generic.

### 5. Concurrency: The "Symmetry" Trap
- **The Mistake:** Fixing a bug in `Get` but forgetting to apply the same logic to `Set` or `Delete`.
- **The Result:** Inconsistent behavior (e.g., `Get` refreshes the LRU order, but `Set` doesn't).
- **The Fix:** Always review all methods that touch the same shared resource (`S.data` and `S.list`) whenever a logic change is made.

### 6. Networking: The "Silence" Problem
- **The Mistake:** Sending responses without a trailing newline (`\n`).
- **The Result:** Clients (like `nc`) didn't display the response because they were waiting for a line-ending signal.
- **The Fix:** Always append `\n` to network responses to signal a complete message.

### 7. Sorting: Indices vs. Values
- **The Mistake:** Using `return i < j` inside `sort.Slice`.
- **The Result:** The slice was "sorted" by its indices rather than the actual hash values, breaking the hash ring logic.
- **The Fix:** Compare the actual values at those indices: `r.nodes[i] < r.nodes[j]`.

### 8. Ring Lifecycle: Missing Initialization
- **The Mistake:** Declaring a map in the `Ring` struct but not initializing it with `make()`.
- **The Result:** Panic on the first `AddNode` call (assignment to entry in nil map).
- **The Fix:** Implement a `New()` constructor to ensure the `nodeMap` is allocated before use.

### 9. The "Proxy" Connection Leak
- **The Mistake:** Using `defer conn.Close()` inside a loop that handles multiple requests.
- **The Result:** Connections to worker servers stayed open until the Coordinator itself crashed, leading to "too many open files" errors.
- **The Fix:** Explicitly call `.Close()` on the worker connection immediately after the response is relayed.

### 10. Premature Response
- **The Mistake:** Sending a success response to the client before the worker server actually processed the request.
- **The Result:** The client received an `OK` even if the worker failed or crashed.
- **The Fix:** Wait for the worker's response and relay that exact response back to the client.

### 11. Library Should Not Print
- **The Mistake:** Using `fmt.Printf` to log warnings inside a library package (`hashring`).
- **The Result:** The library had an unnecessary `fmt` import and printed to stdout, which is useless in production (nobody reads stdout from imports).
- **The Fix:** A library function with nowhere to return an error should `panic()` with a clear message instead of printing silently. `fmt.Printf` became dead code after the switch.

### 12. WaitGroup: "Add Before Go" Rule
- **The Mistake:** Placing `wg.Add(1)` inside `handleConnection` instead of in `Start()` before the `go` keyword.
- **The Result:** A race window where a signal could fire between `go handleConnection(...)` and `wg.Add(1)`, causing `wg.Wait()` to return before that connection finished draining.
- **The Fix:** Always `wg.Add(1)` in the spawning function before `go`, never inside the spawned goroutine.

### 13. WaitGroup: Forgot to Balance All Adds
- **The Mistake:** Adding `wg.Add(1)` inside `Coordinator.Start` without a matching `defer wg.Done()`.
- **The Result:** The WaitGroup counter was permanently +1, causing `wg.Wait()` to block forever on shutdown.
- **The Fix:** Trace the full counter path for every `Add` and `Done`. Every `Add` must have exactly one matching `Done`.

### 14. WaitGroup: Shared vs Separate Concerns
- **The Mistake:** Using a single WaitGroup to track both listener lifecycle and connection draining without verifying the total balance.
- **The Result:** Easy to introduce orphaned Adds or Dones when the two concerns overlap.
- **The Fix:** Trace the counter from main.go through every goroutine. If the logic gets complex, use separate WaitGroups for different concerns.

### 15. Bufered Reader: The Stale Peek
- **The Mistake:** Reading the first byte with `Peek(1)` for RESP detection, then trying to read the same data again with another call.
- **The Fix:** Peek does NOT consume the data — you must call `Read` or `ReadString` afterwards to actually consume it. Use Peek only for inspection.
- **Note:** This is why the dual-mode flow works: `Peek(1)` tells us if it's `*` (RESP) or not (plain text), then we either call `resp.Read(r)` or `r.ReadString('\n')` to consume the data.

### 16. Protocol Translation: One Format, Two Clients
- **The Mistake:** Sending plain text responses to RESP clients, causing the client to hang waiting for proper RESP framing.
- **The Fix:** The coordinator must track which format the client used (`respMode` bool) and encode responses accordingly — `response.Write(conn)` for RESP, string extraction for plain text.
