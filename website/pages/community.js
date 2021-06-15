
import Community from "@teseraio/oss-react-community"
import { CubeIcon, ChatAlt2Icon, ChatIcon } from '@heroicons/react/outline'

const supportLinks = [
    {
        name: 'Community Forum',
        href: 'https://discord.gg/NX6JxWgerk',
        hrefText: 'Ask a question',
        description:
            'Join our Discord community to learn about our latest announcements, chat with the devs and connect with other Ensemble users.',
        icon: ChatAlt2Icon,
    },
    {
        name: 'Office hours',
        href: 'https://calendly.com/ferran-tesera/ensemble-office-hours',
        hrefText: 'Talk with the devs',
        description:
            'If you are stuck with the deployment, wondering how to create your own backend or want to know more about our roadmap, stop by and we will be happy to talk with you.',
        icon: ChatIcon,
    },
    {
        name: 'Github',
        href: 'https://github.com/teseraio/ensemble',
        hrefText: 'Check repository',
        description:
            'Use Github to track the development and report any bugs. If you have any general question, use the forum instead.',
        icon: CubeIcon,
    },
]

export default function IndexPage() {
    return (
        <div>
            <Community supportLinks={supportLinks} />
        </div>
    )
}

IndexPage.getInitialProps = ({ req }) => {
    return {
      title: "Community"
    }
}
