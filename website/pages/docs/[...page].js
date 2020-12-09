
import { getSidebarSlugs, getData } from '../../lib/sidebar'
import sidebarContent from "../../data/sidebar-docs.json"

import DocsPage from '../../components/docs'

export default function Post({postData}) {
    return <DocsPage postData={postData} sidebar={sidebarContent} />
}

const docsPrefix = "/docs/"

export async function getStaticProps({ params }) {
    return getData(params.page, docsPrefix)
}

export async function getStaticPaths() {
    return getSidebarSlugs(sidebarContent, docsPrefix)
}
