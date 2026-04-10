/**
 * Titan MongoDB Native Extension
 * @module @titanpl/mongo
 */

/**
 * Connect to a MongoDB cluster and return a client instance.
 * 
 * @example
 * import { mongo } from "@titanpl/mongo";
 * import { t } from "@titanpl/core";
 * 
 * const client = mongo(t.env.DB_URI);
 * const db = client.db("sample_mflix");
 * const collection = db.collection("comments");
 * 
 * export function getMovies() {
 *   return collection.find({ name: "Mercedes Tyler" });
 * }
 * 
 * @param uri The MongoDB connection string (e.g., mongodb+srv://...)
 * @returns A MongoClient instance
 */
export function mongo(uri: string): MongoClient;

/**
 * Legacy database interface.
 * @deprecated Use the fluent `mongo(uri)` API instead.
 */
export const db: {
    connect(uri: string): MongoClient;
};

/**
 * Represents a connection to a MongoDB cluster.
 */
export class MongoClient {
    private handle: string;

    /**
     * Access a specific database.
     * @param name The name of the database.
     */
    db(name: string): MongoDatabase;

    /**
     * Closes the connection pool and cleans up resources in the native layer.
     */
    close(): { closed: boolean };
}

/**
 * Represents a MongoDB Database.
 */
export class MongoDatabase {
    private handle: string;
    private dbName: string;

    /**
     * Access a specific collection in this database.
     * @param name The name of the collection.
     */
    collection(name: string): MongoCollection;
}

/**
 * Represents a MongoDB Collection with CRUD capabilities.
 */
export class MongoCollection {
    private handle: string;
    private dbName: string;
    private collectionName: string;

    /**
     * Finds documents matching the filter.
     * @param filter The MongoDB query filter.
     * @param options Optional settings like limit and skip.
     */
    find(filter?: object, options?: FindOptions): any[];

    /**
     * Inserts a single document into the collection.
     * @param doc The document to insert.
     */
    insert(doc: object): { insertedId: string };
}

/**
 * Options for finding documents.
 */
export interface FindOptions {
    /**
     * Maximum number of documents to return. Default is 1000.
     */
    limit?: number;
    /**
     * Number of documents to skip.
     */
    skip?: number;
}
