import { log } from "@titanpl/native";
import { conn } from "../db/connect";


export function getuser(req) {
    return conn.find({
        email: "mark_addy@gameofthron.es"
    });
}


