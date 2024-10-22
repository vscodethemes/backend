import path from "path";
import fs from "fs/promises";
import stripComments from "strip-json-comments";
import { trueCasePath } from "true-case-path";
import { convertTheme as tmThemeToJSON } from "tmtheme-to-json";
import * as vsctm from "vscode-textmate";
import TokenMetadata, { Style } from "./token-metadata";
import registry from "./language-registry";
import languages, { Language } from "./languages";
import { ThemeContribute } from "./get-info";
import { unwrapError, normalizeColor, alpha } from "./utils";

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
  language: Language;
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

// Parse the theme and return colors and tokens for each language.
export default async function parseTheme(
  extensionPath: string,
  themeContribute: ThemeContribute
): Promise<Theme> {
  const themePath = await trueCasePath(
    path.resolve(extensionPath, "extension", themeContribute.path)
  );

  const source = await readThemeSource(themePath);

  const displayName = themeContribute.label || source.name;
  if (!displayName) {
    throw new Error(`Theme must have a 'name' defined`);
  }

  let type: ThemeType;
  if (source.type === "light" || themeContribute.uiTheme === "vs") {
    type = "light";
  } else if (source.type === "dark" || themeContribute.uiTheme === "vs-dark") {
    type = "dark";
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
    languageTokens.push({ language: language, tokens });
  }

  const colors = normalizeColors(type, source.colors);

  const theme: Theme = {
    path: themePath,
    displayName,
    type,
    colors,
    languageTokens,
  };

  return theme;
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
    let defaultValueForType;
    if (typeof defaultValue === "string") {
      defaultValueForType = defaultValue;
    } else if (defaultValue) {
      defaultValueForType = defaultValue[type];
    }

    const value = colors[key] || defaultValueForType;
    // Use the default value as the background color in case the value has an alpha channel.
    const backgroundColorForType = backgroundColor || defaultValueForType;

    return normalizeColor(value, backgroundColorForType);
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

  const activityBarForeground = getColorValue(
    "activityBar.foreground",
    {
      dark: "#FFFFFF",
      light: "#FFFFFF",
      hcDark: "#FFFFFF",
      hcLight: editorForeground,
    },
    activityBarBackground
  );
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
    },
    activityBarBackground
  );
  if (!activityBarInActiveForeground) {
    throw new Error(`Missing color value for 'activityBar.inactiveForeground'`);
  }

  const activityBarBorder = getColorValue(
    "activityBar.border",
    {
      dark: undefined,
      light: undefined,
      hcDark: contrastBorder,
      hcLight: contrastBorder,
    },
    editorBackground
  );

  const activityBarActiveBorder = getColorValue(
    "activityBar.activeBorder",
    {
      dark: activityBarForeground,
      light: activityBarForeground,
      hcDark: contrastBorder,
      hcLight: contrastBorder,
    },
    editorBackground
  );
  if (!activityBarActiveBorder) {
    throw new Error(`Missing color value for 'activityBar.activeBorder'`);
  }

  const activityBarActiveBackground = getColorValue(
    "activityBar.activeBackground",
    undefined,
    activityBarBackground
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
    },
    editorBackground
  );

  const tabsContainerBorder = getColorValue(
    "editorGroupHeader.tabsBorder",
    undefined,
    editorBackground
  );

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

  const statusBarBorder = getColorValue(
    "statusBar.border",
    {
      dark: undefined,
      light: undefined,
      hcDark: contrastBorder,
      hcLight: contrastBorder,
    },
    statusBarBackground
  );

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

  const tabBorder = getColorValue(
    "tab.border",
    {
      dark: "#252526",
      light: "#F3F3F3",
      hcDark: contrastBorder,
      hcLight: contrastBorder,
    },
    tabsContainerBackground
  );
  if (!tabBorder) {
    throw new Error(`Missing color value for 'tab.border'`);
  }

  const tabActiveBorder = getColorValue(
    "tab.activeBorder",
    undefined,
    tabActiveBackground
  );

  const tabActiveBorderTop = getColorValue(
    "tab.activeBorderTop",
    {
      dark: undefined,
      light: undefined,
      hcDark: undefined,
      hcLight: "#B5200D",
    },
    tabActiveBackground
  );

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

  const titleBarBorder = getColorValue(
    "titleBar.border",
    {
      dark: undefined,
      light: undefined,
      hcDark: contrastBorder,
      hcLight: contrastBorder,
    },
    titleBarActiveBackground
  );

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

async function readThemeSource(themePath: string): Promise<ThemeSource> {
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
