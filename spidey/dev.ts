#!/usr/bin/env -S deno run -A --watch=static/,routes/

import dev from "$fresh/dev.ts";
import config from "./fresh.config.ts";

import "$std/dotenv/load.ts";

const url = Deno.env.get("API_HOST")!;
console.log("Backend URL:", url)
await dev(import.meta.url, "./main.ts", config);
