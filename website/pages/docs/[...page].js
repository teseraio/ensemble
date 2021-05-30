
import sidebarContent from "../../data/sidebar-docs.json"
import Docs from "@teseraio/oss-react-docs"
import Head from 'next/head'

export default function Post({postData}) {
    return (
        <div>
            <Head>
                <title key="title">{`Docs - ${postData.title} | Ensemble`}</title>
            </Head>
            <Docs.Docs postData={postData} sidebar={sidebarContent} />
        </div>
    )
}

const docsPrefix = "/docs"

export async function getStaticProps({ params }) {
    console.log("-- params --")
    console.log(params)
    
    return Docs.getData(docsPrefix, params)
}

export async function getStaticPaths() {
    const xx =  Docs.getSidebarSlugs(docsPrefix)

    console.log("-- xxxxxx --")
    console.log(xx)
    
    console.log(JSON.stringify(xx))
    return xx
}
