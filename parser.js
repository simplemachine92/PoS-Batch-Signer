var codec = require("json-url")("lzw");
var fs = require("fs");

async function readFile() {
  const data = await fs.promises.readFile(__dirname + "/dat.txt", "utf-8");
  let typed = await codec.compress(JSON.parse(data));
  return typed;
}

(async () => {
  console.log(await readFile());
})();
