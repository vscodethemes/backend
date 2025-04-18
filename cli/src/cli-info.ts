#!/usr/bin/env node

import args from "args";
import path from "path";
import getInfo, { Extension, ThemeContribute } from "./lib/get-info";

// args.option("dir", "Directory of the extension", "");
args.option("dir", "Directory of the extension", process.cwd());

const flags = args.parse(process.argv);

interface InfoResults {
  extension: Extension;
  themeContributes: ThemeContribute[];
}

async function info() {
  if (!flags.dir) {
    throw new Error("Extension directory not provided");
  }

  const dir = path.resolve(flags.dir);
  const { extension, themeContributes } = await getInfo(dir);

  const results: InfoResults = {
    extension,
    themeContributes,
  };

  console.log(JSON.stringify(results));
}

info().catch((err) => {
  console.error(err);
  process.exit(1);
});
