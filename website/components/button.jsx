
import Link from 'next/link'
import clsx from 'clsx';

export default function Button({className, href, children}) {
    return (
        <Link href={href}>
            <a className={clsx("font-semibold p-4 shadow", className)}>
                {children}
            </a>
        </Link>
    )
}
