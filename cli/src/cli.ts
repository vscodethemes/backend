#!/usr/bin/env node

import args from "args";

args.command("info", "Prase extension and output info");
args.command("images", "Generate preview images");

args.parse(process.argv);
