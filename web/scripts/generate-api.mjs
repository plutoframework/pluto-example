import { readFile, writeFile, mkdir, readdir } from "node:fs/promises";
import openapiTS, { astToString } from "openapi-typescript";
import ts from "typescript";

// Resolve the example's local shards without networking; openapi-typescript owns
// schema interpretation. Split its AST by operation group to keep files readable.
const base = new URL("../../openapi/", import.meta.url);
const main = JSON.parse(await readFile(new URL("openapi.json", base), "utf8"));
const auth = JSON.parse(await readFile(new URL("auth.json", base), "utf8"));
const schemas = JSON.parse(
  await readFile(new URL("schemas.json", base), "utf8"),
);
const sources = { "auth.json": auth, "schemas.json": schemas };
function local(value) {
  if (Array.isArray(value)) return value.map(local);
  if (!value || typeof value !== "object") return value;
  if (
    typeof value.$ref === "string" &&
    !value.$ref.startsWith("#/") &&
    !value.$ref.startsWith("./")
  )
    throw new Error("Only local OpenAPI references are allowed");
  if (value.$ref?.startsWith("./")) {
    const [filename, fragment] = value.$ref.slice(2).split("#");
    if (!sources[filename])
      throw new Error("Unexpected external OpenAPI reference");
    if (filename === "schemas.json") return { $ref: fragment };
    const found = fragment
      .slice(1)
      .split("/")
      .reduce(
        (node, key) => node[key.replaceAll("~1", "/").replaceAll("~0", "~")],
        sources[filename],
      );
    return local(found);
  }
  return Object.fromEntries(
    Object.entries(value).map(([key, item]) => [key, local(item)]),
  );
}
const document = local(main);
document.components.schemas = schemas.components.schemas;
const declarations = await openapiTS(document);
const result = new Map();
const header =
  "/** Generated from Pluto OpenAPI. Run npm run api:generate; do not edit. */\n";
for (const declaration of declarations) {
  const name = declaration.name?.text;
  if (!name) continue;
  if (name === "operations") {
    const groups = new Map();
    for (const member of declaration.members) {
      const group = member.name.text.split(".")[0];
      if (!groups.has(group)) groups.set(group, []);
      groups.get(group).push(member);
    }
    for (const [group, members] of groups) {
      const updated = ts.factory.updateInterfaceDeclaration(
        declaration,
        declaration.modifiers,
        ts.factory.createIdentifier(`${group}Operations`),
        declaration.typeParameters,
        declaration.heritageClauses,
        members,
      );
      result.set(
        `operations-${group}.ts`,
        header +
          'import type { components } from "./components";\n' +
          astToString([updated]),
      );
    }
    result.set(
      "operations.ts",
      header +
        [...groups.keys()]
          .map(
            (group) =>
              `import type { ${group}Operations } from "./operations-${group}";`,
          )
          .join("\n") +
        `\nexport interface operations extends ${[...groups.keys()].map((group) => `${group}Operations`).join(", ")} {}\n`,
    );
  } else {
    const imports =
      name === "paths"
        ? 'import type { operations } from "./operations";\n'
        : "";
    result.set(`${name}.ts`, header + imports + astToString([declaration]));
  }
}
const directory = new URL("../src/api/generated/", import.meta.url);
await mkdir(directory, { recursive: true });
const check = process.argv.includes("--check");
for (const [name, content] of result) {
  if (content.split("\n").length > 1000)
    throw new Error(`Generated file exceeds 1000 lines: ${name}`);
  const path = new URL(name, directory);
  if (check) {
    if ((await readFile(path, "utf8").catch(() => "")) !== content)
      throw new Error(`Stale API binding: ${name}; run npm run api:generate`);
  } else await writeFile(path, content);
}
const previous = await readdir(directory);
if (previous.some((name) => !result.has(name)))
  throw new Error(
    "Obsolete generated binding; review and remove it explicitly",
  );
console.log(
  `${check ? "Verified" : "Generated"} ${result.size} API modules from local OpenAPI`,
);
