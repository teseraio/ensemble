import fs from 'fs'
import path from 'path'
import matter from 'gray-matter'

const highlight = require('@mapbox/rehype-prism')

import {withTableOfContents} from './toc'
import renderToString from 'next-mdx-remote/render-to-string'

const components = { }

export async function getData(slug, prefix) {
    const fullPath = path.join("./data/" + prefix, slug.join("/") + ".mdx")
    const fileContents = fs.readFileSync(fullPath, 'utf8')
  
    const matterResult = matter(fileContents)
    
    const mdxSource = await renderToString(matterResult.content, { 
      mdxOptions: {
        remarkPlugins: [
          withTableOfContents,
        ],
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

export function getSidebarSlugs(sidebar, prefix) {
  let params = []
  traverseSidebar(sidebar.sidebar, (route) => {
    if (route.href == undefined) {
      return
    }
    params.push({
      params: {
        page: trimPrefix(route.href, prefix).split("/")
      }
    })
  })

  return {
    paths: params,
    fallback: false,
  }
}

var traverseSidebar = function(routes, fn) {
  for (var route of routes) {
    fn(route)
    if (route.routes != undefined) {
      traverseSidebar(route.routes, fn)
    }
  }
}

function trimPrefix(str, prefix) {
  if (str.startsWith(prefix)) {
      return str.slice(prefix.length)
  } else {
      return str
  }
}
