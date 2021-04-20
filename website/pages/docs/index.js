
import {Docs} from '../../lib/docs'
import sidebarContent from "../../data/sidebar-docs.json"

export default function FirstPost({sidebar}) {
  const postData = {
    title: 'Documentation',
  }
  return <Docs index={Index} postData={postData} sidebar={sidebarContent}/>
}

const Index = () => (
  <div>
    {'Index'}
  </div>
)
