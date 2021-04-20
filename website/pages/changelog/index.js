
import {Changelog, processChangelog} from "../../lib/changelog"

export default function Index(data) {
    return <Changelog {...data} />
}

export async function getStaticProps() {
    return processChangelog()
}
