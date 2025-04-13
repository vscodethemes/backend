import * as templates from "./language-templates";

const defaultLanguages = [
  {
    name: "JavaScript",
    extName: "js",
    scopeName: "source.js",
    grammar: "javascript.tmLanguage.json",
    template: templates.javascript,
    tabName: "main.js",
  },

  {
    name: "CSS",
    extName: "css",
    scopeName: "source.css",
    grammar: "css.tmLanguage.json",
    template: templates.css,
    tabName: "styles.css",
  },

  {
    name: "HTML",
    extName: "html",
    scopeName: "text.html.basic",
    grammar: "html.tmLanguage.json",
    template: templates.html,
    tabName: "index.html",
  },

  {
    name: "Python",
    extName: "py",
    scopeName: "source.python",
    grammar: "MagicPython.tmLanguage.json",
    template: templates.python,
    tabName: "main.py",
  },

  {
    name: "Go",
    extName: "go",
    scopeName: "source.go",
    grammar: "go.tmLanguage.json",
    template: templates.go,
    tabName: "main.go",
  },

  {
    name: "Java",
    extName: "java",
    scopeName: "source.java",
    grammar: "java.tmLanguage.json",
    template: templates.java,
    tabName: "Main.java",
  },

  {
    name: "C++",
    extName: "cpp",
    scopeName: "source.cpp",
    grammar: "cpp.tmLanguage.json",
    template: templates.cpp,
    tabName: "main.cpp",
  },

  {
    name: "PHP",
    extName: "php",
    scopeName: "source.php",
    grammar: "php.tmLanguage.json",
    template: templates.php,
    tabName: "main.php",
  },

  {
    name: "Ruby",
    extName: "rb",
    scopeName: "source.ruby",
    grammar: "ruby.tmLanguage.json",
    template: templates.ruby,
    tabName: "main.rb",
  },

  {
    name: "Rust",
    extName: "rs",
    scopeName: "source.rust",
    grammar: "rust.tmLanguage.json",
    template: templates.rust,
    tabName: "main.rs",
  },

  {
    name: "Elixir",
    extName: "ex",
    scopeName: "source.elixir",
    grammar: "elixir.tmLanguage.json",
    template: templates.elixir,
    tabName: "main.ex",
  },
] as const;

export default defaultLanguages;
export type Language = (typeof defaultLanguages)[number];
