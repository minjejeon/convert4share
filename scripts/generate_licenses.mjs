import { execSync } from 'child_process';
import fs from 'fs';
import path from 'path';

// Configuration
const FRONTEND_DIR = 'frontend';
const OUTPUT_JSON = 'frontend/src/assets/licenses.json';
const OUTPUT_MD = 'THIRD_PARTY_NOTICES.md';

console.log('Starting license generation...');

// Helper: Run command and return output
function run(cmd, cwd) {
    try {
        return execSync(cmd, { cwd, encoding: 'utf8', maxBuffer: 10 * 1024 * 1024 });
    } catch (e) {
        console.error(`Command failed: ${cmd}`, e.message);
        return '';
    }
}

// 1. Gather Frontend Licenses
console.log('Gathering frontend licenses...');
let frontendLicenses = [];
try {
    // We use license-checker to get the file paths
    const jsonStr = run('npx license-checker --production --json', FRONTEND_DIR);
    if (jsonStr) {
        const data = JSON.parse(jsonStr);
        for (const [key, info] of Object.entries(data)) {
            // key is "name@version"
            const lastAt = key.lastIndexOf('@');
            const name = key.substring(0, lastAt);
            const version = key.substring(lastAt + 1);

            let licenseText = 'License text not found.';
            if (info.licenseFile && fs.existsSync(info.licenseFile)) {
                try {
                    licenseText = fs.readFileSync(info.licenseFile, 'utf8');
                } catch (e) {
                    console.error(`Failed to read license file for ${key}: ${e.message}`);
                }
            } else if (info.licenseText) {
                licenseText = info.licenseText;
            }

            frontendLicenses.push({
                name,
                version,
                license: info.licenses,
                repository: info.repository,
                text: licenseText,
                source: 'frontend'
            });
        }
    }
} catch (e) {
    console.error('Frontend generation failed:', e);
}

// 2. Gather Backend Licenses
console.log('Gathering backend licenses...');
let backendLicenses = [];
try {
    // go list -m -json all
    const goListOutput = run('go list -m -json all', '.');
    // The output is a stream of JSON objects, not a single JSON array
    const modules = goListOutput.split('}\n{').map((chunk, index, array) => {
        let str = chunk;
        if (index === 0) str = str + '}';
        else if (index === array.length - 1) str = '{' + str;
        else str = '{' + str + '}';

        // Handle edges if only one item or clean split
        if (!str.startsWith('{')) str = '{' + str;
        if (!str.endsWith('}')) str = str + '}';

        // Clean up any double braces if split logic was imperfect
        try {
            return JSON.parse(str);
        } catch (e) {
            return null;
        }
    }).filter(m => m !== null);

    // Better parsing strategy: split by \n} and reconstruct
    // Or just regex replace.
    // Actually, splitting by `}\n` is safer.

    // Let's re-parse safely
    const goObjects = [];
    let buffer = '';
    const lines = goListOutput.split('\n');
    for (const line of lines) {
        if (line.trim() === '') continue;
        buffer += line + '\n';
        if (line === '}') {
            try {
                goObjects.push(JSON.parse(buffer));
                buffer = '';
            } catch (e) {
                // Not a full object yet?
            }
        }
    }

    for (const mod of goObjects) {
        if (!mod.Dir) continue; // Skip if no directory (e.g. uncached indirect)
        if (mod.Main) continue; // Skip the app itself

        const dir = mod.Dir;
        let licenseText = '';

        // Look for license file
        const potentialFiles = ['LICENSE', 'LICENSE.txt', 'LICENSE.md', 'COPYING', 'COPYING.txt', 'MIT-LICENSE'];

        // Read directory to find matches case-insensitive
        try {
            const files = fs.readdirSync(dir);
            const licenseFile = files.find(f => {
                const upper = f.toUpperCase();
                return potentialFiles.some(pf => upper.startsWith(pf));
            });

            if (licenseFile) {
                licenseText = fs.readFileSync(path.join(dir, licenseFile), 'utf8');
            } else {
                 licenseText = 'License text not found.';
            }
        } catch (e) {
            console.error(`Failed to read module dir ${dir}: ${e.message}`);
        }

        backendLicenses.push({
            name: mod.Path,
            version: mod.Version,
            license: 'Unknown', // Go modules don't explicitly list license type in go.mod
            repository: mod.Path.startsWith('github.com') ? `https://${mod.Path}` : '',
            text: licenseText,
            source: 'backend'
        });
    }

} catch (e) {
    console.error('Backend generation failed:', e);
}

// 3. Output
const allLicenses = [...frontendLicenses, ...backendLicenses];

// JSON
fs.writeFileSync(OUTPUT_JSON, JSON.stringify(allLicenses, null, 2));
console.log(`Wrote ${allLicenses.length} licenses to ${OUTPUT_JSON}`);

// Markdown
let mdContent = '# Third-Party Notices\n\nThis application uses the following third-party software:\n\n';
allLicenses.forEach(pkg => {
    mdContent += `## ${pkg.name} (${pkg.version})\n`;
    if (pkg.repository) mdContent += `Repository: ${pkg.repository}\n\n`;
    mdContent += `License: ${pkg.license || 'See Text'}\n\n`;
    mdContent += '```text\n';
    mdContent += pkg.text.trim();
    mdContent += '\n```\n\n---\n\n';
});

fs.writeFileSync(OUTPUT_MD, mdContent);
console.log(`Wrote markdown to ${OUTPUT_MD}`);
