import t from "@titanpl/route";


t.get("/").action("getuser");

t.start(5100, "Titan Running!");
