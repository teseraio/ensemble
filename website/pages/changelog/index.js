
import {Changelog} from "@teseraio/tesera-oss"

export default function Index(data) {
    return <Changelog.Changelog {...data} />
}

export async function getStaticProps() {
    return Changelog.processChangelog()
}
