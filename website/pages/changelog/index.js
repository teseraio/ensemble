
import Changelog from "@teseraio/oss-react-changelog"
import Head from 'next/head'

export default function Index(data) {
    return (
        <div>
            <Head>
                <title key="title">Changelog | Ensemble</title>
            </Head>
            <Changelog.Changelog {...data} />
        </div>
    )
}

export async function getStaticProps() {
    return Changelog.processChangelog()
}
