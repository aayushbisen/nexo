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

### 🧠 Conceptual Breakthroughs: Stage 3
- **The "Mystery Box" (Reflection)**: Learned that `interface{}` (any) is a box that hides the underlying type, requiring a "Type Assertion" to safely retrieve the original struct.
- **The "Library Catalog" (LRU)**: Understood the hybrid Map+List architecture—using the map as a GPS for instant lookup and the list as a timeline for access order.
- **Symmetry in Concurrency**: Realized that any change to a shared resource (like the LRU list) must be applied consistently across all methods (`Set`, `Get`, `Delete`) to prevent state corruption.

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
