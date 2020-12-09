import React, {useState, useEffect} from "react"
import Octicons from "./octicons.svg"
import Link from 'next/link'

export default function GithubWidget({repo}) {
    const [numStars, setNumStars] = useState(-1)

    useEffect(() => {
        fetch(
            `https://api.github.com/repos/${repo}`, 
            {
                "method": "GET",
                "headers": new Headers({
                    Accept: "application/json"
                })
            }
        )
        .then(res => res.json())
        .then(res => {
            setNumStars(res.stargazers_count)
        })
    }, [])

    return (
        <Link href={`https://github.com/${repo}`}>
            <a className="flex bg-white text-black">
                <span className="github-icon">
                    <Octicons style={{width: '20px', height: '20px'}}/>
                </span>
                {numStars != -1 && numStars != undefined &&
                    <span className="github-stars">{numStars}</span>
                }
            </a>
        </Link>
    )
}
