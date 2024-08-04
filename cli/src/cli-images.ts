#!/usr/bin/env node

import args from "args";
import path from "path";
import fs from "fs/promises";
import slugify from "slugify";
import { ThemeContribute } from "./lib/get-info";
import parseTheme, { Theme, Token } from "./lib/parse-theme";
import { Language } from "./lib/languages";
import renderSvg from "./lib/render-svg";
import renderPng from "./lib/render-png";

args.option("dir", "Directory of the extension", process.cwd());
args.option("label", "Label value of the theme contribute", "");
args.option("uiTheme", "uiTheme value of the theme contribute", "");
args.option("path", "Path value of the theme contribute", "");
args.option(
  "output",
  "Output directory of images",
  path.join(process.cwd(), "images")
);

const flags = args.parse(process.argv);

const dir = path.resolve(flags.dir);
const outputDir = path.resolve(flags.output);

interface ImagesResult {
  theme: Omit<Theme, "languageTokens">;
  languages: LanguageResult[];
}

interface LanguageResult {
  language: Language;
  tokens: Token[][];
  svgPath: string;
  pngPath: string;
}

async function generateImages() {
  if (!flags.path) {
    throw new Error("Path value is required");
  }

  if (!flags.label) {
    throw new Error("Label value is required");
  }

  if (!flags.uiTheme) {
    throw new Error("uiTheme value is required");
  }

  const themeContribute: ThemeContribute = {
    label: flags.label,
    uiTheme: flags.uiTheme,
    path: flags.path,
  };

  const { languageTokens, ...theme } = await parseTheme(dir, themeContribute);

  const languages: LanguageResult[] = [];
  await fs.mkdir(outputDir, { recursive: true });

  for (const { language, tokens } of languageTokens) {
    const svg = renderSvg(theme.displayName, theme.colors, language, tokens);
    const png = await renderPng(svg);

    // Write png to output directory.
    const themeFileName = path.basename(theme.path, ".json");
    const themeSlug = slugify(themeFileName, { lower: true, strict: true });
    const pngFileName = `${themeSlug}-${language.extName}.png`;
    const pngPath = path.join(outputDir, pngFileName);
    const svgFileName = `${themeSlug}-${language.extName}.svg`;
    const svgPath = path.join(outputDir, svgFileName);

    await Promise.all([fs.writeFile(svgPath, svg), fs.writeFile(pngPath, png)]);

    languages.push({ language, tokens, svgPath, pngPath });
  }

  const result: ImagesResult = { theme, languages };

  console.log(JSON.stringify(result));
}

generateImages().catch((err) => {
  console.error(err);
  process.exit(1);
});
