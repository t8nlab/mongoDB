# @titanpl/mongo

A high-performance, native MongoDB extension for the Titan Planet Framework, built with Go.

## Installation

Add the extension to your Titan project's `tanfig.json`:

```json
{
  "name": "your-app",
  "extensions": {
    "allowNative": [
      "@titanpl/core",
      "@titanpl/mongo"
    ]
  }
}
```

## Usage

The extension provides a synchronous, fluent API that mirrors the official MongoDB driver while handling all heavy lifting in the native Go layer.

### Basic Setup

Create a connection in a shared file (e.g., `app/db.js`):

```javascript
import { mongo } from "@titanpl/mongo";
import { t } from "@titanpl/core";

// Connect once at module load
// Ensure DB_URI is set in your .env or Titan environment
const client = mongo(t.env.DB_URI);

export const db = client.db("my_database");
export const users = db.collection("users");
```

### CRUD Operations

```javascript
import { users } from "../db";

export function getProfile(req) {
    const userId = req.query.id;
    
    // Find documents (blocks until Go returns data)
    return users.find({ _id: userId });
}

export function signup(req) {
    const newUser = req.body;
    
    // Insert document
    return users.insert(newUser);
}
```

## API Reference

### `mongo(uri)`
Initializes a native connection pool to the specified MongoDB URI.

### `client.db(name)`
Returns a Database instance.

### `db.collection(name)`
Returns a Collection instance.

### `collection.find(filter, options)`
- `filter`: MongoDB query object. Supports automatic string-to-ObjectID conversion for `_id` fields.
- `options`: `{ limit: number, skip: number }`

### `collection.insert(doc)`
Inserts a document. Returns `{ insertedId: string }`.

## Performance Optimization

The extension is designed for high-concurrency environments with a dedicated Go connection pool.

### Latency Characteristics
- **Initial Connection (10-15s)**: The very first request to a cluster (especially on Atlas) may take some time (e.g., 15s) as the native layer performs SRV DNS resolution and establishes the initial secure TLS handshake. 
- **Subsequent Requests (50-80ms)**: Once the connection is established, subsequent calls use the active pool and typically return in 50-80ms, depending on your index performance and network distance.
- **Release Build**: For production, use `titan build -r`. The release build of the TitanPL engine optimizes native bridge throughput, leading to even lower latencies.

### Best Practices
- **Reuse Connections**: Always initialize your `mongo()` client at the module level (outside of action functions). This ensures you use the internal Go connection pool instead of reconnecting on every request.
- **Indexing**: Ensure your MongoDB collections have proper indexes for the fields you are filtering on.
- **Internal Handling**: This extension uses a C-Shared bridge which minimizes serialization overhead by using direct memory pointers between JS and Go.

## Development

To rebuild the native component:
```bash
npm run build:native
```
Requires Go 1.18+ and `CGO_ENABLED=1`.
