import { mongo } from "@titanpl/mongo";

const client = mongo(t.env.DB_URI);

export const conn = client.db("sample_mflix").collection("users");
