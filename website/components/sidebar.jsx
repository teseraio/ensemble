import Link from 'next/link'
import React from 'react'
import { useRouter } from 'next/router'
import clsx from 'clsx';

import Chevron from '../assets/chevron.svg'

const {useState } = React;

export default function Sidebar({sidebar}) {
    return (
        <div className="pt-10 overflow-y-hidden sidebar">
            <ul>
                {sidebar.map((route, indx) => {
                    const {type} = route
                    if (type == "title") {
                        return <TitleSection key={indx} item={route} />
                    } else if (type == "sep") {
                        return <Separator />
                    }
                    return <Item key={indx} text={route.name} href={route.href} prefix={route.prefix} items={route.routes} />
                })}
            </ul>
        </div>
    )
}

const Separator = () => (
    <div>{'Split'}</div>
)

const TitleSection = ({item}) => {
    return (
        <h5
            className="pt-3 pb-2 font-bold text-xl"
        >
            {item.name}
        </h5>
    )
}

const Item = ({text, href, prefix, items}) => {
    const router = useRouter()
    const path = router.asPath

    // check if the route opens this sidebar
    let checkPath = href
    if (checkPath == undefined) {
        checkPath = prefix
    }
    const isActive = path.startsWith(checkPath)
    
    // check if the sidebar route is open
    const [isOpen, setIsOpen] = useState(isActive)
    const buttonHandler = () => {
        setIsOpen(current => !current)
    }

    return (
        <li
            id="nav"
            className={clsx(
                'relative',
            )}
        >
            <div
            className={clsx('pl-3',
                {
                    'text-main': isActive,
                }
            )}
            >
            {items ? 
                <span>
                    <a className="cursor-pointer block" onClick={buttonHandler}>
                        <Chevron className={clsx('absolute', {'rotate': isOpen})} style={{top: '17px', left: '5px'}}/>
                        <span>{text}</span>
                    </a>
                </span>
            : 
                <Link href={href}>
                    <a className="cursor-pointer">{text}</a>
                </Link>
            }
            </div>
            {items && isOpen &&
                <ul
                    className='pl-8 relative'
                    style={{top: '-3px'}}
                >
                    {items.map((item, indx) => (
                        <Item key={indx} text={item.name} href={item.href} items={item.routes} />
                    ))}
                </ul>
            }
        </li>
    )
}
