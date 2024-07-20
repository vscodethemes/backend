#!/usr/bin/env node

import args from "args";
import path from "path";

args.option("dir", "Directory of the extension", process.cwd());

const flags = args.parse(process.argv);

const dir = path.resolve(flags.dir);
console.log("dir", dir);
