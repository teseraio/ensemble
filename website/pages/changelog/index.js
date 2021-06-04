
import Changelog from "@teseraio/oss-react-changelog"

export default function Index(data) {
    return <Changelog.Changelog {...data} />
}

export async function getStaticProps() {
    return Changelog.processChangelog()
}
