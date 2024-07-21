import path from "path";
import fs from "fs/promises";
import * as oniguruma from "vscode-oniguruma";
import * as vsctm from "vscode-textmate";

// Scope names must match scopeName in {language}.tmLanguage.json
export const scopeMap = {
  javascript: "source.js",
  typescript: "source.ts",
  css: "source.css",
  html: "text.html.basic",
  python: "source.python",
  go: "source.go",
  java: "source.java",
  cpp: "source.cpp",
} as const;

const wasmPath = path.join(require.resolve("vscode-oniguruma"), "../onig.wasm");

const onigLib = fs.readFile(wasmPath).then(({ buffer }) => {
  return oniguruma.loadWASM(buffer).then(() => {
    return {
      createOnigScanner(patterns: any) {
        return new oniguruma.OnigScanner(patterns);
      },
      createOnigString(s: any) {
        return new oniguruma.OnigString(s);
      },
    };
  });
});

export default new vsctm.Registry({
  onigLib,
  loadGrammar: async (scopeName) => {
    let fileName = "";
    if (scopeName === "source.js") {
      fileName = "javascript.tmLanguage.json";
    } else if (scopeName === "source.ts") {
      fileName = "typescript.tmLanguage.json";
    } else if (scopeName === "source.css") {
      fileName = "css.tmLanguage.json";
    } else if (scopeName === "text.html.basic") {
      fileName = "html.tmLanguage.json";
    } else if (scopeName === "source.python") {
      fileName = "MagicPython.tmLanguage.json";
    } else if (scopeName === "source.go") {
      fileName = "go.tmLanguage.json";
    } else if (scopeName === "source.java") {
      fileName = "java.tmLanguage.json";
    } else if (scopeName === "source.cpp") {
      fileName = "cpp.tmLanguage.json";
    }

    if (fileName) {
      const grammarPath = path.join(__dirname, `../../grammars/${fileName}`);
      const data = await fs.readFile(grammarPath);
      return vsctm.parseRawGrammar(data.toString(), grammarPath);
    }

    return null;
  },
});
