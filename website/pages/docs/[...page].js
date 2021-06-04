
import sidebarContent from "../../data/sidebar-docs.json"
import Docs from "@teseraio/oss-react-docs"
import { useRouter } from 'next/router'

export default function Post({postData}) {
    const router = useRouter()

    return (
        <div>
            <Docs.Docs current={router.asPath} postData={postData} sidebar={sidebarContent} />
        </div>
    )
}

const docsPrefix = "/docs"

export async function getStaticProps({ params }) {
    console.log("-- params --")
    console.log(params)
    
    return Docs.getData(docsPrefix, params, "Docs - ")
}

export async function getStaticPaths() {
    const xx =  Docs.getSidebarSlugs(docsPrefix)

    console.log("-- xxxxxx --")
    console.log(xx)
    
    console.log(JSON.stringify(xx))
    return xx
}
