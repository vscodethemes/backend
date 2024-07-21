#!/usr/bin/env node

import args from "args";
import path from "path";
import fs from "fs/promises";
import slugify from "slugify";
import languages from "./lib/languages";
import parseExtension, { Extension, Theme } from "./lib/parse-extension";
import renderSvg from "./lib/render-svg";
import renderPng from "./lib/render-png";

args.option("dir", "Directory of the extension", process.cwd());
args.option(
  "output",
  "Output directory of images",
  path.join(process.cwd(), "images")
);

const flags = args.parse(process.argv);

const dir = path.resolve(flags.dir);
const outputDir = path.resolve(flags.output);

// TODO: Use ora https://www.npmjs.com/package/ora

interface Metadata {
  extension: Extension;
  themes: ThemeMetadata[];
}

interface ThemeMetadata {
  theme: Theme;
  images: ImageMetadata[];
}

interface ImageMetadata {
  language: string;
  paths: Record<string, string>;
}

async function generateImages() {
  const { extension, themes } = await parseExtension(dir);

  await fs.mkdir(outputDir, { recursive: true });

  const metadata: Metadata = {
    extension,
    themes: [],
  };

  for (const theme of themes) {
    const images: ImageMetadata[] = [];
    for (const language of languages) {
      const svg = renderSvg(theme, language);
      const png = await renderPng(svg);

      // Write png to output directory.
      const themeFileName = path.basename(theme.path, ".json");
      const themeSlug = slugify(themeFileName, { lower: true, strict: true });
      const pngFileName = `${themeSlug}-${language.extName}.png`;
      const pngPath = path.join(outputDir, pngFileName);
      const svgFileName = `${themeSlug}-${language.extName}.svg`;
      const svgPath = path.join(outputDir, svgFileName);

      await Promise.all([
        fs.writeFile(svgPath, svg),
        fs.writeFile(pngPath, png),
      ]);

      images.push({
        language: language.extName,
        paths: { svg: svgPath, png: pngPath },
      });
    }

    metadata.themes.push({ theme, images });
  }

  // Write metadata to output directory.
  const metadataPath = path.join(outputDir, "output.json");
  await fs.writeFile(metadataPath, JSON.stringify(metadata));
}

generateImages().catch((err) => {
  console.error(err);
  process.exit(1);
});
