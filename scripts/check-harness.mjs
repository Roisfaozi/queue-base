import assert from "node:assert/strict";
import { lstatSync, readFileSync, readdirSync, readlinkSync, realpathSync } from "node:fs";
import { dirname, isAbsolute, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("../", import.meta.url));
const skills = [
	"qms-domain", "review-findings", "dispatch-prompt", "api-endpoint",
	"auth-tenant-casbin", "go-service", "frontend-surface", "cross-stack-change",
	"verification-before-completion",
];
const documents = [
	"AGENTS.md", "llm/README.md", "llm/cache/project-overview.md",
	"llm/cache/architecture.md", "llm/cache/domain-rules.md", "llm/cache/environment.md",
	"llm/conventions/verification-anti-patterns.md", "llm/workflows/feature.md",
	"llm/workflows/bugfix.md", "llm/workflows/cross-stack-change.md",
	...skills.map((name) => `.agents/skills/${name}/SKILL.md`),
];

function inside(repository, target) {
	const path = relative(repository, target);
	return path !== ".." && !path.startsWith("../") && !isAbsolute(path);
}

export function checkDiscovery(repository) {
	const link = resolve(repository, ".claude/skills");
	assert(lstatSync(link).isSymbolicLink(), "skill discovery must be a symlink");
	assert.equal(readlinkSync(link), "../.agents/skills", "unexpected skill discovery target");
	assert.equal(realpathSync(link), resolve(repository, ".agents/skills"), "skill discovery escaped repository");
}

export function checkMetadata(text, directory, names = new Set()) {
	// ponytail: this checks the existing scalar metadata subset, not general YAML.
	const header = text.replaceAll("\r\n", "\n").match(/^---\n([\s\S]*?)\n---(?:\n|$)/)?.[1];
	assert(header, `${directory}: missing frontmatter`);
	const value = (field) => {
		const rows = [...header.matchAll(new RegExp(`^${field}:[ \\t]*(.*)$`, "gm"))];
		assert.equal(rows.length, 1, `${directory}: expected one ${field}`);
		const result = rows[0][1].trim().replace(/^(['"])(.*)\1$/, "$2").trim();
		assert(result && !["|", ">", "|-", ">-"].includes(result), `${directory}: invalid scalar ${field}`);
		return result;
	};
	const name = value("name");
	assert.equal(name, directory, `${directory}: name must match directory`);
	value("description");
	assert(!names.has(name), `${directory}: duplicate skill name`);
	names.add(name);
	return name;
}

export function checkLinks(repository, source, text) {
	// ponytail: inline Markdown links only; add a Markdown parser if the owned docs adopt reference links.
	for (const match of text.matchAll(/\[[^\n]*?\]\((<[^>\n]+>|[^\s)]+)\)/g)) {
		const link = match[1].replace(/^<|>$/g, "");
		if (/^(https?:|mailto:|#)/i.test(link)) continue;
		assert(!/^[a-z][a-z\d+.-]*:/i.test(link), `${source}: unsupported link ${link}`);
		const path = decodeURIComponent(link.split(/[?#]/, 1)[0]);
		assert(path && !isAbsolute(path), `${source}: non-relative link ${link}`);
		const target = resolve(repository, dirname(source), path);
		assert(inside(repository, target), `${source}: link leaves repository: ${link}`);
		assert(inside(repository, realpathSync(target)), `${source}: symlink leaves repository: ${link}`);
	}
}

function main() {
	checkDiscovery(root);
	assert.match(readFileSync(resolve(root, "CLAUDE.md"), "utf8"), /^@AGENTS\.md\s*$/m, "CLAUDE.md must import AGENTS.md");
	const names = new Set();
	for (const entry of readdirSync(resolve(root, ".agents/skills"), { withFileTypes: true })) {
		assert(entry.isDirectory(), `unexpected skill entry: ${entry.name}`);
		const path = resolve(root, ".agents/skills", entry.name, "SKILL.md");
		assert(inside(root, realpathSync(path)), `skill leaves repository: ${entry.name}`);
		checkMetadata(readFileSync(path, "utf8"), entry.name, names);
	}
	for (const path of documents) {
		const target = resolve(root, path);
		assert(inside(root, realpathSync(target)), `document leaves repository: ${path}`);
		checkLinks(root, path, readFileSync(target, "utf8"));
	}
	const manifest = JSON.parse(readFileSync(resolve(root, "package.json"), "utf8"));
	assert.equal(manifest.scripts["harness:check"], "node --test scripts/check-harness.test.mjs && node scripts/check-harness.mjs");
	console.log(`Harness OK: ${names.size} skill metadata, ${documents.length} linked documents, Claude discovery.`);
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
	try {
		main();
	} catch (error) {
		console.error(`Harness FAIL: ${error.message}`);
		process.exitCode = 1;
	}
}
