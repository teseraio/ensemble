
import sidebarContent from "../../data/sidebar-docs.json"
import { getSidebarSlugs, getData, Docs } from '../../lib/docs'

export default function Post({postData}) {
    return <Docs postData={postData} sidebar={sidebarContent} />
}

const docsPrefix = "/docs/"

export async function getStaticProps({ params }) {
    console.log("-- params --")
    console.log(params)
    
    return getData(docsPrefix, params)
}

export async function getStaticPaths() {
    const xx =  getSidebarSlugs(docsPrefix)

    console.log(JSON.stringify(xx))
    return xx
}
