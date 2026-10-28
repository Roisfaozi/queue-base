import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test } from "node:test";
import { checkLinks, checkMetadata, checkDiscovery } from "./check-harness.mjs";

test("metadata: positive, negative, and CRLF cases", () => {
	for (const row of [
		{ name: "positive", text: "---\nname: queue\ndescription: Queue changes.\n---\n# Queue", valid: true },
		{ name: "edge CRLF", text: "---\r\nname: queue\r\ndescription: Queue changes.\r\n---\r\n", valid: true },
		{ name: "negative missing description", text: "---\nname: queue\n---\n", valid: false },
		{ name: "negative missing header", text: "# Queue", valid: false },
		{ name: "negative duplicate name", text: "---\nname: queue\nname: other\ndescription: Queue.\n---\n", valid: false },
		{ name: "negative blank description", text: "---\nname: queue\ndescription: ''\n---\n", valid: false },
		{ name: "negative blank before field", text: "---\nname: queue\ndescription:\nlicense: MIT\n---\n", valid: false },
	]) {
		if (row.valid) assert.equal(checkMetadata(row.text, "queue"), "queue", row.name);
		else assert.throws(() => checkMetadata(row.text, "queue"), undefined, row.name);
	}
	assert.throws(() => checkMetadata("---\nname: other\ndescription: Queue.\n---\n", "queue"), /name/);
	const names = new Set(["queue"]);
	assert.throws(() => checkMetadata("---\nname: queue\ndescription: Queue.\n---\n", "queue", names), /duplicate/);
});

test("links and discovery: positive, negative, edge, and security cases", () => {
	const directory = mkdtempSync(join(tmpdir(), "queue-harness-"));
	const root = join(directory, "repo");
	mkdirSync(join(root, "docs"), { recursive: true });
	mkdirSync(join(root, ".agents/skills"), { recursive: true });
	mkdirSync(join(root, ".claude"));
	writeFileSync(join(root, "docs/source file.md"), "# Source\n");
	writeFileSync(join(directory, "outside.md"), "# Outside\n");
	symlinkSync("../.agents/skills", join(root, ".claude/skills"));
	symlinkSync(join(directory, "outside.md"), join(root, "docs/escape.md"));
	try {
		checkDiscovery(root);
		for (const row of [
			{ name: "positive", link: "[source](source%20file.md#heading)", valid: true },
			{ name: "edge spaces", link: "[source](<source file.md>)", valid: true },
			{ name: "edge parent", link: "[skills](../.agents/skills/)", valid: true },
			{ name: "edge external", link: "[web](https://example.com/page) [anchor](#heading)", valid: true },
			{ name: "negative missing", link: "[missing](missing.md)", valid: false },
			{ name: "security traversal", link: "[escape](../../outside.md)", valid: false },
			{ name: "security encoded traversal", link: "[escape](%2e%2e/%2e%2e/outside.md)", valid: false },
			{ name: "security symlink", link: "[escape](escape.md)", valid: false },
			{ name: "security absolute", link: `[escape](${join(directory, "outside.md")})`, valid: false },
			{ name: "negative encoding", link: "[broken](bad%zz.md)", valid: false },
		]) {
			if (row.valid) assert.doesNotThrow(() => checkLinks(root, "docs/index.md", row.link), row.name);
			else assert.throws(() => checkLinks(root, "docs/index.md", row.link), undefined, row.name);
		}
		rmSync(join(root, ".claude/skills"));
		symlinkSync(directory, join(root, ".claude/skills"));
		assert.throws(() => checkDiscovery(root), /discovery/);
	} finally {
		rmSync(directory, { recursive: true, force: true });
	}
});
