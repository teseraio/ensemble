
import Community from "@teseraio/oss-react-community"
import { NewspaperIcon, PhoneIcon, SupportIcon } from '@heroicons/react/outline'
import Head from 'next/head'

const supportLinks = [
    {
        name: 'Community Forum',
        href: '#',
        hrefText: 'Contact us',
        description:
            'Varius facilisi mauris sed sit. Non sed et duis dui leo, vulputate id malesuada non. Cras aliquet purus dui laoreet diam sed lacus, fames.',
        icon: PhoneIcon,
    },
    {
        name: 'Office hours',
        href: '#',
        hrefText: 'Contact us',
        description:
            'Varius facilisi mauris sed sit. Non sed et duis dui leo, vulputate id malesuada non. Cras aliquet purus dui laoreet diam sed lacus, fames.',
        icon: SupportIcon,
    },
    {
        name: 'Github',
        href: '#',
        hrefText: 'Contact us',
        description:
            'Varius facilisi mauris sed sit. Non sed et duis dui leo, vulputate id malesuada non. Cras aliquet purus dui laoreet diam sed lacus, fames.',
        icon: NewspaperIcon,
    },
]

export default function IndexPage() {
    return (
        <div>
            <Head>
                <title key="title">Community | Ensemble</title>
            </Head>
            <Community supportLinks={supportLinks} />
        </div>
    )
}
