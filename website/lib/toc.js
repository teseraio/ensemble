
const map = require('unist-util-map')
const is = require('unist-util-is')
const slugify = require('@sindresorhus/slugify')

module.exports.withTableOfContents = (obj) => {
    const contents = []
    return (tree) => {
        const res = map(tree, (node) => {
            if (is(node, 'heading')) {
                const text = stringifyNode(node)
                const slug = slugify(text)
                const level = node.depth

                contents.push({ text, slug, level, children: [] })

                node = {
                    type: 'html',
                    value: `<h${level}><a class="some" id="${slug}"></a><a class="anchor" href="#${slug}"># </a><span>${text}</span></h${level}>`
                }
            }
            return node
        })
        if (obj != undefined) {
            obj.result = buildToc(contents)
        }
        return res
    }
}

function buildToc(headings) {
    for (let i = headings.length-1; i >= 0; i--) {
        // for each item check backwards till you find the parent
        // which is someone with a higher -1 depth
        for (let j = i-1; j >= 0; j--) {
            if (headings[i].level - 1 == headings[j].level) {
                headings[j].children.unshift(headings[i])
                break
            }
        }
    }

    // for the result keep all the level 2 headings
    headings = headings.filter(heading => heading.level == 2)

    return headings
}

function stringifyNode(node) {
    return node.children
        .filter((n) => n.type === 'text')
        .map((n) => n.value)
        .join('')
}
