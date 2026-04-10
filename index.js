import { createExt } from "./utils/native";

const ext = createExt("@titanpl/mongo");

/**
 * Connect to MongoDB and return a client object.
 * @param {string} uri - MongoDB connection string
 * @returns {MongoClient}
 */
export function mongo(uri) {
  const res = ext.call("mongo_connect", { uri });
  return new MongoClient(res.handle);
}

class MongoClient {
  constructor(handle) {
    this.handle = handle;
  }

  /**
   * Get a database instance.
   */
  db(name) {
    return new MongoDatabase(this.handle, name);
  }

  /**
   * Close the connection.
   */
  close() {
    return ext.call("mongo_close", { handle: this.handle });
  }
}

class MongoDatabase {
  constructor(handle, dbName) {
    this.handle = handle;
    this.dbName = dbName;
  }

  /**
   * Get a collection instance.
   */
  collection(name) {
    return new MongoCollection(this.handle, this.dbName, name);
  }
}

class MongoCollection {
  constructor(handle, dbName, collectionName) {
    this.handle = handle;
    this.dbName = dbName;
    this.collectionName = collectionName;
  }

  /**
   * Insert a document.
   */
  insert(doc) {
    return ext.call("mongo_insert", {
      handle: this.handle,
      db: this.dbName,
      collection: this.collectionName,
      doc,
    });
  }

  /**
   * Find documents.
   */
  find(filter = {}, options = {}) {
    return ext.call("mongo_find", {
      handle: this.handle,
      db: this.dbName,
      collection: this.collectionName,
      filter,
      ...options,
    });
  }
}

// Support for old-style connectivity if needed
export const db = {
  connect: mongo,
};