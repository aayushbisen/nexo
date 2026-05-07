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
- **Status:** 📅 Pending

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
