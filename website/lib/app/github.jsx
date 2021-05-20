import React, {useState, useEffect} from "react"
import Octicons from "./octicons.svg"

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
        <a
            href={`https://github.com/${repo}`}
            className="ml-8 inline-flex items-center justify-center px-4 py-2 border border-transparent shadow-sm text-base font-medium text-black bg-white hover:bg-white2"
        >
            <Octicons style={{width: '20px', height: '20px'}}/>
            {numStars != -1 && numStars != undefined &&
                <span className="ml-3 pl-3 border-l">
                    {numStars}
                </span>
            }
        </a>
    )
}