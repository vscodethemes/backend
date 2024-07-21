import path from "path";
import fs from "fs/promises";
import xml2js from "xml2js";
import stripEmoji from "emoji-strip";
import stripComments from "strip-json-comments";
import { trueCasePath } from "true-case-path";
import { convertTheme as tmThemeToJSON } from "tmtheme-to-json";
import * as vsctm from "vscode-textmate";
import * as templates from "./language-templates";
import TokenMetadata, { Style } from "./token-metadata";
import registry, { scopeMap } from "./language-registry";

export interface Extension {
  displayName: string;
  description: string;
  githubLink?: string;
}

export interface Theme {
  path: string;
  name: string;
  source: ThemeSource;
  // TODO: Extract these colors from the source and validate them
  // colors: {
  //   activityBarBackground,
  //   activityBarForeground,
  //   activityBarBorder,
  //   editorBackground,
  //   editorForeground,
  //   editorGroupHeaderTabsBackground,
  //   editorGroupHeaderTabsBorder,
  //   statusBarBackground,
  //   statusBarForeground,
  //   statusBarBorder,
  //   tabActiveBackground,
  //   tabActiveForeground,
  //   tabActiveBorder,
  //   tabBorder,
  //   titleBarActiveBackground,
  //   titleBarActiveForeground,
  //   titleBarBorder,
  // };
  languageTokens: LanguageTokens[];
}

export interface ThemeSource {
  type: string;
  colors: { [key: string]: any };
  tokenColors: TokenColor[];
  name?: string;
  include?: string;
}

export interface LanguageTokens {
  language: string;
  tokens: Token[][];
}

export interface TokenColor {
  name?: string;
  scope?: string | string[];
  settings: {
    foreground?: string;
    background?: string;
    fontStyle?: string;
  };
}

export interface Token {
  text: string;
  style: Style;
}

export interface ThemeContribute {
  label?: string;
  uiTheme: string;
  path: string;
}

export interface ParseResult {
  extension: Extension;
  themes: Theme[];
}

const languages = [
  "javascript",
  "css",
  "html",
  "python",
  "go",
  "java",
  "cpp",
] as const;

export default async function parseExtension(
  dir: string
): Promise<ParseResult> {
  const manifestPath = path.resolve(dir, "./extension.vsixmanifest");
  const manifestXml = await readXml(manifestPath);
  const displayName = parseExtensionDisplayName(manifestXml);
  const description = parseExtensionDescription(manifestXml);
  const githubLink = parseGithubLink(manifestXml);
  const relativePackageJsonPath = parsePackageJsonPath(manifestXml);

  const packageJsonPath = path.resolve(dir, relativePackageJsonPath);
  const packageJson = await readJson(packageJsonPath);
  const themeContributes = parseThemeContributes(packageJson);

  const extension: Extension = {
    displayName,
    description,
    githubLink,
  };

  const themes: Theme[] = [];
  for (const themeContribute of themeContributes) {
    const source = await readThemeSource(dir, themeContribute);
    const name = themeContribute.label || source.name;
    if (!name) {
      throw new Error(`Theme must have a 'name' defined`);
    }

    const languageTokens: LanguageTokens[] = [];
    for (const language of languages) {
      const tokens = await tokenizeTheme(source, language);
      languageTokens.push({ language, tokens });
    }

    themes.push({ path: themeContribute.path, name, source, languageTokens });
  }

  return {
    extension,
    themes,
  };
}

function unwrapError(error: any): string {
  if (error instanceof Error) {
    return error.message;
  }
  return String(error);
}

async function readXml(filePath: string): Promise<any> {
  const text = await fs.readFile(filePath);

  const xml = await new Promise((resolve, reject) => {
    xml2js.parseString(text, { trim: true, normalize: true }, (err, result) => {
      if (err) {
        return reject(
          new Error(`Invalid xml at '${filePath}': ${unwrapError(err)}`)
        );
      }
      resolve(result);
    });
  });

  return xml;
}

async function readJson(filePath: string): Promise<unknown> {
  try {
    const buffer = await fs.readFile(filePath);
    let text = stripComments(buffer.toString());

    // Strip trailing commas.
    text = text.replace(/\,(?=\s*?[\}\]])/g, "");

    return JSON.parse(text);
  } catch (err) {
    throw new Error(`Invalid json at '${filePath}': ${unwrapError(err)}`);
  }
}

async function readThemeSource(
  extensionPath: string,
  themeContribute: ThemeContribute
): Promise<ThemeSource> {
  const themePath = await trueCasePath(
    path.resolve(extensionPath, "extension", themeContribute.path)
  );

  try {
    const themeSource = await readThemeJSON(themePath);

    if (!isPartialTheme(themeSource)) {
      throw new Error(`Path '${themePath}' is invalid`);
    }

    const includes = [themeSource];
    const maxDepth = 10; // The max amount of theme includes to traverse.
    for (let i = 0; i < maxDepth; i += 1) {
      const nextInclude = includes[includes.length - 1]?.include;
      if (!nextInclude) {
        break;
      }
      const includesPath = path.resolve(themePath, "..", nextInclude);
      const includesData = await readThemeJSON(includesPath);

      if (!isPartialTheme(includesData)) {
        throw new Error(`Path '${includesPath}' is invalid`);
      }

      includes.push(includesData);
    }

    // Merge includes together starting from the right-most include.
    const theme: Partial<ThemeSource> = { colors: {}, tokenColors: [] };
    for (let i = includes.length - 1; i >= 0; i -= 1) {
      const include = includes[i];
      if (!include) {
        continue;
      }

      theme.name = include.name || theme.name;
      theme.type = include.type || theme.type;
      // The left-most includes colors takes precedence.
      theme.colors = { ...(theme.colors || {}), ...include.colors };
      let includeTokenColors = include.tokenColors;
      // Convert tmTheme to json.
      if (typeof includeTokenColors === "string") {
        const tmThemePath = path.resolve(themePath, "..", includeTokenColors);
        const tmTheme = await readTMTheme(tmThemePath);
        includeTokenColors = (tmTheme as any).settings;
      }
      // The right-most includes tokenColors gets appended.
      theme.tokenColors = [
        ...(includeTokenColors || []),
        ...(theme.tokenColors || []),
      ];
    }

    if (!isTheme(theme)) {
      throw new Error(
        `Theme must have 'type', 'colors' and 'tokenColors' defined`
      );
    }

    return theme;
  } catch (err) {
    throw new Error(`Invalid theme at ${themePath}: ${unwrapError(err)}`);
  }
}

// References:
// https://github.com/microsoft/vscode-textmate/blob/9e3c5941668cbfcee5095eaec0e58090fda8f316/src/tests/themedTokenizer.ts
// https://github.com/microsoft/vscode-textmate/blob/9e3c5941668cbfcee5095eaec0e58090fda8f316/src/theme.ts
// https://github.com/microsoft/vscode-textmate/blob/9e3c5941668cbfcee5095eaec0e58090fda8f316/src/tests/themedTokenizer.ts#L13
// https://github.com/microsoft/vscode/tree/cf0231eb6e0632a655c71ab8a55b2fa0c960c3e3/extensions/typescript-basics/syntaxes
// https://github.com/microsoft/vscode/blob/94c9ea46838a9a619aeafb7e8afd1170c967bb55/src/vs/editor/common/modes.ts#L148
async function tokenizeTheme(
  theme: ThemeSource,
  language: keyof typeof templates
): Promise<Token[][]> {
  // Use the typescript grammar for javascript
  const grammar = await registry.loadGrammar(scopeMap[language]);
  if (!grammar) throw new Error("grammar file not found.");

  registry.setTheme({
    name: theme.name,
    settings: [
      {
        settings: {
          foreground: theme.colors["editor.foreground"],
        },
      },
      ...theme.tokenColors,
    ],
  });

  const colorMap = registry.getColorMap();
  const lines = templates[language].split("\n");

  let lineTokens: Token[][] = [];
  let state: vsctm.StackElement | null = null;

  for (let i = 0, len = lines.length; i < len; i++) {
    const line = lines[i];
    if (line === undefined) continue;

    const tokenizationResult = grammar.tokenizeLine2(line, state);
    const tokens: Token[] = [];

    for (
      let j = 0, lenJ = tokenizationResult.tokens.length >>> 1;
      j < lenJ;
      j++
    ) {
      let startOffset = tokenizationResult.tokens[j << 1] ?? 0;
      let metadata = tokenizationResult.tokens[(j << 1) + 1] ?? 0;
      let endOffset =
        j + 1 < lenJ ? tokenizationResult.tokens[(j + 1) << 1] : line.length;
      let tokenText = line.substring(startOffset, endOffset);

      tokens.push({
        text: tokenText,
        style: TokenMetadata.getStyleObject(metadata, colorMap),
      });
    }

    lineTokens.push(tokens);

    state = tokenizationResult.ruleStack;
  }

  return lineTokens;
}

function parseMetadata(manifest: any): any {
  const metadata = manifest.PackageManifest.Metadata;
  if (!Array.isArray(metadata) || !metadata[0]) {
    throw new Error("Could not parse metadata from extension manifest");
  }

  return metadata[0];
}

function parseTextContent(node: any): string {
  if (typeof node === "string") return node;
  if (node && typeof node === "object" && "_" in node) return String(node["_"]);
  return "";
}

function parseExtensionDisplayName(manifest: any): string {
  const metadata = parseMetadata(manifest);
  const textContent = parseTextContent(metadata.DisplayName[0]);
  const extensionName = stripEmoji(textContent).trim();
  if (!extensionName) {
    throw new Error("Missing extension name in manifest");
  }

  return extensionName;
}

function parseExtensionDescription(manifest: any): string {
  const metadata = parseMetadata(manifest);
  const textContent = parseTextContent(metadata.Description[0]);
  const extensionDescription = stripEmoji(textContent).trim();
  if (!extensionDescription) {
    throw new Error("Missing extension description in manifest");
  }

  return extensionDescription;
}

function parseProperties(manifest: any): any[] {
  const metadata = parseMetadata(manifest);
  const properties = metadata.Properties;
  if (
    !Array.isArray(properties) ||
    !properties[0] ||
    !Array.isArray(properties[0].Property)
  ) {
    throw new Error("Could not parse properties from extension manifest");
  }

  return properties[0].Property;
}

function parseAttributes(node: any): { [key: string]: string } {
  if (node && typeof node === "object" && "$" in node) return node["$"];
  return {};
}

function parseGithubLink(manifest: any): string | undefined {
  const properties = parseProperties(manifest);

  let githubLink: string | undefined;
  for (const property of properties) {
    const attrs = parseAttributes(property);
    if (attrs.Id === "Microsoft.VisualStudio.Services.Links.GitHub") {
      githubLink = attrs.Value;
    }
  }

  return githubLink;
}

function parseAssets(manifest: any): any[] {
  const assets = manifest.PackageManifest.Assets;
  if (!Array.isArray(assets) || !assets[0] || !Array.isArray(assets[0].Asset)) {
    throw new Error("Could not parse assets from extension manifest");
  }

  return assets[0].Asset;
}

function parsePackageJsonPath(manifest: any): any {
  const assets = parseAssets(manifest);
  let packageJsonPath;

  for (const asset of assets) {
    const attrs = parseAttributes(asset);
    if (attrs.Type === "Microsoft.VisualStudio.Code.Manifest") {
      packageJsonPath = attrs.Path;
    }
  }

  if (!packageJsonPath) {
    throw new Error("Could not find package.json path in extension manifest");
  }

  return packageJsonPath;
}

function parseThemeContributes(packageJson: any) {
  const themeContributesByPath: Record<string, ThemeContribute> = {};
  if (
    packageJson &&
    typeof packageJson === "object" &&
    packageJson.contributes &&
    Array.isArray(packageJson.contributes.themes)
  ) {
    for (const contribute of packageJson.contributes.themes) {
      if (
        contribute &&
        typeof contribute === "object" &&
        contribute.label &&
        contribute.uiTheme &&
        contribute.path
      ) {
        themeContributesByPath[contribute.path] = contribute;
      }
    }
  }

  return Object.values(themeContributesByPath);
}

function isPartialTheme(data: any): data is Partial<ThemeSource> {
  return !!data && typeof data === "object";
}

function isTheme(data: any): data is ThemeSource {
  return (
    !!data &&
    typeof data === "object" &&
    "type" in data &&
    "colors" in data &&
    typeof data.colors === "object" &&
    "tokenColors" in data &&
    Array.isArray(data.tokenColors)
  );
}

async function readThemeJSON(filePath: string): Promise<unknown> {
  const ext = path.extname(filePath);
  if (ext === ".json") {
    return readJson(filePath);
  }
  // TODO: Support .tmTheme?
  // } else if (ext === '.tmTheme') {
  //   return readTMTheme(filePath);
  // }

  throw new Error(`Invalid theme extension at '${filePath}''`);
}

async function readTMTheme(filePath: string): Promise<unknown> {
  try {
    const buffer = await fs.readFile(filePath);
    return tmThemeToJSON(buffer.toString());
  } catch (err) {
    throw new Error(`Invalid json at '${filePath}'': ${unwrapError(err)}`);
  }
}
