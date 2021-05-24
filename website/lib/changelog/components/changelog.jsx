
/*
export default function Changelog({vers}) {
    console.log("--cc")
    console.log(vers)
    
    return (
        <div>
        <div class="bg-white pt-10 pb-20 px-4 sm:px-6 lg:pt-24 lg:pb-28 lg:px-8">
          <div class="relative max-w-4xl mx-auto divide-y-2 divide-gray-200 lg:max-w-7xl">
            <div>
              <h2 class="text-3xl tracking-tight font-extrabold text-gray-900 sm:text-4xl">
                Changelog
              </h2>
            </div>
            <div class="mt-6 pt-10 grid gap-16">

              {vers.map((item, indx) => (
                <div key={indx} className="grid grid-cols-5">
                    <div className="hidden lg:block">
                        {item.version}
                    </div>
                    <div className="col-span-5 lg:col-span-4 text-base">
                        <div dangerouslySetInnerHTML={{__html: item.content}} />
                    </div>
                </div>
              ))}
        
            </div>
          </div>
        </div>
    </div>
    )
}
*/

/* This example requires Tailwind CSS v2.0+ */
import { NewspaperIcon, PhoneIcon, SupportIcon } from '@heroicons/react/outline'

export default function Example({vers}) {
  return (
    <div className="bg-white text-black changelog">
      {/* Header */}
      <div className="relative pb-32">
        <div className="relative container max-w-7xl mx-auto py-24 px-4 sm:py-24 sm:px-6 lg:px-8">
          <h1 className="text-4xl font-extrabold tracking-tight text-black md:text-5xl lg:text-6xl">Changelog</h1>
          <p className="mt-6 max-w-3xl text-xl">
            Varius facilisi mauris sed sit. Non sed et duis dui leo, vulputate id malesuada non. Cras aliquet purus dui
            laoreet diam sed lacus, fames. Dui, amet, nec sit pulvinar.
          </p>
        </div>
      </div>

      {/* Overlapping cards */}
      <section
        className="-mt-32 container max-w-7xl mx-auto relative px-4"
        aria-labelledby="contact-heading"
      >
        {vers.map((item, indx) => (
          <div key={indx} className="grid grid-cols-5 border-t pt-16 pb-10">
              <div className="hidden ml-5 lg:block">
                {item.version}
              </div>
              <div className="col-span-5 lg:col-span-4 text-base">
                <div dangerouslySetInnerHTML={{__html: item.content}} />
              </div>
          </div>
        ))}
      </section>
    </div>
  )
}
