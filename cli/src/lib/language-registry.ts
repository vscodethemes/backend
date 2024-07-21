import path from "path";
import fs from "fs/promises";
import * as oniguruma from "vscode-oniguruma";
import * as vsctm from "vscode-textmate";
import languages from "./languages";

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
    for (const language of languages) {
      if (language.scopeName === scopeName) {
        fileName = language.grammar;
        break;
      }
    }

    if (fileName) {
      const grammarPath = path.join(__dirname, `../../grammars/${fileName}`);
      const data = await fs.readFile(grammarPath);
      return vsctm.parseRawGrammar(data.toString(), grammarPath);
    }

    return null;
  },
});
