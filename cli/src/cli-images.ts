#!/usr/bin/env node

import args from "args";
import path from "path";
import parseExtension from "./lib/parse-extension";

args.option("dir", "Directory of the extension", process.cwd());
args.option("output", "Output directory of image metadata");

const flags = args.parse(process.argv);

console.log(flags);

const dir = path.resolve(flags.dir);

// TODO: Use ora https://www.npmjs.com/package/ora

async function generateImages() {
  const { extension, themes } = await parseExtension(dir);
  console.log(extension);

  for (const theme of themes) {
    // TODO: Generate SVG.
    // const svg = buildSvg(theme);
    // TODO: Generate PNG from SVG with https://github.com/yisibl/resvg-js.
  }
}

generateImages().catch((err) => {
  console.error(err);
  process.exit(1);
});
