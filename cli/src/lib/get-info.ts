import path from "path";
import fs from "fs/promises";
import xml2js from "xml2js";
import stripEmoji from "emoji-strip";
import stripComments from "strip-json-comments";
import { unwrapError } from "./utils";

export interface Extension {
  displayName: string;
  description: string;
  githubLink?: string;
}

export interface ThemeContribute {
  label?: string;
  uiTheme: string;
  path: string;
}

export interface InfoResult {
  extension: Extension;
  themeContributes: ThemeContribute[];
}

// Parse the extension directory and return the extension and themes.
export default async function getInfo(dir: string): Promise<InfoResult> {
  const manifestPath = path.resolve(dir, "./extension.vsixmanifest");
  // Check if the manifest file exists.
  try {
    await fs.access(manifestPath);
  } catch (err) {
    throw new Error(`Could not find extension manifest at '${manifestPath}'`);
  }

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

  return { extension, themeContributes };
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
