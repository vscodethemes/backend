export const css = `html {
  font-size: 16px;
  font-family: 'Open Sans', sans-serif;
}

body {
  margin: 0;
}

*,
*:before,
*:after {
  box-sizing: border-box;
}`;

export const html = `<html lang="en">

<head>
  <title>HTML Template</title>
</head>

<body>
  <main>
    <!-- Page contents -->
    <button id="btn" />
  </main>
</body>

</html>`;

export const javascript = `const btn = document.getElementById('btn');
let count = 0;

function render() {
  btn.innerText = \`Count: \${count}\`;
}

btn.addEventListener('click', () => {
  // Count from 1 to 10.
  if (count < 10) {
    count += 1;
    render();
  }
});
`;

export const python = `import os

"""A string"""

# A comment

class Foo(object):
    def __init__(self):
        num = 42
        print(num)

    @property
    def foo(self):
        return 'bar'
`;

export const go = `type config struct {
    port int
} 

func main() {
    var cfg config
  
    flag.IntVar(&cfg.port, "port", 4000)
    flag.Parse()

    // Start the web server.
    addr := fmt.Printf(":%d", cfg.port)
    log.Fatal(http.ListenAndServe(addr, nil))
}
`;

export const java = `public class Main {
  int num = 1;
  boolean bool = true;  
  String foo = "bar";

  static void printMessage() {
      System.out.println("Hello World!");
  }

  public static void main(String[] args) {
      // Print message to stdout.
      printMessage();
  }
}
`;

export const cpp = `#include <iostream>
#include <fstream>

int main() {
  string line;
  ifstream file;
  
  file.open("myfile.txt");
  
  // Read file line by line.
  while(getline(myfile, line)) {
     printf("%s", line.c_str());
  }
}
`;

export const php = `<?php

class Artist extends Model {
  /**
   * @var string
   */
  protected $table = "artists";

  public function new(string $name): self {
    return self::create([
      "name" => $name,
    ]);
  }
}
`;