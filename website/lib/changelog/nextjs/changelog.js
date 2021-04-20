
import path from 'path'
import fs from 'fs'
import { version } from 'os'
import { isatty } from 'tty'

var remark = require('unified')
var markdown = require('remark-parse')
var html = require('remark-html')

const types = {
    "FEATURES:": "Features"
}

export default function processChangelog() {
    const fullPath = path.join(process.cwd(), "..", "CHANGELOG.md")
    const fileContents = fs.readFileSync(fullPath, 'utf8')

    const lines = fileContents.split("\n")
    let vers = [];

    for (var line of lines) {
        if (line.startsWith("## ")) {
            // new line
            vers.push({
                version: line.substr(3, 5),
                lines: []
            })
        } else {
            if (vers.length != 0) {
                // append a new line
                const val = types[line]
                if (val != null) {
                    line = "### " + val
                }
                vers[vers.length-1].lines.push(line)
            }
        }
    }

    for (var indx in vers) {
        const item = vers[indx]

        let content = ""
        const versionPath = path.join(process.cwd(), "data", "changelog", item.version + ".md")

        if (fs.existsSync(versionPath)) {
            const cc = fs.readFileSync(versionPath, 'utf8')
            content += cc + "\n"
        }

        content += item.lines.join("\n")

        const xx = remark()
        .use(markdown)
        .use(html)
        .processSync(content)

        vers[indx].content = xx.contents
    }

    return {
        props: {
            vers: vers
        },
    }
}
