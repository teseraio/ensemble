
import Docs from "@teseraio/oss-react-docs"
import sidebarContent from "../../data/sidebar-docs.json"

export default function FirstPost({sidebar}) {
  const postData = {
    title: 'Documentation',
  }
  return <Docs.Docs main={Index} postData={postData} sidebar={sidebarContent}/>
}

const Index = () => (
  <div>
    {'Index'}
  </div>
)

FirstPost.getInitialProps = ({ req }) => {
  return {
    title: "Docs"
  }
}
