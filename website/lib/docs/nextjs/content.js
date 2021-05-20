import fs from 'fs'
import path from 'path'
import matter from 'gray-matter'
import renderToString from 'next-mdx-remote/render-to-string'

const highlight = require('@mapbox/rehype-prism')

const components = { }

const contentFolder = "data"

export async function getData(subpath, params) {
  const fullPath = path.join(process.cwd(), contentFolder, subpath, params.page.join("/") + ".mdx")

  const fileContents = fs.readFileSync(fullPath, 'utf8')
  const matterResult = matter(fileContents)
  
  const mdxSource = await renderToString(matterResult.content, { 
    mdxOptions: {
      remarkPlugins: [],
      rehypePlugins: [
        highlight,
      ]
    }, 
    components
  })
  
  const postData = {
    fullPath,
    contentHtml: mdxSource,
    ...matterResult.data
  }

  return {
    props: {
      postData,
      key: fullPath
    }
  }
}

export function getSidebarSlugs(subpath) {
  const fullPath = path.join(process.cwd(), contentFolder, subpath)

  const files = fs.readdirSync(fullPath)

  console.log(fullPath)
  console.log(JSON.stringify(files))

  const paths = files.filter(el => /\.mdx$/.test(el)).map(i => {
      return {
          params: {
              page: i.replace(/\.mdx$/, '').split('/')
          }
      }
  })

  return {
      paths: paths,
      fallback: false
  }
}
