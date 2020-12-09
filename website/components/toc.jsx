import Link from 'next/link'
import React from 'react'

export default function Toc({data}) {
    return (
        <div>
            {data.map((item, indx) => (
                <Item key={indx} item={item} />
            ))}
        </div>
    )
}

const Item = ({item}) => (
    <>
        <li
            className="pl-5"
        >
            <a href={`#${item.slug}`}>
                {item.text}
            </a>
        </li>
        {item.children &&
            <ul
                className="pl-5"
            >
                {item.children.map((item, indx) => (
                    <Item key={indx} item={item} />
                ))}
            </ul>
        }
    </>
)

/*
const List = ({items}) => (
    <ul className="pl-5">
        {items.map((item, indx) => {
            if (item.children != undefined) { 
                return <List key={indx} items={item.children} />
            } else {
                return <Item key={indx} value={item} />
            }
        })}
    </ul>
)

const Item = ({value}) => (
    <li className="ml-4">
        <a href={value.slug}>
            {value.text}
        </a>
    </li>
)
*/
