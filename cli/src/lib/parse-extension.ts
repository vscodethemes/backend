import path from "path";
import fs from "fs/promises";
import xml2js from "xml2js";
import stripEmoji from "emoji-strip";
import stripComments from "strip-json-comments";
import { trueCasePath } from "true-case-path";
import { convertTheme as tmThemeToJSON } from "tmtheme-to-json";
import * as vsctm from "vscode-textmate";
import TokenMetadata, { Style } from "./token-metadata";
import registry from "./language-registry";
import languages from "./languages";
import { unwrapError, normalizeColor, alpha } from "./utils";

export interface Extension {
  displayName: string;
  description: string;
  githubLink?: string;
}

export interface Theme {
  path: string;
  displayName: string;
  type: ThemeType;
  colors: Colors;
  languageTokens: LanguageTokens[];
}

export type ThemeType = "dark" | "light" | "hcDark" | "hcLight";

export interface Colors {
  editorBackground: string;
  editorForeground: string;
  activityBarBackground: string;
  activityBarForeground: string;
  activityBarInActiveForeground: string;
  activityBarBorder?: string;
  activityBarActiveBorder: string;
  activityBarActiveBackground?: string;
  activityBarBadgeBackground: string;
  activityBarBadgeForeground: string;
  tabsContainerBackground?: string;
  tabsContainerBorder?: string;
  statusBarBackground?: string;
  statusBarForeground: string;
  statusBarBorder?: string;
  tabActiveBackground?: string;
  tabInactiveBackground?: string;
  tabActiveForeground: string;
  tabBorder: string;
  tabActiveBorder?: string;
  tabActiveBorderTop?: string;
  titleBarActiveBackground: string;
  titleBarActiveForeground: string;
  titleBarBorder?: string;
}

export interface ThemeSource {
  type: string;
  colors: { [key: string]: string };
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

// Parse the extension directory and return the extension and themes.
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
    try {
      const source = await readThemeSource(dir, themeContribute);

      const displayName = themeContribute.label || source.name;
      if (!displayName) {
        throw new Error(`Theme must have a 'name' defined`);
      }

      let type: ThemeType;
      if (source.type === "dark" || themeContribute.uiTheme === "vs-dark") {
        type = "dark";
      } else if (source.type === "light" || themeContribute.uiTheme === "vs") {
        type = "light";
      } else if (
        source.type === "hc-dark" ||
        themeContribute.uiTheme === "hc-black"
      ) {
        type = "hcDark";
      } else if (
        source.type === "hc-light" ||
        themeContribute.uiTheme === "hc-light"
      ) {
        type = "hcLight";
      } else {
        throw new Error(`Theme 'type' must be one of 'dark' or 'light'`);
      }

      const languageTokens: LanguageTokens[] = [];
      for (const language of languages) {
        const tokens = await tokenizeTheme(source, language);
        languageTokens.push({ language: language.name, tokens });
      }

      const colors = normalizeColors(type, source.colors);

      themes.push({
        path: themeContribute.path,
        displayName,
        type,
        colors,
        languageTokens,
      });
    } catch (err) {
      throw new Error(
        `Error parsing theme '${themeContribute.path}': ${unwrapError(err)}`
      );
    }
  }

  return {
    extension,
    themes,
  };
}

// Extract and normalize the colors from the theme.
function normalizeColors(
  type: ThemeType,
  colors: Record<string, string>
): Colors {
  const getColorValue = (
    key: string,
    defaultValue?: string | Record<ThemeType, string | undefined>,
    backgroundColor?: string
  ) => {
    let value = colors[key];
    if (!value) {
      if (typeof defaultValue === "string") {
        value = defaultValue;
      } else if (defaultValue) {
        value = defaultValue[type];
      }
    }
    return normalizeColor(value, backgroundColor);
  };

  // Defaults pulled from https://github.com/Microsoft/vscode/blob/main/src/vs/workbench/common/theme.ts.
  const foreground = getColorValue("foreground", {
    dark: "#CCCCCC",
    light: "#616161",
    hcDark: "#FFFFFF",
    hcLight: "#292929",
  });
  if (!foreground) {
    throw new Error(`Missing color value for 'foreground'`);
  }

  const contrastBorder = getColorValue("contrastBorder", {
    light: undefined,
    dark: undefined,
    hcDark: "#6FC3DF",
    hcLight: "#0F4A85",
  });

  const editorBackground = getColorValue("editor.background", {
    light: "#FFFFFF",
    dark: "#1E1E1E",
    hcDark: "#000000",
    hcLight: "#FFFFFF",
  });
  if (!editorBackground) {
    throw new Error(`Missing color value for 'editor.background'`);
  }

  const editorForeground = getColorValue("editor.foreground", {
    light: "#333333",
    dark: "#BBBBBB",
    hcDark: "#FFFFFF",
    hcLight: foreground,
  });
  if (!editorForeground) {
    throw new Error(`Missing color value for 'editor.foreground'`);
  }

  const activityBarBackground = getColorValue("activityBar.background", {
    dark: "#333333",
    light: "#2C2C2C",
    hcDark: "#000000",
    hcLight: "#FFFFFF",
  });
  if (!activityBarBackground) {
    throw new Error(`Missing color value for 'activityBar.background'`);
  }

  const activityBarForeground = getColorValue("activityBar.foreground", {
    dark: "#FFFFFF",
    light: "#FFFFFF",
    hcDark: "#FFFFFF",
    hcLight: editorForeground,
  });
  if (!activityBarForeground) {
    throw new Error(`Missing color value for 'activityBar.foreground'`);
  }

  const activityBarInActiveForeground = getColorValue(
    "activityBar.inactiveForeground",
    {
      dark: alpha(activityBarForeground, 0.4),
      light: alpha(activityBarForeground, 0.4),
      hcDark: "#FFFFFF",
      hcLight: editorForeground,
    }
  );
  if (!activityBarInActiveForeground) {
    throw new Error(`Missing color value for 'activityBar.inactiveForeground'`);
  }

  const activityBarBorder = getColorValue("activityBar.border", {
    dark: undefined,
    light: undefined,
    hcDark: contrastBorder,
    hcLight: contrastBorder,
  });

  const activityBarActiveBorder = getColorValue("activityBar.activeBorder", {
    dark: activityBarForeground,
    light: activityBarForeground,
    hcDark: contrastBorder,
    hcLight: contrastBorder,
  });
  if (!activityBarActiveBorder) {
    throw new Error(`Missing color value for 'activityBar.activeBorder'`);
  }

  const activityBarActiveBackground = getColorValue(
    "activityBar.activeBackground"
  );

  const activityBarBadgeBackground = getColorValue(
    "activityBarBadge.background",
    {
      dark: "#007ACC",
      light: "#007ACC",
      hcDark: "#000000",
      hcLight: "#0F4A85",
    }
  );
  if (!activityBarBadgeBackground) {
    throw new Error(`Missing color value for 'activityBarBadge.background'`);
  }

  const activityBarBadgeForeground = getColorValue(
    "activityBarBadge.foreground",
    "#FFFFFF"
  );
  if (!activityBarBadgeForeground) {
    throw new Error(`Missing color value for 'activityBarBadge.foreground'`);
  }

  const tabsContainerBackground = getColorValue(
    "editorGroupHeader.tabsBackground",
    {
      dark: "#252526",
      light: "#F3F3F3",
      hcDark: undefined,
      hcLight: undefined,
    }
  );

  const tabsContainerBorder = getColorValue("editorGroupHeader.tabsBorder");

  const statusBarBackground = getColorValue("statusBar.background", {
    dark: "#007ACC",
    light: "#007ACC",
    hcDark: undefined,
    hcLight: undefined,
  });

  const statusBarForeground = getColorValue("statusBar.foreground", {
    dark: "#FFFFFF",
    light: "#FFFFFF",
    hcDark: "#FFFFFF",
    hcLight: editorForeground,
  });
  if (!statusBarForeground) {
    throw new Error(`Missing color value for 'statusBar.foreground'`);
  }

  const statusBarBorder = getColorValue("statusBar.border", {
    dark: undefined,
    light: undefined,
    hcDark: contrastBorder,
    hcLight: contrastBorder,
  });

  const tabActiveBackground = getColorValue(
    "tab.activeBackground",
    editorBackground
  );

  const tabInactiveBackground = getColorValue("tab.inactiveBackground", {
    dark: "#2D2D2D",
    light: "#ECECEC",
    hcDark: undefined,
    hcLight: undefined,
  });

  const tabActiveForeground = getColorValue("tab.activeForeground", {
    dark: "#FFFFFF",
    light: "#333333",
    hcDark: "#FFFFFF",
    hcLight: "#292929",
  });
  if (!tabActiveForeground) {
    throw new Error(`Missing color value for 'tab.activeForeground'`);
  }

  const tabBorder = getColorValue("tab.border", {
    dark: "#252526",
    light: "#F3F3F3",
    hcDark: contrastBorder,
    hcLight: contrastBorder,
  });
  if (!tabBorder) {
    throw new Error(`Missing color value for 'tab.border'`);
  }

  const tabActiveBorder = getColorValue("tab.activeBorder");

  const tabActiveBorderTop = getColorValue("tab.activeBorderTop", {
    dark: undefined,
    light: undefined,
    hcDark: undefined,
    hcLight: "#B5200D",
  });

  const titleBarActiveBackground = getColorValue("titleBar.activeBackground", {
    dark: "#3C3C3C",
    light: "#DDDDDD",
    hcDark: "#000000",
    hcLight: "#FFFFFF",
  });
  if (!titleBarActiveBackground) {
    throw new Error(`Missing color value for 'titleBar.activeBackground'`);
  }

  const titleBarActiveForeground = getColorValue("titleBar.activeForeground", {
    dark: "#CCCCCC",
    light: "#333333",
    hcDark: "#FFFFFF",
    hcLight: "#292929",
  });
  if (!titleBarActiveForeground) {
    throw new Error(`Missing color value for 'titleBar.activeForeground'`);
  }

  const titleBarBorder = getColorValue("titleBar.border", {
    dark: undefined,
    light: undefined,
    hcDark: contrastBorder,
    hcLight: contrastBorder,
  });

  return {
    editorBackground,
    editorForeground,
    activityBarBackground,
    activityBarForeground,
    activityBarInActiveForeground,
    activityBarBorder,
    activityBarActiveBorder,
    activityBarActiveBackground,
    activityBarBadgeBackground,
    activityBarBadgeForeground,
    tabsContainerBackground,
    tabsContainerBorder,
    statusBarBackground,
    statusBarForeground,
    statusBarBorder,
    tabActiveBackground,
    tabInactiveBackground,
    tabActiveForeground,
    tabBorder,
    tabActiveBorder,
    tabActiveBorderTop,
    titleBarActiveBackground,
    titleBarActiveForeground,
    titleBarBorder,
  };
}

// Tokenize the theme for a given language.
//
// References:
// https://github.com/microsoft/vscode-textmate/blob/9e3c5941668cbfcee5095eaec0e58090fda8f316/src/tests/themedTokenizer.ts
// https://github.com/microsoft/vscode-textmate/blob/9e3c5941668cbfcee5095eaec0e58090fda8f316/src/theme.ts
// https://github.com/microsoft/vscode-textmate/blob/9e3c5941668cbfcee5095eaec0e58090fda8f316/src/tests/themedTokenizer.ts#L13
// https://github.com/microsoft/vscode/tree/cf0231eb6e0632a655c71ab8a55b2fa0c960c3e3/extensions/typescript-basics/syntaxes
// https://github.com/microsoft/vscode/blob/94c9ea46838a9a619aeafb7e8afd1170c967bb55/src/vs/editor/common/modes.ts#L148
async function tokenizeTheme(
  theme: ThemeSource,
  language: (typeof languages)[number]
): Promise<Token[][]> {
  const grammar = await registry.loadGrammar(language.scopeName);
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
  const lines = language.template.split("\n");

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

// File reading functions.

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

// Manifest parser functions.

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

// Package JSON parser functions.

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
